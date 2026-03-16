import { beforeEach, describe, expect, it, vi } from "vitest";

import { completeWithProvider } from "./client";

describe("AI client provider routing", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("routes anthropic providers to the messages endpoint", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      json: async () => ({
        content: [{ type: "text", text: "ok" }],
      }),
    } as Response);

    await completeWithProvider({
      provider: "anthropic",
      apiKey: "test-key",
      messages: [{ role: "user", content: "hello" }],
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.anthropic.com/v1/messages",
      expect.any(Object),
    );
  });

  it("routes google providers to the generateContent endpoint", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      json: async () => ({
        candidates: [{ content: { parts: [{ text: "ok" }] } }],
      }),
    } as Response);

    await completeWithProvider({
      provider: "google",
      apiKey: "test-key",
      messages: [{ role: "user", content: "hello" }],
    });

    expect(fetchMock.mock.calls[0]?.[0]).toContain(
      "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=test-key",
    );
  });

  it("routes openai-compatible providers through chat completions", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      json: async () => ({
        choices: [{ message: { content: "ok" } }],
      }),
    } as Response);

    await completeWithProvider({
      provider: "qwen",
      apiKey: "test-key",
      messages: [{ role: "user", content: "hello" }],
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
      expect.any(Object),
    );
  });
});
