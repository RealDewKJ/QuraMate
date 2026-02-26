<p align="center">
  <img src=".github/logo.png" alt="QuraMate" width="120" />
</p>

<h1 align="center">QuraMate</h1>

<p align="center">
  <strong>A modern, open-source database management tool. Connect. Query. Manage.</strong>
</p>

<p align="center">
  <a href="https://github.com/RealDewKJ/QuraMate/releases/latest"><img src="https://img.shields.io/github/v/release/RealDewKJ/QuraMate?style=flat-square&color=blue" alt="Release" /></a>
  <a href="https://github.com/RealDewKJ/QuraMate/blob/main/LICENSE"><img src="https://img.shields.io/github/license/RealDewKJ/QuraMate?style=flat-square" alt="License" /></a>
  <a href="https://github.com/RealDewKJ/QuraMate/releases"><img src="https://img.shields.io/github/downloads/RealDewKJ/QuraMate/total?style=flat-square&color=green" alt="Downloads" /></a>
  <a href="https://github.com/RealDewKJ/QuraMate/stargazers"><img src="https://img.shields.io/github/stars/RealDewKJ/QuraMate?style=flat-square" alt="Stars" /></a>
  <a href="https://www.npmjs.com/package/quramate"><img src="https://img.shields.io/npm/v/quramate?style=flat-square&color=cb3837" alt="npm" /></a>
</p>

<p align="center">
  <a href="#features">Features</a> ·
  <a href="#supported-databases">Databases</a> ·
  <a href="#install">Install</a> ·
  <a href="#build-from-source">Build</a> ·
  <a href="#contributing">Contributing</a>
</p>

---

## What is QuraMate?

