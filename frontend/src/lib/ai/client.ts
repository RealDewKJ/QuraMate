import {
  AI_PROVIDER_DEFAULT_CONFIGS,
  type AIProviderConfigMap,
  type AIProviderId,
} from "./config";

export type AIMessageRole = "system" | "user" | "assistant";

export interface AIMessage {
  role: AIMessageRole;
  content: string;
}

export interface AICompletionInput {
  provider: AIProviderId;
  apiKey: string;
  model?: string;
  baseURL?: string;
  messages: AIMessage[];
  temperature?: number;
  maxTokens?: number;
  signal?: AbortSignal;
}

export interface AICompletionResult {
  text: string;
  raw: unknown;
}

export class AIProviderError extends Error {
  constructor(
    message: string,
    public readonly provider: AIProviderId,
    public readonly status?: number,
    public readonly details?: unknown,
  ) {
    super(message);
    this.name = "AIProviderError";
  }
}

const trimTrailingSlash = (value: string) => value.replace(/\/+$/, "");

const resolveProviderConfig = (provider: AIProviderId, model?: string, baseURL?: string) => {
  const defaults = AI_PROVIDER_DEFAULT_CONFIGS[provider];
  return {
    model: model?.trim() || defaults.model,
    baseURL: trimTrailingSlash(baseURL?.trim() || defaults.baseURL),
  };
};

const toOpenAIStyleMessages = (messages: AIMessage[]) =>
  messages.map((message) => ({ role: message.role, content: message.content }));

const callJson = async (provider: AIProviderId, url: string, init: RequestInit): Promise<unknown> => {
  const response = await fetch(url, init);
  const json = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new AIProviderError("AI provider request failed", provider, response.status, json);
  }
  return json;
};

const extractTextFromUnknown = (value: unknown): string => {
  if (typeof value === "string") return value;
  if (Array.isArray(value)) {
    return value
      .map((item) => {
        if (typeof item === "string") return item;
        if (item && typeof item === "object") {
          const textValue = (item as { text?: unknown }).text;
          if (typeof textValue === "string") return textValue;
        }
        return "";
      })
      .join("")
      .trim();
  }
  return "";
};

const completeOpenAICompatible = async (input: AICompletionInput): Promise<AICompletionResult> => {
  const { baseURL, model } = resolveProviderConfig(input.provider, input.model, input.baseURL);
  const body: Record<string, unknown> = {
    model,
    messages: toOpenAIStyleMessages(input.messages),
  };
  if (typeof input.temperature === "number") body.temperature = input.temperature;
  if (typeof input.maxTokens === "number") body.max_tokens = input.maxTokens;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (input.apiKey?.trim()) {
    headers.Authorization = `Bearer ${input.apiKey}`;
  }

  const json = (await callJson(input.provider, `${baseURL}/chat/completions`, {
    method: "POST",
    headers,
    body: JSON.stringify(body),
    signal: input.signal,
  })) as {
    choices?: Array<{
      text?: string;
      finish_reason?: string;
      message?: { content?: unknown; reasoning_content?: string };
      delta?: { content?: unknown };
    }>;
    error?: unknown;
  };

  const choice = json.choices?.[0];
  const text =
    extractTextFromUnknown(choice?.message?.content) ||
    extractTextFromUnknown(choice?.delta?.content) ||
    (choice?.message?.reasoning_content || "").trim() ||
    (choice?.text || "").trim();

  if (!text) {
    throw new AIProviderError("Provider returned no text content", input.provider, undefined, {
      finishReason: choice?.finish_reason,
      error: json.error,
      choice: choice || null,
    });
  }
  return { text, raw: json };
};

