import {
  AI_PROVIDER_DEFAULT_CONFIGS,
  AI_PROVIDER_IDS,
  type AIProviderConfigMap,
  type AIProviderId,
} from "./config";

export type AIProviderConnectionStatus = "configured" | "not_configured" | "error";
export type AIProviderLastTestResult = "success" | "failure" | null;

export interface AIProviderState {
  status: AIProviderConnectionStatus;
  lastTestResult: AIProviderLastTestResult;
  lastTestMessage: string;
  lastTestedAt: string | null;
  effectiveEndpointOrigin: string | null;
}

export type AIProviderStateMap = Record<AIProviderId, AIProviderState>;

export interface AiSettingsSnapshot {
  provider: AIProviderId;
  providerConfigs: AIProviderConfigMap;
  providerState: AIProviderStateMap;
  allowCustomBaseURL: boolean;
  shareSchemaContext: boolean;
  shareQueryHistory: boolean;
  shareResultSample: boolean;
  shareExecutionPlan: boolean;
}

const cloneDefaults = <T>(value: T): T => JSON.parse(JSON.stringify(value)) as T;

export const createDefaultProviderState = (): AIProviderState => ({
  status: "not_configured",
  lastTestResult: null,
  lastTestMessage: "",
  lastTestedAt: null,
  effectiveEndpointOrigin: null,
});

export const createDefaultProviderStateMap = (): AIProviderStateMap =>
  AI_PROVIDER_IDS.reduce((acc, providerId) => {
    acc[providerId] = createDefaultProviderState();
    return acc;
  }, {} as AIProviderStateMap);

export const createDefaultAiSettingsSnapshot = (): AiSettingsSnapshot => ({
  provider: "openai",
  providerConfigs: cloneDefaults(AI_PROVIDER_DEFAULT_CONFIGS),
  providerState: createDefaultProviderStateMap(),
  allowCustomBaseURL: false,
  shareSchemaContext: false,
  shareQueryHistory: false,
  shareResultSample: false,
  shareExecutionPlan: false,
});

const trimTrailingSlash = (value: string) => value.replace(/\/+$/, "");

export const normalizeProviderId = (value: unknown): AIProviderId => {
  const normalized = typeof value === "string" ? value.trim().toLowerCase() : "";
  return (AI_PROVIDER_IDS as readonly string[]).includes(normalized)
    ? (normalized as AIProviderId)
    : "openai";
};

export const getEffectiveEndpointOrigin = (value: string): string | null => {
  try {
    return new URL(trimTrailingSlash(value.trim())).origin.toLowerCase();
  } catch {
    return null;
  }
};

export const normalizeProviderState = (value: unknown): AIProviderState => {
  const defaults = createDefaultProviderState();
  if (!value || typeof value !== "object") {
    return defaults;
  }

  const candidate = value as Partial<AIProviderState>;
  const lastTestResult =
    candidate.lastTestResult === "success" || candidate.lastTestResult === "failure"
      ? candidate.lastTestResult
      : null;
  const status =
    candidate.status === "configured" ||
    candidate.status === "not_configured" ||
    candidate.status === "error"
      ? candidate.status
      : defaults.status;

  return {
    status,
    lastTestResult,
    lastTestMessage:
      typeof candidate.lastTestMessage === "string" ? candidate.lastTestMessage.trim() : "",
    lastTestedAt:
      typeof candidate.lastTestedAt === "string" && candidate.lastTestedAt.trim()
        ? candidate.lastTestedAt
        : null,
    effectiveEndpointOrigin:
      typeof candidate.effectiveEndpointOrigin === "string" &&
      candidate.effectiveEndpointOrigin.trim()
        ? candidate.effectiveEndpointOrigin.trim().toLowerCase()
        : null,
  };
};

export const normalizeAiSettings = (input: unknown): AiSettingsSnapshot => {
  const defaults = createDefaultAiSettingsSnapshot();
  if (!input || typeof input !== "object") {
    return defaults;
  }

  const raw = input as {
    provider?: unknown;
    providerConfigs?: Partial<AIProviderConfigMap>;
    providerState?: Partial<AIProviderStateMap>;
    baseURL?: unknown;
    model?: unknown;
    allowCustomBaseURL?: unknown;
    shareSchemaContext?: unknown;
    shareQueryHistory?: unknown;
    shareResultSample?: unknown;
    shareExecutionPlan?: unknown;
  };

  const provider = normalizeProviderId(raw.provider);
  const normalizedConfigs = cloneDefaults(AI_PROVIDER_DEFAULT_CONFIGS);
  for (const providerId of AI_PROVIDER_IDS) {
    const existing = raw.providerConfigs?.[providerId];
    if (existing && typeof existing === "object") {
      const candidate = existing as { baseURL?: unknown; model?: unknown };
      if (typeof candidate.baseURL === "string" && candidate.baseURL.trim()) {
        normalizedConfigs[providerId].baseURL = trimTrailingSlash(candidate.baseURL.trim());
      }
      if (typeof candidate.model === "string" && candidate.model.trim()) {
        normalizedConfigs[providerId].model = candidate.model.trim();
      }
    }
  }

  if (typeof raw.baseURL === "string" && raw.baseURL.trim()) {
    normalizedConfigs[provider].baseURL = trimTrailingSlash(raw.baseURL.trim());
  }
  if (typeof raw.model === "string" && raw.model.trim()) {
    normalizedConfigs[provider].model = raw.model.trim();
  }

  const providerState = createDefaultProviderStateMap();
  for (const providerId of AI_PROVIDER_IDS) {
    providerState[providerId] = normalizeProviderState(raw.providerState?.[providerId]);
    providerState[providerId].effectiveEndpointOrigin =
      getEffectiveEndpointOrigin(normalizedConfigs[providerId].baseURL);
  }

  return {
    provider,
    providerConfigs: normalizedConfigs,
    providerState,
    allowCustomBaseURL: raw.allowCustomBaseURL === true,
    shareSchemaContext: raw.shareSchemaContext === true,
    shareQueryHistory: raw.shareQueryHistory === true,
    shareResultSample: raw.shareResultSample === true,
    shareExecutionPlan: raw.shareExecutionPlan === true,
  };
};

export const parseAiSettings = (settingsJson: string): AiSettingsSnapshot => {
  if (!settingsJson) {
    return createDefaultAiSettingsSnapshot();
  }

  try {
    const parsed = JSON.parse(settingsJson) as { ai?: unknown };
    return normalizeAiSettings(parsed?.ai);
  } catch {
    return createDefaultAiSettingsSnapshot();
  }
};

export const mergeAiSettingsIntoUserSettings = (
  settingsJson: string,
  aiSettings: AiSettingsSnapshot,
): string => {
  let parsed: Record<string, unknown> = {};
  if (settingsJson) {
    try {
      parsed = JSON.parse(settingsJson) as Record<string, unknown>;
    } catch {
      parsed = {};
    }
  }

  const nextSettings = {
    ...parsed,
    ai: {
      provider: aiSettings.provider,
      providerConfigs: cloneDefaults(aiSettings.providerConfigs),
      providerState: cloneDefaults(aiSettings.providerState),
      allowCustomBaseURL: aiSettings.allowCustomBaseURL,
      shareSchemaContext: aiSettings.shareSchemaContext,
      shareQueryHistory: aiSettings.shareQueryHistory,
      shareResultSample: aiSettings.shareResultSample,
      shareExecutionPlan: aiSettings.shareExecutionPlan,
    },
  };

  return JSON.stringify(nextSettings);
};
