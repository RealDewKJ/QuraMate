<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import AppRecoveryFeedback from './app/AppRecoveryFeedback.vue';
import AppRecoveryModal from './app/AppRecoveryModal.vue';
import AppShellBackground from './app/AppShellBackground.vue';
import AppTopTabBar from './app/AppTopTabBar.vue';
import DbConnection from './DbConnection.vue';
import DbDashboard from './DbDashboard.vue';
import UpdateNotification from './UpdateNotification.vue';
import { useAppShell } from '../composables/useAppShell';

const updateNotificationRef = ref<InstanceType<typeof UpdateNotification> | null>(null);

const currentVersion = computed(() => updateNotificationRef.value?.currentVersion || '');
const isCheckingUpdates = computed(() => updateNotificationRef.value?.checking || false);

const handleRequestUpdateCheck = () => {
  updateNotificationRef.value?.manualCheck();
};

const handleRecoveryConnectionSelection = (payload: { sessionKey: string; selected: boolean }) => {
  setRecoveryConnectionSelection(payload.sessionKey, payload.selected);
};

const {
  activeConnection,
  activeTabId,
  connections,
  initializeAppShell,
  openConnectionTab,
  openHomeTab,
  isRecoveringSession,
  pendingSqlFile,
  recoveryCandidates,
  recoveryErrors,
  rememberRecoveryChoice,
  setRecoveryConnectionSelection,
  setRememberRecoveryChoice,
  isRecoveryModalVisible,
  selectAllRecoveryConnections,
  dismissRecoveryPrompt,
  restoreSelectedConnections,
  clearRecoverySelections,
  closeConnectionTab,
  focusExistingConnection,
  registerConnection,
  updateConnectionConfig,
  t,
} = useAppShell();

onMounted(async () => {
  await initializeAppShell();
});
</script>

<template>
  <div class="relative flex h-screen flex-col overflow-hidden bg-background text-foreground">
    <AppShellBackground />

    <AppTopTabBar
      :active-tab-id="activeTabId"
      :connections="connections"
      :current-version="currentVersion"
      :is-checking-updates="isCheckingUpdates"
      :new-connection-label="t('common.newConnection')"
      :check-for-updates-label="t('common.checkForUpdates')"
      @request-update-check="handleRequestUpdateCheck"
      @open-home-tab="openHomeTab"
      @open-connection-tab="openConnectionTab"
    />

    <div class="relative z-10 flex-1 overflow-hidden">
      <div v-show="activeTabId === null" class="h-full overflow-auto">
        <DbConnection
          :active-connections="connections"
          :pending-sql-file="pendingSqlFile"
          @connected="registerConnection"
          @connection-exists="focusExistingConnection"
          @connection-updated="updateConnectionConfig"
        />
      </div>

      <KeepAlive :max="5">
        <DbDashboard
          v-if="activeConnection"
          :key="activeConnection.id"
          :connection-id="activeConnection.id"
          :connection-name="activeConnection.name"
          :session-key="activeConnection.sessionKey"
          :db-type="activeConnection.config.type"
          :is-read-only="activeConnection.config.readOnly"
          @disconnect="closeConnectionTab"
        />
      </KeepAlive>
    </div>

    <AppRecoveryModal
      v-if="isRecoveryModalVisible"
      :candidates="recoveryCandidates"
      :remember-recovery-choice="rememberRecoveryChoice"
      :title="t('common.recovery.title')"
      :description="t('common.recovery.description')"
      :select-all-label="t('common.recovery.selectAll')"
      :unselect-all-label="t('common.recovery.unselectAll')"
      :remember-choice-label="t('common.recovery.rememberChoice')"
      :skip-label="t('common.recovery.skip')"
      :recover-label="t('common.recovery.recoverSelected')"
      @restore-selected-connections="restoreSelectedConnections"
      @select-all-recovery-connections="selectAllRecoveryConnections"
      @dismiss-recovery-prompt="dismissRecoveryPrompt"
      @set-recovery-connection-selection="handleRecoveryConnectionSelection"
      @clear-recovery-selections="clearRecoverySelections"
      @set-remember-recovery-choice="setRememberRecoveryChoice"
    />

    <AppRecoveryFeedback
      :is-recovering-session="isRecoveringSession"
      :recovering-label="t('common.recovery.recoveringSelected')"
      :error-title="t('common.recovery.someFailed')"
      :errors="recoveryErrors"
    />

    <UpdateNotification ref="updateNotificationRef" />
  </div>
</template>
