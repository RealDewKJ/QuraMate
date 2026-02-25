import { LoadAIProviderKey, LoadSetting } from "../../wailsjs/go/main/App";
import {
  AI_PROVIDER_DEFAULT_CONFIGS,
  AI_PROVIDER_IDS,
  type AIProviderConfigMap,
  type AIProviderId,
} from "../lib/ai/config";
import {
  AIProviderError,
  completeWithProvider,
  type AIMessage,
  type AICompletionResult,
} from "../lib/ai/client";

export interface AiSettingsSnapshot {
  provider: AIProviderId;
  providerConfigs: AIProviderConfigMap;
}

const deepCloneDefaults = (): AIProviderConfigMap =>
  JSON.parse(JSON.stringify(AI_PROVIDER_DEFAULT_CONFIGS)) as AIProviderConfigMap;

const parseAiSettings = (settingsJson: string): AiSettingsSnapshot => {
  const defaults: AiSettingsSnapshot = {
    provider: "openai",
    providerConfigs: deepCloneDefaults(),
  };
  if (!settingsJson) return defaults;

  try {
    const parsed = JSON.parse(settingsJson) as {
      ai?: { provider?: string; providerConfigs?: Partial<AIProviderConfigMap> };
    };
    const provider = parsed.ai?.provider;
    if (provider && (AI_PROVIDER_IDS as readonly string[]).includes(provider)) {
      defaults.provider = provider as AIProviderId;
    }

    const configured = parsed.ai?.providerConfigs;
    if (configured && typeof configured === "object") {
      AI_PROVIDER_IDS.forEach((providerId) => {
        const providerConfig = configured[providerId];
        if (!providerConfig) return;
        if (typeof providerConfig.baseURL === "string" && providerConfig.baseURL.trim()) {
          defaults.providerConfigs[providerId].baseURL = providerConfig.baseURL.trim();
        }
        if (typeof providerConfig.model === "string" && providerConfig.model.trim()) {
          defaults.providerConfigs[providerId].model = providerConfig.model.trim();
        }
      });
    }
  } catch {
    return defaults;
  }

  return defaults;
};

export const loadAiSettings = async (): Promise<AiSettingsSnapshot> => {
  const rawSettings = await LoadSetting("user_settings");
  return parseAiSettings(rawSettings);
};

export const completeWithSavedProvider = async (
  messages: AIMessage[],
  options?: { provider?: AIProviderId; temperature?: number; maxTokens?: number; signal?: AbortSignal },
): Promise<AICompletionResult> => {
  const settings = await loadAiSettings();
  const provider = options?.provider || settings.provider;
  const apiKey = await LoadAIProviderKey(provider);
  if (provider !== "local" && !apiKey?.trim()) {
    throw new AIProviderError("No API key saved for this provider", provider);
  }

  const runtimeConfig = settings.providerConfigs[provider] || AI_PROVIDER_DEFAULT_CONFIGS[provider];
  return completeWithProvider({
    provider,
    apiKey,
    baseURL: runtimeConfig.baseURL,
    model: runtimeConfig.model,
    messages,
    temperature: options?.temperature,
    maxTokens: options?.maxTokens,
    signal: options?.signal,
  });
};
