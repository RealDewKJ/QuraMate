# Localization Architecture

QuraMate now uses `vue-i18n` with namespace-based JSON files stored under `frontend/locales/`.

## Supported languages

- English (`en`) is the source language
- Thai (`th`)
- Chinese, Simplified (`zh`)
- Japanese (`ja`)
- Vietnamese (`vi`)

Japanese is recommended because Japan remains one of Asia's strongest software-buying markets and users expect polished, native-language UX. Vietnamese is recommended because Vietnam's software market is growing quickly, mobile and SaaS adoption are rising fast, and localized products can differentiate earlier there than in more saturated markets.

## File layout

```text
frontend/locales/
  en/
    common.json
    auth.json
    batch.json
  th/
    common.json
    auth.json
    batch.json
  zh/
    common.json
    auth.json
    batch.json
  ja/
    common.json
    auth.json
    batch.json
  vi/
    common.json
    auth.json
    batch.json
```

Namespaces keep translations modular:

- `common.json`: shared UI labels, dialogs, buttons, settings text
- `auth.json`: authentication and session strings
- `batch.json`: batch operations, progress, and pluralized count messages

## Runtime usage

Developers only write English source strings in `frontend/locales/en/*.json`.

Use keys from code:

```ts
const { t } = useI18n();

t("common.save");
t("common.settings.title");
t("auth.loginTitle");
t("batch.itemsProcessed", 3);
```

Interpolation:

```ts
t("common.recovery.failureItem", {
  name: "Primary DB",
  message: "Connection failed",
});
```

Pluralization:

```json
{
  "itemsProcessed_one": "{count} item processed",
  "itemsProcessed_other": "{count} items processed"
}
```

```ts
t("batch.itemsProcessed", 1);
t("batch.itemsProcessed", 12);
```

Language switching:

```ts
import { setAppLocale } from "@/i18n";

setAppLocale("ja");
```

Fallback handling:

- runtime locale is loaded from `user_settings.general.language`
- local storage is used as an early-start fallback
- `en` is the required fallback locale for every missing key

## Translation workflow

```text
Developer -> Git push -> TMS sync -> AI translate -> Review -> Pull request -> Merge
```

Recommended TMS setup:

1. Connect the repository branch `develop` to Crowdin, Lokalise, or Phrase.
2. Configure English files as the only source files.
3. Enable translation memory, glossary, and AI pre-translation for new keys.
4. Auto-export translated files back to the repo as a pull request into `develop`.
5. Merge translation PRs through the normal review pipeline.

## Example Crowdin mapping

The repo includes a `crowdin.yml` example:

```yml
files:
  - source: /frontend/locales/en/*.json
    translation: /frontend/locales/%two_letters_code%/%original_file_name%
```

Equivalent mappings can be used in Phrase or Lokalise.

## Git strategy

- feature branches: developers add new English keys only
- `develop`: TMS sync target and translation PR base branch
- `main`: release branch

Flow:

1. Merge feature work into `develop`
2. TMS imports updated English strings
3. AI creates draft translations for `th`, `zh`, `ja`, and `vi`
4. Optional review happens in the TMS for sensitive content
5. TMS opens an automated translation PR into `develop`
6. CI validates locale completeness before merge

## CI validation

Run:

```bash
cd frontend
npm run validate:locales
```

This checks:

- missing keys
- extra keys not found in English
- placeholder mismatches like `{{name}}`
- invalid JSON

## Best practices

Correct:

```ts
t("common.save");
t("common.recovery.failureItem", { name, message });
```

Incorrect:

```ts
"Save";
"Welcome " + user.name;
count + " items processed";
```

Rules:

- never hardcode UI text
- never concatenate translatable sentence fragments
- use placeholders such as `{{name}}`
- keep keys stable
- prefer semantic identifiers over English sentences as keys
- review high-risk strings even if AI translation is enabled

## Architecture diagram

```text
Developer
  |
  v
Git Repository
  |
  v
Translation Platform (Crowdin / Lokalise / Phrase)
  |
  v
AI Translation + Translation Memory + Glossary
  |
  v
Translation Review
  |
  v
Automatic PR to Repository
  |
  v
Application Runtime using vue-i18n
```
