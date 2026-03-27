import { computed, nextTick, onBeforeUnmount, onMounted, ref, type ComponentPublicInstance, type Ref } from 'vue';
import { onKeyStroke, useToggle } from '@vueuse/core';

import { LoadSetting } from '../../wailsjs/go/app/App';
import type { ActiveConnection, ConnectionConfig, ConnectionInputMode } from './useConnectionForm';

export interface DatabaseTypeOption {
  value: ConnectionConfig['type'];
  label: string;
  disabled?: boolean;
}

export interface DatabaseTypeUiController {
  highlightedIndex: number;
  isOpen: boolean;
  setButtonRef: (element: Element | ComponentPublicInstance | null) => void;
  setMenuRef: (element: Element | ComponentPublicInstance | null) => void;
  setHighlightedIndex: (index: number) => void;
  toggleMenu: () => Promise<void>;
  selectOption: (type: ConnectionConfig['type']) => void;
  handleTriggerKeydown: (event: KeyboardEvent) => Promise<void>;
  handleOptionKeydown: (event: KeyboardEvent) => Promise<void>;
}

export const inputModeOptions: ReadonlyArray<{ value: ConnectionInputMode; label: string }> = [
  { value: 'fields', label: 'Fields' },
  { value: 'connectionString', label: 'Direct Connection String' },
];

export const databaseTypeOptions: ReadonlyArray<DatabaseTypeOption> = [
  { value: 'postgres', label: 'PostgreSQL' },
  { value: 'supabase', label: 'Supabase' },
  { value: 'mysql', label: 'MySQL' },
  { value: 'mariadb', label: 'MariaDB' },
  { value: 'mssql', label: 'MSSQL' },
  { value: 'sqlite', label: 'SQLite' },
  { value: 'duckdb', label: 'DuckDB' },
  { value: 'greenplum', label: 'Greenplum' },
  { value: 'redshift', label: 'Redshift' },
  { value: 'cockroachdb', label: 'CockroachDB' },
  { value: 'databend', label: 'Databend' },
  { value: 'libsql', label: 'LibSQL' },
  { value: 'd1', label: 'Cloudflare D1', disabled: true },
];

const databaseTypeLogoByType: Partial<Record<ConnectionConfig['type'], string>> = {
  postgres:
    'https://cdn.brandfetch.io/idjSeCeMle/theme/dark/logo.svg?c=1bxid64Mup7aczewSAYMX&t=1772301524158',
  supabase:
    'https://cdn.brandfetch.io/idsSceG8fK/w/436/h/449/theme/dark/symbol.png?c=1bxid64Mup7aczewSAYMX&t=1668081497517',
  mysql:
    'https://cdn.brandfetch.io/idBdG8DdKe/theme/dark/logo.svg?c=1bxid64Mup7aczewSAYMX&t=1667573657581',
  mariadb:
    'https://cdn.brandfetch.io/idxldSTiIy/theme/dark/symbol.svg?c=1bxid64Mup7aczewSAYMX&t=1668081833820',
  mssql:
    'https://img.icons8.com/?size=100&id=laYYF3dV0Iew&format=png&color=000000',
  sqlite:
    'https://cdn.brandfetch.io/idDk4skBSf/theme/dark/logo.svg?c=1bxid64Mup7aczewSAYMX&t=1770964278385',
  duckdb:
    'https://cdn.brandfetch.io/idYamjE8Qv/theme/dark/logo.svg?c=1bxid64Mup7aczewSAYMX&t=1742917607612',
  greenplum:
    'https://e7.pngegg.com/pngimages/634/19/png-clipart-greenplum-database-massively-parallel-data-warehouse-logo-others-miscellaneous-text.png',
  redshift:
    'https://upload.wikimedia.org/wikipedia/commons/thumb/7/73/Amazon-Redshift-Logo.svg/40px-Amazon-Redshift-Logo.svg.png?_=20200715222800',
  cockroachdb:
    'https://cdn.brandfetch.io/idZWR8KHHs/w/400/h/400/theme/dark/icon.jpeg?c=1bxid64Mup7aczewSAYMX&t=1698087041619',
  databend:
    'https://cdn.brandfetch.io/idBSJmwM3M/theme/light/logo.svg?c=1bxid64Mup7aczewSAYMX&t=1757968429246',
  libsql:
    'https://cdn.brandfetch.io/idMTyZGFbT/w/180/h/180/theme/dark/logo.png?c=1bxid64Mup7aczewSAYMX&t=1689372293515',
  d1:
    'https://cdn.brandfetch.io/idJ3Cg8ymG/theme/dark/logo.svg?c=1bxid64Mup7aczewSAYMX&t=1667589504295',
};

export const getDatabaseTypeLogo = (type: string): string =>
  databaseTypeLogoByType[type as ConnectionConfig['type']] || '';

