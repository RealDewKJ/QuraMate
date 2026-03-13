<script lang="ts" setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from "vue";
import { useToggle, onKeyStroke } from "@vueuse/core";
import { LoadSetting } from "../../wailsjs/go/app/App";
import SettingsDialog from "./SettingsDialog.vue";
import Toast from "./Toast.vue";
import SavedConnectionsModal from "./connection/SavedConnectionsModal.vue";
import {
  useConnectionForm,
  type ConnectionConfig,
  type ActiveConnection,
} from "../composables/useConnectionForm";

const toastRef = ref<InstanceType<typeof Toast> | null>(null);
const trustServerCertificateDefault = ref(true);

const applyTrustServerCertificateDefault = (value: boolean) => {
  trustServerCertificateDefault.value = value;
  if (config.type === "mssql") {
    config.trustServerCertificate = value;
  }
};

const handleSettingsSave = (nextSettings?: any) => {
  const nextValue =
    nextSettings?.general?.trustSqlServerCertificateByDefault !== false;
  applyTrustServerCertificateDefault(nextValue);
};

const loadConnectionDefaults = async () => {
  try {
    const savedSettingsJson = await LoadSetting("user_settings");
    if (!savedSettingsJson) return;
    const parsed = JSON.parse(savedSettingsJson);
    applyTrustServerCertificateDefault(
      parsed?.general?.trustSqlServerCertificateByDefault !== false,
    );
  } catch (e) {
    console.error("Failed to load connection defaults:", e);
  }
};

const props = defineProps<{
  activeConnections: ActiveConnection[];
  pendingSqlFile?: { path: string; name: string; content: string } | null;
}>();

const emit = defineEmits<{
  connected: [conn: ActiveConnection];
  "connection-exists": [id: string];
  "connection-updated": [update: { id: string; config: ConnectionConfig }];
}>();

const {
  config,
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
  (conn) => emit("connected", conn),
  (id) => emit("connection-exists", id),
  (update) => emit("connection-updated", update),
  () => trustServerCertificateDefault.value,
);

const [showSettings, toggleSettings] = useToggle(false);
const [showSavedModal, toggleSavedModal] = useToggle(false);
const [showPassword, togglePassword] = useToggle(false);
const [showSshPassword, toggleSshPassword] = useToggle(false);
const savedConnectionsAnnouncement = ref("");
const databaseTypeButtonRef = ref<HTMLButtonElement | null>(null);
const databaseTypeMenuRef = ref<HTMLElement | null>(null);
const isDatabaseTypeMenuOpen = ref(false);
const highlightedDatabaseTypeIndex = ref(0);
let databaseTypeTypeahead = "";
let databaseTypeTypeaheadTimeout: ReturnType<typeof setTimeout> | null = null;
const supportsSsh = computed(
  () => !["sqlite", "duckdb", "libsql"].includes(config.type),
);
const primaryFieldClass =
  "flex min-h-[44px] h-auto w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm text-foreground ring-offset-background transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50";
const primaryFieldWithIconClass = `${primaryFieldClass} h-11 pr-10`;
const iconButtonFieldClass =
  "inline-flex items-center justify-center whitespace-nowrap rounded-full border border-border/80 bg-background/95 text-sm font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50";
const ghostPillButtonClass =
  "inline-flex items-center justify-center whitespace-nowrap rounded-full border border-border/80 bg-background/95 text-sm font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50";
const primaryPillButtonClass =
  "inline-flex items-center justify-center whitespace-nowrap rounded-full bg-primary px-4 text-sm font-semibold text-primary-foreground shadow-sm transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50";
const destructivePillButtonClass =
  "inline-flex items-center justify-center whitespace-nowrap rounded-full border border-destructive/60 bg-background/95 text-sm font-medium text-destructive shadow-sm transition-colors hover:bg-destructive/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50";
