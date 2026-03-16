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
export type AIProviderProtocol = "openai-compatible" | "anthropic" | "google";
export type AIProviderSupportStatus = "stable" | "experimental";

export interface AIProviderDefinition {
  id: AIProviderId;
  value: AIProviderId;
  label: string;
  protocol: AIProviderProtocol;
  requiresApiKey: boolean;
  supportsCustomBaseURL: boolean;
  supportsChat: boolean;
  supportsSqlOnly: boolean;
  supportsVision: boolean;
  status: AIProviderSupportStatus;
  defaultBaseURL: string;
  defaultModel: string;
  modelOptions: string[];
}

export const AI_PROVIDER_DEFINITIONS: AIProviderDefinition[] = [
  {
    id: "openai",
    value: "openai",
    label: "OpenAI (ChatGPT)",
    protocol: "openai-compatible",
    requiresApiKey: true,
    supportsCustomBaseURL: true,
    supportsChat: true,
    supportsSqlOnly: true,
    supportsVision: false,
    status: "stable",
    defaultBaseURL: "https://api.openai.com/v1",
    defaultModel: "gpt-4o-mini",
    modelOptions: ["gpt-4o-mini", "gpt-4.1-mini", "gpt-4.1"],
  },
  {
    id: "anthropic",
    value: "anthropic",
    label: "Anthropic (Claude)",
    protocol: "anthropic",
    requiresApiKey: true,
    supportsCustomBaseURL: true,
    supportsChat: true,
    supportsSqlOnly: true,
    supportsVision: false,
    status: "stable",
    defaultBaseURL: "https://api.anthropic.com",
    defaultModel: "claude-3-5-sonnet-latest",
    modelOptions: [
      "claude-3-5-sonnet-latest",
      "claude-3-5-haiku-latest",
      "claude-3-opus-latest",
    ],
  },
  {
    id: "google",
    value: "google",
    label: "Google (Gemini)",
    protocol: "google",
    requiresApiKey: true,
    supportsCustomBaseURL: true,
    supportsChat: true,
    supportsSqlOnly: true,
    supportsVision: false,
    status: "stable",
    defaultBaseURL: "https://generativelanguage.googleapis.com",
    defaultModel: "gemini-2.0-flash",
    modelOptions: ["gemini-2.0-flash", "gemini-2.0-pro", "gemini-1.5-pro"],
  },
  {
    id: "glm",
    value: "glm",
    label: "GLM (Zhipu AI)",
    protocol: "openai-compatible",
    requiresApiKey: true,
    supportsCustomBaseURL: true,
    supportsChat: true,
    supportsSqlOnly: true,
    supportsVision: false,
    status: "experimental",
    defaultBaseURL: "https://open.bigmodel.cn/api/paas/v4",
    defaultModel: "glm-4.7",
    modelOptions: ["glm-4.5", "glm-4.6", "glm-4.7", "glm-5-turbo", "glm-5"],
  },
  {
    id: "kimi",
    value: "kimi",
    label: "Kimi (Moonshot AI)",
    protocol: "openai-compatible",
    requiresApiKey: true,
    supportsCustomBaseURL: true,
    supportsChat: true,
    supportsSqlOnly: true,
    supportsVision: false,
    status: "experimental",
    defaultBaseURL: "https://api.moonshot.cn/v1",
    defaultModel: "moonshot-v1-8k",
    modelOptions: ["moonshot-v1-8k", "moonshot-v1-32k", "moonshot-v1-128k"],
  },
  {
    id: "qwen",
    value: "qwen",
    label: "Qwen (Alibaba)",
    protocol: "openai-compatible",
    requiresApiKey: true,
    supportsCustomBaseURL: true,
    supportsChat: true,
    supportsSqlOnly: true,
    supportsVision: false,
    status: "experimental",
    defaultBaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
    defaultModel: "qwen-plus",
    modelOptions: ["qwen-plus", "qwen-turbo", "qwen-max"],
  },
  {
    id: "minimax",
    value: "minimax",
    label: "MiniMax",
    protocol: "openai-compatible",
    requiresApiKey: true,
    supportsCustomBaseURL: true,
    supportsChat: true,
    supportsSqlOnly: true,
    supportsVision: false,
    status: "experimental",
    defaultBaseURL: "https://api.minimax.chat/v1",
    defaultModel: "abab6.5s-chat",
    modelOptions: ["abab6.5s-chat", "abab6.5-chat", "abab5.5-chat"],
  },
  {
    id: "local",
    value: "local",
    label: "Local (Ollama / Llama.cpp)",
    protocol: "openai-compatible",
    requiresApiKey: false,
    supportsCustomBaseURL: true,
    supportsChat: true,
    supportsSqlOnly: true,
    supportsVision: false,
    status: "stable",
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

export const AI_PROVIDER_DEFAULT_CONFIGS: AIProviderConfigMap =
  AI_PROVIDER_DEFINITIONS.reduce((acc, provider) => {
    acc[provider.id] = {
      baseURL: provider.defaultBaseURL,
      model: provider.defaultModel,
    };
    return acc;
  }, {} as AIProviderConfigMap);

export const AI_PROVIDER_DEFINITION_MAP: Record<
  AIProviderId,
  AIProviderDefinition
> = AI_PROVIDER_DEFINITIONS.reduce(
  (acc, provider) => {
    acc[provider.id] = provider;
    return acc;
  },
  {} as Record<AIProviderId, AIProviderDefinition>,
);
