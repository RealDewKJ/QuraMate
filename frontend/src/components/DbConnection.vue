<template>
    <div class="flex flex-col items-center justify-center min-h-screen p-4 transition-colors duration-300"
        :class="{ 'dark': isDark }">
        <div class="absolute top-4 right-4">
            <button @click="toggleTheme"
                class="p-2 rounded-full hover:bg-accent hover:text-accent-foreground transition-colors">
                <svg v-if="isDark" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24"
                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-sun">
                    <circle cx="12" cy="12" r="4" />
                    <path d="M12 2v2" />
                    <path d="M12 20v2" />
                    <path d="m4.93 4.93 1.41 1.41" />
                    <path d="m17.66 17.66 1.41 1.41" />
                    <path d="M2 12h2" />
                    <path d="M20 12h2" />
                    <path d="m6.34 17.66-1.41-1.41" />
                    <path d="m19.07 4.93-1.41-1.41" />
                </svg>
                <svg v-else xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-moon">
                    <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z" />
                </svg>
            </button>
        </div>

        <div class="w-full max-w-md space-y-8 bg-card text-card-foreground p-8 rounded-xl border shadow-lg">
            <div class="text-center space-y-2">
                <div
                    class="h-12 w-12 bg-primary rounded-lg flex items-center justify-center text-primary-foreground mx-auto mb-4">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-database">
                        <ellipse cx="12" cy="5" rx="9" ry="3" />
                        <path d="M3 5V19A9 3 0 0 0 21 19V5" />
                        <path d="M3 12A9 3 0 0 0 21 12" />
                    </svg>
                </div>
                <h2 class="text-3xl font-bold tracking-tight">VaultDB</h2>
                <p class="text-muted-foreground text-sm">Connect to your database to start managing data.</p>
            </div>

            <div class="space-y-4">
                <div class="space-y-4 animate-in fade-in slide-in-from-top-4 duration-300">
                    <div class="space-y-2">
                        <label class="text-sm font-medium leading-none" for="connName">Connection Name
                            (Optional)</label>
                        <input v-model="config.name" id="connName" type="text" placeholder="My Database"
                            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                    </div>

                    <div class="space-y-2">
                        <label
                            class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                            for="dbType">Database Type</label>
                        <div class="relative">
                            <select v-model="config.type" id="dbType"
                                class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 appearance-none">
                                <option value="postgres">PostgreSQL</option>
                                <option value="mysql">MySQL</option>
                                <option value="mssql">MSSQL</option>
                                <option value="sqlite">SQLite</option>
                            </select>
                            <div
                                class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-input-foreground">
                                <svg class="h-4 w-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"
                                    fill="currentColor">
                                    <path fill-rule="evenodd"
                                        d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                                        clip-rule="evenodd" />
                                </svg>
                            </div>
                        </div>
                    </div>

                    <div v-if="config.type !== 'sqlite'"
                        class="space-y-4 animate-in fade-in slide-in-from-top-4 duration-300">
                        <div class="grid grid-cols-2 gap-4">
                            <div class="space-y-2">
                                <label class="text-sm font-medium leading-none" for="host">Host</label>
                                <input v-model="config.host" id="host" type="text" placeholder="localhost"
                                    class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                            </div>
                            <div class="space-y-2">
                                <label class="text-sm font-medium leading-none" for="port">Port</label>
                                <input v-model.number="config.port" id="port" type="number"
                                    class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                            </div>
                        </div>
                        <div class="space-y-2">
                            <label class="text-sm font-medium leading-none" for="user">User</label>
                            <input v-model="config.user" id="user" type="text"
                                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                        </div>
                        <div class="space-y-2">
                            <label class="text-sm font-medium leading-none" for="password">Password</label>
                            <input v-model="config.password" id="password" type="password"
                                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                        </div>
                        <div class="space-y-2">
                            <label class="text-sm font-medium leading-none" for="database">Database Name</label>
                            <input v-model="config.database" id="database" type="text"
                                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                        </div>
                    </div>

                    <div v-else class="space-y-4 animate-in fade-in slide-in-from-top-4 duration-300">
                        <div class="space-y-2">
                            <label class="text-sm font-medium leading-none" for="filepath">Database File Path</label>
                            <input v-model="config.database" id="filepath" type="text" placeholder="/path/to/db.sqlite"
                                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                        </div>
                    </div>

                    <div class="flex gap-2 mt-6">
                        <button @click="connect" :class="{ 'opacity-50 cursor-not-allowed': isLoading }"
                            :disabled="isLoading"
                            class="flex-1 inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-10 px-4 py-2">
                            <span v-if="isLoading" class="mr-2">
                                <svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none"
                                    viewBox="0 0 24 24">
                                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor"
                                        stroke-width="4">
                                    </circle>
                                    <path class="opacity-75" fill="currentColor"
                                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
                                    </path>
                                </svg>
                            </span>
                            {{ isLoading ? 'Connecting...' : 'Connect' }}
                        </button>
                        <button @click="testConnection" :class="{ 'opacity-50 cursor-not-allowed': isTesting }"
                            :disabled="isLoading || isTesting"
                            class="flex-1 inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2">
                            <span v-if="isTesting" class="mr-2">
                                <svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none"
                                    viewBox="0 0 24 24">
                                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor"
                                        stroke-width="4">
                                    </circle>
                                    <path class="opacity-75" fill="currentColor"
                                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
                                    </path>
                                </svg>
                            </span>
                            {{ isTesting ? 'Testing...' : 'Test Connection' }}
                        </button>
                        <button @click="showSavedModal = true"
                            class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 w-10">
                            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-history">
                                <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74-2.74L3 12" />
                                <path d="M3 3v9h9" />
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
                        {{ error }}
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

            <!-- Saved Connections Modal -->
            <div v-if="showSavedModal"
                class="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm transition-all duration-100 animate-in fade-in">
                <div
                    class="fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-6 shadow-lg duration-200 sm:rounded-lg md:w-full animate-in fade-in zoom-in-95 slide-in-from-left-1/2 slide-in-from-top-48">
                    <div class="flex flex-col space-y-1.5 text-center sm:text-left">
                        <h2 class="text-lg font-semibold leading-none tracking-tight">Saved Connections</h2>
                        <p class="text-sm text-muted-foreground">Select a connection to load its details.</p>
                    </div>
                    <div class="grid gap-4 py-4 max-h-[60vh] overflow-y-auto">
                        <div v-if="savedConnections.length === 0"
                            class="text-center text-muted-foreground text-sm py-8">
                            No saved connections found.
                        </div>
                        <div v-else class="space-y-2">
                            <div v-for="(conn, index) in savedConnections" :key="index"
                                class="flex items-center justify-between p-3 rounded-lg border bg-card hover:bg-accent hover:text-accent-foreground transition-colors cursor-pointer group"
                                @click="selectConnection(conn)">
                                <div class="flex items-center gap-3 overflow-hidden">
                                    <div
                                        class="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center text-primary">
                                        <svg v-if="conn.type === 'postgres'" xmlns="http://www.w3.org/2000/svg"
                                            width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                            stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                                            class="lucide lucide-database">
                                            <ellipse cx="12" cy="5" rx="9" ry="3" />
                                            <path d="M3 5V19A9 3 0 0 0 21 19V5" />
                                            <path d="M3 12A9 3 0 0 0 21 12" />
                                        </svg>
                                        <svg v-else-if="conn.type === 'mysql'" xmlns="http://www.w3.org/2000/svg"
                                            width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                            stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                                            class="lucide lucide-database">
                                            <ellipse cx="12" cy="5" rx="9" ry="3" />
                                            <path d="M3 5V19A9 3 0 0 0 21 19V5" />
                                            <path d="M3 12A9 3 0 0 0 21 12" />
                                        </svg>
                                        <svg v-else-if="conn.type === 'sqlite'" xmlns="http://www.w3.org/2000/svg"
                                            width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                            stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                                            class="lucide lucide-file-code">
                                            <path d="M10 12.5 8 15l2 2.5" />
                                            <path d="m14 12.5 2 2.5-2 2.5" />
                                            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                                            <path d="M14 2v6h6" />
                                        </svg>
                                        <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round"
                                            class="lucide lucide-database">
                                            <ellipse cx="12" cy="5" rx="9" ry="3" />
                                            <path d="M3 5V19A9 3 0 0 0 21 19V5" />
                                            <path d="M3 12A9 3 0 0 0 21 12" />
                                        </svg>
                                    </div>
                                    <div class="flex flex-col truncate text-left">
                                        <span class="text-sm font-medium truncate">{{ getConnectionLabel(conn) }}</span>
                                        <span class="text-xs text-muted-foreground truncate">{{ conn.host }}:{{
                                            conn.port
                                            }}</span>
                                    </div>
                                </div>
                                <button @click.stop="removeConnection(index)"
                                    class="p-2 rounded-md hover:bg-destructive hover:text-destructive-foreground transition-colors opacity-0 group-hover:opacity-100">
                                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                        stroke-linejoin="round" class="lucide lucide-trash-2">
                                        <path d="M3 6h18" />
                                        <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                                        <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
                                        <line x1="10" x2="10" y1="11" y2="17" />
                                        <line x1="14" x2="14" y1="11" y2="17" />
                                    </svg>
                                </button>
                            </div>
                        </div>
                    </div>
                    <div class="flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2">
                        <button @click="showSavedModal = false"
                            class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2">
                            Cancel
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive, watch, onMounted, computed } from 'vue';
import { ConnectDB, TestConnection } from '../../wailsjs/go/main/App';