QuraMate is a lightweight, cross-platform desktop database client built with [Wails](https://wails.io/) (Go + Vue). It lets you connect to multiple databases simultaneously, browse schemas, write and execute SQL, design table structures, visualize ER diagrams, and manage your data — all from a single, beautiful interface.

No cloud. No accounts. Your data stays on your machine.

## Features

- **Multi-connection** — open multiple database connections in parallel, each with its own tab workspace
- **SQL Editor** — syntax-highlighted editor with auto-complete, beautifier, and selected-text execution
- **Streaming Results** — query results stream in real-time with virtual scrolling for millions of rows
- **Table Designer** — visually modify columns, types, nullability, defaults, and primary keys
- **ER Diagrams** — auto-generated entity-relationship diagrams with foreign key visualization
- **Query Analysis** — built-in `EXPLAIN` support for query execution plan inspection
- **Inline Editing** — click any cell to edit, with type-aware inputs and row insertion
- **SSH Tunneling** — connect to remote databases through SSH with password or key-based auth
- **Read-Only Mode** — prevent accidental writes with a per-connection read-only toggle
- **Dark / Light Theme** — system-aware theme with manual toggle
- **Saved Connections** — persist and recall connection configurations locally
- **Import / Export** — CSV and JSON export from any result set

## Supported Databases

QuraMate supports **11 database engines** out of the box — no plugins, no extensions.

### Server-based

| Database             | Protocol    | Default Port |
| -------------------- | ----------- | ------------ |
| PostgreSQL           | `pgx`       | 5432         |
| MySQL                | `mysql`     | 3306         |
| MariaDB              | `mysql`     | 3306         |
| Microsoft SQL Server | `sqlserver` | 1433         |
| Greenplum            | `pgx`       | 5432         |
| Amazon Redshift      | `pgx`       | 5432         |
| CockroachDB          | `pgx`       | 26257        |
| Databend             | `mysql`     | 3307         |

### File-based

| Database | Protocol |
| -------- | -------- |
| SQLite   | `sqlite` |
| LibSQL   | `sqlite` |

## Install

### npm / pnpm / yarn / bun

Install globally via any JavaScript package manager:

```bash
# npm
npm install -g quramate

# pnpm
pnpm add -g quramate

# yarn
yarn global add quramate

# bun
bun add -g quramate
```

Then launch:

```bash
quramate
```

### Homebrew (macOS)

```bash
brew tap RealDewKJ/tap
brew install --cask quramate
```

### Direct Download

Pre-built binaries are available on the [Releases](https://github.com/RealDewKJ/QuraMate/releases/latest) page.

| Platform | Architecture | Download                                                                              |
| -------- | ------------ | ------------------------------------------------------------------------------------- |
| Windows  | x64          | [QuraMate-amd64-installer.exe](https://github.com/RealDewKJ/QuraMate/releases/latest) |
| macOS    | Universal    | [QuraMate-macOS-universal.zip](https://github.com/RealDewKJ/QuraMate/releases/latest) |
| macOS    | Universal    | [QuraMate-macOS-universal.dmg](https://github.com/RealDewKJ/QuraMate/releases/latest) |

### macOS installation notes

The macOS build is currently ad-hoc signed in CI (not notarized). On first run, Gatekeeper may block the app.

If blocked:

1. Move `QuraMate.app` to `/Applications`.
2. Right-click the app and choose **Open** once.
3. If needed, run:

```bash
xattr -dr com.apple.quarantine /Applications/QuraMate.app
```

### Coming Soon

| Package Manager | Platform | Command                    |
| --------------- | -------- | -------------------------- |
| winget          | Windows  | `winget install QuraMate`  |
| Scoop           | Windows  | `scoop install quramate`   |
| Chocolatey      | Windows  | `choco install quramate`   |
| Snap            | Linux    | `snap install quramate`    |
| Flatpak         | Linux    | `flatpak install quramate` |
| AUR             | Arch     | `yay -S quramate`          |

## Build from Source

### Prerequisites

- [Go](https://golang.org/dl/) ≥ 1.21
- [Node.js](https://nodejs.org/) ≥ 18
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2

### Steps

```bash
# Clone the repository
git clone https://github.com/RealDewKJ/QuraMate.git
cd QuraMate

# Install frontend dependencies
cd frontend && npm install && cd ..

# Run in development mode
wails dev

# Build production binary
wails build
```

The output binary will be at `build/bin/QuraMate.app` (macOS) or `build/bin/QuraMate.exe` (Windows).

## Release pipeline and signing

- Windows installer signing is handled through SignPath in GitHub Actions.
- SignPath Foundation enrollment and project configuration are a manual prerequisite.
- After SignPath is approved and `SIGNPATH_*` repository secrets/variables are configured, enable the `publish-release` job in `.github/workflows/release.yml` to publish signed artifacts to GitHub Releases.
- The npm package is auto-published via `.github/workflows/npm-publish.yml` when a GitHub Release is published. Requires `NPM_TOKEN` secret.

## Tech Stack

| Layer             | Technology         |
| ----------------- | ------------------ |
| Backend           | Go                 |
| Frontend          | Vue 3 + TypeScript |
| Desktop Framework | Wails v2           |
| Styling           | Tailwind CSS       |
| SQL Editor        | CodeMirror         |
| SQL Formatter     | sql-formatter      |

## Project Structure

```
QuraMate/
├── db.go                 # Database connection & query engine
├── app.go                # Wails application bindings
├── main.go               # Entry point
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── DbConnection.vue    # Connection form
│   │   │   └── DbDashboard.vue     # Main workspace
│   │   └── assets/
│   └── wailsjs/          # Auto-generated Go bindings
├── npm/                   # npm wrapper package
├── homebrew/              # Homebrew Cask definition
├── build/
│   ├── appicon.png
│   └── bin/               # Build output
└── wails.json             # Wails configuration
```

## Contributing

Contributions are welcome! Whether it's bug reports, feature requests, or pull requests — all are appreciated.

1. Fork the repository
2. Create your feature branch (`git checkout -b feat/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feat/amazing-feature`)
5. Open a Pull Request

Please follow the [Conventional Commits](https://www.conventionalcommits.org/) specification for commit messages.

## Feedback & Support

Found a bug? Have a feature request? Please [open an issue](https://github.com/RealDewKJ/QuraMate/issues/new).

## License

This project is open source. See the [LICENSE](LICENSE) file for details.

---

<p align="center">
  Built with ❤️ using <a href="https://wails.io">Wails</a> + <a href="https://vuejs.org">Vue</a> + <a href="https://go.dev">Go</a>
</p>
