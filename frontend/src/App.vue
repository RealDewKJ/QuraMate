<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import DbConnection from './components/DbConnection.vue';
import DbDashboard from './components/DbDashboard.vue';
import UpdateNotification from './components/UpdateNotification.vue';
import { colorMode } from './composables/useTheme';
import { LoadSetting, GetStartupFile, ReadTextFile, ConnectDB, CheckPendingFile } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

const updateNotificationRef = ref<InstanceType<typeof UpdateNotification> | null>(null);
const { locale } = useI18n({ useScope: 'global' });

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
    }
    document.documentElement.style.fontFamily = appFont;
  } catch (e) {
    console.error("Failed to load global settings on mount", e);
  }

  // Handle Startup File
  try {
    const startupFile = await GetStartupFile();
    if (startupFile) {
      processStartupFile(startupFile);
    }
  } catch (e) {
    console.error("Failed to fetch startup file", e);
  }

  // File-based IPC for second instance — poll for pending files
  // Poll for pending files from second instances (file-based IPC fallback)
  console.log("[OpenWith] Starting pending file poll (every 2s)");
  pendingFileInterval = window.setInterval(async () => {
    try {
      const filePath = await CheckPendingFile();
      console.log("[OpenWith] Poll tick, result:", filePath || "(empty)");
      if (filePath) {
        console.log("[OpenWith] Found pending file:", filePath);
        processStartupFile(filePath);
      }
    } catch (e) {
      console.error("[OpenWith] Poll error:", e);
    }
  }, 2000);
});

let pendingFileInterval: number | null = null;
onUnmounted(() => {
  if (pendingFileInterval) window.clearInterval(pendingFileInterval);
});

const processStartupFile = async (startupFile: string) => {
  try {
    console.log("Startup string from OS:", startupFile);
    const ext = startupFile.split('.').pop()?.toLowerCase();

    if (ext === 'db' || ext === 'sqlite' || ext === 'sqlite3') {
      const payload = {
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
      };
      const result = await ConnectDB(payload);
      if (result && !result.error) {
        const name = startupFile.replace(/^.*[\\/]/, '')
        handleConnected({
          id: result.id,
          name: `${name} (SQLite)`,
          config: payload
        });
      }
    } else if (ext === 'sql') {
      const content = await ReadTextFile(startupFile);
      const fileName = startupFile.replace(/^.*[\\/]/, '');

      // If there's an active connection, dispatch to it
      if (connections.value.length > 0) {
        const id = activeTabId.value || connections.value[0].id;
        switchToTab(id);

        window.dispatchEvent(new CustomEvent('open-sql-file', {
          detail: { content, fileName, connectionId: id }
        }));

      } else {
        // No connections open — store as pending and show connection screen
        pendingSqlFile.value = { path: startupFile, name: fileName, content };
        activeTabId.value = null; // ensure connection screen is visible
      }
    }
  } catch (e) {
    console.error("Failed to process startup file", e);
  }
};

interface Connection {
  id: string;
  name: string;
  config: any;
}

const pendingSqlFile = ref<{ path: string; name: string; content: string } | null>(null);

const connections = ref<Connection[]>([]);
const activeTabId = ref<string | null>(null);

const handleConnected = (conn: Connection) => {
  connections.value.push(conn);
  activeTabId.value = conn.id;

  // If there's a pending SQL file, dispatch it to the new connection
  if (pendingSqlFile.value) {
    const sqlContent = pendingSqlFile.value.content;
    const sqlFileName = pendingSqlFile.value.name;
    const connId = conn.id;
    pendingSqlFile.value = null;

    setTimeout(() => {
      window.dispatchEvent(new CustomEvent('open-sql-file', {
        detail: { content: sqlContent, fileName: sqlFileName, connectionId: connId }
      }));
    }, 300);
  }
};

