# SQL Notebooks Roadmap

Last updated: 2026-03-19
Status: In Progress
Owner: Product + Frontend + Backend

## Goal

`SQL Notebooks` should evolve QuraMate from a query runner into a structured SQL workspace where users can write, run, explain, document, and reuse database knowledge in one place.

The feature should feel native to the existing dashboard rather than a separate mini-app. The first versions should reuse the current query execution and results flow, then gradually add notebook-specific structure, runbook behavior, and team knowledge features.

## Product Position

`SQL Notebooks` is the primary workspace for:

- ad hoc querying
- troubleshooting runbooks
- reusable data investigation workflows

`Runbooks` should not survive as a competing top-level feature. It should become a support layer inside the `SQL Notebooks` experience.

In the UI, that means notebooks remain the main place where work is assembled, run, and documented.

It should solve these jobs well:

- save a query with explanation and context
- run investigative SQL in repeatable steps
- document why a query exists and when to use it
- keep a team-ready operational playbook close to the database

## Core Principles

- Start with a thin integration on top of the current dashboard and query workspace.
- Reuse QuraMate's current execution, result rendering, and connection model wherever possible.
- Treat notebooks as structured data, not just saved editor text.
- Keep production safety visible with environment and write-risk guardrails.
- Prefer progressive disclosure so notebook UX stays easy for first-time users.
- Ship in phases that can be validated with real usage before expanding scope.

## Current System Anchors

The first implementation should extend these existing areas instead of replacing them:

- `frontend/src/components/DbDashboard.vue`
- `frontend/src/components/dashboard/DbQueryWorkspacePane.vue`
- `frontend/src/components/dashboard/DbQueryResultsPane.vue`
- `frontend/src/composables/useQueryExecution.ts`
- `frontend/src/composables/useTabs.ts`
- `frontend/src/types/dashboard.ts`

## Proposed Architecture

### Frontend feature layout

```text
frontend/src/
  components/
    notebooks/
      SqlNotebooksWorkspace.vue
      SqlNotebooksSidebar.vue
      SqlNotebookHeader.vue
      SqlNotebookEditor.vue
      SqlNotebookCellList.vue
      SqlNotebookCellItem.vue
      SqlNotebookVariablesPanel.vue
      SqlNotebookRunPanel.vue
      cells/
        SqlCellEditor.vue
        MarkdownCellEditor.vue
        RunbookCellEditor.vue
  composables/
    useSqlNotebooks.ts
    useSqlNotebookEditor.ts
    useSqlNotebookRunner.ts
  types/
    sqlNotebook.ts
```

### Responsibility map

- `SqlNotebooksWorkspace.vue`: feature shell inside the dashboard, wires notebook library, editor, and side panels.
- `SqlNotebooksSidebar.vue`: notebook tree, recent items, favorites, tags, and create/open actions.
- `SqlNotebookEditor.vue`: composition surface for the selected notebook and active cell state.
- `SqlNotebookCellList.vue`: renders and reorders notebook cells.
- `SqlNotebookCellItem.vue`: per-cell frame, actions, state badge, and type-specific rendering.
- `SqlCellEditor.vue`: SQL authoring and run actions using the current query execution path.
- `MarkdownCellEditor.vue`: notes, instructions, and explanations.
- `RunbookCellEditor.vue`: operational steps with checklist and expected outcome fields.
- `useSqlNotebooks.ts`: notebook loading, saving, listing, filters, and metadata state.
- `useSqlNotebookEditor.ts`: selection state, insert/move/duplicate/delete cell actions.
- `useSqlNotebookRunner.ts`: cell execution, run-all flow, variables, and result binding.

### Data model

The first typed model should cover:

```text
Notebook
  id
  title
  description
  tags[]
  connectionScope
  cells[]
  createdAt
  updatedAt
  lastOpenedAt

NotebookCell
  id
  type: sql | | runbook
  title
  content
  collapsed
  executionState
  lastRunAt

NotebookVariable
  key
  label
  value
  required
  type
```

## UI Structure

The feature should live as a mode within the database dashboard, not as a separate top-level app route.

### Entry point

Add a dashboard mode switch or tab group for:

- `Query`
- `SQL Notebooks`
- `History`

### Workspace layout

```text
| Notebook Library | Notebook Canvas | Context Panel |
```

- Left: notebooks, collections, tags, favorites, recent.
- Center: cell-based notebook editor.
- Right: quick context only, with variables and snapshots available without overwhelming the main canvas.

### Cell types for V1

- `sql`
- `markdown`

### Cell types for later phases

