<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import DbConnection from './components/DbConnection.vue';
import DbDashboard from './components/DbDashboard.vue';
import UpdateNotification from './components/UpdateNotification.vue';
import { colorMode } from './composables/useTheme';
import { getConnectionLabel, type ConnectionConfig } from './composables/useConnectionForm';
import { LoadSetting, SaveSetting, GetStartupFile, ReadTextFile, ConnectDB } from '../wailsjs/go/app/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

const updateNotificationRef = ref<InstanceType<typeof UpdateNotification> | null>(null);
const { locale } = useI18n({ useScope: 'global' });

const APP_CONNECTION_SESSION_KEY = 'app_connection_session_v1';
const APP_CONNECTION_SESSION_VERSION = 1;
const APP_CONNECTION_RECOVERY_PREF_KEY = 'app_connection_recovery_pref_v1';
const APP_CONNECTION_RECOVERY_PREF_VERSION = 1;

interface Connection {
  id: string;
  name: string;
  config: ConnectionConfig;
  sessionKey: string;
}

interface PersistedConnectionEntry {
  name: string;
  sessionKey: string;
  config: ConnectionConfig;
}

interface PersistedConnectionSession {
  version: number;
  activeSessionKey: string | null;
  connections: PersistedConnectionEntry[];
}

interface RecoveryCandidate extends PersistedConnectionEntry {
  selected: boolean;
}

interface RecoveryPreference {
  version: number;
  autoRecover: boolean;
  selectedSessionKeys: string[];
}

const pendingSqlFile = ref<{ path: string; name: string; content: string } | null>(null);

const connections = ref<Connection[]>([]);
const activeTabId = ref<string | null>(null);

const showRecoveryModal = ref(false);
const isRecoveringSession = ref(false);
const recoveryCandidates = ref<RecoveryCandidate[]>([]);
const recoveryPreferredActiveSessionKey = ref<string | null>(null);
const recoveryErrors = ref<string[]>([]);
const rememberRecoveryChoice = ref(false);

let appSessionSaveTimer: ReturnType<typeof setTimeout> | null = null;

const hashString = (value: string): string => {
  let hash = 5381;
  for (let i = 0; i < value.length; i += 1) {
    hash = ((hash << 5) + hash) ^ value.charCodeAt(i);
  }
  return (hash >>> 0).toString(16);
};

const buildConnectionSessionKey = (config: Partial<ConnectionConfig>, displayName?: string): string => {
  const type = String(config.type || '').toLowerCase();
  const normalizedName = String(displayName || config.name || '').trim().toLowerCase();

  if (type === 'sqlite' || type === 'duckdb' || type === 'libsql') {
    const id = `file|${type}|${String(config.database || '').trim().toLowerCase()}|${normalizedName}`;
    return `conn_${hashString(id)}`;
  }

  const sshSection = config.sshEnabled
    ? `|ssh|${String(config.sshHost || '').trim().toLowerCase()}|${String(config.sshPort || 22)}|${String(config.sshUser || '').trim().toLowerCase()}|${String(config.sshKeyFile || '').trim().toLowerCase()}`
    : '|ssh|off';

  const id = [
    'net',
    type,
    String(config.host || '').trim().toLowerCase(),
    String(config.port || ''),
    String(config.user || '').trim().toLowerCase(),
    String(config.database || '').trim().toLowerCase(),
    normalizedName,
    sshSection,
  ].join('|');

  return `conn_${hashString(id)}`;
};

const defaultConfigValues = (): ConnectionConfig => ({
  id: '',
  name: '',
  type: 'postgres',
  host: 'localhost',
  port: 5432,
  user: 'postgres',
  password: '',
  database: 'postgres',
  readOnly: false,
  sshEnabled: false,
  sshHost: '',
  sshPort: 22,
  sshUser: '',
  sshPassword: '',
  sshKeyFile: '',
  encoding: '',
});