const props = defineProps<{
    activeConnections: any[]
}>();

const emit = defineEmits(['connected', 'connection-exists']);

const config = reactive({
    name: '',
    type: 'postgres',
    host: 'localhost',
    port: 5432,
    user: 'postgres',
    password: '',
    database: 'postgres'
});

const error = ref('');
const testSuccess = ref('');
const isLoading = ref(false);
const isTesting = ref(false);
const isDark = ref(false);
const showSavedModal = ref(false);
const savedConnections = ref<any[]>([]);

const toggleTheme = () => {
    isDark.value = !isDark.value;
    if (isDark.value) {
        document.documentElement.classList.add('dark');
    } else {
        document.documentElement.classList.remove('dark');
    }
};

watch(() => config.type, (newType) => {
    if (newType === 'postgres') config.port = 5432;
    else if (newType === 'mysql') config.port = 3306;
    else if (newType === 'mssql') config.port = 1433;
});

const connect = async () => {
    error.value = '';
    testSuccess.value = '';
    isLoading.value = true;

    // Check for existing connection
    const existing = props.activeConnections.find(c =>
        c.config.type === config.type &&
        c.config.host === config.host &&
        c.config.port === config.port &&
        c.config.user === config.user &&
        c.config.database === config.database
    );

    if (existing) {
        emit('connection-exists', existing.id);
        isLoading.value = false;
        return;
    }

    try {
        const result = await ConnectDB(config);
        if (result.id) {
            saveConnection(JSON.parse(JSON.stringify(config)));
            emit('connected', {
                id: result.id,
                name: config.name || getConnectionLabel(config),
                config: { ...config }
            });
        } else {
            error.value = result.error || 'Unknown error';
        }
    } catch (e: any) {
        error.value = e.toString();
    } finally {
        isLoading.value = false;
    }
};