- `runbook`
- `result snapshot`
- `chart`
- `checklist`

## Delivery Roadmap

## Phase 1: Foundation

Objective: create a usable notebook experience with minimal risk by reusing the existing query engine.

Scope:

- notebook data model and local persistence
- notebook sidebar with list, create, rename, delete
- notebook editor with ordered cells
- SQL cell and Markdown cell
- run single SQL cell
- route results into the existing results pane
- convert current query tab content into a notebook cell
- keep notebook creation and save flows close to the current query workspace

Success criteria:

- users can create and save notebooks
- users can run SQL from a notebook without leaving the dashboard
- notebooks feel faster to use than managing ad hoc SQL manually

## Phase 2: Structured Reuse

Objective: make notebooks repeatable for day-to-day analysis work.

Scope:

- notebook variables
- notebook metadata: tags, environment, purpose, owner
- duplicate notebook and template support
- result snapshot save and compare
- recent notebooks and favorites
- search across notebook titles, tags, and cell content

Success criteria:

- users can reuse the same notebook for repeated investigations
- variable-driven notebooks reduce copy-paste editing
- common operational workflows become searchable assets

## Phase 3: Runbook Mode

Objective: turn notebooks into guided operational flows.

Scope:

- runbook cell type
- step status: pending, running, verified, skipped
- expected result and rollback notes
- write-risk warnings for destructive SQL
- environment-aware guardrails for production
- run selected cells or run all executable steps

Success criteria:

- users can execute incident or maintenance workflows step by step
- production-facing tasks feel safer and more explicit
- notebooks become reliable team runbooks, not just personal notes

## Phase 4: Team Knowledge Layer

Objective: make notebooks collaborative and explainable.

Scope:

- comments or annotations per cell
- AI explain and summarize on SQL cells
- export to Markdown, SQL bundle, and report-friendly output
- import/export notebooks
- richer sharing model if local collaboration is added later
- support simple redacted share codes as a stepping stone toward true cross-device sharing

Success criteria:

- notebooks become long-lived operational knowledge
- onboarding and handoff become easier
- teams can review and improve shared SQL workflows over time

## Phase 5: Sharing and Peer Transfer

Objective: let users move notebook knowledge between QuraMate installs without requiring a full collaboration backend first.

Scope:

- create a temporary `SessionShare` code that can be joined from another machine through a relay or shared service
- keep sharing focused on one collaborative path instead of maintaining a separate legacy share-code flow in the product UI
- optionally export notebook content as a redacted bundle when needed
- share via QR, one-time transfer code, or local file bundle
- optional local-network peer discovery for QuraMate users on the same LAN
- trust prompt before accepting shared SQL content
- strip or redact connection-specific metadata before share
- import shared content as draft/template by default, not auto-run content
- explore external share surfaces such as email-safe bundles, temporary signed links, and chat-friendly package flows without exposing live connection details

Current implementation progress:

- SQL Notebook sharing UI now prioritizes `SessionShare` as the primary cross-device path
- app client now talks to a configurable relay URL for live session create/join
- self-hostable relay server scaffold now lives in `cmd/sql-notebook-share-relay`
- recommended hosted path now has a `Cloudflare Worker + D1` scaffold in `cloudflare/sql-notebook-share-relay`
- remaining work is mostly around trust review UX, deployment guidance, expiry management polish, and optional external share surfaces

Success criteria:

- users can hand off notebooks to another QuraMate user with low friction
- shared content can move between machines without copying raw files manually
- imported SQL feels safe, explicit, and connection-agnostic by default
- the app supports lightweight team knowledge exchange before full cloud sync exists

## Phase 6: Live Share Sessions

Objective: support temporary codeshare-style collaboration while keeping SQL Notebook local-first by default.

Scope:

- create a temporary live session code for a notebook
- let another QuraMate user join that notebook session with the same code
- sync notebook edits in near real time while the session is active
- stop live syncing automatically when the session expires
- keep each user's local notebook copy even after expiry or disconnect
- make expiry stop only the remote link, not the local notebook data
- add presence/session-state UI such as `live`, `peer connected`, and `session ended`
- preserve safety by not auto-running SQL on peer machines
- evaluate Cloudflare Worker + Durable Objects as the default hosted implementation

Success criteria:

- two users can collaborate on the same notebook with one temporary code
- edits propagate while the session is active and connected
- when the session expires, sharing stops but each user keeps their local notebook copy
- users can clearly tell whether they are in a live session or editing locally

## Implementation Strategy

### V1 integration path