const normalizeConnectionConfig = (config: Partial<ConnectionConfig>): ConnectionConfig => {
  return {
    ...defaultConfigValues(),
    ...config,
    type: config.type || 'postgres',
    host: config.host || '',
    port: typeof config.port === 'number' ? config.port : Number(config.port || 0),
    user: config.user || '',
    password: config.password || '',
    database: config.database || '',
    readOnly: !!config.readOnly,
    sshEnabled: !!config.sshEnabled,
    sshHost: config.sshHost || '',
    sshPort: typeof config.sshPort === 'number' ? config.sshPort : Number(config.sshPort || 22),
    sshUser: config.sshUser || '',
    sshPassword: config.sshPassword || '',
    sshKeyFile: config.sshKeyFile || '',
    encoding: config.encoding || '',
    name: config.name || '',
    id: config.id || '',
  };
};

const saveAppConnectionSession = async () => {
  const activeConnection = connections.value.find((conn) => conn.id === activeTabId.value) || null;
  const payload: PersistedConnectionSession = {
    version: APP_CONNECTION_SESSION_VERSION,
    activeSessionKey: activeConnection?.sessionKey || null,
    connections: connections.value.map((conn) => ({
      name: conn.name,
      sessionKey: conn.sessionKey,
      config: { ...conn.config },
    })),
  };

  try {
    await SaveSetting(APP_CONNECTION_SESSION_KEY, JSON.stringify(payload));
  } catch (e) {
    console.error('Failed to persist app connection session', e);
  }
};

const scheduleSaveAppConnectionSession = () => {
  if (appSessionSaveTimer) {
    clearTimeout(appSessionSaveTimer);
  }

  appSessionSaveTimer = setTimeout(() => {
    appSessionSaveTimer = null;
    void saveAppConnectionSession();
  }, 350);
};

const clearAppConnectionSession = async () => {
  try {
    await SaveSetting(APP_CONNECTION_SESSION_KEY, '');
  } catch (e) {
    console.error('Failed to clear app connection session', e);
  }
};

const loadRecoveryPreference = async (): Promise<RecoveryPreference | null> => {
  try {
    const raw = await LoadSetting(APP_CONNECTION_RECOVERY_PREF_KEY);
    if (!raw) return null;

    const parsed = JSON.parse(raw) as Partial<RecoveryPreference>;
    if (parsed.version !== APP_CONNECTION_RECOVERY_PREF_VERSION || typeof parsed.autoRecover !== 'boolean') {
      return null;
    }

    return {
      version: APP_CONNECTION_RECOVERY_PREF_VERSION,
      autoRecover: parsed.autoRecover,
      selectedSessionKeys: Array.isArray(parsed.selectedSessionKeys)
        ? parsed.selectedSessionKeys.filter((k): k is string => typeof k === 'string')
        : [],
    };
  } catch (e) {
    console.error('Failed to load recovery preference', e);
    return null;
  }
};

const saveRecoveryPreference = async (autoRecover: boolean, selectedSessionKeys: string[]) => {
  const payload: RecoveryPreference = {
    version: APP_CONNECTION_RECOVERY_PREF_VERSION,
    autoRecover,
    selectedSessionKeys,
  };

  try {
    await SaveSetting(APP_CONNECTION_RECOVERY_PREF_KEY, JSON.stringify(payload));
  } catch (e) {
    console.error('Failed to save recovery preference', e);
  }
};

const clearRecoveryPreference = async () => {
  try {
    await SaveSetting(APP_CONNECTION_RECOVERY_PREF_KEY, '');
  } catch (e) {
    console.error('Failed to clear recovery preference', e);
  }
};

watch([connections, activeTabId], scheduleSaveAppConnectionSession, { deep: true });

const activeConn = computed(() =>
  connections.value.find((c) => c.id === activeTabId.value) ?? null
);

const hasAnyRecoverySelection = computed(() =>
  recoveryCandidates.value.some((candidate) => candidate.selected)
);

const switchToTab = (id: string) => {
  activeTabId.value = id;
};

const switchToHome = () => {
  activeTabId.value = null;
};

