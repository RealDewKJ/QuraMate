export const AI_PROVIDER_IDS = [
  "openai",
  "anthropic",
  "google",
  "glm",
  "kimi",
  "qwen",
  "minimax",
  "local",
] as const;

export type AIProviderId = (typeof AI_PROVIDER_IDS)[number];

export interface AIProviderDefinition {
  value: AIProviderId;
  label: string;
  defaultBaseURL: string;
  defaultModel: string;
  modelOptions: string[];
}

export const AI_PROVIDER_DEFINITIONS: AIProviderDefinition[] = [
  {
    value: "openai",
    label: "OpenAI (ChatGPT)",
    defaultBaseURL: "https://api.openai.com/v1",
    defaultModel: "gpt-4o-mini",
    modelOptions: ["gpt-4o-mini", "gpt-4.1-mini", "gpt-4.1"],
  },
  {
    value: "anthropic",
    label: "Anthropic (Claude)",
    defaultBaseURL: "https://api.anthropic.com",
    defaultModel: "claude-3-5-sonnet-latest",
    modelOptions: [
      "claude-3-5-sonnet-latest",
      "claude-3-5-haiku-latest",
      "claude-3-opus-latest",
    ],
  },
  {
    value: "google",
    label: "Google (Gemini)",
    defaultBaseURL: "https://generativelanguage.googleapis.com",
    defaultModel: "gemini-2.0-flash",
    modelOptions: ["gemini-2.0-flash", "gemini-2.0-pro", "gemini-1.5-pro"],
  },
  {
    value: "glm",
    label: "GLM (Zhipu AI)",
    defaultBaseURL: "https://open.bigmodel.cn/api/paas/v4",
    defaultModel: "glm-4-flash",
    modelOptions: ["glm-4-flash", "glm-4-plus", "glm-4-air", "glm-4.7"],
  },
  {
    value: "kimi",
    label: "Kimi (Moonshot AI)",
    defaultBaseURL: "https://api.moonshot.cn/v1",
    defaultModel: "moonshot-v1-8k",
    modelOptions: ["moonshot-v1-8k", "moonshot-v1-32k", "moonshot-v1-128k"],
  },
  {
    value: "qwen",
    label: "Qwen (Alibaba)",
    defaultBaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
    defaultModel: "qwen-plus",
    modelOptions: ["qwen-plus", "qwen-turbo", "qwen-max"],
  },
  {
    value: "minimax",
    label: "MiniMax",
    defaultBaseURL: "https://api.minimax.chat/v1",
    defaultModel: "abab6.5s-chat",
    modelOptions: ["abab6.5s-chat", "abab6.5-chat", "abab5.5-chat"],
  },
  {
    value: "local",
    label: "Local (Ollama / Llama.cpp)",
    defaultBaseURL: "http://localhost:11434/v1",
    defaultModel: "llama3.1:8b",
    modelOptions: ["llama3.1:8b", "llama3.1:70b", "qwen2.5:7b", "mistral:7b"],
  },
];

export interface AIProviderRuntimeConfig {
  baseURL: string;
  model: string;
}

export type AIProviderConfigMap = Record<AIProviderId, AIProviderRuntimeConfig>;

export const AI_PROVIDER_DEFAULT_CONFIGS: AIProviderConfigMap = AI_PROVIDER_DEFINITIONS.reduce(
  (acc, provider) => {
    acc[provider.value] = {
      baseURL: provider.defaultBaseURL,
      model: provider.defaultModel,
    };
    return acc;
  },
  {} as AIProviderConfigMap,
);
