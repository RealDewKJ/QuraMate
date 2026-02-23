<template>
    <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center">
        <!-- Overlay -->
        <div class="fixed inset-0 bg-background/80 backdrop-blur-sm transition-opacity" @click="close"></div>

        <!-- Dialog -->
        <div
            class="relative z-50 flex w-full max-w-[1000px] h-[700px] flex-col rounded-xl border border-border bg-card text-card-foreground shadow-lg overflow-hidden animate-in fade-in zoom-in-95">
            <!-- Header -->
            <div class="flex flex-col space-y-1.5 p-6 border-b border-border">
                <h2 class="text-2xl font-semibold leading-none tracking-tight">Settings</h2>
                <p class="text-sm text-muted-foreground">
                    Manage your app preferences, appearance, and connections.
                </p>
                <button @click="close"
                    class="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground">
                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-x">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                    <span class="sr-only">Close</span>
                </button>
            </div>

            <!-- Content Layout -->
            <div class="flex flex-1 overflow-hidden">
                <!-- Sidebar Navigation -->
                <div class="w-48 border-r border-border bg-muted/20 p-4 shrink-0 overflow-y-auto">
                    <nav class="flex flex-col space-y-1">
                        <button v-for="tab in tabs" :key="tab.id" @click="activeTab = tab.id"
                            class="flex items-center gap-2 px-3 py-2 text-sm font-medium rounded-md text-left transition-colors"
                            :class="activeTab === tab.id ? 'bg-secondary text-secondary-foreground' : 'text-muted-foreground hover:bg-muted hover:text-foreground'">
                            <component :is="getIcon(tab.icon)" class="h-4 w-4" />
                            {{ tab.label }}
                        </button>
                    </nav>
                </div>

                <!-- Tab Content -->
                <div class="flex-1 overflow-y-auto p-6 bg-background">
                    <!-- General Tab -->
                    <div v-if="activeTab === 'general'" class="space-y-6">
                        <div>
                            <h3 class="text-lg font-medium">General</h3>
                            <p class="text-sm text-muted-foreground mb-4">Basic application settings.</p>

                            <div class="space-y-4">
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        Data Editor Language
                                    </label>
                                    <p class="text-xs text-muted-foreground">Select the language for the data editor
                                        interface.</p>
                                    <select v-model="settings.general.language"
                                        class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 max-w-sm">
                                        <option value="en">English (Default)</option>
                                        <option value="th">Thai (Coming Soon)</option>
                                    </select>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Appearance Tab -->
                    <div v-if="activeTab === 'appearance'" class="space-y-6">
                        <div>
                            <h3 class="text-lg font-medium">Appearance</h3>
                            <p class="text-sm text-muted-foreground mb-4">Customize the look and feel of QuraMate.</p>

                            <div class="space-y-4">
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        Theme
                                    </label>
                                    <div class="grid grid-cols-3 gap-4 pt-2">
                                        <button @click="setTheme('light')"
                                            class="flex flex-col items-center gap-2 rounded-lg border-2 p-2 hover:bg-accent cursor-pointer transition-all"
                                            :class="settings.appearance.theme === 'light' ? 'border-primary' : 'border-transparent'">
                                            <div
                                                class="items-center rounded-md border-2 border-muted p-1 bg-[#ecedef] w-full">
                                                <div class="space-y-2 rounded-sm bg-[#ffffff] p-2">
                                                    <div class="space-y-2 rounded-md bg-[#ecedef] p-2 shadow-sm">
                                                        <div class="h-2 w-20 rounded-lg bg-[#ffffff]"></div>
                                                        <div class="h-2 w-full rounded-lg bg-[#ffffff]"></div>
                                                    </div>
                                                    <div
                                                        class="flex items-center space-x-2 rounded-md bg-[#ecedef] p-2 shadow-sm">
                                                        <div class="h-4 w-4 rounded-full bg-[#ffffff]"></div>
                                                        <div class="h-2 w-full rounded-lg bg-[#ffffff]"></div>
                                                    </div>
                                                </div>
                                            </div>
                                            <span class="text-sm font-medium">Light</span>
                                        </button>

                                        <button @click="setTheme('dark')"
                                            class="flex flex-col items-center gap-2 rounded-lg border-2 p-2 hover:bg-accent cursor-pointer transition-all"
                                            :class="settings.appearance.theme === 'dark' ? 'border-primary' : 'border-transparent'">
                                            <div
                                                class="items-center rounded-md border-2 border-muted p-1 bg-slate-950 w-full">
                                                <div class="space-y-2 rounded-sm bg-slate-800 p-2">
                                                    <div class="space-y-2 rounded-md bg-slate-950 p-2 shadow-sm">
                                                        <div class="h-2 w-20 rounded-lg bg-slate-800"></div>
                                                        <div class="h-2 w-full rounded-lg bg-slate-800"></div>
                                                    </div>
                                                    <div
                                                        class="flex items-center space-x-2 rounded-md bg-slate-950 p-2 shadow-sm">
                                                        <div class="h-4 w-4 rounded-full bg-slate-800"></div>
                                                        <div class="h-2 w-full rounded-lg bg-slate-800"></div>
                                                    </div>
                                                </div>
                                            </div>
                                            <span class="text-sm font-medium">Dark</span>
                                        </button>

                                        <button @click="setTheme('system')"
                                            class="flex flex-col items-center gap-2 rounded-lg border-2 p-2 hover:bg-accent cursor-pointer transition-all"
                                            :class="settings.appearance.theme === 'system' ? 'border-primary' : 'border-transparent'">
                                            <div
                                                class="items-center rounded-md border-2 border-muted p-1 bg-gradient-to-br from-[#ecedef] from-50% to-slate-950 to-50% w-full">
                                                <div
                                                    class="space-y-2 rounded-sm bg-background p-2 border border-border">
                                                    <div class="space-y-2 rounded-md bg-muted p-2 shadow-sm">
                                                        <div class="h-2 w-20 rounded-lg bg-background"></div>
                                                        <div class="h-2 w-full rounded-lg bg-background"></div>
                                                    </div>
                                                    <div
                                                        class="flex items-center space-x-2 rounded-md bg-muted p-2 shadow-sm">
                                                        <div class="h-4 w-4 rounded-full bg-background"></div>
                                                        <div class="h-2 w-full rounded-lg bg-background"></div>
                                                    </div>
                                                </div>
                                            </div>
                                            <span class="text-sm font-medium">System</span>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- SQL Editor Tab -->
                    <div v-if="activeTab === 'editor'" class="space-y-6">
                        <div>
                            <h3 class="text-lg font-medium">SQL Editor</h3>
                            <p class="text-sm text-muted-foreground mb-4">Configure the SQL query editor environment.
                            </p>

                            <div class="space-y-4">
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        Font Family
                                    </label>
                                    <select v-model="settings.editor.fontFamily"
                                        class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 max-w-sm">
                                        <option value="'JetBrains Mono', monospace">JetBrains Mono</option>
                                        <option value="'Fira Code', monospace">Fira Code</option>
                                        <option value="'Cascadia Code', monospace">Cascadia Code</option>
                                        <option value="Consolas, monospace">Consolas</option>
                                        <option value="Courier New, monospace">Courier New</option>
                                    </select>
                                </div>

                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        Font Size
                                    </label>
                                    <div class="flex items-center gap-4 max-w-sm">
                                        <input type="range" v-model="settings.editor.fontSize" min="10" max="24"
                                            class="w-full accent-primary">
                                        <span class="text-sm font-mono w-12 text-right">{{ settings.editor.fontSize
                                        }}px</span>
                                    </div>
                                </div>

                                <!-- Preview Box -->
                                <div class="mt-6 border border-border rounded-md overflow-hidden bg-card">
                                    <div
                                        class="bg-muted px-3 py-1.5 border-b border-border text-xs font-medium text-muted-foreground">
                                        Preview</div>
                                    <div class="p-4 bg-[var(--vscode-editor-background,#1e1e1e)] overflow-hidden">
                                        <pre :style="{ fontFamily: settings.editor.fontFamily, fontSize: `${settings.editor.fontSize}px` }"
                                            class="text-[var(--vscode-editor-foreground,#d4d4d4)]"><span class="text-[#569cd6]">SELECT</span>
  id,
  username,
  created_at