const selectAllRecoveryCandidates = () => {
  recoveryCandidates.value.forEach((candidate) => {
    candidate.selected = true;
  });
};

const unselectAllRecoveryCandidates = () => {
  recoveryCandidates.value.forEach((candidate) => {
    candidate.selected = false;
  });
};

const handleConnected = (conn: { id: string; name: string; config: ConnectionConfig; sessionKey?: string }) => {
  const normalizedConfig = normalizeConnectionConfig(conn.config || {});
  const sessionKey = conn.sessionKey || buildConnectionSessionKey(normalizedConfig, conn.name);

  const existingIndex = connections.value.findIndex((c) => c.sessionKey === sessionKey);
  if (existingIndex !== -1) {
    connections.value[existingIndex] = {
      ...connections.value[existingIndex],
      id: conn.id,
      name: conn.name || connections.value[existingIndex].name,
      config: normalizedConfig,
      sessionKey,
    };
    activeTabId.value = conn.id;
    return;
  }

  connections.value.push({
    id: conn.id,
    name: conn.name,
    config: normalizedConfig,
    sessionKey,
  });
  activeTabId.value = conn.id;

  if (pendingSqlFile.value) {
    const sqlContent = pendingSqlFile.value.content;
    const sqlFileName = pendingSqlFile.value.name;
    const sqlFilePath = pendingSqlFile.value.path;
    const connId = conn.id;
    pendingSqlFile.value = null;

    setTimeout(() => {
      window.dispatchEvent(new CustomEvent('open-sql-file', {
        detail: { content: sqlContent, fileName: sqlFileName, filePath: sqlFilePath, connectionId: connId }
      }));
    }, 300);
  }
};

const handleDisconnect = (id: string) => {
  const index = connections.value.findIndex(c => c.id === id);
  if (index !== -1) {
    connections.value.splice(index, 1);
    if (activeTabId.value === id) {
      if (connections.value.length > 0) {
        activeTabId.value = connections.value[connections.value.length - 1].id;
      } else {
        activeTabId.value = null;
      }
    }
  }
};

const handleConnectionExists = (id: string) => {
  switchToTab(id);
};

const handleConnectionUpdate = (update: { id: string, config: ConnectionConfig }) => {
  const conn = connections.value.find(c => c.id === update.id);
  if (conn) {
    conn.config = normalizeConnectionConfig(update.config);
  }
};

const processStartupFile = async (startupFile: string) => {
  try {
    console.log('Startup string from OS:', startupFile);
    const ext = startupFile.split('.').pop()?.toLowerCase();

    if (ext === 'db' || ext === 'sqlite' || ext === 'sqlite3') {
      const payload = normalizeConnectionConfig({
        type: 'sqlite',
        host: '',
        port: 0,
        user: '',
        password: '',
        database: startupFile,
        readOnly: false,
        sshEnabled: false,
        sshHost: '',
        sshPort: 22,
        sshUser: '',
        sshPassword: '',
        sshKeyFile: '',
        id: '',
      });

      const result = await ConnectDB(payload);
      if (result && !result.error) {
        const name = startupFile.replace(/^.*[\\/]/, '');
        handleConnected({
          id: result.id,
          name: `${name} (SQLite)`,
          config: payload,
          sessionKey: buildConnectionSessionKey(payload, `${name} (SQLite)`),
        });
      }
    } else if (ext === 'sql') {
      const content = await ReadTextFile(startupFile);
      const fileName = startupFile.replace(/^.*[\\/]/, '');

      if (connections.value.length > 0) {
        const id = activeTabId.value || connections.value[0].id;
        switchToTab(id);

        window.dispatchEvent(new CustomEvent('open-sql-file', {
          detail: { content, fileName, filePath: startupFile, connectionId: id }
        }));
      } else {
        pendingSqlFile.value = { path: startupFile, name: fileName, content };
        activeTabId.value = null;
      }
    }
  } catch (e) {
    console.error('Failed to process startup file', e);
  }
};

