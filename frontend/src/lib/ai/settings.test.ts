import { describe, expect, it } from "vitest";

import {
  createDefaultAiSettingsSnapshot,
  parseAiSettings,
  normalizeAiSettings,
} from "./settings";

describe("AI settings normalization", () => {
  it("loads defaults when settings are missing", () => {
    expect(parseAiSettings("")).toEqual(createDefaultAiSettingsSnapshot());
  });

  it("merges partial provider configs and strips legacy key fields", () => {
    const settings = parseAiSettings(
      JSON.stringify({
        ai: {
          provider: "google",
          apiKey: "legacy-inline-key",
          apiKeys: {
            openai: "legacy-openai-key",
          },
          providerConfigs: {
            google: {
              baseURL: "https://example.com/custom/ ",
              model: " gemini-custom ",
            },
          },
        },
      }),
    );

    expect(settings.provider).toBe("google");
    expect(settings.providerConfigs.google.baseURL).toBe("https://example.com/custom");
    expect(settings.providerConfigs.google.model).toBe("gemini-custom");
    expect(settings.providerState.google.effectiveEndpointOrigin).toBe("https://example.com");
  });

  it("preserves provider state and normalizes invalid provider selection", () => {
    const settings = normalizeAiSettings({
      provider: "invalid-provider",
      providerState: {
        openai: {
          status: "error",
          lastTestResult: "failure",
          lastTestMessage: "bad request",
          lastTestedAt: "2026-03-16T12:00:00.000Z",
          effectiveEndpointOrigin: "https://api.openai.com",
        },
      },
    });

    expect(settings.provider).toBe("openai");
    expect(settings.providerState.openai.status).toBe("error");
    expect(settings.providerState.openai.lastTestMessage).toBe("bad request");
  });
});