const completeAnthropic = async (input: AICompletionInput): Promise<AICompletionResult> => {
  const { baseURL, model } = resolveProviderConfig(input.provider, input.model, input.baseURL);
  const body: Record<string, unknown> = {
    model,
    max_tokens: typeof input.maxTokens === "number" ? input.maxTokens : 1024,
    messages: input.messages
      .filter((message) => message.role !== "system")
      .map((message) => ({
        role: message.role === "assistant" ? "assistant" : "user",
        content: message.content,
      })),
  };
  if (typeof input.temperature === "number") body.temperature = input.temperature;

  const systemMessages = input.messages
    .filter((message) => message.role === "system")
    .map((message) => message.content.trim())
    .filter(Boolean);
  if (systemMessages.length > 0) {
    body.system = systemMessages.join("\n");
  }

  const json = (await callJson(input.provider, `${baseURL}/v1/messages`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "x-api-key": input.apiKey,
      "anthropic-version": "2023-06-01",
    },
    body: JSON.stringify(body),
    signal: input.signal,
  })) as {
    content?: Array<{ type?: string; text?: string }>;
    stop_reason?: string;
    error?: unknown;
  };

  const text = (json.content || [])
    .filter((item) => item.type === "text")
    .map((item) => item.text || "")
    .join("")
    .trim();

  if (!text) {
    throw new AIProviderError("Provider returned no text content", input.provider, undefined, {
      stopReason: json.stop_reason,
      error: json.error,
      contentBlocks: json.content || [],
    });
  }

  return { text, raw: json };
};

const completeGoogle = async (input: AICompletionInput): Promise<AICompletionResult> => {
  const { baseURL, model } = resolveProviderConfig(input.provider, input.model, input.baseURL);
  const contents = input.messages
    .filter((message) => message.role !== "system")
    .map((message) => ({
      role: message.role === "assistant" ? "model" : "user",
      parts: [{ text: message.content }],
    }));

  const body: Record<string, unknown> = { contents };
  if (typeof input.temperature === "number" || typeof input.maxTokens === "number") {
    body.generationConfig = {
      ...(typeof input.temperature === "number" ? { temperature: input.temperature } : {}),
      ...(typeof input.maxTokens === "number" ? { maxOutputTokens: input.maxTokens } : {}),
    };
  }

  const systemMessages = input.messages
    .filter((message) => message.role === "system")
    .map((message) => message.content.trim())
    .filter(Boolean);
  if (systemMessages.length > 0) {
    body.systemInstruction = {
      parts: [{ text: systemMessages.join("\n") }],
    };
  }

  const endpoint = `${baseURL}/v1beta/models/${encodeURIComponent(model)}:generateContent?key=${encodeURIComponent(
    input.apiKey,
  )}`;
  const json = (await callJson(input.provider, endpoint, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    signal: input.signal,
  })) as {
    candidates?: Array<{ content?: { parts?: Array<{ text?: string }> }; finishReason?: string }>;
    promptFeedback?: { blockReason?: string };
    error?: unknown;
  };

  const text = (json.candidates?.[0]?.content?.parts || [])
    .map((part) => part.text || "")
    .join("")
    .trim();

  if (!text) {
    throw new AIProviderError("Provider returned no text content", input.provider, undefined, {
      finishReason: json.candidates?.[0]?.finishReason,
      blockReason: json.promptFeedback?.blockReason,
      error: json.error,
      candidate: json.candidates?.[0] || null,
    });
  }

  return { text, raw: json };
};

export const completeWithProvider = async (input: AICompletionInput): Promise<AICompletionResult> => {
  const allowEmptyApiKey = input.provider === "local";
  if (!allowEmptyApiKey && !input.apiKey?.trim()) {
    throw new AIProviderError("API key is required", input.provider);
  }
  if (!input.messages?.length) {
    throw new AIProviderError("messages is required", input.provider);
  }

  if (input.provider === "anthropic") {
    return completeAnthropic(input);
  }
  if (input.provider === "google") {
    return completeGoogle(input);
  }
  return completeOpenAICompatible(input);
};

export interface AIClientSettings {
  provider: AIProviderId;
  providerConfigs: AIProviderConfigMap;
}

export const completeWithUserSettings = async (
  settings: AIClientSettings,
  apiKey: string,
  messages: AIMessage[],
  options?: Pick<AICompletionInput, "temperature" | "maxTokens" | "signal">,
) => {
  const activeConfig =
    settings.providerConfigs[settings.provider] || AI_PROVIDER_DEFAULT_CONFIGS[settings.provider];
  return completeWithProvider({
    provider: settings.provider,
    apiKey,
    model: activeConfig.model,
    baseURL: activeConfig.baseURL,
    messages,
    temperature: options?.temperature,
    maxTokens: options?.maxTokens,
    signal: options?.signal,
  });
};
