import {
  LoadAIProviderKey,
  LoadSetting,
  LogClientEvent,
} from "../../../wailsjs/go/app/App";
import {
  AI_PROVIDER_DEFAULT_CONFIGS,
  AI_PROVIDER_DEFINITION_MAP,
  type AIProviderDefinition,
  type AIProviderId,
} from "./config";
import {
  completeWithProvider,
  AIProviderError,
  type AICompletionInput,
  type AICompletionResult,
  type AIMessage,
} from "./client";
import {
  getEffectiveEndpointOrigin,
  parseAiSettings,
  type AiSettingsSnapshot,
} from "./settings";

// Shared timeout for provider connectivity tests and runtime completions.
const AI_OPERATION_TIMEOUT_MS = 60000;

export type AIOperationType = "test" | "copilot" | "explain";
export type AIRequiredCapability = "chat" | "sql-only" | "vision";

export interface AIResolvedProviderContext {
  provider: AIProviderId;
  definition: AIProviderDefinition;
  apiKey: string;
  baseURL: string;
  model: string;
  endpointOrigin: string | null;
  settings: AiSettingsSnapshot;
}

export interface AITestResult {
  ok: boolean;
  provider: AIProviderId;
  latencyMs: number;
  message: string;
  details: string;
  endpointOrigin: string | null;
  testedAt: string;
}

export interface AICompletionWithMetadata extends AICompletionResult {
  provider: AIProviderId;
  latencyMs: number;
  endpointOrigin: string | null;
}

const trimTrailingSlash = (value: string) => value.replace(/\/+$/, "");

const mergeAbortSignals = (signal?: AbortSignal, timeoutMs?: number) => {
  if (!signal && !timeoutMs) {
    return { signal: undefined as AbortSignal | undefined, cleanup: () => {} };
  }

  const controller = new AbortController();
  let timeoutId: ReturnType<typeof setTimeout> | null = null;

  const abortFromSignal = () => controller.abort(signal?.reason);
  if (signal) {
    if (signal.aborted) {
      controller.abort(signal.reason);
    } else {
      signal.addEventListener("abort", abortFromSignal, { once: true });
    }
  }

  if (timeoutMs && timeoutMs > 0) {
    timeoutId = setTimeout(() => {
      controller.abort(new Error(`AI request timed out after ${timeoutMs}ms`));
    }, timeoutMs);
  }

  return {
    signal: controller.signal,
    cleanup: () => {
      if (signal) {
        signal.removeEventListener("abort", abortFromSignal);
      }
      if (timeoutId) {
        clearTimeout(timeoutId);
      }
    },
  };
};

const getCapabilitySupport = (
  definition: AIProviderDefinition,
  capability?: AIRequiredCapability,
): boolean => {
  if (!capability) {
    return true;
  }
  if (capability === "chat") {
    return definition.supportsChat;
  }
  if (capability === "sql-only") {
    return definition.supportsSqlOnly;
  }
  return definition.supportsVision;
};

const getCustomEndpointState = (provider: AIProviderId, baseURL: string): boolean => {
  const defaults = AI_PROVIDER_DEFAULT_CONFIGS[provider];
  const currentOrigin = getEffectiveEndpointOrigin(baseURL);
  const defaultOrigin = getEffectiveEndpointOrigin(defaults.baseURL);
  if (!currentOrigin || !defaultOrigin) {
    return true;
  }
  return currentOrigin !== defaultOrigin;
};

const logAiEvent = async (
  operation: AIOperationType,
  context: {
    provider: AIProviderId;
    protocol: string;
    model: string;
    endpointOrigin: string | null;
    latencyMs: number;
    ok: boolean;
    errorClass?: string;
    status?: number;
  },
) => {
  const payload = {
    scope: "ai",
    operation,
    provider: context.provider,
    protocol: context.protocol,
    model: context.model,
    endpointOrigin: context.endpointOrigin,
    latencyMs: context.latencyMs,
    ok: context.ok,
    errorClass: context.errorClass || null,
    status: context.status ?? null,
  };

  try {
    await LogClientEvent("INFO", `[AI] ${JSON.stringify(payload)}`);
  } catch {
    // Best-effort logging only.
  }
};

export const loadPersistedAiSettings = async (): Promise<AiSettingsSnapshot> => {
  const rawSettings = await LoadSetting("user_settings");
  return parseAiSettings(rawSettings);
};

