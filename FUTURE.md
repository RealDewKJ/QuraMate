# Future Tasks

Pending distribution and packaging tasks that are planned but not yet implemented.

## Package Managers

### Windows

- [ ] **Winget** — Publish to Windows Package Manager (`winget install QuraMate`)
  - Requires: manifest submission to [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs)
  - Needs: signed installer (from SignPath)

- [ ] **Scoop** — Create Scoop bucket (`scoop install quramate`)
  - Requires: separate repo `RealDewKJ/scoop-bucket` with a JSON manifest
  - Needs: portable zip build (non-installer)

- [ ] **Chocolatey** — Publish to Chocolatey (`choco install quramate`)
  - Requires: Chocolatey account + `.nuspec` package definition
  - Needs: signed installer

### Linux

- [ ] **Linux builds** — Add `build-linux` job to `release.yml`
  - CI dependencies: `libgtk-3-dev`, `libwebkit2gtk-4.0-dev`, `pkg-config`
  - Targets: `linux/amd64` (and optionally `arm64`)
  - Output: `.tar.gz` archive

- [ ] **Snap** — Publish to Snap Store (`snap install quramate`)
  - Requires: `snapcraft.yaml` definition
  - Needs: Linux binary + desktop file

- [ ] **Flatpak** — Publish to Flathub (`flatpak install quramate`)
  - Requires: Flatpak manifest + Flathub submission
  - Needs: Linux binary + desktop file + AppStream metadata

- [ ] **AUR** — Publish to Arch User Repository (`yay -S quramate`)
  - Requires: `PKGBUILD` file
  - Can build from source or use pre-built binary (`quramate-bin`)

## Signing & Notarization

- [ ] **macOS Notarization** — Apple notarization for Gatekeeper bypass
  - Requires: Apple Developer account ($99/year)
  - Eliminates the `xattr -dr com.apple.quarantine` workaround

## CI/CD Enhancements

- [ ] **Auto-update Homebrew Cask** — CI step to generate and push updated Cask with SHA256 on release
- [ ] **Automatic changelog** — Generate changelog from conventional commits