const databaseTypeOptions = [
  { value: "postgres", label: "PostgreSQL" },
  { value: "mysql", label: "MySQL" },
  { value: "mariadb", label: "MariaDB" },
  { value: "mssql", label: "MSSQL" },
  { value: "sqlite", label: "SQLite" },
  { value: "duckdb", label: "DuckDB" },
  { value: "greenplum", label: "Greenplum" },
  { value: "redshift", label: "Redshift" },
  { value: "cockroachdb", label: "CockroachDB" },
  { value: "databend", label: "Databend" },
  { value: "libsql", label: "LibSQL" },
] as const;
const selectedDatabaseTypeLabel = computed(
  () =>
    databaseTypeOptions.find((option) => option.value === config.type)?.label ??
    "Select database type",
);
const selectedDatabaseTypeIndex = computed(() =>
  Math.max(
    0,
    databaseTypeOptions.findIndex((option) => option.value === config.type),
  ),
);
const showSqlServerSettings = computed(
  () =>
    config.type === "mssql" ||
    props.activeConnections.some(
      (conn) => (conn.config.type || "").toLowerCase() === "mssql",
    ),
);

onKeyStroke("Escape", () => {
  showSettings.value = false;
  isDatabaseTypeMenuOpen.value = false;
});

const closeDatabaseTypeMenu = () => {
  isDatabaseTypeMenuOpen.value = false;
};

const clearDatabaseTypeTypeahead = () => {
  databaseTypeTypeahead = "";
  if (databaseTypeTypeaheadTimeout) {
    clearTimeout(databaseTypeTypeaheadTimeout);
    databaseTypeTypeaheadTimeout = null;
  }
};

const focusHighlightedDatabaseType = async () => {
  await nextTick();
  const optionEls =
    databaseTypeMenuRef.value?.querySelectorAll<HTMLElement>("[data-db-type-option]");
  optionEls?.[highlightedDatabaseTypeIndex.value]?.focus();
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

const selectDatabaseType = (type: ConnectionConfig["type"]) => {
  config.type = type;
  highlightedDatabaseTypeIndex.value = selectedDatabaseTypeIndex.value;
  closeDatabaseTypeMenu();
  nextTick(() => databaseTypeButtonRef.value?.focus());
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
    databaseTypeTypeahead = "";
    databaseTypeTypeaheadTimeout = null;
  }, 500);

  const matchIndex = databaseTypeOptions.findIndex((option) =>
    option.label.toLowerCase().startsWith(databaseTypeTypeahead),
  );

  if (matchIndex >= 0) {
    highlightedDatabaseTypeIndex.value = matchIndex;
    if (!isDatabaseTypeMenuOpen.value) {
      await openDatabaseTypeMenu();
    } else {
      await focusHighlightedDatabaseType();
    }
  }
};

const handleDatabaseTypeTriggerKeydown = async (event: KeyboardEvent) => {
  if (event.key === "ArrowDown" || event.key === "ArrowUp") {
    event.preventDefault();
    if (!isDatabaseTypeMenuOpen.value) {
      await openDatabaseTypeMenu();
    } else {
      moveDatabaseTypeHighlight(event.key === "ArrowDown" ? 1 : -1);
      await focusHighlightedDatabaseType();
    }
    return;
  }

  if (event.key === "Enter" || event.key === " ") {
    event.preventDefault();
    await toggleDatabaseTypeMenu();
    return;
  }

  await handleDatabaseTypeTypeahead(event.key);
};

const handleDatabaseTypeOptionKeydown = async (event: KeyboardEvent) => {
  if (event.key === "ArrowDown" || event.key === "ArrowUp") {
    event.preventDefault();
    moveDatabaseTypeHighlight(event.key === "ArrowDown" ? 1 : -1);
    await focusHighlightedDatabaseType();
    return;
  }

  if (event.key === "Enter" || event.key === " ") {
    event.preventDefault();
    const option = databaseTypeOptions[highlightedDatabaseTypeIndex.value];
    if (option) {
      selectDatabaseType(option.value);
    }
    return;
  }

  if (event.key === "Escape") {
    event.preventDefault();
    closeDatabaseTypeMenu();
    databaseTypeButtonRef.value?.focus();
    return;
  }

  if (event.key === "Home") {
    event.preventDefault();
    highlightedDatabaseTypeIndex.value = 0;
    await focusHighlightedDatabaseType();
    return;
  }

  if (event.key === "End") {
    event.preventDefault();
    highlightedDatabaseTypeIndex.value = databaseTypeOptions.length - 1;
    await focusHighlightedDatabaseType();
    return;
  }

  await handleDatabaseTypeTypeahead(event.key);
};

