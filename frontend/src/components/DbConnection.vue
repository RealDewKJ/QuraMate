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
                <div class="flex justify-center mb-4">
                    <div
                        class="h-12 w-12 bg-primary rounded-lg flex items-center justify-center text-primary-foreground">
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-database">
                            <ellipse cx="12" cy="5" rx="9" ry="3" />
                            <path d="M3 5V19A9 3 0 0 0 21 19V5" />
                            <path d="M3 12A9 3 0 0 0 21 12" />
                        </svg>
                    </div>
                </div>
                <h2 class="text-3xl font-bold tracking-tight">VaultDB</h2>
                <p class="text-muted-foreground text-sm">Connect to your database to start managing data.</p>
            </div>

            <div class="space-y-4">
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

                <button @click="connect" :class="{ 'opacity-50 cursor-not-allowed': isLoading }" :disabled="isLoading"
                    class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-10 px-4 py-2 w-full mt-6">
                    <span v-if="isLoading" class="mr-2">
                        <svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none"
                            viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4">
                            </circle>
                            <path class="opacity-75" fill="currentColor"
                                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
                            </path>
                        </svg>
                    </span>
                    {{ isLoading ? 'Connecting...' : 'Connect' }}
                </button>

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
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive, watch, onMounted } from 'vue';
import { ConnectDB } from '../../wailsjs/go/main/App';

const emit = defineEmits(['connected']);

const config = reactive({
    type: 'postgres',
    host: 'localhost',
    port: 5432,
    user: 'postgres',
    password: '',
    database: 'postgres'
});

const error = ref('');
const isLoading = ref(false);
const isDark = ref(false);

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
    isLoading.value = true;
    try {
        const result = await ConnectDB(config);
        if (result === 'Success') {
            emit('connected');
        } else {
            error.value = result;
        }
    } catch (e: any) {
        error.value = e.toString();
    } finally {
        isLoading.value = false;
    }
};

onMounted(() => {
    // Check system preference
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
        // Default to dark if system is dark, or let user toggle
        // isDark.value = true;
        // document.documentElement.classList.add('dark');
    }
});
</script>
