import { DurableObject } from "cloudflare:workers";

interface Env {
  NOTEBOOK_SHARES_DB: D1Database;
  LIVE_NOTEBOOK_SESSIONS: DurableObjectNamespace<LiveNotebookSession>;
  DEFAULT_EXPIRY_DAYS?: string;
}

interface CreateShareRequest {
  scope: "notebook" | "sql";
  senderLabel?: string;
  payload: unknown;
}

interface ResolveShareRecord {
  code: string;
  expires_at: string;
  sender_label: string;
  payload_json: string;
}

interface ReusableShareRecord {
  code: string;
  expires_at: string;
}

interface CreateLiveSessionRequest {
  senderLabel?: string;
  payload: unknown;
}

interface ResolveLiveSessionRecord {
  code: string;
  session_id: string;
  sender_label: string;
  payload_json: string;
  expires_at: string;
}

interface LiveSessionMessage {
  type:
    | "presence"
    | "snapshot"
    | "patch"
    | "ack"
    | "session-expired"
    | "session-ended";
  payload?: unknown;
}

const corsHeaders = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Headers": "Content-Type",
  "Access-Control-Allow-Methods": "GET,POST,OPTIONS",
} as const;

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    try {
      if (request.method === "OPTIONS") {
        return new Response(null, {
          status: 204,
          headers: corsHeaders,
        });
      }

      const url = new URL(request.url);

      if (request.method === "POST" && url.pathname === "/api/notebook-shares") {
        return handleCreateShare(request, env);
      }

      if (
        request.method === "GET" &&
        url.pathname.startsWith("/api/notebook-shares/")
      ) {
        return handleResolveShare(url.pathname, env);
      }

      if (
        request.method === "POST" &&
        url.pathname === "/api/notebook-live-sessions"
      ) {
        return handleCreateLiveSession(request, env);
      }

      if (
        request.method === "POST" &&
        /^\/api\/notebook-live-sessions\/[^/]+\/join$/.test(url.pathname)
      ) {
        return handleJoinLiveSession(url.pathname, env);
      }

      if (
        request.method === "GET" &&
        /^\/api\/notebook-live-sessions\/[^/]+\/socket$/.test(url.pathname)
      ) {
        return handleLiveSessionSocket(url.pathname, request, env);
      }

      return json(
        {
          error: "Route not found.",
        },
        404,
      );
    } catch (error) {
      return json(
        {
          error: `Worker error: ${toErrorMessage(error)}`,
        },
        500,
      );
    }
  },
};

async function handleCreateShare(
  request: Request,
  env: Env,
): Promise<Response> {
  let body: CreateShareRequest;

  try {
    body = (await request.json()) as CreateShareRequest;
  } catch {
    return json({ error: "Invalid share payload." }, 400);
  }

  if (body.scope !== "notebook" && body.scope !== "sql") {
    return json({ error: "Scope must be notebook or sql." }, 400);
  }

  if (body.payload === undefined || body.payload === null) {
    return json({ error: "Payload is required." }, 400);
  }

  const payloadJson = JSON.stringify(body.payload);
  if (!payloadJson.trim()) {
    return json({ error: "Payload is required." }, 400);
  }

  if (payloadJson.length > 1_000_000) {
    return json({ error: "Payload is too large." }, 413);
  }

  const normalizedPayloadJson = normalizePayloadForReuse(body.payload);

  const now = new Date();
  const expiresAt = new Date(now);
  expiresAt.setUTCDate(now.getUTCDate() + getExpiryDays(env));

  const reusableShare = await env.NOTEBOOK_SHARES_DB.prepare(
    `
      SELECT
        code,
        expires_at
      FROM notebook_shares
      WHERE scope = ? AND payload_json = ? AND expires_at > ?
      ORDER BY expires_at DESC
      LIMIT 1
    `,
  )
    .bind(body.scope, normalizedPayloadJson, now.toISOString())
    .first<ReusableShareRecord>();

  if (reusableShare) {
    return json({
      code: reusableShare.code,
      expiresAt: reusableShare.expires_at,
    });
  }

  const code = await generateUniqueCode(env.NOTEBOOK_SHARES_DB);

  await env.NOTEBOOK_SHARES_DB.prepare(
    `
      INSERT INTO notebook_shares (
        code,
        scope,
        sender_label,
        payload_json,
        created_at,
        expires_at
      ) VALUES (?, ?, ?, ?, ?, ?)
    `,
  )
    .bind(
      code,
      body.scope,
      (body.senderLabel || "").trim(),
      normalizedPayloadJson,
      now.toISOString(),
      expiresAt.toISOString(),
    )
    .run();

  void cleanupExpiredShares(env.NOTEBOOK_SHARES_DB);

  return json({
    code,
    expiresAt: expiresAt.toISOString(),
  });
}

