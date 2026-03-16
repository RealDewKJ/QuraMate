import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../../../wailsjs/go/app/App", () => ({
  LoadAIProviderKey: vi.fn(),
  LoadSetting: vi.fn(),
  LogClientEvent: vi.fn(),
}));

vi.mock("./client", () => ({
  AIProviderError: class AIProviderError extends Error {
    provider: string;
    status?: number;
    details?: unknown;

    constructor(message: string, provider: string, status?: number, details?: unknown) {
      super(message);
      this.name = "AIProviderError";
      this.provider = provider;
      this.status = status;
      this.details = details;
    }
  },
  completeWithProvider: vi.fn(),
}));

import { LoadAIProviderKey, LoadSetting } from "../../../wailsjs/go/app/App";
import { completeWithProvider } from "./client";
import {
  completeWithSavedProvider,
  resolveSavedProviderContext,
  testSavedProviderConnection,
} from "./service";

const mockedLoadSetting = vi.mocked(LoadSetting);
const mockedLoadAIProviderKey = vi.mocked(LoadAIProviderKey);
const mockedCompleteWithProvider = vi.mocked(completeWithProvider);

describe("AI runtime service", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedLoadSetting.mockResolvedValue(
      JSON.stringify({
        ai: {
          provider: "openai",
          providerConfigs: {
            openai: {
              baseURL: "https://api.openai.com/v1",
              model: "gpt-4o-mini",
            },
            google: {
              baseURL: "https://generativelanguage.googleapis.com",
              model: "gemini-2.0-flash",
            },
            local: {
              baseURL: "http://localhost:11434/v1",
              model: "llama3.1:8b",
            },
          },
          providerState: {},
          allowCustomBaseURL: false,
        },
      }),
    );
    mockedLoadAIProviderKey.mockImplementation(async (provider) =>
      provider === "local" ? "" : "persisted-key",
    );
    mockedCompleteWithProvider.mockResolvedValue({
      text: "CONNECTION_OK",
      raw: {},
    });
  });

  it("blocks custom endpoints when opt-in is disabled", async () => {
    mockedLoadSetting.mockResolvedValue(
      JSON.stringify({
        ai: {
          provider: "openai",
          providerConfigs: {
            openai: {
              baseURL: "https://proxy.example.com/v1",
              model: "gpt-4o-mini",
            },
          },
          allowCustomBaseURL: false,
        },
      }),
    );

    await expect(resolveSavedProviderContext("openai", "chat")).rejects.toThrow(
      "Custom AI endpoints are disabled",
    );
  });

  it("allows local providers without an API key", async () => {
    const context = await resolveSavedProviderContext("local", "chat");
    expect(context.provider).toBe("local");
    expect(context.apiKey).toBe("");
  });

  it("uses persisted settings for runtime completions", async () => {
    await completeWithSavedProvider(
      [{ role: "user", content: "hello" }],
      { provider: "openai", operation: "copilot" },
    );

    expect(mockedCompleteWithProvider).toHaveBeenCalledWith(
      expect.objectContaining({
        provider: "openai",
        apiKey: "persisted-key",
        baseURL: "https://api.openai.com/v1",
        model: "gpt-4o-mini",
      }),
    );
  });

  it("persists provider test flow through the shared runtime service", async () => {
    const result = await testSavedProviderConnection("openai");

    expect(result.ok).toBe(true);
    expect(result.provider).toBe("openai");
    expect(mockedCompleteWithProvider).toHaveBeenCalledTimes(1);
  });
});