const handleDocumentClick = (event: MouseEvent) => {
  const target = event.target as HTMLElement | null;
  if (!target?.closest("[data-db-type-menu]")) {
    closeDatabaseTypeMenu();
  }
};

const handleSelectConn = (conn: ConnectionConfig) => {
  showSavedModal.value = false;
  selectConnection(conn);
};

const handleEditConn = (conn: ConnectionConfig) => {
  showSavedModal.value = false;
  editConnection(conn);
};

const handleRemoveConn = (conn: ConnectionConfig) => {
  const index = savedConnections.value.findIndex((c) => c.id === conn.id);
  if (index !== -1) {
    removeConnection(index);
    const displayName = conn.name || conn.database || conn.host || conn.type;
    savedConnectionsAnnouncement.value = `Deleted saved connection ${displayName}.`;
  }
};

onMounted(async () => {
  await loadConnectionDefaults();
  migrateSavedConnections();
  document.addEventListener("click", handleDocumentClick);
});

onBeforeUnmount(() => {
  document.removeEventListener("click", handleDocumentClick);
  clearDatabaseTypeTypeahead();
});
</script>

<template>
  <div class="flex flex-col items-center justify-center min-h-screen p-4 transition-colors duration-300">
    <div class="absolute top-4 right-4">
      <button @click="toggleSettings(true)" aria-label="Open settings"
        class="p-2 rounded-full hover:bg-accent hover:text-accent-foreground transition-colors">
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
          stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
          class="lucide lucide-settings">
          <path
            d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
          <circle cx="12" cy="12" r="3" />
        </svg>
      </button>
    </div>

    <div v-if="pendingSqlFile"
      class="w-full max-w-md mb-3 p-3 bg-primary/10 border border-primary/30 rounded-lg flex items-center gap-3 animate-in fade-in slide-in-from-top-4 duration-300">
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
        class="text-primary flex-shrink-0">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
        <path d="M14 2v6h6" />
        <path d="M10 12.5 8 15l2 2.5" />
        <path d="m14 12.5 2 2.5-2 2.5" />
      </svg>
      <div class="min-w-0">
        <p class="text-sm font-medium text-primary truncate">📄 {{ pendingSqlFile.name }}</p>
        <p class="text-xs text-muted-foreground">Connect to a database to open this SQL file</p>
      </div>
    </div>

    <div class="w-full max-w-md space-y-4 bg-card text-card-foreground p-6 rounded-xl border shadow-lg">
      <div class="text-center space-y-1">
        <div class="flex items-center justify-center mx-auto mb-2">
          <img src="../assets/images/new-icon.png" class="w-20 h-20 object-contain" alt="QuraMate Icon" />
        </div>
        <p class="text-muted-foreground text-sm">
          QuraMate - Connect to your database to start managing data.
        </p>
      </div>

      <div v-if="isLoading && isQuickConnecting && !error"
        class="py-12 flex flex-col items-center justify-center space-y-4 animate-in fade-in duration-500">
        <div class="relative w-16 h-16">
          <div class="absolute inset-0 rounded-full border-4 border-primary/20"></div>
          <div class="absolute inset-0 rounded-full border-4 border-primary border-t-transparent animate-spin"></div>
        </div>
        <div class="text-center">
          <p class="font-medium text-lg">
            Connecting to {{ connectionLabel }}...
          </p>
          <p class="text-sm text-muted-foreground">
            Please wait while we establish a secure connection.
          </p>
        </div>
        <button @click="cancelConnection"
          :class="`${destructivePillButtonClass} h-11 px-4`">
          Cancel
        </button>
      </div>

      <div v-else class="space-y-3">
        <div class="space-y-3 animate-in fade-in slide-in-from-top-4 duration-300">
          <div class="space-y-2">
            <label class="text-sm font-medium leading-none" for="connName">Connection Name (Optional)</label>
            <input v-model="config.name" id="connName" type="text" placeholder="My Database"
              :class="primaryFieldClass" />
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
              for="dbType">Database Type</label>
            <div ref="databaseTypeMenuRef" class="relative" data-db-type-menu>
              <button ref="databaseTypeButtonRef" id="dbType" type="button"
                :aria-expanded="isDatabaseTypeMenuOpen ? 'true' : 'false'" aria-haspopup="listbox"
                @click="toggleDatabaseTypeMenu" @keydown="handleDatabaseTypeTriggerKeydown"
                :class="`${primaryFieldClass} justify-between hover:bg-accent/40`">
                <span>{{ selectedDatabaseTypeLabel }}</span>
                <svg class="h-4 w-4 text-muted-foreground transition-transform duration-150"
                  :class="isDatabaseTypeMenuOpen ? 'rotate-180' : ''" xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                  <path fill-rule="evenodd"
                    d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                    clip-rule="evenodd" />
                </svg>
              </button>
              <div v-if="isDatabaseTypeMenuOpen" role="listbox" aria-labelledby="dbType"
                class="absolute left-0 top-full z-50 mt-2 w-full overflow-hidden rounded-2xl border border-border/80 bg-popover/95 py-1 text-popover-foreground shadow-xl ring-1 ring-black/5 backdrop-blur animate-in fade-in zoom-in-95 duration-100">
                <button v-for="(option, index) in databaseTypeOptions" :key="option.value" type="button"
                  data-db-type-option role="option" :tabindex="highlightedDatabaseTypeIndex === index ? 0 : -1"
                  :aria-selected="config.type === option.value ? 'true' : 'false'" @click="selectDatabaseType(option.value)"
                  @focus="highlightedDatabaseTypeIndex = index" @keydown="handleDatabaseTypeOptionKeydown"
                  class="flex w-full items-center justify-between px-3 py-2 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground"
                  :class="[
                    highlightedDatabaseTypeIndex === index ? 'bg-accent/50 text-accent-foreground' : '',
                    config.type === option.value ? 'bg-accent/70 text-accent-foreground' : ''
                  ]">
                  <span>{{ option.label }}</span>
                  <svg v-if="config.type === option.value" xmlns="http://www.w3.org/2000/svg" width="14" height="14"
                    viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                    stroke-linejoin="round" class="lucide lucide-check">
                    <path d="M20 6 9 17l-5-5" />
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <div v-if="
            config.type !== 'sqlite' &&
            config.type !== 'duckdb' &&
            config.type !== 'libsql'
          " class="space-y-3 animate-in fade-in slide-in-from-top-4 duration-300">
            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <label class="text-sm font-medium leading-none" for="host">Host</label>
                <input v-model="config.host" id="host" type="text" placeholder="localhost"
                  :class="primaryFieldClass" />
                <p v-if="fieldErrors.host" class="text-xs text-destructive">{{ fieldErrors.host }}</p>
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium leading-none" for="port">Port</label>
                <input v-model.number="config.port" id="port" type="number"
                  :class="primaryFieldClass" />
                <p v-if="fieldErrors.port" class="text-xs text-destructive">{{ fieldErrors.port }}</p>
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <label class="text-sm font-medium leading-none" for="user">User</label>
                <input v-model="config.user" id="user" type="text"
                  :class="primaryFieldClass" />
                <p v-if="fieldErrors.user" class="text-xs text-destructive">{{ fieldErrors.user }}</p>
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium leading-none" for="password">Password</label>
                <div class="relative">
                  <input v-model="config.password" id="password" :type="showPassword ? 'text' : 'password'"
                    :class="primaryFieldWithIconClass" />
                  <button type="button" @click="togglePassword()" aria-label="Toggle password visibility"
                    class="absolute inset-y-0 right-0 flex items-center pr-3 text-muted-foreground hover:text-foreground">
                    <svg v-if="!showPassword" xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                      viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                      stroke-linejoin="round" class="lucide lucide-eye">
                      <path
                        d="M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0" />
                      <circle cx="12" cy="12" r="3" />
                    </svg>
                    <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                      fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                      class="lucide lucide-eye-off">
                      <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24" />
                      <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68" />
                      <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61" />
                      <line x1="2" x2="22" y1="2" y2="22" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
            <div class="space-y-2">
              <label class="text-sm font-medium leading-none" for="database">Database Name</label>
              <input v-model="config.database" id="database" type="text"
                :class="primaryFieldClass" />
              <p v-if="fieldErrors.database" class="text-xs text-destructive">{{ fieldErrors.database }}</p>
            </div>
          </div>

          <div v-else class="space-y-3 animate-in fade-in slide-in-from-top-4 duration-300">
            <div class="space-y-2">
              <label class="text-sm font-medium leading-none" for="filepath">Database File Path</label>
              <div class="flex items-center space-x-2">
                <input v-model="config.database" id="filepath" type="text" placeholder="/path/to/db.sqlite"
                  :class="primaryFieldClass" />
                <button type="button" @click="handleSelectSqliteFile"
                  :class="`${iconButtonFieldClass} h-11 px-4`">
                  Browse...
                </button>
              </div>
              <p v-if="fieldErrors.database" class="text-xs text-destructive">{{ fieldErrors.database }}</p>
            </div>
          </div>

          <div class="space-y-4 animate-in fade-in slide-in-from-top-4 duration-300"></div>

          <!-- SSH Tunnel Config -->
          <div v-if="supportsSsh"
            class="space-y-3 animate-in fade-in slide-in-from-top-4 duration-300 border-t pt-3 border-border">
            <div class="flex items-center space-x-2">
              <input type="checkbox" id="sshEnabled" v-model="config.sshEnabled"
                class="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary" />
              <label for="sshEnabled"
                class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                Use SSH Tunnel
              </label>
            </div>

            <div v-if="config.sshEnabled" class="space-y-3 pl-4 border-l-2 border-border/50 ml-1">
              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                  <label class="text-sm font-medium leading-none" for="sshHost">SSH Host</label>
                  <input v-model="config.sshHost" id="sshHost" type="text" placeholder="ssh.example.com"
                    :class="primaryFieldClass" />
                  <p v-if="fieldErrors.sshHost" class="text-xs text-destructive">{{ fieldErrors.sshHost }}</p>
                </div>
                <div class="space-y-2">
                  <label class="text-sm font-medium leading-none" for="sshPort">SSH Port</label>
                  <input v-model.number="config.sshPort" id="sshPort" type="number" placeholder="22"
                    :class="primaryFieldClass" />
                  <p v-if="fieldErrors.sshPort" class="text-xs text-destructive">{{ fieldErrors.sshPort }}</p>
                </div>
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium leading-none" for="sshUser">SSH User</label>
                <input v-model="config.sshUser" id="sshUser" type="text"
                  :class="primaryFieldClass" />
                <p v-if="fieldErrors.sshUser" class="text-xs text-destructive">{{ fieldErrors.sshUser }}</p>
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium leading-none" for="sshPassword">SSH Password</label>
                <div class="relative">
                  <input v-model="config.sshPassword" id="sshPassword" :type="showSshPassword ? 'text' : 'password'"
                    :class="primaryFieldWithIconClass" />
                  <button type="button" @click="toggleSshPassword()" aria-label="Toggle SSH password visibility"
                    class="absolute inset-y-0 right-0 flex items-center pr-3 text-muted-foreground hover:text-foreground">
                    <svg v-if="!showSshPassword" xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                      viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                      stroke-linejoin="round" class="lucide lucide-eye">
                      <path
                        d="M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0" />
                      <circle cx="12" cy="12" r="3" />
                    </svg>
                    <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                      fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                      class="lucide lucide-eye-off">
                      <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24" />
                      <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68" />
                      <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61" />
                      <line x1="2" x2="22" y1="2" y2="22" />
                    </svg>
                  </button>
                </div>
                <p v-if="fieldErrors.sshAuth" class="text-xs text-destructive">{{ fieldErrors.sshAuth }}</p>
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium leading-none" for="sshKeyFile">SSH Key File (Optional)</label>
                <input v-model="config.sshKeyFile" id="sshKeyFile" type="text" placeholder="/path/to/private_key"
                  :class="primaryFieldClass" />
              </div>
            </div>
          </div>

          <div class="flex items-center space-x-2 animate-in fade-in slide-in-from-top-4 duration-300">
            <input type="checkbox" id="readOnly" v-model="config.readOnly"
              class="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary" />
            <label for="readOnly"
              class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">Read
              Only Mode</label>
          </div>

          <div class="flex gap-2 mt-4">
            <button v-if="!isTesting" @click="connect"
              :disabled="isLoading"
              :class="[primaryPillButtonClass, 'h-11 flex-1', { 'opacity-50 cursor-not-allowed': isLoading }]">
              <span v-if="isLoading" class="mr-2">
                <svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
                  </path>
                </svg>
              </span>
              {{ isLoading ? "Connecting..." : "Connect" }}
            </button>
            <button v-if="!isLoading" @click="testConnection"
              :disabled="isLoading || isTesting"
              :class="[ghostPillButtonClass, 'h-11 flex-1', { 'opacity-50 cursor-not-allowed': isTesting }]">
              <span v-if="isTesting" class="mr-2">
                <svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
                  </path>
                </svg>
              </span>
              {{ isTesting ? "Testing..." : "Test Connection" }}
            </button>
            <button v-if="isLoading || isTesting" @click="cancelConnection"
              :class="`${destructivePillButtonClass} h-11 px-4`">
              Cancel
            </button>
            <button v-if="!isLoading && !isTesting" @click="toggleSavedModal(true)" aria-label="Open saved connections"
              :class="`${ghostPillButtonClass} h-11 w-11`">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                class="lucide lucide-clock">
                <circle cx="12" cy="12" r="10" />
                <polyline points="12 6 12 12 16 14" />
              </svg>
            </button>
          </div>

          <div v-if="error"
            class="bg-destructive/15 text-destructive text-sm p-3 rounded-md flex items-center gap-2 animate-in fade-in zoom-in duration-300">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
              stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
              class="lucide lucide-alert-circle">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" x2="12" y1="8" y2="12" />
              <line x1="12" x2="12.01" y1="16" y2="16" />
            </svg>
            <div class="flex-1">
              <p>{{ error }}</p>
              <div v-if="canTrustCurrentSshHost" class="mt-2 space-y-2">
                <button v-if="!sshHostKeyInfo" @click="loadCurrentSshHostKeyInfo" :disabled="isLoadingHostKeyInfo"
                  :class="`${destructivePillButtonClass} h-8 px-3 text-xs disabled:cursor-not-allowed`">
                  {{ isLoadingHostKeyInfo ? "Loading..." : "Show SSH Fingerprint" }}
                </button>
                <div v-else class="text-xs rounded-md border border-destructive/30 bg-destructive/5 p-2">
                  <p><span class="font-medium">Host:</span> {{ sshHostKeyInfo.pattern }}</p>
                  <p><span class="font-medium">Type:</span> {{ sshHostKeyInfo.keyType }}</p>
                  <p class="break-all"><span class="font-medium">Fingerprint:</span> {{ sshHostKeyInfo.fingerprint }}
                  </p>
                  <button @click="copyCurrentSshFingerprint"
                    :class="`${destructivePillButtonClass} mt-2 h-8 px-3 text-xs`">
                    Copy Fingerprint
                  </button>
                  <div class="mt-2 space-y-1">
                    <label for="expectedSshFingerprint" class="font-medium">Expected Fingerprint (Optional)</label>
                    <input id="expectedSshFingerprint" v-model="expectedSshFingerprint" type="text"
                      placeholder="SHA256:..."
                      class="w-full rounded-md border border-destructive/40 bg-background px-2 py-1 text-xs focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1" />
                    <p v-if="expectedSshFingerprint.trim()" :class="isFingerprintMatch ? 'text-green-600' : 'text-destructive'">
                      {{ isFingerprintMatch ? "Fingerprint matches expected value." : "Fingerprint does not match expected value." }}
                    </p>
                    <p v-if="pinnedSshFingerprint" class="break-all text-muted-foreground">
                      Pinned fingerprint: {{ pinnedSshFingerprint }}
                    </p>
                    <p v-if="isPinnedFingerprintMismatch" class="text-destructive">
                      Fingerprint differs from previously trusted host key for this host.
                    </p>
                    <div v-if="isPinnedFingerprintMismatch" class="mt-2 rounded-md border border-destructive/30 bg-background/70 p-2 space-y-1">
                      <label for="sshRotationReason" class="font-medium">Rotation Reason</label>
                      <input id="sshRotationReason" v-model="sshRotationReason" type="text"
                        placeholder="Explain why this host key changed"
                        class="w-full rounded-md border border-destructive/30 bg-background px-2 py-1 text-xs focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1" />
                      <label for="sshRotationConfirmText" class="font-medium">Type ROTATE to confirm</label>
                      <input id="sshRotationConfirmText" v-model="sshRotationConfirmText" type="text"
                        placeholder="ROTATE"
                        class="w-full rounded-md border border-destructive/30 bg-background px-2 py-1 text-xs focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1" />
                      <button @click="acceptPinnedFingerprintRotation"
                        :class="`${destructivePillButtonClass} h-8 px-3 text-xs`">
                        Accept Rotation
                      </button>
                    </div>
                  </div>
                  <button @click="trustCurrentSshHost" :disabled="isTrustingHost || isFingerprintMismatch || isPinnedFingerprintMismatch"
                    :class="`${destructivePillButtonClass} mt-2 h-8 px-3 text-xs disabled:cursor-not-allowed`">
                    {{ isTrustingHost ? "Trusting..." : "Trust SSH Host Key" }}
                  </button>
                </div>
                <div v-if="sshTrustAudit.length" class="mt-3 rounded-md border border-destructive/30 bg-background/70 p-2">
                  <div class="mb-1 flex items-center justify-between gap-2">
                    <p class="font-medium">Recent Trusted Hosts</p>
                    <div class="flex items-center gap-1">
                      <button @click="importSshTrustAudit"
                        :class="`${destructivePillButtonClass} h-7 px-2 text-[11px]`">
                        Import
                      </button>
                      <button @click="exportSshTrustAudit"
                        :class="`${destructivePillButtonClass} h-7 px-2 text-[11px]`">
                        Export
                      </button>
                      <button @click="clearSshTrustAudit"
                        :class="`${destructivePillButtonClass} h-7 px-2 text-[11px]`">
                        Clear
                      </button>
                    </div>
                  </div>
                  <input v-model="sshTrustAuditSearch" type="text" placeholder="Search host or fingerprint"
                    class="mb-1 w-full rounded-md border border-destructive/30 bg-background px-2 py-1 text-[11px] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1" />
                  <div v-for="entry in filteredSshTrustAudit.slice(0, 3)" :key="`${entry.pattern}-${entry.trustedAt}`"
                    class="py-1 border-t border-destructive/20 first:border-t-0">
                    <p class="break-all">{{ entry.pattern }}</p>
                    <p class="break-all text-[11px]">{{ entry.fingerprint }}</p>
                    <p class="text-[11px] text-muted-foreground">{{ new Date(entry.trustedAt).toLocaleString() }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-if="testSuccess"
            class="bg-green-500/15 text-green-600 text-sm p-3 rounded-md flex items-center gap-2 animate-in fade-in zoom-in duration-300">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
              stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
              class="lucide lucide-check-circle">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
              <polyline points="22 4 12 14.01 9 11.01" />
            </svg>
            {{ testSuccess }}
          </div>
        </div>
      </div>
    </div>

    <SavedConnectionsModal :isOpen="showSavedModal" :connections="savedConnections"
      :status-announcement="savedConnectionsAnnouncement"
      @close="toggleSavedModal(false)"
      @select="handleSelectConn" @edit="handleEditConn" @remove="handleRemoveConn" />

    <Toast ref="toastRef" />
    <SettingsDialog
      :isOpen="showSettings"
      :showSqlServerSettings="showSqlServerSettings"
      @close="toggleSettings(false)"
      @save="handleSettingsSave"
    />
  </div>
</template>

<style scoped>
/* Hide the native browser password reveal eye icon (especially in Edge and Chrome) */
input[type="password"]::-ms-reveal,
input[type="password"]::-ms-clear {
  display: none;
}

input[type="password"]::-webkit-contacts-auto-fill-button,
input[type="password"]::-webkit-credentials-auto-fill-button {
  visibility: hidden;
  position: absolute;
  right: 0;
}
</style>