async function handleResolveShare(
  pathname: string,
  env: Env,
): Promise<Response> {
  const code = normalizeCode(pathname.replace("/api/notebook-shares/", ""));
  if (!code) {
    return json({ error: "Share code is required." }, 400);
  }

  const record = await env.NOTEBOOK_SHARES_DB.prepare(
    `
      SELECT
        code,
        expires_at,
        sender_label,
        payload_json
      FROM notebook_shares
      WHERE code = ?
    `,
  )
    .bind(code)
    .first<ResolveShareRecord>();

  if (!record) {
    return json({ error: "Share code not found." }, 404);
  }

  const expiresAt = new Date(record.expires_at);
  if (Number.isNaN(expiresAt.getTime())) {
    return json({ error: "Stored share code expiry is invalid." }, 500);
  }

  if (Date.now() > expiresAt.getTime()) {
    await env.NOTEBOOK_SHARES_DB.prepare(
      "DELETE FROM notebook_shares WHERE code = ?",
    )
      .bind(code)
      .run();

    return json(
      {
        error: "This share code expired. Ask the sender to generate a new one.",
      },
      410,
    );
  }

  return json({
    code: record.code,
    expiresAt: record.expires_at,
    senderLabel: record.sender_label,
    payload: JSON.parse(record.payload_json),
  });
}

async function handleCreateLiveSession(
  request: Request,
  env: Env,
): Promise<Response> {
  let body: CreateLiveSessionRequest;

  try {
    body = (await request.json()) as CreateLiveSessionRequest;
  } catch {
    return json({ error: "Invalid live session payload." }, 400);
  }

  if (body.payload === undefined || body.payload === null) {
    return json({ error: "Payload is required." }, 400);
  }

  const payloadJson = normalizePayloadForReuse(body.payload);
  if (!payloadJson.trim()) {
    return json({ error: "Payload is required." }, 400);
  }

  if (payloadJson.length > 1_000_000) {
    return json({ error: "Payload is too large." }, 413);
  }

  const now = new Date();
  const expiresAt = new Date(now);
  expiresAt.setUTCDate(now.getUTCDate() + getExpiryDays(env));

  const reusableSession = await env.NOTEBOOK_SHARES_DB.prepare(
    `
      SELECT
        code,
        session_id,
        expires_at
      FROM live_notebook_sessions
      WHERE payload_json = ? AND expires_at > ?
      ORDER BY expires_at DESC
      LIMIT 1
    `,
  )
    .bind(payloadJson, now.toISOString())
    .first<{ code: string; session_id: string; expires_at: string }>();

  if (reusableSession) {
    return json({
      code: reusableSession.code,
      expiresAt: reusableSession.expires_at,
      sessionId: reusableSession.session_id,
    });
  }

  const code = await generateUniqueCode(env.NOTEBOOK_SHARES_DB, "live_notebook_sessions");
  const sessionId = env.LIVE_NOTEBOOK_SESSIONS.idFromName(code);
  const stub = env.LIVE_NOTEBOOK_SESSIONS.get(sessionId);

  await stub.fetch("https://live-session.internal/bootstrap", {
    method: "POST",
    body: JSON.stringify({
      code,
      expiresAt: expiresAt.toISOString(),
    }),
  });

  await env.NOTEBOOK_SHARES_DB.prepare(
    `
      INSERT INTO live_notebook_sessions (
        code,
        session_id,
        sender_label,
        payload_json,
        created_at,
        expires_at
      ) VALUES (?, ?, ?, ?, ?, ?)
    `,
  )
    .bind(
      code,
      sessionId.toString(),
      (body.senderLabel || "").trim(),
      payloadJson,
      now.toISOString(),
      expiresAt.toISOString(),
    )
    .run();

  return json({
    code,
    expiresAt: expiresAt.toISOString(),
    sessionId: sessionId.toString(),
  });
}

async function handleJoinLiveSession(
  pathname: string,
  env: Env,
): Promise<Response> {
  const code = normalizeCode(
    pathname.replace("/api/notebook-live-sessions/", "").replace("/join", ""),
  );
  if (!code) {
    return json({ error: "Live session code is required." }, 400);
  }

  const record = await loadLiveSessionRecord(code, env);
  if (record instanceof Response) {
    return record;
  }

  return json({
    code: record.code,
    sessionId: record.session_id,
    senderLabel: record.sender_label,
    expiresAt: record.expires_at,
    payload: JSON.parse(record.payload_json),
  });
}

async function handleLiveSessionSocket(
  pathname: string,
  request: Request,
  env: Env,
): Promise<Response> {
  const code = normalizeCode(
    pathname.replace("/api/notebook-live-sessions/", "").replace("/socket", ""),
  );
  if (!code) {
    return json({ error: "Live session code is required." }, 400);
  }

  const record = await loadLiveSessionRecord(code, env);
  if (record instanceof Response) {
    return record;
  }

  const sessionId = env.LIVE_NOTEBOOK_SESSIONS.idFromString(record.session_id);
  const stub = env.LIVE_NOTEBOOK_SESSIONS.get(sessionId);

  return stub.fetch(
    new Request("https://live-session.internal/connect", request),
  );
}