1. Extend the dashboard with a mode switch for notebooks.
2. Introduce notebook types and local storage contracts.
3. Render a notebook workspace beside the current query workspace.
4. Reuse query execution and results rendering for notebook SQL cells.
5. Keep query-to-notebook save flows tightly connected to the current dashboard workspace.

### Persistence recommendation

Start with local persistence in the existing settings or local app storage layer so the UI can validate the feature quickly.

After usage is proven, evaluate a dedicated local table for:

- notebooks
- notebook cells
- notebook variables
- notebook run snapshots

### Safety model

From the first interactive version, include:

- read-only awareness
- visible environment label
- confirmation for destructive SQL in protected environments
- optional later policy for notebook-level execution restrictions

## Work Tracker

Use this table to keep the feature moving in visible steps.

| ID | Work item | Area | Status | Notes |
| --- | --- | --- | --- | --- |
| SN-01 | Confirm product scope and V1 boundaries | Product | Completed | Ship `sql` + `markdown` cells first |
| SN-02 | Define notebook TypeScript models | Frontend | Completed | Added `frontend/src/types/sqlNotebook.ts` |
| SN-03 | Add notebook mode entry in dashboard | Frontend | Completed | Query mode remains intact and runnable |
| SN-04 | Scaffold notebook workspace components | Frontend | Completed | Added feature shell, sidebar, editor, and cell editors |
| SN-05 | Add local persistence contract | Frontend/Backend | Completed | Local-first persistence is keyed by dashboard session/connection |
| SN-06 | Reuse query execution for SQL cells | Frontend | Completed | SQL notebook cells can run through the existing query/results flow |
| SN-07 | Harden query -> notebook save flow | Frontend | Completed | Query-to-notebook save, refresh sync, and unsaved-switch warning dialog are wired |
| SN-08 | Add notebook search, recent, favorites | Frontend | Completed | Search, favorites, duplicate, sorting, and sidebar filter chips are available |
| SN-09 | Add variables and parameter substitution | Frontend/Backend | Completed | Variable form replaces prompts and SQL placeholders resolve before run |
| SN-09A | Add notebook templates and create-from-template flow | Frontend | Completed | Templates can be marked, filtered, and used to create a fresh notebook run |
| SN-09B | Add snapshot compare workflow | Frontend | Completed | Snapshot compare now supports explicit key selection, row-order aware matching, and row-level value previews |
| SN-10 | Add runbook mode and guardrails | Frontend/Backend | In Progress | Runbook cells, structured approver status, evidence fields, template presets, sqlite-backed preset favorites, destructive SQL guardrails, and reduced-density UX are wired in the notebook workspace |
| SN-11 | Add export/import and AI explain | Frontend | In Progress | Notebook sharing now supports local-only share codes plus summary export; AI explain remains next |
| SN-12 | Add notebook sharing bundles and safe import flow | Frontend/Backend | Completed | Team sharing now supports notebook or SQL-cell summary export plus local sqlite-backed share codes for same-device or same-profile workflows |
| SN-13 | Add cross-device share code relay | Frontend/Backend | Planned | Requires a relay or shared service so a code generated on one machine resolves on another machine |
| SN-14 | Add QR, transfer code, and optional LAN peer discovery | Frontend/Backend | Planned | Phase 5 advanced sharing |
| SN-15 | Add live share sessions with temporary real-time sync | Frontend/Backend | Planned | Phase 6 codeshare-style collaboration that stops on expiry but leaves local copies intact |

## Risks

- If notebooks are implemented as a plain text blob, the feature will feel like saved files instead of a real workspace.
- If results are not tightly integrated with the current query result flow, users will feel context switching friction.
- If production safeguards are delayed too long, users may not trust runbook-style execution.
- If the first version tries to include templates, comments, AI, variables, and runbooks at once, delivery will slow down.
- If sharing includes connection details or secrets, users may leak sensitive environment information accidentally.
- If imported SQL can run immediately without review, peer transfer will create avoidable trust and safety risks.

## Next Build Steps

- add rollback completion and sign-off checkpoints to runbook execution
- refine snapshot diff with exportable compare views and smarter row pairing for edge cases
- polish template gallery with richer preset categories and onboarding copy
- define cross-device share-code relay API, expiry rules, and trust UX
- prototype external sharing options such as signed handoff links, bundle exchange, and peer transfer flows beyond the simple local share code
- prototype safe import UX before adding QR or LAN discovery

## Update Rules

When work starts, keep this file current by updating:

- `Last updated`
- the `Status` line
- the `Work Tracker` table
- a short note in the relevant phase section when scope changes