export const getDatabaseTypeIconKind = (type: string): 'postgres' | 'mysql' | 'mssql' | 'file' | 'cloud' | 'generic' => {
  const normalized = (type || '').toLowerCase();
  if (['postgres', 'supabase', 'greenplum', 'redshift', 'cockroachdb'].includes(normalized)) {
    return normalized === 'supabase' ? 'cloud' : 'postgres';
  }
  if (['mysql', 'mariadb', 'databend'].includes(normalized)) {
    return 'mysql';
  }
  if (normalized === 'mssql') {
    return 'mssql';
  }
  if (['sqlite', 'duckdb', 'libsql'].includes(normalized)) {
    return 'file';
  }
  if (normalized === 'd1') {
    return 'cloud';
  }
  return 'generic';
};

interface UseDbConnectionUiOptions {
  config: ConnectionConfig;
  activeConnections: () => ActiveConnection[];
  trustServerCertificateDefault: Ref<boolean>;
}

export function useDbConnectionUi(options: UseDbConnectionUiOptions) {
  const { config, activeConnections, trustServerCertificateDefault } = options;

  const [showSettings] = useToggle(false);
  const [showSavedModal] = useToggle(false);
  const [showPassword] = useToggle(false);
  const [showSshPassword] = useToggle(false);
  const savedConnectionsAnnouncement = ref('');

  const databaseTypeButtonRef = ref<HTMLButtonElement | null>(null);
  const databaseTypeMenuRef = ref<HTMLElement | null>(null);
  const isDatabaseTypeMenuOpen = ref(false);
  const highlightedDatabaseTypeIndex = ref(0);

  let databaseTypeTypeahead = '';
  let databaseTypeTypeaheadTimeout: ReturnType<typeof setTimeout> | null = null;

  const selectedDatabaseTypeLabel = computed(
    () => databaseTypeOptions.find((option) => option.value === config.type)?.label ?? 'Select database type',
  );

  const selectedDatabaseTypeIndex = computed(() =>
    Math.max(0, databaseTypeOptions.findIndex((option) => option.value === config.type)),
  );

  const showSqlServerSettings = computed(
    () => config.type === 'mssql' || activeConnections().some((connection) => (connection.config.type || '').toLowerCase() === 'mssql'),
  );

  const applyTrustServerCertificateDefault = (value: boolean) => {
    trustServerCertificateDefault.value = value;
    if (config.type === 'mssql') {
      config.trustServerCertificate = value;
    }
  };

  const handleSettingsSave = (nextSettings?: { general?: { trustSqlServerCertificateByDefault?: boolean } }) => {
    const nextValue = nextSettings?.general?.trustSqlServerCertificateByDefault !== false;
    applyTrustServerCertificateDefault(nextValue);
  };

  const loadConnectionDefaults = async () => {
    try {
      const savedSettingsJson = await LoadSetting('user_settings');
      if (!savedSettingsJson) {
        return;
      }

      const parsed = JSON.parse(savedSettingsJson) as { general?: { trustSqlServerCertificateByDefault?: boolean } };
      applyTrustServerCertificateDefault(parsed?.general?.trustSqlServerCertificateByDefault !== false);
    } catch (error) {
      console.error('Failed to load connection defaults:', error);
    }
  };

  const openSettings = () => {
    showSettings.value = true;
  };

  const closeSettings = () => {
    showSettings.value = false;
  };

  const openSavedConnections = () => {
    showSavedModal.value = true;
  };

  const closeSavedConnections = () => {
    showSavedModal.value = false;
  };

  const togglePasswordVisibility = () => {
    showPassword.value = !showPassword.value;
  };

  const toggleSshPasswordVisibility = () => {
    showSshPassword.value = !showSshPassword.value;
  };

  const clearDatabaseTypeTypeahead = () => {
    databaseTypeTypeahead = '';
    if (databaseTypeTypeaheadTimeout) {
      clearTimeout(databaseTypeTypeaheadTimeout);
      databaseTypeTypeaheadTimeout = null;
    }
  };

  const closeDatabaseTypeMenu = () => {
    isDatabaseTypeMenuOpen.value = false;
  };

  const focusHighlightedDatabaseType = async () => {
    await nextTick();
    const optionElements = databaseTypeMenuRef.value?.querySelectorAll<HTMLElement>('[data-db-type-option]');
    optionElements?.[highlightedDatabaseTypeIndex.value]?.focus();
  };

  const openDatabaseTypeMenu = async () => {
    highlightedDatabaseTypeIndex.value = selectedDatabaseTypeIndex.value;
    isDatabaseTypeMenuOpen.value = true;
    await focusHighlightedDatabaseType();
  };

  const toggleDatabaseTypeMenu = async () => {
    if (isDatabaseTypeMenuOpen.value) {
      closeDatabaseTypeMenu();
      return;
    }

    await openDatabaseTypeMenu();
  };

  const selectDatabaseType = (type: ConnectionConfig['type']) => {
    const option = databaseTypeOptions.find((item) => item.value === type);
    if (option?.disabled) {
      return;
    }

    config.type = type;
    highlightedDatabaseTypeIndex.value = selectedDatabaseTypeIndex.value;
    closeDatabaseTypeMenu();
    void nextTick(() => databaseTypeButtonRef.value?.focus());
  };

  const moveDatabaseTypeHighlight = (direction: 1 | -1) => {
    const total = databaseTypeOptions.length;
    highlightedDatabaseTypeIndex.value =
      (highlightedDatabaseTypeIndex.value + direction + total) % total;
  };

  const handleDatabaseTypeTypeahead = async (key: string) => {
    if (key.length !== 1 || !/\S/.test(key)) {
      return;
    }

    databaseTypeTypeahead = `${databaseTypeTypeahead}${key.toLowerCase()}`;
    if (databaseTypeTypeaheadTimeout) {
      clearTimeout(databaseTypeTypeaheadTimeout);
    }

    databaseTypeTypeaheadTimeout = setTimeout(() => {
      databaseTypeTypeahead = '';
      databaseTypeTypeaheadTimeout = null;
    }, 500);

    const matchIndex = databaseTypeOptions.findIndex((option) =>
      option.label.toLowerCase().startsWith(databaseTypeTypeahead),
    );

    if (matchIndex < 0) {
      return;
    }

    highlightedDatabaseTypeIndex.value = matchIndex;
    if (!isDatabaseTypeMenuOpen.value) {
      await openDatabaseTypeMenu();
      return;
    }

    await focusHighlightedDatabaseType();
  };

  const handleDatabaseTypeTriggerKeydown = async (event: KeyboardEvent) => {
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault();
      if (!isDatabaseTypeMenuOpen.value) {
        await openDatabaseTypeMenu();
      } else {
        moveDatabaseTypeHighlight(event.key === 'ArrowDown' ? 1 : -1);
        await focusHighlightedDatabaseType();
      }
      return;
    }

    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      await toggleDatabaseTypeMenu();
      return;
    }

    await handleDatabaseTypeTypeahead(event.key);
  };

  const handleDatabaseTypeOptionKeydown = async (event: KeyboardEvent) => {
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault();
      moveDatabaseTypeHighlight(event.key === 'ArrowDown' ? 1 : -1);
      await focusHighlightedDatabaseType();
      return;
    }

    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      const option = databaseTypeOptions[highlightedDatabaseTypeIndex.value];
      if (option && !option.disabled) {
        selectDatabaseType(option.value);
      }
      return;
    }

    if (event.key === 'Escape') {
      event.preventDefault();
      closeDatabaseTypeMenu();
      databaseTypeButtonRef.value?.focus();
      return;
    }

    if (event.key === 'Home') {
      event.preventDefault();
      highlightedDatabaseTypeIndex.value = 0;
      await focusHighlightedDatabaseType();
      return;
    }

    if (event.key === 'End') {
      event.preventDefault();
      highlightedDatabaseTypeIndex.value = databaseTypeOptions.length - 1;
      await focusHighlightedDatabaseType();
      return;
    }

    await handleDatabaseTypeTypeahead(event.key);
  };

  const handleDocumentClick = (event: MouseEvent) => {
    const target = event.target as HTMLElement | null;
    if (!target?.closest('[data-db-type-menu]')) {
      closeDatabaseTypeMenu();
    }
  };

  const setHighlightedDatabaseTypeIndex = (index: number) => {
    highlightedDatabaseTypeIndex.value = index;
  };

  const setDatabaseTypeButtonRef = (element: Element | ComponentPublicInstance | null) => {
    databaseTypeButtonRef.value = element instanceof HTMLButtonElement ? element : null;
  };

  const setDatabaseTypeMenuRef = (element: Element | ComponentPublicInstance | null) => {
    databaseTypeMenuRef.value = element instanceof HTMLElement ? element : null;
  };

  onKeyStroke('Escape', () => {
    closeSettings();
    closeDatabaseTypeMenu();
  });

  onMounted(async () => {
    await loadConnectionDefaults();
    document.addEventListener('click', handleDocumentClick);
  });

  onBeforeUnmount(() => {
    document.removeEventListener('click', handleDocumentClick);
    clearDatabaseTypeTypeahead();
  });

  const databaseTypeUi = computed<DatabaseTypeUiController>(() => ({
    highlightedIndex: highlightedDatabaseTypeIndex.value,
    isOpen: isDatabaseTypeMenuOpen.value,
    setButtonRef: setDatabaseTypeButtonRef,
    setMenuRef: setDatabaseTypeMenuRef,
    setHighlightedIndex: setHighlightedDatabaseTypeIndex,
    toggleMenu: toggleDatabaseTypeMenu,
    selectOption: selectDatabaseType,
    handleTriggerKeydown: handleDatabaseTypeTriggerKeydown,
    handleOptionKeydown: handleDatabaseTypeOptionKeydown,
  }));

  return {
    closeSavedConnections,
    closeSettings,
    databaseTypeOptions,
    databaseTypeUi,
    handleSettingsSave,
    inputModeOptions,
    openSavedConnections,
    openSettings,
    savedConnectionsAnnouncement,
    selectedDatabaseTypeLabel,
    setSavedConnectionsAnnouncement: (value: string) => {
      savedConnectionsAnnouncement.value = value;
    },
    showPassword,
    showSavedModal,
    showSettings,
    showSqlServerSettings,
    showSshPassword,
    togglePasswordVisibility,
    toggleSshPasswordVisibility,
    trustServerCertificateDefault,
  };
}
