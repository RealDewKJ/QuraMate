import { computed, onUnmounted, ref, watch } from 'vue';

import { type ConnectionConfig } from './useConnectionForm';
import {
  APP_CONNECTION_SESSION_KEY,
  APP_CONNECTION_SESSION_VERSION,
  buildConnectionSessionKey,
  normalizeConnectionConfig,
  type PersistedConnectionEntry,
  type PersistedConnectionSession,
} from './appConnectionSessionShared';
import { useAppConnectionRecovery } from './useAppConnectionRecovery';
import { ConnectDB, GetStartupFile, ReadTextFile, SaveSetting } from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';

interface AppConnection {
  id: string;
  name: string;
  config: ConnectionConfig;
  sessionKey: string;
}

interface PendingSqlFile {
  path: string;
  name: string;
  content: string;
}

interface UseAppConnectionSessionOptions {
  t: (key: string, params?: Record<string, unknown>) => string;
  getTrustSqlServerCertificateByDefault: () => boolean;
}

type ConnectedPayload = {
  id: string;
  name: string;
  config: ConnectionConfig;
  sessionKey?: string;
};

export function useAppConnectionSession(options: UseAppConnectionSessionOptions) {
  const { t, getTrustSqlServerCertificateByDefault } = options;

  const pendingSqlFile = ref<PendingSqlFile | null>(null);
  const connections = ref<AppConnection[]>([]);
  const activeTabId = ref<string | null>(null);

  const activeConnection = computed(() =>
    connections.value.find((connection) => connection.id === activeTabId.value) ?? null,
  );

  let appSessionSaveTimer: ReturnType<typeof setTimeout> | null = null;
  let disposeOpenFileListener: (() => void) | null = null;

  const buildSessionEntry = (connection: AppConnection): PersistedConnectionEntry => ({
    name: connection.name,
    sessionKey: connection.sessionKey,
    config: { ...connection.config },
  });

  const dispatchOpenSqlFile = (file: PendingSqlFile, connectionId: string) => {
    window.dispatchEvent(new CustomEvent('open-sql-file', {
      detail: {
        content: file.content,
        fileName: file.name,
        filePath: file.path,
        connectionId,
      },
    }));
  };

  const saveAppConnectionSession = async () => {
    const currentActiveConnection = connections.value.find((connection) => connection.id === activeTabId.value) ?? null;
    const payload: PersistedConnectionSession = {
      version: APP_CONNECTION_SESSION_VERSION,
      activeSessionKey: currentActiveConnection?.sessionKey || null,
      connections: connections.value.map(buildSessionEntry),
    };

    try {
      await SaveSetting(APP_CONNECTION_SESSION_KEY, JSON.stringify(payload));
    } catch (error) {
      console.error('Failed to persist app connection session', error);
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

  const openConnectionTab = (id: string) => {
    activeTabId.value = id;
  };

  const openHomeTab = () => {
    activeTabId.value = null;
  };

  const registerConnection = (connection: ConnectedPayload) => {
    const normalizedConfig = normalizeConnectionConfig(connection.config || {}, getTrustSqlServerCertificateByDefault);
    const sessionKey = connection.sessionKey || buildConnectionSessionKey(normalizedConfig, connection.name);

    const existingIndex = connections.value.findIndex((item) => item.sessionKey === sessionKey);
    if (existingIndex !== -1) {
      connections.value[existingIndex] = {
        ...connections.value[existingIndex],
        id: connection.id,
        name: connection.name || connections.value[existingIndex].name,
        config: normalizedConfig,
        sessionKey,
      };
      activeTabId.value = connection.id;
      return;
    }

    connections.value.push({
      id: connection.id,
      name: connection.name,
      config: normalizedConfig,
      sessionKey,
    });
    activeTabId.value = connection.id;

    if (!pendingSqlFile.value) {
      return;
    }

    const file = pendingSqlFile.value;
    pendingSqlFile.value = null;

    setTimeout(() => {
      dispatchOpenSqlFile(file, connection.id);
    }, 300);
  };

  const closeConnectionTab = (id: string) => {
    const index = connections.value.findIndex((connection) => connection.id === id);
    if (index === -1) {
      return;
    }

    connections.value.splice(index, 1);

    if (activeTabId.value !== id) {
      return;
    }

    activeTabId.value = connections.value.length > 0
      ? connections.value[connections.value.length - 1].id
      : null;
  };

  const focusExistingConnection = (id: string) => {
    openConnectionTab(id);
  };

  const updateConnectionConfig = (update: { id: string; config: ConnectionConfig }) => {
    const connection = connections.value.find((item) => item.id === update.id);
    if (!connection) {
      return;
    }

    connection.config = normalizeConnectionConfig(update.config, getTrustSqlServerCertificateByDefault);
  };

  const processStartupFile = async (startupFile: string) => {
    try {
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
          trustServerCertificate: false,
          sshEnabled: false,
          sshHost: '',
          sshPort: 22,
          sshUser: '',
          sshPassword: '',
          sshKeyFile: '',
          id: '',
        }, getTrustSqlServerCertificateByDefault);

        const result = await ConnectDB(payload);
        if (result && result.id && !result.error) {
          const fileName = startupFile.replace(/^.*[\\/]/, '');
          const displayName = `${fileName} (SQLite)`;
          registerConnection({
            id: result.id,
            name: displayName,
            config: payload,
            sessionKey: buildConnectionSessionKey(payload, displayName),
          });
        }

        return;
      }

      if (ext !== 'sql') {
        return;
      }

      const content = await ReadTextFile(startupFile);
      const fileName = startupFile.replace(/^.*[\\/]/, '');

      if (connections.value.length === 0) {
        pendingSqlFile.value = { path: startupFile, name: fileName, content };
        activeTabId.value = null;
        return;
      }

      const targetConnectionId = activeTabId.value || connections.value[0].id;
      openConnectionTab(targetConnectionId);
      dispatchOpenSqlFile({ path: startupFile, name: fileName, content }, targetConnectionId);
    } catch (error) {
      console.error('Failed to process startup file', error);
    }
  };

  const connectionRecovery = useAppConnectionRecovery({
    t,
    getTrustSqlServerCertificateByDefault,
    registerConnection,
    openConnectionTab,
  });

  const initializeConnectionSession = async () => {
    await connectionRecovery.initializeConnectionRecovery();

    try {
      const startupFile = await GetStartupFile();
      if (startupFile) {
        await processStartupFile(startupFile);
      }
    } catch (error) {
      console.error('Failed to fetch startup file', error);
    }

    disposeOpenFileListener = EventsOn('app:open-file', (filePath: string) => {
      if (filePath) {
        void processStartupFile(filePath);
      }
    });
  };

  watch([connections, activeTabId], scheduleSaveAppConnectionSession, { deep: true });

  onUnmounted(() => {
    if (appSessionSaveTimer) {
      clearTimeout(appSessionSaveTimer);
      appSessionSaveTimer = null;
    }

    disposeOpenFileListener?.();
    void saveAppConnectionSession();
  });

  return {
    activeConnection,
    activeTabId,
    connections,
    openConnectionTab,
    openHomeTab,
    registerConnection,
    focusExistingConnection,
    updateConnectionConfig,
    closeConnectionTab,
    initializeConnectionSession,
    pendingSqlFile,
    ...connectionRecovery,
  };
}
