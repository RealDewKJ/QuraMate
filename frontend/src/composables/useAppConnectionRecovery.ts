import { ref } from 'vue';

import { getConnectionLabel, type ConnectionConfig } from './useConnectionForm';
import {
  APP_CONNECTION_RECOVERY_PREF_KEY,
  APP_CONNECTION_RECOVERY_PREF_VERSION,
  APP_CONNECTION_SESSION_KEY,
  APP_CONNECTION_SESSION_VERSION,
  buildConnectionSessionKey,
  normalizeConnectionConfig,
  type PersistedConnectionEntry,
  type PersistedConnectionSession,
  type RecoveryCandidate,
  type RecoveryPreference,
} from './appConnectionSessionShared';
import { ConnectDB, LoadSetting, SaveSetting } from '../../wailsjs/go/app/App';

interface ConnectedPayload {
  id: string;
  name: string;
  config: ConnectionConfig;
  sessionKey?: string;
}

interface UseAppConnectionRecoveryOptions {
  t: (key: string, params?: Record<string, unknown>) => string;
  getTrustSqlServerCertificateByDefault: () => boolean;
  registerConnection: (connection: ConnectedPayload) => void;
  openConnectionTab: (id: string) => void;
}

export function useAppConnectionRecovery(options: UseAppConnectionRecoveryOptions) {
  const { t, getTrustSqlServerCertificateByDefault, registerConnection, openConnectionTab } = options;

  const isRecoveryModalVisible = ref(false);
  const isRecoveringSession = ref(false);
  const recoveryCandidates = ref<RecoveryCandidate[]>([]);
  const recoveryPreferredActiveSessionKey = ref<string | null>(null);
  const recoveryErrors = ref<string[]>([]);
  const rememberRecoveryChoice = ref(false);

  const loadRecoveryPreference = async (): Promise<RecoveryPreference | null> => {
    try {
      const raw = await LoadSetting(APP_CONNECTION_RECOVERY_PREF_KEY);
      if (!raw) {
        return null;
      }

      const parsed = JSON.parse(raw) as Partial<RecoveryPreference>;
      if (parsed.version !== APP_CONNECTION_RECOVERY_PREF_VERSION || typeof parsed.autoRecover !== 'boolean') {
        return null;
      }

      return {
        version: APP_CONNECTION_RECOVERY_PREF_VERSION,
        autoRecover: parsed.autoRecover,
        selectedSessionKeys: Array.isArray(parsed.selectedSessionKeys)
          ? parsed.selectedSessionKeys.filter((key): key is string => typeof key === 'string')
          : [],
      };
    } catch (error) {
      console.error('Failed to load recovery preference', error);
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
    } catch (error) {
      console.error('Failed to save recovery preference', error);
    }
  };

  const clearRecoveryPreference = async () => {
    try {
      await SaveSetting(APP_CONNECTION_RECOVERY_PREF_KEY, '');
    } catch (error) {
      console.error('Failed to clear recovery preference', error);
    }
  };

  const clearAppConnectionSession = async () => {
    try {
      await SaveSetting(APP_CONNECTION_SESSION_KEY, '');
    } catch (error) {
      console.error('Failed to clear app connection session', error);
    }
  };

  const selectAllRecoveryConnections = () => {
    recoveryCandidates.value.forEach((candidate) => {
      candidate.selected = true;
    });
  };

  const clearRecoverySelections = () => {
    recoveryCandidates.value.forEach((candidate) => {
      candidate.selected = false;
    });
  };

  const setRecoveryConnectionSelection = (sessionKey: string, selected: boolean) => {
    const candidate = recoveryCandidates.value.find((item) => item.sessionKey === sessionKey);
    if (!candidate) {
      return;
    }

    candidate.selected = selected;
  };

  const setRememberRecoveryChoice = (value: boolean) => {
    rememberRecoveryChoice.value = value;
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
      try {
        const result = await ConnectDB(entry.config);
        if (result && result.id && !result.error) {
          registerConnection({
            id: result.id,
            name: entry.name || getConnectionLabel(entry.config),
            config: entry.config,
            sessionKey: entry.sessionKey,
          });
          recoveredIdBySessionKey.set(entry.sessionKey, result.id);
          continue;
        }

        recoveryErrors.value.push(t('common.recovery.failureItem', {
          name: entry.name,
          message: result?.error || t('common.errors.unknownConnectionError'),
        }));
      } catch (error: unknown) {
        const message = error instanceof Error
          ? error.message
          : String(error || t('common.errors.connectionFailed'));

        recoveryErrors.value.push(t('common.recovery.failureItem', {
          name: entry.name,
          message,
        }));
      }
    }

    const preferredId = recoveryPreferredActiveSessionKey.value
      ? recoveredIdBySessionKey.get(recoveryPreferredActiveSessionKey.value)
      : null;

    if (preferredId) {
      openConnectionTab(preferredId);
    } else if (recoveredIdBySessionKey.size > 0) {
      const [firstRecoveredId] = recoveredIdBySessionKey.values();
      openConnectionTab(firstRecoveredId);
    }

    isRecoveringSession.value = false;
    recoveryCandidates.value = [];
    recoveryPreferredActiveSessionKey.value = null;
  };

  const openRecoveryPrompt = async () => {
    try {
      const raw = await LoadSetting(APP_CONNECTION_SESSION_KEY);
      if (!raw) {
        return;
      }

      const parsed = JSON.parse(raw) as Partial<PersistedConnectionSession>;
      if (parsed.version !== APP_CONNECTION_SESSION_VERSION || !Array.isArray(parsed.connections) || parsed.connections.length === 0) {
        return;
      }

      recoveryPreferredActiveSessionKey.value = parsed.activeSessionKey || null;

      const candidates = parsed.connections
        .filter((entry): entry is PersistedConnectionEntry => !!entry?.config)
        .map((entry) => ({
          name: entry.name || getConnectionLabel(entry.config),
          sessionKey: entry.sessionKey || buildConnectionSessionKey(entry.config, entry.name),
          config: normalizeConnectionConfig(entry.config, getTrustSqlServerCertificateByDefault),
        }));

      if (candidates.length === 0) {
        return;
      }

      const preference = await loadRecoveryPreference();
      if (preference?.autoRecover) {
        const preferredKeys = new Set(preference.selectedSessionKeys);
        const autoTargets = preferredKeys.size > 0
          ? candidates.filter((candidate) => preferredKeys.has(candidate.sessionKey))
          : candidates;

        if (autoTargets.length > 0) {
          await recoverConnections(autoTargets);
          return;
        }
      }

      rememberRecoveryChoice.value = false;
      recoveryCandidates.value = candidates.map((candidate) => ({ ...candidate, selected: true }));
      isRecoveryModalVisible.value = true;
    } catch (error) {
      console.error('Failed to load connection recovery session', error);
    }
  };

  const dismissRecoveryPrompt = async () => {
    isRecoveryModalVisible.value = false;
    recoveryCandidates.value = [];
    recoveryPreferredActiveSessionKey.value = null;
    recoveryErrors.value = [];
    rememberRecoveryChoice.value = false;
    await clearAppConnectionSession();
    await clearRecoveryPreference();
  };

  const restoreSelectedConnections = async () => {
    const targets = recoveryCandidates.value.filter((candidate) => candidate.selected);
    isRecoveryModalVisible.value = false;

    if (rememberRecoveryChoice.value) {
      await saveRecoveryPreference(true, targets.map((target) => target.sessionKey));
    } else {
      await clearRecoveryPreference();
    }

    await recoverConnections(targets);
  };

  const initializeConnectionRecovery = async () => {
    await openRecoveryPrompt();
  };

  return {
    dismissRecoveryPrompt,
    initializeConnectionRecovery,
    isRecoveringSession,
    isRecoveryModalVisible,
    recoveryCandidates,
    recoveryErrors,
    rememberRecoveryChoice,
    restoreSelectedConnections,
    selectAllRecoveryConnections,
    clearRecoverySelections,
    setRecoveryConnectionSelection,
    setRememberRecoveryChoice,
  };
}
