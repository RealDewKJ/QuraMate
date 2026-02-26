#!/usr/bin/env node

/**
 * QuraMate npm postinstall script.
 * Downloads the correct platform binary from GitHub Releases.
 */

const https = require("https");
const http = require("http");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");
const os = require("os");

const pkg = require("./package.json");
const VERSION = pkg.version;
const GITHUB_REPO = "RealDewKJ/QuraMate";
const BASE_URL = `https://github.com/${GITHUB_REPO}/releases/download/v${VERSION}`;

/**
 * Maps the current platform/arch to a GitHub Release asset name.
 */
function getAssetInfo() {
  const platform = os.platform();
  const arch = os.arch();

  switch (platform) {
    case "win32":
      return {
        asset: `QuraMate-amd64-installer.exe`,
        binary: "QuraMate.exe",
        isInstaller: true,
      };
    case "darwin":
      return {
        asset: `QuraMate-macOS-universal.zip`,
        binary: "QuraMate.app",
        isZip: true,
      };
    default:
      console.error(`\n  ❌ Unsupported platform: ${platform} (${arch})`);
      console.error(`  QuraMate currently supports: Windows (x64), macOS (Universal)`);
      console.error(`  Visit https://github.com/${GITHUB_REPO}/releases for manual download.\n`);
      process.exit(1);
  }
}

/**
 * Downloads a file, following redirects.
 */
function download(url, dest) {
  return new Promise((resolve, reject) => {
    const lib = url.startsWith("https") ? https : http;
    const file = fs.createWriteStream(dest);

    const request = lib.get(url, (response) => {
      // Follow redirects (GitHub returns 302)
      if (response.statusCode === 301 || response.statusCode === 302) {
        file.close();
        fs.unlinkSync(dest);
        return download(response.headers.location, dest).then(resolve).catch(reject);
      }

      if (response.statusCode !== 200) {
        file.close();
        fs.unlinkSync(dest);
        return reject(new Error(`Download failed: HTTP ${response.statusCode} for ${url}`));
      }

      const totalBytes = parseInt(response.headers["content-length"], 10);
      let downloadedBytes = 0;

      response.on("data", (chunk) => {
        downloadedBytes += chunk.length;
        if (totalBytes) {
          const pct = ((downloadedBytes / totalBytes) * 100).toFixed(1);
          process.stdout.write(`\r  ⬇  Downloading QuraMate... ${pct}%`);
        }
      });

      response.pipe(file);

      file.on("finish", () => {
        file.close();
        process.stdout.write("\n");
        resolve();
      });
    });

    request.on("error", (err) => {
      fs.unlink(dest, () => {});
      reject(err);
    });
  });
}

/**
 * Extracts a zip file (macOS).
 */
function extractZip(zipPath, destDir) {
  if (os.platform() === "darwin") {
    execSync(`ditto -x -k "${zipPath}" "${destDir}"`, { stdio: "inherit" });
  } else {
    // Fallback: use tar on Linux, PowerShell on Windows
    if (os.platform() === "win32") {
      execSync(`powershell -Command "Expand-Archive -Path '${zipPath}' -DestinationPath '${destDir}' -Force"`, { stdio: "inherit" });
    } else {
      execSync(`unzip -o "${zipPath}" -d "${destDir}"`, { stdio: "inherit" });
    }
  }
}

async function main() {
  console.log(`\n  📦 QuraMate v${VERSION} — postinstall\n`);

  const info = getAssetInfo();
  const url = `${BASE_URL}/${info.asset}`;
  const binDir = path.join(__dirname, "bin");
  const downloadDir = path.join(__dirname, ".download");

  // Ensure directories exist
  fs.mkdirSync(binDir, { recursive: true });
  fs.mkdirSync(downloadDir, { recursive: true });

  const downloadPath = path.join(downloadDir, info.asset);

  try {
    await download(url, downloadPath);
    console.log(`  ✅ Downloaded ${info.asset}`);

    if (info.isZip) {
      console.log("  📂 Extracting...");
      extractZip(downloadPath, binDir);
      console.log("  ✅ Extracted successfully");
    } else if (info.isInstaller) {
      // For Windows, copy the installer to bin/
      const destPath = path.join(binDir, info.binary);
      fs.copyFileSync(downloadPath, destPath);
      console.log(`  ✅ Binary placed at ${destPath}`);
    }

    // Make unix binaries executable
    if (os.platform() !== "win32") {
      const binScript = path.join(binDir, "quramate");
      if (fs.existsSync(binScript)) {
        fs.chmodSync(binScript, 0o755);
      }
    }

    // Cleanup
    fs.rmSync(downloadDir, { recursive: true, force: true });

    console.log(`\n  🎉 QuraMate installed successfully!`);
    console.log(`  Run 'quramate' to launch.\n`);
  } catch (err) {
    console.error(`\n  ❌ Installation failed: ${err.message}`);
    console.error(`  You can download manually from: https://github.com/${GITHUB_REPO}/releases\n`);
    // Don't fail the npm install — the binary just won't be available
    process.exit(0);
  }
}

main();
