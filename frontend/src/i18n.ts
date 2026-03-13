import { createI18n } from "vue-i18n";

export const supportedLocales = ["en", "th", "zh", "ja", "vi"] as const;
export type SupportedLocale = (typeof supportedLocales)[number];

export const localeNamespaces = ["common", "auth", "batch"] as const;
export type LocaleNamespace = (typeof localeNamespaces)[number];

type LocaleMessageSchema = Record<string, unknown>;
type LocaleMessages = Record<SupportedLocale, Record<LocaleNamespace, LocaleMessageSchema>>;

const DEFAULT_LOCALE: SupportedLocale = "en";
const STORAGE_KEY = "language";

const localeFiles = import.meta.glob("../locales/*/*.json", {
  eager: true,
  import: "default",
}) as Record<string, LocaleMessageSchema>;

const buildMessages = (): LocaleMessages => {
  const messages = Object.fromEntries(
    supportedLocales.map((locale) => [
      locale,
      Object.fromEntries(localeNamespaces.map((namespace) => [namespace, {}])),
    ]),
  ) as LocaleMessages;

  for (const [filePath, contents] of Object.entries(localeFiles)) {
    const match = filePath.match(/locales\/([^/]+)\/([^/]+)\.json$/);
    if (!match) {
      continue;
    }

    const [, locale, namespace] = match;
    if (
      supportedLocales.includes(locale as SupportedLocale) &&
      localeNamespaces.includes(namespace as LocaleNamespace)
    ) {
      messages[locale as SupportedLocale][namespace as LocaleNamespace] = contents;
    }
  }

  return messages;
};

export const localeOptions: Array<{ value: SupportedLocale; labelKey: string }> = [
  { value: "en", labelKey: "common.settings.languages.en" },
  { value: "th", labelKey: "common.settings.languages.th" },
  { value: "zh", labelKey: "common.settings.languages.zh" },
  { value: "ja", labelKey: "common.settings.languages.ja" },
  { value: "vi", labelKey: "common.settings.languages.vi" },
];

export const normalizeLocale = (value?: string | null): SupportedLocale => {
  if (!value) {
    return DEFAULT_LOCALE;
  }

  const normalized = value.toLowerCase();

  if (supportedLocales.includes(normalized as SupportedLocale)) {
    return normalized as SupportedLocale;
  }

  const primaryCode = normalized.split(/[-_]/)[0];
  if (supportedLocales.includes(primaryCode as SupportedLocale)) {
    return primaryCode as SupportedLocale;
  }

  if (normalized.startsWith("zh")) {
    return "zh";
  }

  return DEFAULT_LOCALE;
};

export const getStoredLocale = (): SupportedLocale => {
  const savedLocale =
    typeof window !== "undefined" ? window.localStorage.getItem(STORAGE_KEY) : null;
  const browserLocale =
    typeof navigator !== "undefined" ? navigator.language : DEFAULT_LOCALE;

  return normalizeLocale(savedLocale || browserLocale);
};

const applyDocumentLocale = (locale: SupportedLocale) => {
  if (typeof document === "undefined") {
    return;
  }

  document.documentElement.lang = locale;
};

const messages = buildMessages();
const initialLocale = getStoredLocale();

const i18n = createI18n({
  legacy: false,
  locale: initialLocale,
  fallbackLocale: DEFAULT_LOCALE,
  availableLocales: [...supportedLocales],
  globalInjection: true,
  messages,
  missingWarn: import.meta.env.DEV,
  fallbackWarn: import.meta.env.DEV,
});

applyDocumentLocale(initialLocale);

export const setAppLocale = (nextLocale: string): SupportedLocale => {
  const normalizedLocale = normalizeLocale(nextLocale);
  i18n.global.locale.value = normalizedLocale;

  if (typeof window !== "undefined") {
    window.localStorage.setItem(STORAGE_KEY, normalizedLocale);
  }

  applyDocumentLocale(normalizedLocale);
  return normalizedLocale;
};

export default i18n;