const testConnection = async () => {
    error.value = '';
    testSuccess.value = '';
    isTesting.value = true;
    try {
        const result = await TestConnection(JSON.parse(JSON.stringify(config)));
        if (result === 'Success') {
            testSuccess.value = 'Connection successful!';
        } else {
            error.value = result;
        }
    } catch (e: any) {
        error.value = e.toString();
    } finally {
        isTesting.value = false;
    }
};

const saveConnection = (newConfig: any) => {
    const exists = savedConnections.value.some(c =>
        c.type === newConfig.type &&
        c.host === newConfig.host &&
        c.port === newConfig.port &&
        c.user === newConfig.user &&
        c.database === newConfig.database &&
        c.name === newConfig.name
    );

    if (!exists) {
        savedConnections.value.push(newConfig);
        localStorage.setItem('savedConnections', JSON.stringify(savedConnections.value));
    }
};

const removeConnection = (index: number) => {
    savedConnections.value.splice(index, 1);
    localStorage.setItem('savedConnections', JSON.stringify(savedConnections.value));
};

const selectConnection = (conn: any) => {
    config.name = conn.name || '';
    Object.assign(config, conn);
    showSavedModal.value = false;
};

const getConnectionLabel = (conn: any) => {
    if (conn.name) return conn.name;
    if (conn.type === 'sqlite') return `SQLite: ${conn.database}`;
    return `${conn.user}@${conn.host}:${conn.port}/${conn.database} (${conn.type})`;
};

onMounted(() => {
    // Check system preference
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
        // Default to dark if system is dark, or let user toggle
        // isDark.value = true;
        // document.documentElement.classList.add('dark');
    }

    const saved = localStorage.getItem('savedConnections');
    if (saved) {
        try {
            savedConnections.value = JSON.parse(saved);
        } catch (e) {
            console.error('Failed to load saved connections', e);
        }
    }
});
</script>
