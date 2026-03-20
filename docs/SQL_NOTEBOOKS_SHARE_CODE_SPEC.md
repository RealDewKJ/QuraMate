# SQL Notebook SessionShare Spec

Last updated: 2026-03-19
Status: In Progress
Owner: Product + Frontend + Backend
Related: `docs/SQL_NOTEBOOKS_ROADMAP.md`

## Purpose

Define the first real `SessionShare across machines` flow for SQL Notebooks.

This spec now focuses on the live `SessionShare` path. The product UI should not continue carrying a separate legacy share-code flow alongside it.

Current implementation status in repo:

- SQL Notebook share/import modals now focus on `SessionShare`
- Cloudflare Worker routes now support live session create/join/socket flows
- a self-hostable relay server is included at `cmd/sql-notebook-share-relay`
- a `Cloudflare Worker + D1` relay scaffold is included at `cloudflare/sql-notebook-share-relay`
- trust review is still minimal and should be expanded in follow-up work

## Problem Statement

Users want to:

- click `Share`
- receive a short `SessionShare` code
- send that code to a teammate
- let the teammate paste the code into QuraMate on another machine
- join the shared notebook session safely

The current app can redact notebook data and persist local data, but it does not yet have a shared relay or any cross-device retrieval path.

## Goals

- Support a short human-readable `SessionShare` code for notebooks.
- Make the code joinable from another machine.
- Keep connection-specific details redacted by default.
- Import shared content as draft content only.
- Avoid requiring a full collaboration product or user account system for V1.

## Non-Goals

- Live collaboration
- Notebook sync between users
- Shared editing
- Automatic execution after import
- Cross-tenant org permissions model
- Permanent public notebook hosting

## Future Mode: Live Share Session

This is the next collaboration layer after basic share codes.

Target behavior:

- one user creates a live session code
- another user joins with the same code
- while the session is active, notebook edits propagate to connected peers in near real time
- when the session expires, live syncing stops
- each participant keeps their own local notebook copy
- local notebook data must not be deleted when the live session ends

This mode should feel closer to `codeshare.io` than to one-time import/export.

## User Experience

### Sender flow

1. User opens `Share` on a notebook.
2. User chooses `SessionShare`.
3. App uploads a redacted payload to a relay service and starts a live sync channel.
4. App returns a short code such as `QN-4KF9D2`.
5. User copies the code and sends it to another person.

### Receiver flow

1. User opens `Import` in SQL Notebooks.
2. User pastes the `SessionShare` code.
3. App joins the code from the relay service.
4. Notebook is imported locally as a draft.
5. While the session remains active, remote notebook edits continue to sync.

## Product Rules

- Imported content must never auto-run.
- Connection name must be redacted by default.
- Execution history should not be shared in V1.
- Result snapshots should be excluded by default in V1.
- SessionShare codes should expire automatically.
- Re-opening the same unchanged notebook session should reuse the same active code until it expires.
- Imported notebooks should default to:
  - `isFavorite = false`
  - `isTemplate = false`
  - empty snapshots
  - idle execution state

Rules for future live sessions:

- live sessions must never auto-run SQL on other peers
- every participant keeps a local notebook copy
- session expiry stops remote propagation only
- session expiry must not delete any local notebook data
- owner can optionally end the session early
- peers should be able to disconnect and keep working locally

## Payload Shape

The relay should store a redacted payload shaped roughly like this:

```text
SharePayload
  id
  code
  scope: notebook | sql
  createdAt
  expiresAt
  senderLabel?
  sourceDbType
  redacted: true
  notebook
    title
    description
    tags[]
    variables[]
    metadata
    cells[]
```

Notes:

- `connectionName` is excluded.
- `connectionId` is excluded.
- local result caches are excluded.
- execution timestamps are excluded.

## Share Code Format

Recommended V1 format:

- prefix: `QN-`
- body: 6 to 8 uppercase alphanumeric chars
- examples:
  - `QN-4KF9D2`
  - `QN-X82M7QP`

Requirements:

- easy to read over chat
- low collision probability
- not derived from notebook content

## Relay Architecture

V1 requires a small shared service.

Implemented in repo:

- `cmd/sql-notebook-share-relay/main.go`
- `internal/notebookshare/relay/relay.go`

### Minimum service responsibilities

- accept redacted notebook payloads
- generate unique share codes
- store payloads with expiry
- resolve payload by code
- delete or expire old payloads

### Storage options

Preferred V1:

- small server-side sqlite or postgres table

Preferred no-server deployment:

- Cloudflare Worker + D1

Preferred live-session deployment:

- Cloudflare Worker + Durable Objects

Table concept:

```text
shared_notebooks
  id
  code
  payload_json
  created_at
  expires_at
  sender_label
  scope
  source_db_type
```

## API Draft

### Create share

`POST /api/notebook-shares`