<span class="text-[#569cd6]">FROM</span>
  users
<span class="text-[#569cd6]">WHERE</span>
  status = <span class="text-[#ce9178]">'active'</span>;</pre>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- AI Tab -->
                    <div v-if="activeTab === 'ai'" class="space-y-6">
                        <div>
                            <h3 class="text-lg font-medium">Artificial Intelligence</h3>
                            <p class="text-sm text-muted-foreground mb-4">Configure AI providers for QuraMate's smart
                                features.</p>

                            <div class="p-4 mb-4 text-sm text-amber-800 rounded-lg bg-amber-50 dark:bg-amber-950/30 dark:text-amber-400 flex items-start gap-3 border border-amber-200 dark:border-amber-900/50"
                                role="alert">
                                <svg class="flex-shrink-0 inline w-4 h-4 mt-0.5" aria-hidden="true"
                                    xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
                                    <path
                                        d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5ZM9.5 4a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3ZM12 15H8a1 1 0 0 1 0-2h1v-3H8a1 1 0 0 1 0-2h2a1 1 0 0 1 1 1v4h1a1 1 0 0 1 0 2Z" />
                                </svg>
                                <span class="sr-only">Info</span>
                                <div>
                                    <span class="font-medium">Coming Soon!</span> AI features are currently in
                                    development. You can save your keys now for when the feature goes live.
                                </div>
                            </div>

                            <div class="space-y-4">
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        AI Provider
                                    </label>
                                    <select v-model="settings.ai.provider"
                                        class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 max-w-sm">
                                        <option value="openai">OpenAI (ChatGPT)</option>
                                        <option value="anthropic">Anthropic (Claude)</option>
                                        <option value="google">Google (Gemini)</option>
                                        <option value="local">Local (Ollama / Llama.cpp)</option>
                                    </select>
                                </div>

                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        API Key
                                    </label>
                                    <div class="relative max-w-md">
                                        <input :type="showAiKey ? 'text' : 'password'" v-model="settings.ai.apiKey"
                                            placeholder="sk-..."
                                            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 pr-10">
                                        <button @click="showAiKey = !showAiKey"
                                            class="absolute right-0 top-0 h-full px-3 text-muted-foreground hover:text-foreground">
                                            <svg v-if="!showAiKey" xmlns="http://www.w3.org/2000/svg" width="16"
                                                height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                                stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                                                class="lucide lucide-eye">
                                                <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" />
                                                <circle cx="12" cy="12" r="3" />
                                            </svg>
                                            <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                                                viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                                stroke-linecap="round" stroke-linejoin="round"
                                                class="lucide lucide-eye-off">
                                                <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24" />
                                                <path
                                                    d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68" />
                                                <path
                                                    d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61" />
                                                <line x1="2" x2="22" y1="2" y2="22" />
                                            </svg>
                                        </button>
                                    </div>
                                    <p class="text-[10px] text-muted-foreground">Your API key is stored locally and
                                        securely.</p>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Logs Tab -->
                    <div v-if="activeTab === 'logs'" class="space-y-6 flex flex-col h-full">
                        <div>
                            <h3 class="text-lg font-medium">Application Logs</h3>
                            <p class="text-sm text-muted-foreground">View local logs for debugging and system
                                monitoring.</p>
                        </div>

                        <div
                            class="flex-1 min-h-[300px] border border-border rounded-md bg-[#1e1e1e] text-[#cccccc] font-mono text-xs overflow-auto flex flex-col">
                            <div
                                class="bg-[#2d2d2d] border-b border-[#3c3c3c] px-3 py-1.5 flex justify-between items-center shrink-0">
                                <span class="font-medium text-[10px] text-[#858585]">quramate.log</span>
                                <div class="flex gap-2">
                                    <button class="hover:text-white transition-colors" title="Refresh Logs"
                                        @click="refreshLogs">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round"
                                            class="lucide lucide-refresh-cw"
                                            :class="{ 'animate-spin': isRefreshingLogs }">
                                            <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
                                            <path d="M3 3v5h5" />
                                        </svg>
                                    </button>
                                    <button class="hover:text-white transition-colors" title="Clear View">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round"
                                            class="lucide lucide-trash-2">
                                            <path d="M3 6h18" />
                                            <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                                            <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
                                            <line x1="10" x2="10" y1="11" y2="17" />
                                            <line x1="14" x2="14" y1="11" y2="17" />
                                        </svg>
                                    </button>
                                </div>
                            </div>
                            <div class="p-3 overflow-auto flex-1 space-y-1">
                                <!-- Placeholder mock logs -->
                                <div class="text-[#858585]">[2024-03-20 10:15:32] <span
                                        class="text-[#ce9178]">INFO</span> Application started. Version {{ appVersion }}
                                </div>
                                <div class="text-[#858585]">[2024-03-20 10:15:33] <span
                                        class="text-[#ce9178]">INFO</span> Initializing local storage and settings...
                                </div>
                                <div class="text-[#858585]">[2024-03-20 10:16:01] <span class="text-[#b5cea8]">DB</span>
                                    Successfully connected to PostgreSQL database at localhost:5432.</div>
                                <div class="text-[#858585]">[2024-03-20 10:20:15] <span
                                        class="text-[#569cd6]">QERY</span> Executing query: SELECT * FROM users LIMIT
                                    100...</div>
                                <div class="text-[#858585]">[2024-03-20 10:20:15] <span
                                        class="text-[#569cd6]">QERY</span> Query returned 100 rows in 24ms.</div>
                                <div v-if="hasMockError" class="text-[#858585]">[2024-03-20 10:25:40] <span
                                        class="text-[#f44747]">ERROR</span> Connection timeout trying to reach AI
                                    Provider API.</div>
                            </div>
                        </div>

                        <div class="flex justify-between items-center text-xs text-muted-foreground mt-2">
                            <span>Logs are stored locally on your machine.</span>
                            <button class="text-primary hover:underline">Open Log Folder</button>
                        </div>
                    </div>

                    <!-- Info Tab -->
                    <div v-if="activeTab === 'info'" class="space-y-6">
                        <div>
                            <h3 class="text-lg font-medium">About QuraMate</h3>
                            <p class="text-sm text-muted-foreground mb-6">System information and credits.</p>

                            <div
                                class="flex flex-col items-center justify-center p-8 border border-border rounded-xl bg-card shadow-sm text-center">
                                <div
                                    class="w-24 h-24 rounded-2xl mb-6 flex items-center justify-center shadow-lg shadow-black/5 rotate-3 transition-transform hover:rotate-6 overflow-hidden bg-white/5 p-1 border border-border">
                                    <img src="../assets/images/new-icon.png" alt="QuraMate Logo"
                                        class="w-full h-full object-contain" />
                                </div>

                                <h1 class="text-3xl font-bold tracking-tight mb-2">QuraMate</h1>
                                <p
                                    class="text-sm font-mono bg-muted/50 px-3 py-1 rounded-full text-foreground/80 mb-6 border border-border">
                                    Version {{ appVersion }}</p>

                                <p class="text-muted-foreground max-w-sm mb-8 text-sm leading-relaxed">
                                    A modern, lightweight database management tool designed for developers. Built with
                                    Vue 3, Tailwind CSS, and Go.
                                </p>

                                <div class="grid grid-cols-2 gap-4 w-full">
                                    <div
                                        class="flex flex-col items-center p-3 rounded-lg bg-muted/30 border border-border border-dashed">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round"
                                            class="lucide lucide-github mb-2 text-foreground/70">
                                            <path
                                                d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36-.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4" />
                                            <path d="M9 18c-4.51 2-5-2-7-2" />
                                        </svg>
                                        <a href="https://github.com/RealDewKJ/QuraMate" target="_blank"
                                            class="text-sm font-medium hover:underline hover:text-primary">Source
                                            Code</a>
                                    </div>
                                    <div
                                        class="flex flex-col items-center p-3 rounded-lg bg-muted/30 border border-border border-dashed">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round"
                                            class="lucide lucide-globe mb-2 text-foreground/70">
                                            <circle cx="12" cy="12" r="10" />
                                            <path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20" />
                                            <path d="M2 12h20" />
                                        </svg>
                                        <a href="#"
                                            class="text-sm font-medium hover:underline hover:text-primary">Website</a>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Footer -->
            <div class="flex items-center justify-end border-t border-border p-4 bg-muted/20">
                <div class="flex gap-2">
                    <button @click="close"
                        class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2">
                        Cancel
                    </button>
                    <button @click="save"
                        class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-10 px-4 py-2">
                        Save Changes
                    </button>
                </div>
            </div>
        </div>
        <Toast ref="toastRef" />
    </div>
