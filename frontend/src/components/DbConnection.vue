<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import ConnectionActionsBar from './connection/ConnectionActionsBar.vue';
import ConnectionDetailsForm from './connection/ConnectionDetailsForm.vue';
import ConnectionHeroCard from './connection/ConnectionHeroCard.vue';
import ConnectionSecurityPanel from './connection/ConnectionSecurityPanel.vue';
import ConnectionSshSection from './connection/ConnectionSshSection.vue';
import SavedConnectionsModal from './connection/SavedConnectionsModal.vue';
import SettingsDialog from './SettingsDialog.vue';
import Toast from './Toast.vue';
import {
  useConnectionForm,
  type ActiveConnection,
  type ConnectionConfig,
  type ConnectionInputMode,
} from '../composables/useConnectionForm';
import { useDbConnectionUi } from '../composables/useDbConnectionUi';

const props = defineProps<{
  activeConnections: ActiveConnection[];
  pendingSqlFile?: { path: string; name: string; content: string } | null;
}>();

const emit = defineEmits<{
  (e: 'connected', conn: ActiveConnection): void;
  (e: 'connection-exists', id: string): void;
  (e: 'connection-updated', update: { id: string; config: ConnectionConfig }): void;
}>();

const trustServerCertificateDefault = ref(true);

const {
  config,
  connectionInputMode,
  connectionString,
  supportsConnectionString,
  connectionStringPlaceholder,
  error,
  testSuccess,
  isLoading,
  isTesting,
  isTrustingHost,
  isLoadingHostKeyInfo,
  sshHostKeyInfo,
  expectedSshFingerprint,
  sshRotationReason,
  sshRotationConfirmText,
  isFingerprintMatch,
  isFingerprintMismatch,
  isPinnedFingerprintMismatch,
  pinnedSshFingerprint,
  sshTrustAudit,
  sshTrustAuditSearch,
  filteredSshTrustAudit,
  canTrustCurrentSshHost,
  isQuickConnecting,
  savedConnections,
  connectionLabel,
  fieldErrors,
  handleSelectSqliteFile,
  cancelConnection,
  connect,
  testConnection,
  removeConnection,
  selectConnection,
  editConnection,
  migrateSavedConnections,
  loadCurrentSshHostKeyInfo,
  trustCurrentSshHost,
  acceptPinnedFingerprintRotation,
  copyCurrentSshFingerprint,
  clearSshTrustAudit,
  exportSshTrustAudit,
  importSshTrustAudit,
} = useConnectionForm(
  () => props.activeConnections,
  (conn) => emit('connected', conn),
  (id) => emit('connection-exists', id),
  (update) => emit('connection-updated', update),
  () => trustServerCertificateDefault.value,
);

const {
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
  setSavedConnectionsAnnouncement,
  showPassword,
  showSavedModal,
  showSettings,
  showSqlServerSettings,
  showSshPassword,
  togglePasswordVisibility,
  toggleSshPasswordVisibility,
} = useDbConnectionUi({
  config,
  activeConnections: () => props.activeConnections,
  trustServerCertificateDefault,
});

const supportsSsh = computed(
  () => !['sqlite', 'duckdb', 'libsql', 'd1'].includes(config.type),
);

const updateConfig = (patch: Partial<ConnectionConfig>) => {
  Object.assign(config, patch);
};

const updateConnectionInputMode = (mode: ConnectionInputMode) => {
  connectionInputMode.value = mode;
};

const updateConnectionString = (value: string) => {
  connectionString.value = value;
};

const updateExpectedSshFingerprint = (value: string) => {
  expectedSshFingerprint.value = value;
};

const updateSshRotationReason = (value: string) => {
  sshRotationReason.value = value;
};

const updateSshRotationConfirmText = (value: string) => {
  sshRotationConfirmText.value = value;
};

const updateSshTrustAuditSearch = (value: string) => {
  sshTrustAuditSearch.value = value;
};

const handleSelectDatabaseType = (type: ConnectionConfig['type']) => {
  databaseTypeUi.value.selectOption(type);
};

const handleSelectSavedConnection = (connection: ConnectionConfig) => {
  closeSavedConnections();
  selectConnection(connection);
};

const handleEditSavedConnection = (connection: ConnectionConfig) => {
  closeSavedConnections();
  editConnection(connection);
};

const handleRemoveSavedConnection = (connection: ConnectionConfig) => {
  const index = savedConnections.value.findIndex((item) => item.id === connection.id);
  if (index === -1) {
    return;
  }

  removeConnection(index);
  const displayName = connection.name || connection.database || connection.host || connection.type;
  setSavedConnectionsAnnouncement(`Deleted saved connection ${displayName}.`);
};

onMounted(() => {
  migrateSavedConnections();
});

const toastRef = ref<InstanceType<typeof Toast> | null>(null);
</script>

