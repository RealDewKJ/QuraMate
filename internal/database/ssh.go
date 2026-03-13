package database

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const sshDialTimeout = 15 * time.Second

func existingKnownHostsFiles() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return nil
	}

	candidates := []string{
		filepath.Join(homeDir, ".ssh", "known_hosts"),
		filepath.Join(homeDir, ".ssh", "known_hosts2"),
	}

	files := make([]string, 0, len(candidates))
	for _, path := range candidates {
		info, statErr := os.Stat(path)
		if statErr == nil && !info.IsDir() {
			files = append(files, path)
		}
	}

	return files
}

func buildSSHHostKeyCallback() (ssh.HostKeyCallback, error) {
	knownHostFiles := existingKnownHostsFiles()
	if len(knownHostFiles) == 0 {
		return nil, fmt.Errorf("no SSH known_hosts file found (expected ~/.ssh/known_hosts). Please add the SSH host key first")
	}

	baseCallback, err := knownhosts.New(knownHostFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to load known_hosts: %w", err)
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := baseCallback(hostname, remote, key)
		if err == nil {
			return nil
		}

		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) {
			gotFingerprint := ssh.FingerprintSHA256(key)
			if len(keyErr.Want) == 0 {
				return fmt.Errorf("ssh host key for %s is not trusted (fingerprint: %s); add it to ~/.ssh/known_hosts", hostname, gotFingerprint)
			}
			return fmt.Errorf("ssh host key mismatch for %s (got: %s); check ~/.ssh/known_hosts", hostname, gotFingerprint)
		}

		return err
	}, nil
}

func proxyBidirectional(localConn net.Conn, remoteConn net.Conn) {
	var closeOnce sync.Once
	closeAll := func() {
		_ = localConn.Close()
		_ = remoteConn.Close()
	}

	relay := func(dst net.Conn, src net.Conn, done chan<- struct{}) {
		_, _ = io.Copy(dst, src)
		closeOnce.Do(closeAll)
		done <- struct{}{}
	}

	done := make(chan struct{}, 2)
	go relay(localConn, remoteConn, done)
	go relay(remoteConn, localConn, done)

	<-done
	<-done
}

func (d *Database) openSSHTunnel(config DBConfig) (string, int, error) {
	if strings.TrimSpace(config.SSHHost) == "" {
		return "", 0, fmt.Errorf("ssh host is required")
	}
	if config.SSHPort <= 0 || config.SSHPort > 65535 {
		return "", 0, fmt.Errorf("ssh port must be between 1 and 65535")
	}
	if strings.TrimSpace(config.SSHUser) == "" {
		return "", 0, fmt.Errorf("ssh user is required")
	}
	if strings.TrimSpace(config.SSHKeyFile) == "" && config.SSHPassword == "" {
		return "", 0, fmt.Errorf("ssh password or ssh key file is required")
	}

	hostKeyCallback, err := buildSSHHostKeyCallback()
	if err != nil {
		return "", 0, err
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.SSHUser,
		HostKeyCallback: hostKeyCallback,
		Timeout:         sshDialTimeout,
	}

	if strings.TrimSpace(config.SSHKeyFile) != "" {
		key, err := os.ReadFile(config.SSHKeyFile)
		if err != nil {
			return "", 0, fmt.Errorf("unable to read private key: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return "", 0, fmt.Errorf("unable to parse private key: %v", err)
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(config.SSHPassword)}
	}

	sshAddr := fmt.Sprintf("%s:%d", config.SSHHost, config.SSHPort)
	client, err := ssh.Dial("tcp", sshAddr, sshConfig)
	if err != nil {
		return "", 0, fmt.Errorf("failed to dial ssh: %w", err)
	}
	d.sshClient = client

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		client.Close()
		d.sshClient = nil
		return "", 0, fmt.Errorf("failed to start local listener: %w", err)
	}
	d.sshListener = listener

	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				d.log("INFO", fmt.Sprintf("SSH tunnel listener stopped: %v", err))
				return
			}

			go func(lc net.Conn) {
				remoteAddr := fmt.Sprintf("%s:%d", config.Host, config.Port)
				remoteConn, err := client.Dial("tcp", remoteAddr)
				if err != nil {
					d.log("ERROR", fmt.Sprintf("SSH tunnel dial error: %v", err))
					_ = lc.Close()
					return
				}

				proxyBidirectional(lc, remoteConn)
			}(localConn)
		}
	}()

	tcpAddr := listener.Addr().(*net.TCPAddr)
	return "localhost", tcpAddr.Port, nil
}