</template>

<script setup>
import { ref, reactive, h, onMounted, watch } from 'vue';
import Toast from './Toast.vue';

const props = defineProps({
    isOpen: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['close', 'save']);
const toastRef = ref(null);

// Tabs configuration
const tabs = [
    { id: 'general', label: 'General', icon: 'Settings' },
    { id: 'appearance', label: 'Appearance', icon: 'Palette' },
    { id: 'editor', label: 'SQL Editor', icon: 'Type' },
    { id: 'ai', label: 'AI Provider', icon: 'Bot' },
    { id: 'logs', label: 'Logs', icon: 'FileText' },
    { id: 'info', label: 'About', icon: 'Info' }
];

const activeTab = ref('general');
const showAiKey = ref(false);
const isRefreshingLogs = ref(false);
const hasMockError = ref(true);
const appVersion = ref('1.0.0-alpha'); // Hardcoded until we get Wails backend hooked up

// Deep copy of settings for the form
const settings = reactive({
    general: {
        language: 'en'
    },
    appearance: {
        theme: 'system'
    },
    editor: {
        fontFamily: "'JetBrains Mono', monospace",
        fontSize: 14
    },
    ai: {
        provider: 'openai',
        apiKey: ''
    }
});

// Load actual theme on mount to show correct active state
onMounted(() => {
    // Try to load settings from localStorage or similar here
    // For now we check the html class documentElement
    if (document.documentElement.classList.contains('dark')) {
        settings.appearance.theme = 'dark';
    } else {
        settings.appearance.theme = 'light';
    }
});

const close = () => {
    emit('close');
};

const save = () => {
    // Apply theme immediately on save
    if (settings.appearance.theme === 'dark') {
        document.documentElement.classList.add('dark');
    } else if (settings.appearance.theme === 'light') {
        document.documentElement.classList.remove('dark');
    } else {
        // System handling would go here, simplified for now
        if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
            document.documentElement.classList.add('dark');
        } else {
            document.documentElement.classList.remove('dark');
        }
    }

    // TODO: Save to localStorage or backend
    toastRef.value?.success('Settings saved successfully!');
    emit('save', { ...settings });
    emit('close');
};