async function loadLiveSessionRecord(
  code: string,
  env: Env,
): Promise<ResolveLiveSessionRecord | Response> {
  const record = await env.NOTEBOOK_SHARES_DB.prepare(
    `
      SELECT
        code,
        session_id,
        sender_label,
        payload_json,
        expires_at
      FROM live_notebook_sessions
      WHERE code = ?
    `,
  )
    .bind(code)
    .first<ResolveLiveSessionRecord>();

  if (!record) {
    return json({ error: "Live session code not found." }, 404);
  }

  const expiresAt = new Date(record.expires_at);
  if (Number.isNaN(expiresAt.getTime())) {
    return json({ error: "Stored live session expiry is invalid." }, 500);
  }

  if (Date.now() > expiresAt.getTime()) {
    await env.NOTEBOOK_SHARES_DB.prepare(
      "DELETE FROM live_notebook_sessions WHERE code = ?",
    )
      .bind(code)
      .run();

    return json(
      {
        error: "This live session expired. Reconnect with a new code.",
      },
      410,
    );
  }

  return record;
}

async function generateUniqueCode(
  database: D1Database,
  tableName = "notebook_shares",
): Promise<string> {
  for (let index = 0; index < 20; index += 1) {
    const code = randomCode();
    const existing = await database
      .prepare(`SELECT code FROM ${tableName} WHERE code = ?`)
      .bind(code)
      .first<{ code: string }>();

    if (!existing) {
      return code;
    }
  }

  throw new Error("Could not generate unique share code.");
}

function randomCode(): string {
  const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789";
  let body = "";
  for (let index = 0; index < 6; index += 1) {
    body += alphabet[Math.floor(Math.random() * alphabet.length)];
  }
  return `QN-${body}`;
}

function normalizeCode(value: string): string {
  return value.trim().toUpperCase();
}

function normalizePayloadForReuse(payload: unknown): string {
  if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
    return JSON.stringify(payload);
  }

  const normalizedPayload = {
    ...(payload as Record<string, unknown>),
    exportedAt: "",
  };

  return JSON.stringify(normalizedPayload);
}

function getExpiryDays(env: Env): number {
  const value = Number(env.DEFAULT_EXPIRY_DAYS || "7");
  if (!Number.isFinite(value) || value <= 0) {
    return 7;
  }
  return Math.min(Math.floor(value), 30);
}

async function cleanupExpiredShares(database: D1Database): Promise<void> {
  await database
    .prepare("DELETE FROM notebook_shares WHERE expires_at < ?")
    .bind(new Date().toISOString())
    .run();
}

function json(payload: unknown, status = 200): Response {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      "Content-Type": "application/json",
      ...corsHeaders,
    },
  });
}

function toErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }

  return String(error);
}

export class LiveNotebookSession extends DurableObject {
  private sockets = new Set<WebSocket>();
  private expiresAt = "";
  private code = "";

  constructor(ctx: DurableObjectState, env: Env) {
    super(ctx, env);
  }

  override async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);

    if (request.method === "POST" && url.pathname === "/bootstrap") {
      const body = (await request.json()) as { code?: string; expiresAt?: string };
      this.code = body.code || "";
      this.expiresAt = body.expiresAt || "";
      return json({ ok: true });
    }

    if (url.pathname === "/connect" && request.headers.get("Upgrade") === "websocket") {
      if (this.isExpired()) {
        return json({ error: "This live session expired. Reconnect with a new code." }, 410);
      }

      const pair = new WebSocketPair();
      const client = pair[0];
      const server = pair[1];

      this.ctx.acceptWebSocket(server);
      this.sockets.add(server);
      this.broadcast({
        type: "presence",
        payload: {
          code: this.code,
          peers: this.sockets.size,
        },
      });

      return new Response(null, {
        status: 101,
        webSocket: client,
      });
    }

    return json({ error: "Route not found." }, 404);
  }

  override webSocketMessage(socket: WebSocket, message: string | ArrayBuffer): void {
    if (this.isExpired()) {
      socket.send(JSON.stringify({ type: "session-expired" satisfies LiveSessionMessage["type"] }));
      socket.close(4001, "Session expired");
      this.sockets.delete(socket);
      return;
    }

    const raw = typeof message === "string" ? message : new TextDecoder().decode(message);

    let parsed: LiveSessionMessage | null = null;
    try {
      parsed = JSON.parse(raw) as LiveSessionMessage;
    } catch {
      socket.send(JSON.stringify({ type: "ack", payload: { ok: false, error: "Invalid message" } }));
      return;
    }

    for (const peer of this.sockets) {
      if (peer !== socket) {
        peer.send(JSON.stringify(parsed));
      }
    }

    socket.send(JSON.stringify({ type: "ack", payload: { ok: true } }));
  }

  override webSocketClose(socket: WebSocket): void {
    this.sockets.delete(socket);
    this.broadcast({
      type: "presence",
      payload: {
        code: this.code,
        peers: this.sockets.size,
      },
    });
  }

  private isExpired(): boolean {
    if (!this.expiresAt) {
      return false;
    }

    const expiresAt = new Date(this.expiresAt);
    return Number.isNaN(expiresAt.getTime()) ? false : Date.now() > expiresAt.getTime();
  }

  private broadcast(message: LiveSessionMessage): void {
    const serialized = JSON.stringify(message);
    for (const socket of this.sockets) {
      socket.send(serialized);
    }
  }
}