const recoverConnections = async (targets: PersistedConnectionEntry[]) => {
  if (targets.length === 0) {
    await clearAppConnectionSession();
    recoveryCandidates.value = [];
    recoveryPreferredActiveSessionKey.value = null;
    return;
  }

  isRecoveringSession.value = true;
  recoveryErrors.value = [];

  const recoveredIdBySessionKey = new Map<string, string>();

  for (const entry of targets) {
    if (connections.value.some((conn) => conn.sessionKey === entry.sessionKey)) {
      continue;
    }

    try {
      const result = await ConnectDB(entry.config);
      if (result && result.id && !result.error) {
        handleConnected({
          id: result.id,
          name: entry.name || getConnectionLabel(entry.config),
          config: entry.config,
          sessionKey: entry.sessionKey,
        });
        recoveredIdBySessionKey.set(entry.sessionKey, result.id);
      } else {
        recoveryErrors.value.push(`${entry.name}: ${result?.error || 'Unknown connection error'}`);
      }
    } catch (e: any) {
      recoveryErrors.value.push(`${entry.name}: ${e?.message || e?.toString?.() || 'Connection failed'}`);
    }
  }

  const preferredId = recoveryPreferredActiveSessionKey.value
    ? recoveredIdBySessionKey.get(recoveryPreferredActiveSessionKey.value)
    : null;

  if (preferredId) {
    activeTabId.value = preferredId;
  } else if (connections.value.length > 0 && !activeTabId.value) {
    activeTabId.value = connections.value[0].id;
  }

  isRecoveringSession.value = false;
  recoveryCandidates.value = [];
  recoveryPreferredActiveSessionKey.value = null;
};

const openRecoveryPrompt = async () => {
  try {
    const raw = await LoadSetting(APP_CONNECTION_SESSION_KEY);
    if (!raw) return;

    const parsed = JSON.parse(raw) as Partial<PersistedConnectionSession>;
    if (parsed.version !== APP_CONNECTION_SESSION_VERSION || !Array.isArray(parsed.connections) || parsed.connections.length === 0) {
      return;
    }

    recoveryPreferredActiveSessionKey.value = parsed.activeSessionKey || null;
    const candidates = parsed.connections
      .filter((entry) => !!entry?.config)
      .map((entry) => ({
        name: entry.name || getConnectionLabel(entry.config),
        sessionKey: entry.sessionKey || buildConnectionSessionKey(entry.config, entry.name),
        config: normalizeConnectionConfig(entry.config),
      }));

    if (candidates.length === 0) return;

    const preference = await loadRecoveryPreference();
    if (preference?.autoRecover) {
      const preferredKeys = new Set(preference.selectedSessionKeys);
      const autoTargets = preferredKeys.size > 0
        ? candidates.filter((item) => preferredKeys.has(item.sessionKey))
        : candidates;
      if (autoTargets.length > 0) {
        await recoverConnections(autoTargets);
        return;
      }
    }

    rememberRecoveryChoice.value = false;
    recoveryCandidates.value = candidates.map((candidate) => ({ ...candidate, selected: true }));
    showRecoveryModal.value = true;
  } catch (e) {
    console.error('Failed to load connection recovery session', e);
  }
};

const skipRecoveryPrompt = async () => {
  showRecoveryModal.value = false;
  recoveryCandidates.value = [];
  recoveryPreferredActiveSessionKey.value = null;
  recoveryErrors.value = [];
  rememberRecoveryChoice.value = false;
  await clearAppConnectionSession();
  await clearRecoveryPreference();
};

const recoverSelectedConnections = async () => {
  const targets = recoveryCandidates.value.filter((item) => item.selected);
  showRecoveryModal.value = false;

  if (rememberRecoveryChoice.value) {
    await saveRecoveryPreference(true, targets.map((target) => target.sessionKey));
  } else {
    await clearRecoveryPreference();
  }

  await recoverConnections(targets);
};