const setTheme = (theme) => {
    settings.appearance.theme = theme;
};

const refreshLogs = () => {
    isRefreshingLogs.value = true;
    hasMockError.value = !hasMockError.value;
    setTimeout(() => {
        isRefreshingLogs.value = false;
    }, 1000);
};

// Simple Lucide icon renderer for this component
const getIcon = (name) => {
    const icons = {
        Settings: h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
            h('path', { d: 'M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z' }),
            h('circle', { cx: '12', cy: '12', r: '3' })
        ]),
        Palette: h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
            h('circle', { cx: '13.5', cy: '6.5', r: '.5', fill: 'currentColor' }),
            h('circle', { cx: '17.5', cy: '10.5', r: '.5', fill: 'currentColor' }),
            h('circle', { cx: '8.5', cy: '7.5', r: '.5', fill: 'currentColor' }),
            h('circle', { cx: '6.5', cy: '12.5', r: '.5', fill: 'currentColor' }),
            h('path', { d: 'M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.648-.746 1.648-1.688 0-.437-.18-.835-.437-1.125-.29-.289-.438-.652-.438-1.125a1.64 1.64 0 0 1 1.668-1.668h1.996c3.051 0 5.555-2.503 5.555-5.554C21.965 6.012 17.461 2 12 2z' })
        ]),
        Type: h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
            h('polyline', { points: '4 7 4 4 20 4 20 7' }),
            h('line', { x1: '9', x2: '15', y1: '20', y2: '20' }),
            h('line', { x1: '12', x2: '12', y1: '4', y2: '20' })
        ]),
        Bot: h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
            h('path', { d: 'M12 8V4H8' }),
            h('rect', { width: '16', height: '12', x: '4', y: '8', rx: '2' }),
            h('path', { d: 'M2 14h2' }),
            h('path', { d: 'M20 14h2' }),
            h('path', { d: 'M15 13v2' }),
            h('path', { d: 'M9 13v2' })
        ]),
        FileText: h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
            h('path', { d: 'M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z' }),
            h('path', { d: 'M14 2v4a2 2 0 0 0 2 2h4' }),
            h('path', { d: 'M10 9H8' }),
            h('path', { d: 'M16 13H8' }),
            h('path', { d: 'M16 17H8' })
        ]),
        Info: h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
            h('circle', { cx: '12', cy: '12', r: '10' }),
            h('path', { d: 'M12 16v-4' }),
            h('path', { d: 'M12 8h.01' })
        ])
    };
    return icons[name];
};
</script>