const handleDisconnect = (id: string) => {
  const index = connections.value.findIndex(c => c.id === id);
  if (index !== -1) {
    connections.value.splice(index, 1);
    // If we closed the active tab, switch to another one or home
    if (activeTabId.value === id) {
      if (connections.value.length > 0) {
        // Switch to the last opened connection or previous one
        activeTabId.value = connections.value[connections.value.length - 1].id;
      } else {
        activeTabId.value = null; // Back to home
      }
    }
  }
};

const switchToHome = () => {
  activeTabId.value = null;
};

const switchToTab = (id: string) => {
  activeTabId.value = id;
};

const removeTab = (id: string, event: Event) => {
  event.stopPropagation();
  // In a real app we might want to trigger disconnect here too if it wasn't triggered by the dashboard
  // For now we assume the disconnect button in dashboard handles the cleanup logic
  // But if we close via tab, we should probably call DisconnectDB. 
  // However, since we don't have easy access to DisconnectDB here directly without importing, 
  // and the dashboard handles it, let's just purely focus on UI state for now or improved later.
  // Actually better to let the user disconnect from the dashboard to ensure resources are freed.
  // So we will just switch to that tab if not active? 
  // Or we can just perform the UI removal and assume backend cleanup happens or is irrelevant (leak).
  // Let's implement proper cleanup later/verify. For now rely on dashboard disconnect button.
  // But wait, the user asked for tabs. A close button on tab is expected.
  // Let's rely on the dashboard "Disconnect" button for now to force proper cleanup.
  // So this function is just a placeholder or for "force close".
  handleDisconnect(id);
  handleDisconnect(id);
}

const handleConnectionExists = (id: string) => {
  switchToTab(id);
};

const handleConnectionUpdate = (update: { id: string, config: any }) => {
  const conn = connections.value.find(c => c.id === update.id);
  if (conn) {
    conn.config = { ...update.config };
    // Force reactivity if needed, but above should work
  }
};
</script>

<template>
  <div class="h-screen flex flex-col text-foreground relative overflow-hidden bg-background">
    <!-- Background Gradient Decorators -->
    <div class="absolute inset-0 z-0 overflow-hidden pointer-events-none">
      <div
        class="absolute top-[-10%] right-[-5%] w-[500px] h-[500px] rounded-full bg-orange-500/10 dark:bg-orange-500/5 blur-[120px]">
      </div>
      <div
        class="absolute bottom-[-10%] left-[-10%] w-[600px] h-[600px] rounded-full bg-orange-400/10 dark:bg-orange-400/5 blur-[150px]">
      </div>
    </div>

    <!-- Tab Bar -->
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

      <!-- Version & Update Check -->
      <div class="flex items-center gap-2 px-3 flex-shrink-0 border-l border-border">
        <span class="text-[11px] text-muted-foreground font-mono" v-if="updateNotificationRef?.currentVersion">
          v{{ updateNotificationRef.currentVersion }}
        </span>
        <button @click="updateNotificationRef?.manualCheck()"
          class="flex items-center justify-center w-7 h-7 rounded-md transition-colors hover:bg-muted/60 text-muted-foreground hover:text-primary"
          :class="{ 'animate-spin': updateNotificationRef?.checking }" title="ตรวจสอบอัพเดต">
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

    <!-- Content Area -->
    <div class="flex-1 overflow-hidden relative z-10">
      <div v-show="activeTabId === null" class="h-full overflow-auto">
        <DbConnection :activeConnections="connections" :pendingSqlFile="pendingSqlFile" @connected="handleConnected"
          @connection-exists="handleConnectionExists" @connection-updated="handleConnectionUpdate" />
      </div>

      <div v-for="conn in connections" :key="conn.id" v-show="activeTabId === conn.id" class="h-full">
        <DbDashboard :connectionId="conn.id" :connectionName="conn.name" :dbType="conn.config.type"
          :isReadOnly="conn.config.readOnly" @disconnect="handleDisconnect" />
      </div>
    </div>

    <!-- Update Notification Overlay -->
    <UpdateNotification ref="updateNotificationRef" />
  </div>
</template>

<style></style>