Request:

```json
{
  "scope": "notebook",
  "senderLabel": "Alice",
  "payload": {
    "redacted": true,
    "sourceDbType": "postgres",
    "notebook": {}
  }
}
```

Response:

```json
{
  "code": "QN-4KF9D2",
  "expiresAt": "2026-03-26T10:00:00Z"
}
```

### Resolve share

`GET /api/notebook-shares/{code}`

Response:

```json
{
  "code": "QN-4KF9D2",
  "expiresAt": "2026-03-26T10:00:00Z",
  "payload": {
    "redacted": true,
    "sourceDbType": "postgres",
    "notebook": {}
  }
}
```

### Optional delete share

`DELETE /api/notebook-shares/{code}`

This can be deferred if expiry is enough for V1.

## Future Live Session API Draft

### Create live session

`POST /api/notebook-live-sessions`

Request:

```json
{
  "senderLabel": "Alice",
  "payload": {
    "redacted": true,
    "notebook": {}
  }
}
```

Response:

```json
{
  "code": "QN-LIVE82",
  "expiresAt": "2026-03-26T10:00:00Z",
  "sessionId": "session_123"
}
```

### Join live session

`POST /api/notebook-live-sessions/{code}/join`

Response:

```json
{
  "sessionId": "session_123",
  "expiresAt": "2026-03-26T10:00:00Z",
  "payload": {
    "redacted": true,
    "notebook": {}
  }
}
```

### Sync channel

Recommended transport:

- WebSocket
- Durable Object hibernation WebSocket support on Cloudflare

Messages:

- `presence`
- `snapshot`
- `patch`
- `ack`
- `session-expired`
- `session-ended`

Patch model:

- notebook title/description updates
- cell insert/delete/move
- cell content updates
- metadata updates

Do not sync:

- live SQL execution to peers by default
- local-only UI state such as open panels or scroll position
- local snapshots unless explicitly shared later

Conflict strategy for V1 live mode:

- last-write-wins for text payloads
- coarse notebook patch broadcast first
- more advanced CRDT/OT behavior can come later if needed

## Running The Relay

Local/dev example:

```bash
go run ./cmd/sql-notebook-share-relay
```

Optional environment variables:

- `QURAMATE_SHARE_RELAY_PORT`
- `QURAMATE_SHARE_RELAY_DB`

Default local address:

- `http://localhost:8787`

## Cloudflare Deployment

Preferred hosted option:

- deploy `cloudflare/sql-notebook-share-relay`

Guide:

- `docs/CLOUDFLARE_SQL_NOTEBOOK_SHARE_RELAY.md`

## Security and Safety

### Required

- server stores only redacted notebook payloads
- no secrets or connection names in stored payload
- import requires explicit user confirmation
- imported SQL is always inactive until user runs it
- payload size limits
- expiry required for every share

### Recommended

- basic rate limiting on create and resolve
- signed server-generated ids rather than client-generated ids
- audit logging for create and resolve events
- sender label optional, never trusted as identity

## Expiry Rules

Recommended V1 default:

- 7 days expiry

Optional future:

- 1 day
- 7 days
- 30 days

Expired codes should return a clear error:

- `This share code expired. Ask the sender to generate a new one.`

## UI Changes Needed

### Frontend

- `Share` modal:
  - replace local-only wording with `Share Across Devices`
  - show code and expiry after successful create
- `Import` modal:
  - add resolve loading state
  - add trust review before final import
- notebook import:
  - reuse existing local import normalization logic

### Backend

- add relay client methods or relay endpoints
- define payload serializer with redaction rules
- add config for relay base URL

## Failure States

Receiver-side UX should handle:

- invalid code
- expired code
- relay unavailable
- malformed payload
- unsafe payload rejected by validator

Suggested messages:

- `Share code not found.`
- `This share code expired.`
- `Could not reach the sharing service right now.`
- `This shared notebook is invalid and could not be imported.`

## Rollout Plan

### Step 1

- keep current local-only share code as internal prototype
- add clear naming so it is not confused with cross-device sharing

### Step 2

- implement relay-backed create and resolve endpoints
- switch UI wording to `Share Across Devices`

### Step 3

- add trust review screen before import
- add expiry and better error handling

### Step 4

- consider QR representation of the share code
- consider one-click copy link if a hosted relay exists

## Open Questions

- Should relay be hosted by QuraMate or self-hostable first?
- Should sender label be free text or tied to a future identity model?
- Should SQL-cell-only shares become standalone notebooks on import or import as one-cell drafts?
- Should snapshots be optionally included later behind a warning?

## Recommended Decision

For the first real cross-device release:

- use a small relay service
- 7-day expiry
- notebook and SQL-cell share supported
- import as local draft only
- no snapshots
- no auto-run
- no auth requirement in V1, but build the API so auth can be added later
