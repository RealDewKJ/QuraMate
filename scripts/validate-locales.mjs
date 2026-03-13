import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const repoRoot = path.resolve(__dirname, "..");
const localesRoot = path.join(repoRoot, "frontend", "locales");
const sourceLocale = "en";
const targetLocales = ["th", "zh", "ja", "vi"];
const namespaces = ["common", "auth", "batch"];

const flattenObject = (value, prefix = "") => {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    return prefix ? { [prefix]: value } : {};
  }

  return Object.entries(value).reduce((acc, [key, nestedValue]) => {
    const nextPrefix = prefix ? `${prefix}.${key}` : key;
    return { ...acc, ...flattenObject(nestedValue, nextPrefix) };
  }, {});
};

const getPlaceholders = (value) => {
  const matches = String(value).match(/{{?\s*[\w.]+\s*}}?/g) || [];
  return matches.map((match) => match.replace(/[{}]/g, "").trim()).sort();
};

const readNamespace = async (locale, namespace) => {
  const filePath = path.join(localesRoot, locale, `${namespace}.json`);
  const raw = await fs.readFile(filePath, "utf8");
  return JSON.parse(raw);
};

const errors = [];

for (const namespace of namespaces) {
  const sourceMessages = await readNamespace(sourceLocale, namespace);
  const sourceFlat = flattenObject(sourceMessages);
  const sourceKeys = Object.keys(sourceFlat).sort();

  for (const locale of targetLocales) {
    const localeMessages = await readNamespace(locale, namespace);
    const localeFlat = flattenObject(localeMessages);
    const localeKeys = Object.keys(localeFlat).sort();

    const missingKeys = sourceKeys.filter((key) => !(key in localeFlat));
    const extraKeys = localeKeys.filter((key) => !(key in sourceFlat));

    if (missingKeys.length > 0) {
      errors.push(
        `[${locale}/${namespace}] Missing keys: ${missingKeys.join(", ")}`,
      );
    }

    if (extraKeys.length > 0) {
      errors.push(`[${locale}/${namespace}] Extra keys: ${extraKeys.join(", ")}`);
    }

    for (const key of sourceKeys) {
      if (!(key in localeFlat)) {
        continue;
      }

      const sourcePlaceholders = getPlaceholders(sourceFlat[key]);
      const localePlaceholders = getPlaceholders(localeFlat[key]);
      if (sourcePlaceholders.join("|") !== localePlaceholders.join("|")) {
        errors.push(
          `[${locale}/${namespace}] Placeholder mismatch for "${key}": expected [${sourcePlaceholders.join(", ")}] but found [${localePlaceholders.join(", ")}]`,
        );
      }
    }
  }
}

if (errors.length > 0) {
  console.error("Locale validation failed:\n");
  for (const error of errors) {
    console.error(`- ${error}`);
  }
  process.exit(1);
}

console.log("Locale validation passed.");