const applyAppearancePreferences = (appearance?: {
  usePointerCursors?: boolean;
}) => {
  const root = document.documentElement;
  const usePointerCursors = appearance?.usePointerCursors !== false;

  root.classList.toggle('pref-pointer-cursors', usePointerCursors);
  root.classList.toggle('pref-disable-pointer-cursors', !usePointerCursors);
};

onMounted(async () => {
  try {
    const savedSettingsJson = await LoadSetting('user_settings');
    let appFont = 'system-ui, sans-serif';
    if (savedSettingsJson) {
      const parsed = JSON.parse(savedSettingsJson);
      if (parsed.general && parsed.general.language) {
        locale.value = parsed.general.language;
      }
      if (parsed.appearance && parsed.appearance.appFont) {
        appFont = parsed.appearance.appFont;
      }
      applyAppearancePreferences(parsed.appearance);
    } else {
      applyAppearancePreferences();
    }
    document.documentElement.style.fontFamily = appFont;
  } catch (e) {
    console.error('Failed to load global settings on mount', e);
  }

  await openRecoveryPrompt();

  try {
    const startupFile = await GetStartupFile();
    if (startupFile) {
      processStartupFile(startupFile);
    }
  } catch (e) {
    console.error('Failed to fetch startup file', e);
  }

  EventsOn('app:open-file', (filePath: string) => {
    if (filePath) {
      processStartupFile(filePath);
    }
  });
});

onUnmounted(() => {
  if (appSessionSaveTimer) {
    clearTimeout(appSessionSaveTimer);
    appSessionSaveTimer = null;
  }
  void saveAppConnectionSession();
});
</script>

