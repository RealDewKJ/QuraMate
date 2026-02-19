# VaultDB

VaultDB is a modern, cross-platform database client. It is easy to connect to your databases, manage tables, view data, write SQL, and run queries.

## Feedback

This repository is used to collect VaultDB bugs and feedback. If you encounter any bugs or have any suggestions or new feature requests, please open an issue.

## Supported Databases

- MySQL
- PostgreSQL
- SQL Server (MSSQL)
- SQLite

## Supported Platform

- Windows
- macOS
- Linux

## Development Setup

### Prerequisites

- [Go](https://go.dev/dl/) (1.21+)
- [Node.js](https://nodejs.org/) (18+) with npm
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2

### Fedora Linux (Rawhide / Newer Versions)

Fedora 39+ ships with WebKitGTK 4.1 and has deprecated the older 4.0 API. You must use the `webkit2_41` build tag.

1. **Install the development libraries:**
   ```bash
   sudo dnf install webkit2gtk4.1-devel gtk3-devel
   ```

2. **Verify the packages are installed correctly:**
   ```bash
   pkg-config --modversion webkit2gtk-4.1
   ```

3. **Clean any stale build artifacts** (if switching between WebKit versions):
   ```bash
   rm -f VaultDB
   ```

4. **Run in development mode:**
   ```bash
   wails dev -tags webkit2_41
   ```

5. **Build for production:**
   ```bash
   wails build -tags webkit2_41
   ```

> **Note:** The `-tags webkit2_41` flag tells the Go compiler to use the WebKitGTK 4.1 bindings instead of the default 4.0. This is required on Fedora 39+ where `webkit2gtk-4.0` is no longer available.

## Download

- [Releases](https://github.com/yourusername/vaultdb/releases)

## Contact Us

If you have any other questions or need further assistance, please open an issue on this repository.
