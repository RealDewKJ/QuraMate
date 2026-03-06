# QuraMate Features

QuraMate is a cross-platform desktop database client built with Go (Wails) and Vue.
It focuses on practical daily workflows for database engineers, developers, and analysts.

## What QuraMate Can Do

- Connect to multiple databases in parallel and switch sessions quickly.
- Support major engines: PostgreSQL, MySQL, MariaDB, SQL Server, SQLite, LibSQL, and others in the same UI.
- Browse schemas and objects (tables, views, stored procedures, functions).
- Write and run SQL in a Monaco-based editor with syntax highlighting and formatting.
- Execute selected SQL only, or run full scripts with streaming result handling.
- Save SQL files, reopen them, and continue editing from the same query tab.
- Use keyboard shortcuts for common actions (run, save, refresh, new/close tab, format).
- Edit table data inline with context-aware cell editors.
- Paste tabular data from clipboard/Excel with validation and preview before insert.
- Insert rows in batch with transactional safety and clear inserted/skipped feedback.
- Design and modify table structures visually (columns, keys, indexes, foreign keys).
- Generate and view ER diagrams from relationship metadata.
- Track activity and long-running queries via a built-in activity monitor.
- Cancel active queries and inspect execution/fetch timing stats.
- Use optional read-only mode to reduce accidental write operations.
- Export/import table data in common formats such as CSV, JSON, SQL, and Excel.
- Store preferences and query history locally.
- Check for app updates from GitHub releases.

## Platform and Architecture

- Desktop app (Windows/macOS/Linux targets via Wails build pipeline).
- Backend: Go
- Frontend: Vue 3 + TypeScript + Tailwind CSS
- Local-first workflow: no mandatory cloud account required.