<template>
  <div class="h-screen flex flex-col text-foreground relative overflow-hidden bg-background">
    <div class="absolute inset-0 z-0 overflow-hidden pointer-events-none">
      <div
        class="absolute top-[-10%] right-[-5%] w-[500px] h-[500px] rounded-full bg-orange-500/10 dark:bg-orange-500/5 blur-[120px]">
      </div>
      <div
        class="absolute bottom-[-10%] left-[-10%] w-[600px] h-[600px] rounded-full bg-orange-400/10 dark:bg-orange-400/5 blur-[150px]">
      </div>
    </div>

    <div class="flex items-center bg-background/80 backdrop-blur-sm border-b border-border z-10">
      <div class="flex items-center overflow-x-auto flex-1">
        <button @click="switchToHome"
          class="flex items-center px-4 py-3 text-sm font-medium border-r border-border transition-colors hover:bg-muted/50 focus:outline-none"
          :class="{ 'bg-background text-primary border-b-2 border-b-primary': activeTabId === null, 'text-muted-foreground': activeTabId !== null }">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
            class="lucide lucide-plus mr-2">
            <path d="M5 12h14" />
            <path d="M12 5v14" />
          </svg>
          New Connection
        </button>

        <div v-for="conn in connections" :key="conn.id" @click="switchToTab(conn.id)"
          class="flex items-center px-4 py-3 text-sm font-medium border-r border-border cursor-pointer transition-colors hover:bg-muted/50 select-none group min-w-[150px] max-w-[250px]"
          :class="{ 'bg-background text-primary border-b-2 border-b-primary': activeTabId === conn.id, 'text-muted-foreground': activeTabId !== conn.id }">
          <div class="flex items-center truncate mr-2">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
              stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
              class="lucide lucide-database mr-2 flex-shrink-0">
              <ellipse cx="12" cy="5" rx="9" ry="3" />
              <path d="M3 5V19A9 3 0 0 0 21 19V5" />
              <path d="M3 12A9 3 0 0 0 21 12" />
            </svg>
            <span class="truncate">{{ conn.name }}</span>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-2 px-3 flex-shrink-0 border-l border-border">
        <span class="text-[11px] text-muted-foreground font-mono" v-if="updateNotificationRef?.currentVersion">
          v{{ updateNotificationRef.currentVersion }}
        </span>
        <button @click="updateNotificationRef?.manualCheck()"
          class="flex items-center justify-center w-7 h-7 rounded-md transition-colors hover:bg-muted/60 text-muted-foreground hover:text-primary"
          :class="{ 'animate-spin': updateNotificationRef?.checking }" title="Check for updates">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
            <path d="M3 3v5h5" />
            <path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16" />
            <path d="M21 21v-5h-5" />
          </svg>
        </button>
      </div>
    </div>

    <div class="flex-1 overflow-hidden relative z-10">
      <div v-show="activeTabId === null" class="h-full overflow-auto">
        <DbConnection :activeConnections="connections" :pendingSqlFile="pendingSqlFile" @connected="handleConnected"
          @connection-exists="handleConnectionExists" @connection-updated="handleConnectionUpdate" />
      </div>

      <KeepAlive :max="5">
        <DbDashboard v-if="activeConn" :key="activeConn.id" :connectionId="activeConn.id"
          :connectionName="activeConn.name" :sessionKey="activeConn.sessionKey" :dbType="activeConn.config.type"
          :isReadOnly="activeConn.config.readOnly" @disconnect="handleDisconnect" />
      </KeepAlive>
    </div>

    <div v-if="showRecoveryModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm p-4">
      <div class="w-full max-w-xl rounded-xl border border-border bg-card text-card-foreground shadow-xl p-6 space-y-4">
        <div>
          <h2 class="text-lg font-semibold">Restore Previous Connections</h2>
          <p class="text-sm text-muted-foreground">Select the connections to recover from the previous session.</p>
        </div>

        <div class="flex items-center justify-end gap-2">
          <button @click="selectAllRecoveryCandidates"
            class="inline-flex items-center justify-center rounded-md text-xs font-medium border border-input bg-background hover:bg-accent h-8 px-3 py-1.5">
            Select All
          </button>
          <button @click="unselectAllRecoveryCandidates"
            class="inline-flex items-center justify-center rounded-md text-xs font-medium border border-input bg-background hover:bg-accent h-8 px-3 py-1.5">
            Unselect All
          </button>
        </div>

        <div class="max-h-[340px] overflow-y-auto border border-border rounded-md">
          <label v-for="candidate in recoveryCandidates" :key="candidate.sessionKey"
            class="flex items-start gap-3 p-3 border-b border-border last:border-b-0 hover:bg-muted/40 cursor-pointer">
            <input v-model="candidate.selected" type="checkbox" class="mt-1 h-4 w-4" />
            <div class="min-w-0">
              <div class="text-sm font-medium truncate">{{ candidate.name }}</div>
              <div class="text-xs text-muted-foreground truncate">{{ candidate.config.type }} � {{ candidate.config.database || candidate.config.host }}</div>
            </div>
          </label>
        </div>

        <label class="flex items-center gap-2 text-sm text-muted-foreground">
          <input v-model="rememberRecoveryChoice" type="checkbox" class="h-4 w-4" />
          Remember my choice (auto-recover next time without asking)
        </label>

        <div class="flex justify-end gap-2">
          <button @click="skipRecoveryPrompt"
            class="inline-flex items-center justify-center rounded-md text-sm font-medium border border-input bg-background hover:bg-accent h-9 px-4 py-2">
            Skip
          </button>
          <button @click="recoverSelectedConnections" :disabled="!hasAnyRecoverySelection"
            class="inline-flex items-center justify-center rounded-md text-sm font-medium bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 h-9 px-4 py-2">
            Recover Selected
          </button>
        </div>
      </div>
    </div>

    <div v-if="isRecoveringSession"
      class="fixed bottom-4 right-4 z-40 rounded-md border border-border bg-card px-4 py-2 text-sm shadow">
      Recovering selected connections...
    </div>

    <div v-if="recoveryErrors.length > 0"
      class="fixed bottom-4 left-4 z-40 max-w-lg rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive">
      <div class="font-medium mb-1">Some connections could not be recovered:</div>
      <ul class="list-disc pl-5">
        <li v-for="item in recoveryErrors" :key="item">{{ item }}</li>
      </ul>
    </div>

    <UpdateNotification ref="updateNotificationRef" />
  </div>
</template>

<style></style>