<template>
  <div class="flex min-h-screen flex-col items-center justify-center p-4 transition-colors duration-300">
    <div class="absolute right-4 top-4">
      <button
        type="button"
        aria-label="Open settings"
        class="rounded-full p-2 transition-colors hover:bg-accent hover:text-accent-foreground"
        @click="openSettings"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
          <circle cx="12" cy="12" r="3" />
        </svg>
      </button>
    </div>

    <div class="w-full max-w-md space-y-4 rounded-xl border bg-card p-6 text-card-foreground shadow-lg">
      <ConnectionHeroCard
        :connection-label="connectionLabel"
        :error="error"
        :is-loading="isLoading"
        :is-quick-connecting="isQuickConnecting"
        :pending-sql-file="pendingSqlFile"
        @cancel-connection="cancelConnection"
      />

      <div v-if="!(isLoading && isQuickConnecting && !error)" class="space-y-3">
        <ConnectionDetailsForm
          :config="config"
          :field-errors="fieldErrors"
          :connection-input-mode="connectionInputMode"
          :supports-connection-string="supportsConnectionString"
          :connection-string="connectionString"
          :connection-string-placeholder="connectionStringPlaceholder"
          :database-type-options="databaseTypeOptions"
          :database-type-ui="databaseTypeUi"
          :input-mode-options="inputModeOptions"
          :selected-database-type-label="selectedDatabaseTypeLabel"
          :show-password="showPassword"
          @select-database-type="handleSelectDatabaseType"
          @select-sqlite-file="handleSelectSqliteFile"
          @toggle-password="togglePasswordVisibility"
          @update-config="updateConfig"
          @update-connection-input-mode="updateConnectionInputMode"
          @update-connection-string="updateConnectionString"
        />

        <ConnectionSshSection
          :config="config"
          :field-errors="fieldErrors"
          :show-ssh-password="showSshPassword"
          :supports-ssh="supportsSsh"
          @toggle-ssh-password="toggleSshPasswordVisibility"
          @update-config="updateConfig"
        />

        <ConnectionSecurityPanel
          :can-trust-current-ssh-host="canTrustCurrentSshHost"
          :config="config"
          :error="error"
          :expected-ssh-fingerprint="expectedSshFingerprint"
          :filtered-ssh-trust-audit="filteredSshTrustAudit"
          :is-fingerprint-match="isFingerprintMatch"
          :is-fingerprint-mismatch="isFingerprintMismatch"
          :is-loading-host-key-info="isLoadingHostKeyInfo"
          :is-pinned-fingerprint-mismatch="isPinnedFingerprintMismatch"
          :is-trusting-host="isTrustingHost"
          :pinned-ssh-fingerprint="pinnedSshFingerprint"
          :ssh-host-key-info="sshHostKeyInfo"
          :ssh-rotation-confirm-text="sshRotationConfirmText"
          :ssh-rotation-reason="sshRotationReason"
          :ssh-trust-audit="sshTrustAudit"
          :ssh-trust-audit-search="sshTrustAuditSearch"
          @accept-pinned-fingerprint-rotation="acceptPinnedFingerprintRotation"
          @clear-ssh-trust-audit="clearSshTrustAudit"
          @copy-current-ssh-fingerprint="copyCurrentSshFingerprint"
          @export-ssh-trust-audit="exportSshTrustAudit"
          @import-ssh-trust-audit="importSshTrustAudit"
          @load-current-ssh-host-key-info="loadCurrentSshHostKeyInfo"
          @trust-current-ssh-host="trustCurrentSshHost"
          @update-config="updateConfig"
          @update-expected-ssh-fingerprint="updateExpectedSshFingerprint"
          @update-ssh-rotation-confirm-text="updateSshRotationConfirmText"
          @update-ssh-rotation-reason="updateSshRotationReason"
          @update-ssh-trust-audit-search="updateSshTrustAuditSearch"
        />

        <ConnectionActionsBar
          :is-loading="isLoading"
          :is-testing="isTesting"
          :test-success="testSuccess"
          @cancel-connection="cancelConnection"
          @connect="connect"
          @open-saved-connections="openSavedConnections"
          @test-connection="testConnection"
        />
      </div>
    </div>

    <SavedConnectionsModal
      :is-open="showSavedModal"
      :connections="savedConnections"
      :status-announcement="savedConnectionsAnnouncement"
      @close="closeSavedConnections"
      @select="handleSelectSavedConnection"
      @edit="handleEditSavedConnection"
      @remove="handleRemoveSavedConnection"
    />

    <Toast ref="toastRef" />

    <SettingsDialog
      :is-open="showSettings"
      :show-sql-server-settings="showSqlServerSettings"
      @close="closeSettings"
      @save="handleSettingsSave"
    />
  </div>
</template>