export const resolveSavedProviderContext = async (
  providerOverride?: AIProviderId,
  capability?: AIRequiredCapability,
): Promise<AIResolvedProviderContext> => {
  const settings = await loadPersistedAiSettings();
  const provider = providerOverride || settings.provider;
  const definition = AI_PROVIDER_DEFINITION_MAP[provider];
  const apiKey = (await LoadAIProviderKey(provider)).trim();
  const providerConfig = settings.providerConfigs[provider] || AI_PROVIDER_DEFAULT_CONFIGS[provider];
  const baseURL = trimTrailingSlash(providerConfig.baseURL || "");
  const model = (providerConfig.model || "").trim();

  if (!definition) {
    throw new AIProviderError("Unsupported provider", provider);
  }
  if (!getCapabilitySupport(definition, capability)) {
    throw new AIProviderError(
      `Provider "${definition.label}" does not support this operation`,
      provider,
    );
  }
  if (definition.requiresApiKey && !apiKey) {
    throw new AIProviderError("No API key saved for this provider", provider);
  }
  if (!model) {
    throw new AIProviderError("Model is required", provider);
  }
  if (!baseURL) {
    throw new AIProviderError("Base URL is required", provider);
  }
  if (!getEffectiveEndpointOrigin(baseURL)) {
    throw new AIProviderError("Base URL is invalid", provider);
  }

  const isCustomEndpoint = getCustomEndpointState(provider, baseURL);
  if (isCustomEndpoint && !definition.supportsCustomBaseURL) {
    throw new AIProviderError(
      `Provider "${definition.label}" does not support custom endpoint hosts`,
      provider,
    );
  }
  if (isCustomEndpoint && !settings.allowCustomBaseURL) {
    throw new AIProviderError(
      "Custom AI endpoints are disabled. Enable them in Settings before using a non-default provider URL.",
      provider,
    );
  }

  return {
    provider,
    definition,
    apiKey,
    baseURL,
    model,
    endpointOrigin: getEffectiveEndpointOrigin(baseURL),
    settings,
  };
};

export const completeWithSavedProvider = async (
  messages: AIMessage[],
  options?: {
    provider?: AIProviderId;
    temperature?: number;
    maxTokens?: number;
    signal?: AbortSignal;
    operation?: AIOperationType;
    capability?: AIRequiredCapability;
    timeoutMs?: number;
  },
): Promise<AICompletionWithMetadata> => {
  const operation = options?.operation || "copilot";
  const startedAt = performance.now();
  const context = await resolveSavedProviderContext(
    options?.provider,
    options?.capability || "chat",
  );
  const { signal, cleanup } = mergeAbortSignals(
    options?.signal,
    options?.timeoutMs ?? AI_OPERATION_TIMEOUT_MS,
  );

  try {
    const result = await completeWithProvider({
      provider: context.provider,
      apiKey: context.apiKey,
      baseURL: context.baseURL,
      model: context.model,
      messages,
      temperature: options?.temperature,
      maxTokens: options?.maxTokens,
      signal,
    });

    const latencyMs = Math.round(performance.now() - startedAt);
    await logAiEvent(operation, {
      provider: context.provider,
      protocol: context.definition.protocol,
      model: context.model,
      endpointOrigin: context.endpointOrigin,
      latencyMs,
      ok: true,
    });

    return {
      ...result,
      provider: context.provider,
      latencyMs,
      endpointOrigin: context.endpointOrigin,
    };
  } catch (error) {
    const latencyMs = Math.round(performance.now() - startedAt);
    await logAiEvent(operation, {
      provider: context.provider,
      protocol: context.definition.protocol,
      model: context.model,
      endpointOrigin: context.endpointOrigin,
      latencyMs,
      ok: false,
      errorClass: error instanceof Error ? error.name : "UnknownError",
      status: error instanceof AIProviderError ? error.status : undefined,
    });
    throw error;
  } finally {
    cleanup();
  }
};

export const testSavedProviderConnection = async (
  providerOverride?: AIProviderId,
): Promise<AITestResult> => {
  const testedAt = new Date().toISOString();
  const completion = await completeWithSavedProvider(
    [{ role: "user", content: "Reply with exactly: CONNECTION_OK" }],
    {
      provider: providerOverride,
      operation: "test",
      capability: "chat",
      temperature: 0,
      maxTokens: 12,
    },
  );

  return {
    ok: true,
    provider: completion.provider,
    latencyMs: completion.latencyMs,
    message: completion.text || "Connected successfully",
    details: "",
    endpointOrigin: completion.endpointOrigin,
    testedAt,
  };
};

export const toProviderErrorTestResult = (
  provider: AIProviderId,
  error: unknown,
  latencyMs: number,
  endpointOrigin: string | null,
): AITestResult => {
  const message = error instanceof Error ? error.message : String(error);
  const details =
    error instanceof AIProviderError && error.details
      ? JSON.stringify(error.details).slice(0, 500)
      : "";

  return {
    ok: false,
    provider,
    latencyMs,
    message,
    details,
    endpointOrigin,
    testedAt: new Date().toISOString(),
  };
};

export const buildCompletionInputForProvider = (
  context: AIResolvedProviderContext,
  messages: AIMessage[],
  options?: Pick<AICompletionInput, "temperature" | "maxTokens" | "signal">,
): AICompletionInput => ({
  provider: context.provider,
  apiKey: context.apiKey,
  model: context.model,
  baseURL: context.baseURL,
  messages,
  temperature: options?.temperature,
  maxTokens: options?.maxTokens,
  signal: options?.signal,
});
