import { ref } from 'vue';

import { hydrateAppLocale, persistAppLocale } from '../i18n';
import { LoadSetting } from '../../wailsjs/go/app/App';

interface AppAppearanceSettings {
  appFont?: string;
  usePointerCursors?: boolean;
}

interface AppSettings {
  appearance?: AppAppearanceSettings;
  general?: {
    language?: string;
    trustSqlServerCertificateByDefault?: boolean;
  };
}

interface UseAppBootstrapOptions {
  locale: { value: string };
}

const applyAppearancePreferences = (appearance?: AppAppearanceSettings) => {
  const root = document.documentElement;
  const usePointerCursors = appearance?.usePointerCursors !== false;

  root.classList.toggle('pref-pointer-cursors', usePointerCursors);
  root.classList.toggle('pref-disable-pointer-cursors', !usePointerCursors);
};

export function useAppBootstrap(options: UseAppBootstrapOptions) {
  const { locale } = options;
  const trustSqlServerCertificateByDefault = ref(true);

  const initializeAppBootstrap = async () => {
    try {
      await hydrateAppLocale();
      const savedSettingsJson = await LoadSetting('user_settings');
      let appFont = 'system-ui, sans-serif';

      if (savedSettingsJson) {
        const parsed = JSON.parse(savedSettingsJson) as AppSettings;
        trustSqlServerCertificateByDefault.value =
          parsed.general?.trustSqlServerCertificateByDefault !== false;

        if (parsed.general?.language) {
          locale.value = await persistAppLocale(parsed.general.language);
        }

        if (parsed.appearance?.appFont) {
          appFont = parsed.appearance.appFont;
        }

        applyAppearancePreferences(parsed.appearance);
      } else {
        trustSqlServerCertificateByDefault.value = true;
        applyAppearancePreferences();
      }

      document.documentElement.style.fontFamily = appFont;
    } catch (error) {
      console.error('Failed to load global settings on mount', error);
    }
  };

  return {
    initializeAppBootstrap,
    trustSqlServerCertificateByDefault,
  };
}
