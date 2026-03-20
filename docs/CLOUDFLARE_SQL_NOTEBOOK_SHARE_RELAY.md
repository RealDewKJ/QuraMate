# Cloudflare SQL Notebook Share Relay

Last updated: 2026-03-19

This is the recommended `no always-on machine` deployment path for QuraMate SQL Notebook share codes.

The app already supports:

- generating a cross-device share code against a relay URL
- resolving that code from another machine

This document explains how to deploy the relay on Cloudflare so both machines can use the same share service without keeping your own server online.

## What This Deploys

Location:

- `cloudflare/sql-notebook-share-relay`

Stack:

- Cloudflare Workers
- Cloudflare D1
- Cloudflare Durable Objects for future live sessions

The relay stores only redacted notebook payloads and returns short codes like `QN-ABC123`.

## Prerequisites

- Cloudflare account
- `npm`
- `wrangler`

## Setup

From the relay folder:

```bash
cd cloudflare/sql-notebook-share-relay
npm install
```

Create the D1 database:

```bash
npx wrangler d1 create sql-notebook-share-relay
```

Cloudflare will return a `database_id`.

Put that value into:

- `cloudflare/sql-notebook-share-relay/wrangler.toml`

Replace:

```toml
database_id = "REPLACE_WITH_D1_DATABASE_ID"
```

Apply the schema:

```bash
npx wrangler d1 execute sql-notebook-share-relay --remote --file=./schema.sql
```

Deploy the Worker:

```bash
npm run deploy
```

After deploy, Cloudflare will give you a Worker URL such as:

```text
https://sql-notebook-share-relay.<your-subdomain>.workers.dev
```

Put that URL into:

- `frontend/src/constants/sqlNotebookShare.ts`

## Use In QuraMate

On the sending machine:

1. Open SQL Notebook share.
2. Choose `Code`.
3. Click `Generate Code`.
4. Send the generated code to your teammate.

On the receiving machine:

1. Open SQL Notebook import.
2. Paste the code.
3. Click `Import`.

## Optional Custom Domain

You can map the Worker to your own domain or subdomain, for example:

- `https://share.example.com`

If you do that, use that URL in QuraMate instead of the default `workers.dev` address.

## Notes

- imported notebooks still come in as local drafts
- SQL does not auto-run after import
- expiry defaults to 7 days
- `DEFAULT_EXPIRY_DAYS` can be changed in `wrangler.toml`
- live session scaffolding now exists in the Worker for future codeshare-style sync, but the desktop UI is not wired to it yet

## Recommended Next Improvements

- add trust review UI before final import
- add rate limiting
- add sender label display in import review
- add optional revoke/delete endpoint
