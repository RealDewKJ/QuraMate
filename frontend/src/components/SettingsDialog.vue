<template>
    <div
        v-if="isOpen"
        class="fixed inset-0 z-50 flex items-center justify-center"
    >
        <!-- Overlay -->
        <div
            class="fixed inset-0 bg-background/80 backdrop-blur-sm transition-opacity"
            @click="close"
        ></div>

        <!-- Dialog -->
        <div
            class="relative z-50 flex w-full max-w-[1000px] h-[700px] flex-col rounded-xl border border-border bg-card text-card-foreground shadow-lg overflow-hidden animate-in fade-in zoom-in-95"
        >
            <!-- Header -->
            <div class="flex flex-col space-y-1.5 p-6 border-b border-border">
                <h2 class="text-2xl font-semibold leading-none tracking-tight">
                    Settings
                </h2>
                <p class="text-sm text-muted-foreground">
                    Manage your app preferences, appearance, and connections.
                </p>
                <button
                    @click="close"
                    class="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground"
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
                        class="lucide lucide-x"
                    >
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                    <span class="sr-only">Close</span>
                </button>
            </div>

            <!-- Content Layout -->
            <div class="flex flex-1 overflow-hidden">
                <!-- Sidebar Navigation -->
                <div
                    class="w-48 border-r border-border bg-muted/20 p-4 shrink-0 overflow-y-auto"
                >
                    <nav class="flex flex-col space-y-1">
                        <button
                            v-for="tab in tabs"
                            :key="tab.id"
                            @click="activeTab = tab.id"
                            class="flex items-center gap-2 px-3 py-2 text-sm font-medium rounded-md text-left transition-colors"
                            :class="
                                activeTab === tab.id
                                    ? 'bg-secondary text-secondary-foreground'
                                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                            "
                        >
                            <component
                                :is="getIcon(tab.icon)"
                                class="h-4 w-4"
                            />
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
                            <p class="text-sm text-muted-foreground mb-4">
                                Basic application settings.
                            </p>

                            <div class="space-y-4">
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        Data Editor Language
                                    </label>
                                    <p class="text-xs text-muted-foreground">
                                        Select the language for the data editor
                                        interface.
                                    </p>
                                    <select
                                        v-model="settings.general.language"
                                        class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 max-w-sm"
                                    >
                                        <option value="en">
                                            English (Default)
                                        </option>
                                        <option value="th">
                                            Thai (Coming Soon)
                                        </option>
                                    </select>
                                </div>

                                <div
                                    class="grid gap-2 pt-4 border-t border-border"
                                >
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        Safe Mode
                                    </label>
                                    <div class="flex items-center space-x-2">
                                        <button
                                            type="button"
                                            role="switch"
                                            :aria-checked="
                                                settings.general.enableSafeMode
                                            "
                                            @click="
                                                settings.general.enableSafeMode =
                                                    !settings.general
                                                        .enableSafeMode
                                            "
                                            class="peer inline-flex h-[24px] w-[44px] shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50"
                                            :class="
                                                settings.general.enableSafeMode
                                                    ? 'bg-primary'
                                                    : 'bg-input'
                                            "
                                        >
                                            <span
                                                class="pointer-events-none block h-5 w-5 rounded-full bg-background shadow-lg ring-0 transition-transform"
                                                :class="
                                                    settings.general
                                                        .enableSafeMode
                                                        ? 'translate-x-5'
                                                        : 'translate-x-0'
                                                "
                                            >
                                            </span>
                                        </button>
                                        <span
                                            class="text-sm text-muted-foreground"
                                            >Warn before executing UPDATE or
                                            DELETE queries without a WHERE
                                            clause</span
                                        >
                                    </div>
                                </div>

                                <div class="grid gap-2 pt-4 border-t border-border">
                                    <label class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        Query History
                                    </label>
                                    <div class="flex items-center space-x-2">
                                        <button
                                            type="button"
                                            role="switch"
                                            :aria-checked="settings.general.enableQueryHistory"
                                            @click="settings.general.enableQueryHistory = !settings.general.enableQueryHistory"
                                            class="peer inline-flex h-[24px] w-[44px] shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50"
                                            :class="settings.general.enableQueryHistory ? 'bg-primary' : 'bg-input'"
                                        >
                                            <span
                                                class="pointer-events-none block h-5 w-5 rounded-full bg-background shadow-lg ring-0 transition-transform"
                                                :class="settings.general.enableQueryHistory ? 'translate-x-5' : 'translate-x-0'"
                                            >
                                            </span>
                                        </button>
                                        <span class="text-sm text-muted-foreground">Save executed SQL statements to local query history</span>
                                    </div>
                                </div>

                                <div class="grid gap-2">
                                    <label class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        Query History Retention
                                    </label>
                                    <p class="text-xs text-muted-foreground">
                                        Automatically remove non-favorite history older than this period.
                                    </p>
                                    <select
                                        v-model.number="settings.general.queryHistoryRetentionDays"
                                        :disabled="!settings.general.enableQueryHistory"
                                        class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 max-w-sm"
                                    >
                                        <option :value="7">7 days</option>
                                        <option :value="30">30 days</option>
                                        <option :value="90">90 days</option>
                                        <option :value="180">180 days</option>
                                        <option :value="365">365 days</option>
                                    </select>
                                </div>

                                <!-- <div class="grid gap-2 pt-4 border-t border-border">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        Query Performance Logs
                                    </label>
                                    <div class="flex items-center space-x-2">
                                        <button type="button" role="switch"
                                            :aria-checked="settings.general.enablePerfLogs"
                                            @click="settings.general.enablePerfLogs = !settings.general.enablePerfLogs"
                                            class="peer inline-flex h-[24px] w-[44px] shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50"
                                            :class="settings.general.enablePerfLogs ? 'bg-primary' : 'bg-input'">
                                            <span
                                                class="pointer-events-none block h-5 w-5 rounded-full bg-background shadow-lg ring-0 transition-transform"
                                                :class="settings.general.enablePerfLogs ? 'translate-x-5' : 'translate-x-0'">
                                            </span>
                                        </button>
                                        <span class="text-sm text-muted-foreground">Write lightweight query timing and
                                            row-count logs to the browser console</span>
                                    </div>
                                </div> -->

                            </div>
                        </div>
                    </div>

                    <!-- Keybindings Tab -->
                    <div v-if="activeTab === 'keybindings'" class="space-y-6">
                        <div>
                            <h3 class="text-lg font-medium">Keybindings</h3>
                            <p class="text-sm text-muted-foreground mb-4">
                                System shortcuts are fixed and cannot be edited.
                            </p>

                            <div class="space-y-4">
                                <div
                                    v-for="group in keybindingGroups"
                                    :key="group.id"
                                    class="rounded-lg border border-border bg-card/40"
                                >
                                    <div
                                        class="px-4 py-3 border-b border-border bg-muted/30"
                                    >
                                        <h4 class="text-sm font-semibold">
                                            {{ group.label }}
                                        </h4>
                                    </div>
                                    <div class="divide-y divide-border">
                                        <div
                                            v-for="binding in group.items"
                                            :key="`${group.id}-${binding.key}`"
                                            class="px-4 py-3 flex items-center justify-between gap-4"
                                        >
                                            <div class="min-w-0">
                                                <p
                                                    class="text-sm font-medium text-foreground"
                                                >
                                                    {{ binding.action }}
                                                </p>
                                                <p
                                                    class="text-xs text-muted-foreground"
                                                >
                                                    {{ binding.description }}
                                                </p>
                                            </div>
                                            <kbd
                                                class="shrink-0 rounded-md border border-input bg-background px-2 py-1 text-xs font-mono text-foreground"
                                                >{{ binding.key }}</kbd
                                            >
                                        </div>
                                    </div>
                                </div>

                            </div>
                        </div>
                    </div>

                    <!-- Appearance Tab -->
                    <div v-if="activeTab === 'appearance'" class="space-y-6">
                        <div>
                            <h3 class="text-lg font-medium">Appearance</h3>
                            <p class="text-sm text-muted-foreground mb-4">
                                Customize the look and feel of QuraMate.
                            </p>

                            <div class="space-y-4">
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        Theme
                                    </label>
                                    <div class="flex flex-wrap gap-2 pt-2">
                                        <button
                                            @click="setTheme('light')"
                                            class="inline-flex items-center gap-2 rounded-md border px-3 py-2 text-sm transition-colors"
                                            :class="
                                                settings.appearance.theme ===
                                                'light'
                                                    ? 'border-primary bg-primary/10 text-primary'
                                                    : 'border-border bg-background text-foreground hover:bg-accent'
                                            "
                                        >
                                            <svg
                                                xmlns="http://www.w3.org/2000/svg"
                                                width="16"
                                                height="16"
                                                viewBox="0 0 24 24"
                                                fill="none"
                                                stroke="currentColor"
                                                stroke-width="2"
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                            >
                                                <circle cx="12" cy="12" r="4" />
                                                <path d="M12 2v2" />
                                                <path d="M12 20v2" />
                                                <path d="m4.93 4.93 1.41 1.41" />
                                                <path d="m17.66 17.66 1.41 1.41" />
                                                <path d="M2 12h2" />
                                                <path d="M20 12h2" />
                                                <path d="m6.34 17.66-1.41 1.41" />
                                                <path d="m19.07 4.93-1.41 1.41" />
                                            </svg>
                                            <span>Light</span>
                                        </button>

                                        <button
                                            @click="setTheme('dark')"
                                            class="inline-flex items-center gap-2 rounded-md border px-3 py-2 text-sm transition-colors"
                                            :class="
                                                settings.appearance.theme ===
                                                'dark'
                                                    ? 'border-primary bg-primary/10 text-primary'
                                                    : 'border-border bg-background text-foreground hover:bg-accent'
                                            "
                                        >
                                            <svg
                                                xmlns="http://www.w3.org/2000/svg"
                                                width="16"
                                                height="16"
                                                viewBox="0 0 24 24"
                                                fill="none"
                                                stroke="currentColor"
                                                stroke-width="2"
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                            >
                                                <path
                                                    d="M12 3a6 6 0 1 0 9 9 9 9 0 1 1-9-9"
                                                />
                                            </svg>
                                            <span>Dark</span>
                                        </button>

                                        <button
                                            @click="setTheme('system')"
                                            class="inline-flex items-center gap-2 rounded-md border px-3 py-2 text-sm transition-colors"
                                            :class="
                                                settings.appearance.theme ===
                                                'system'
                                                    ? 'border-primary bg-primary/10 text-primary'
                                                    : 'border-border bg-background text-foreground hover:bg-accent'
                                            "
                                        >
                                            <svg
                                                xmlns="http://www.w3.org/2000/svg"
                                                width="16"
                                                height="16"
                                                viewBox="0 0 24 24"
                                                fill="none"
                                                stroke="currentColor"
                                                stroke-width="2"
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                            >
                                                <rect
                                                    width="18"
                                                    height="12"
                                                    x="3"
                                                    y="4"
                                                    rx="2"
                                                />
                                                <path d="M8 20h8" />
                                                <path d="M12 16v4" />
                                            </svg>
                                            <span>System</span>
                                        </button>
                                    </div>
                                </div>

                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none"
                                    >
                                        Use pointer cursors
                                    </label>
                                    <p class="text-xs text-muted-foreground">
                                        Change the cursor to a pointer when
                                        hovering over interactive elements.
                                    </p>
                                    <div class="flex items-center space-x-2">
                                        <button
                                            type="button"
                                            role="switch"
                                            :aria-checked="
                                                settings.appearance
                                                    .usePointerCursors
                                            "
                                            @click="
                                                settings.appearance.usePointerCursors =
                                                    !settings.appearance
                                                        .usePointerCursors
                                            "
                                            class="peer inline-flex h-[24px] w-[44px] shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50"
                                            :class="
                                                settings.appearance
                                                    .usePointerCursors
                                                    ? 'bg-primary'
                                                    : 'bg-input'
                                            "
                                        >
                                            <span
                                                class="pointer-events-none block h-5 w-5 rounded-full bg-background shadow-lg ring-0 transition-transform"
                                                :class="
                                                    settings.appearance
                                                        .usePointerCursors
                                                        ? 'translate-x-5'
                                                        : 'translate-x-0'
                                                "
                                            />
                                        </button>
                                    </div>
                                </div>

                                <div
                                    class="grid gap-2 mt-4 pt-4 border-t border-border"
                                >
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        App Font
                                    </label>
                                    <select
                                        v-model="settings.appearance.appFont"
                                        class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 max-w-sm"
                                    >
                                        <option value="system-ui, sans-serif">
                                            System UI (Default)
                                        </option>
                                        <option value="'Sarabun', sans-serif">
                                            Sarabun
                                        </option>
                                        <option value="'Inter', sans-serif">
                                            Inter
                                        </option>
                                        <option value="'Roboto', sans-serif">
                                            Roboto
                                        </option>
                                        <option value="'Open Sans', sans-serif">
                                            Open Sans
                                        </option>
                                    </select>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- SQL Editor Tab -->
                    <div v-if="activeTab === 'editor'" class="space-y-6">
                        <div>
                            <h3 class="text-lg font-medium">SQL Editor</h3>
                            <p class="text-sm text-muted-foreground mb-4">
                                Configure the SQL query editor environment.
                            </p>

                            <div class="space-y-4">
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        Font Family
                                    </label>
                                    <select
                                        v-model="settings.editor.fontFamily"
                                        class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 max-w-sm"
                                    >
                                        <option
                                            value="'JetBrains Mono', monospace"
                                        >
                                            JetBrains Mono
                                        </option>
                                        <option value="'Fira Code', monospace">
                                            Fira Code
                                        </option>
                                        <option
                                            value="'Cascadia Code', monospace"
                                        >
                                            Cascadia Code
                                        </option>
                                        <option value="Consolas, monospace">
                                            Consolas
                                        </option>
                                        <option value="Courier New, monospace">
                                            Courier New
                                        </option>
                                    </select>
                                </div>

                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        Font Size
                                    </label>
                                    <div
                                        class="flex items-center gap-4 max-w-sm"
                                    >
                                        <input
                                            type="range"
                                            v-model="settings.editor.fontSize"
                                            min="10"
                                            max="24"
                                            class="w-full accent-primary"
                                        />
                                        <span
                                            class="text-sm font-mono w-12 text-right"
                                            >{{
                                                settings.editor.fontSize
                                            }}px</span
                                        >
                                    </div>
                                </div>

                                <!-- Preview Box -->
                                <div
                                    class="mt-6 border border-border rounded-md overflow-hidden bg-card"
                                >
                                    <div
                                        class="bg-muted px-3 py-1.5 border-b border-border text-xs font-medium text-muted-foreground"
                                    >
                                        Preview
                                    </div>
                                    <div
                                        class="p-4 bg-[var(--vscode-editor-background,#1e1e1e)] overflow-hidden"
                                    >
                                        <pre
                                            :style="{
                                                fontFamily:
                                                    settings.editor.fontFamily,
                                                fontSize: `${settings.editor.fontSize}px`,
                                            }"
                                            class="text-[var(--vscode-editor-foreground,#d4d4d4)]"
                                        ><span class="text-[#569cd6]">SELECT</span>
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
                            <h3 class="text-lg font-medium">
                                Artificial Intelligence
                            </h3>
                            <p class="text-sm text-muted-foreground mb-4">
                                Configure AI providers for QuraMate's smart
                                features.
                            </p>

                            <div
                                class="p-4 mb-4 text-sm text-amber-800 rounded-lg bg-amber-50 dark:bg-amber-950/30 dark:text-amber-400 flex items-start gap-3 border border-amber-200 dark:border-amber-900/50"
                                role="alert"
                            >
                                <svg
                                    class="flex-shrink-0 inline w-4 h-4 mt-0.5"
                                    aria-hidden="true"
                                    xmlns="http://www.w3.org/2000/svg"
                                    fill="currentColor"
                                    viewBox="0 0 20 20"
                                >
                                    <path
                                        d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5ZM9.5 4a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3ZM12 15H8a1 1 0 0 1 0-2h1v-3H8a1 1 0 0 1 0-2h2a1 1 0 0 1 1 1v4h1a1 1 0 0 1 0 2Z"
                                    />
                                </svg>
                                <span class="sr-only">Info</span>
                                <div>
                                    <span class="font-medium"
                                        >Coming Soon!</span
                                    >
                                    AI features are currently in development.
                                    You can save your keys now for when the
                                    feature goes live.
                                </div>
                            </div>

                            <div class="space-y-4">
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        AI Provider
                                    </label>
                                    <select
                                        v-model="settings.ai.provider"
                                        class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 max-w-sm"
                                    >
                                        <option
                                            v-for="provider in aiProviders"
                                            :key="provider.value"
                                            :value="provider.value"
                                        >
                                            {{ provider.label }}
                                        </option>
                                    </select>
                                </div>

                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        API Key
                                    </label>
                                    <div class="relative max-w-md">
                                        <input
                                            :type="
                                                showAiKey ? 'text' : 'password'
                                            "
                                            v-model="currentProviderApiKey"
                                            placeholder="sk-..."
                                            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 pr-10"
                                        />
                                        <button
                                            @click="showAiKey = !showAiKey"
                                            class="absolute right-0 top-0 h-full px-3 text-muted-foreground hover:text-foreground"
                                        >
                                            <svg
                                                v-if="!showAiKey"
                                                xmlns="http://www.w3.org/2000/svg"
                                                width="16"
                                                height="16"
                                                viewBox="0 0 24 24"
                                                fill="none"
                                                stroke="currentColor"
                                                stroke-width="2"
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                class="lucide lucide-eye"
                                            >
                                                <path
                                                    d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"
                                                />
                                                <circle cx="12" cy="12" r="3" />
                                            </svg>
                                            <svg
                                                v-else
                                                xmlns="http://www.w3.org/2000/svg"
                                                width="16"
                                                height="16"
                                                viewBox="0 0 24 24"
                                                fill="none"
                                                stroke="currentColor"
                                                stroke-width="2"
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                class="lucide lucide-eye-off"
                                            >
                                                <path
                                                    d="M9.88 9.88a3 3 0 1 0 4.24 4.24"
                                                />
                                                <path
                                                    d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68"
                                                />
                                                <path
                                                    d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61"
                                                />
                                                <line
                                                    x1="2"
                                                    x2="22"
                                                    y1="2"
                                                    y2="22"
                                                />
                                            </svg>
                                        </button>
                                    </div>
                                    <p
                                        class="text-[10px] text-muted-foreground"
                                    >
                                        API key จะถูกเก็บแยกตาม provider ใน OS
                                        Keychain/Credential Vault ของเครื่องนี้
                                        และจะไม่ถูกบันทึกลงไฟล์ settings
                                        database.
                                    </p>
                                </div>

                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        Base URL
                                    </label>
                                    <input
                                        v-model="currentProviderBaseURL"
                                        placeholder="https://api.example.com/v1"
                                        class="flex h-10 w-full max-w-xl rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                    />
                                    <p
                                        class="text-[10px] text-muted-foreground"
                                    >
                                        ใช้ endpoint เฉพาะ provider นี้เท่านั้น
                                        (แก้ได้แยกกันทุก provider)
                                    </p>
                                </div>

                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        Model
                                    </label>
                                    <select
                                        v-model="selectedProviderModelOption"
                                        class="flex h-10 w-full max-w-xl rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                    >
                                        <option
                                            v-for="model in currentProviderModelOptions"
                                            :key="model"
                                            :value="model"
                                        >
                                            {{ model }}
                                        </option>
                                        <option
                                            :value="CUSTOM_MODEL_OPTION_VALUE"
                                        >
                                            Custom model...
                                        </option>
                                    </select>
                                    <input
                                        v-if="
                                            selectedProviderModelOption ===
                                            CUSTOM_MODEL_OPTION_VALUE
                                        "
                                        v-model="currentProviderCustomModel"
                                        placeholder="custom-model-id"
                                        class="flex h-10 w-full max-w-xl rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                    />
                                    <p
                                        class="text-[10px] text-muted-foreground"
                                    >
                                        รายการ model จะเปลี่ยนตาม provider
                                        ที่เลือก และรองรับ custom model
                                    </p>
                                </div>

                                <div class="pt-2 flex flex-col gap-2">
                                    <button
                                        @click="testProviderConnection"
                                        :disabled="!canTestProvider"
                                        class="inline-flex w-fit items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-9 px-4 py-2"
                                    >
                                        <svg
                                            v-if="isTestingProvider"
                                            class="animate-spin mr-2 h-4 w-4"
                                            xmlns="http://www.w3.org/2000/svg"
                                            fill="none"
                                            viewBox="0 0 24 24"
                                        >
                                            <circle
                                                class="opacity-25"
                                                cx="12"
                                                cy="12"
                                                r="10"
                                                stroke="currentColor"
                                                stroke-width="4"
                                            ></circle>
                                            <path
                                                class="opacity-75"
                                                fill="currentColor"
                                                d="M4 12a8 8 0 0 1 8-8v4a4 4 0 0 0-4 4H4z"
                                            ></path>
                                        </svg>
                                        <svg
                                            v-else
                                            xmlns="http://www.w3.org/2000/svg"
                                            width="16"
                                            height="16"
                                            viewBox="0 0 24 24"
                                            fill="none"
                                            stroke="currentColor"
                                            stroke-width="2"
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            class="lucide lucide-plug mr-2"
                                        >
                                            <path d="M12 22v-5" />
                                            <path d="M9 8V2" />
                                            <path d="M15 8V2" />
                                            <path
                                                d="M18 8a6 6 0 0 1-12 0h12Z"
                                            />
                                        </svg>
                                        {{
                                            isTestingProvider
                                                ? "Testing..."
                                                : "Test Provider Connection"
                                        }}
                                    </button>

                                    <div
                                        v-if="providerTestResult"
                                        class="rounded-md border p-3 text-xs"
                                        :class="
                                            providerTestResult.ok
                                                ? 'border-emerald-400/40 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
                                                : 'border-red-400/40 bg-red-500/10 text-red-700 dark:text-red-300'
                                        "
                                    >
                                        <p class="font-semibold">
                                            {{
                                                providerTestResult.ok
                                                    ? "Connection Successful"
                                                    : "Connection Failed"
                                            }}
                                            <span class="ml-2 opacity-80"
                                                >({{
                                                    providerTestResult.latencyMs
                                                }}ms)</span
                                            >
                                        </p>
                                        <p class="mt-1 break-all">
                                            {{ providerTestResult.message }}
                                        </p>
                                        <p
                                            v-if="providerTestResult.details"
                                            class="mt-1 opacity-90 break-all"
                                        >
                                            {{ providerTestResult.details }}
                                        </p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- System Logs Tab -->
                    <div
                        v-if="activeTab === 'logs'"
                        class="space-y-6 flex flex-col h-full"
                    >
                        <div class="flex items-center justify-between">
                            <div>
                                <h3 class="text-lg font-medium">System Logs</h3>
                                <p class="text-sm text-muted-foreground">
                                    Monitor application events and errors.
                                </p>
                            </div>
                            <button
                                @click="clearLogs"
                                class="inline-flex items-center justify-center rounded-md text-xs font-medium border border-input bg-background hover:bg-accent hover:text-accent-foreground h-8 px-3"
                            >
                                Clear Logs
                            </button>
                        </div>

                        <div
                            class="border border-border rounded-md bg-muted/30 overflow-hidden flex flex-col min-h-[400px]"
                        >
                            <div
                                class="flex-1 overflow-auto p-3 font-mono text-[11px] space-y-1"
                            >
                                <div
                                    v-if="appLogs.length === 0"
                                    class="text-center text-muted-foreground py-4"
                                >
                                    No logs recorded yet.
                                </div>
                                <div
                                    v-for="(log, i) in appLogs"
                                    :key="i"
                                    class="flex gap-2"
                                >
                                    <span
                                        class="text-muted-foreground whitespace-nowrap"
                                        >{{ log.time }}</span
                                    >
                                    <span
                                        :class="{
                                            'text-red-500 font-bold':
                                                log.level === 'ERROR',
                                            'text-blue-500 font-bold':
                                                log.level === 'INFO',
                                            'text-amber-500 font-bold':
                                                log.level === 'WARN',
                                        }"
                                        >[{{ log.level }}]</span
                                    >
                                    <span
                                        class="text-foreground/80 break-all"
                                        >{{ log.message }}</span
                                    >
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Changelogs Tab -->
                    <div
                        v-if="activeTab === 'changelogs'"
                        class="space-y-6 flex flex-col h-full"
                    >
                        <div>
                            <h3 class="text-lg font-medium">Changelogs</h3>
                            <p class="text-sm text-muted-foreground">
                                Keep track of new features, improvements, and
                                bug fixes.
                            </p>
                        </div>

                        <div class="flex-1 overflow-auto pr-2 space-y-3">
                            <div
                                v-for="log in changelogs"
                                :key="log.version"
                                class="border border-border rounded-md overflow-hidden bg-card"
                            >
                                <button
                                    @click="toggleVersion(log.version)"
                                    class="w-full flex items-center justify-between p-4 bg-muted/30 hover:bg-muted/50 transition-colors text-left focus:outline-none"
                                >
                                    <div class="flex items-center gap-3">
                                        <h4 class="font-medium text-base">
                                            {{ log.version }}
                                        </h4>
                                        <span
                                            class="text-xs text-muted-foreground bg-background border border-border px-2 py-0.5 rounded-full"
                                            >{{ log.date }}</span
                                        >
                                    </div>
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="16"
                                        height="16"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="transition-transform duration-200 text-muted-foreground"
                                        :class="
                                            expandedVersion === log.version
                                                ? 'rotate-180'
                                                : ''
                                        "
                                    >
                                        <path d="m6 9 6 6 6-6" />
                                    </svg>
                                </button>

                                <div
                                    v-show="expandedVersion === log.version"
                                    class="p-4 border-t border-border bg-background space-y-2"
                                >
                                    <ul class="space-y-2.5">
                                        <li
                                            v-for="(
                                                change, index
                                            ) in log.changes"
                                            :key="index"
                                            class="flex gap-3 text-sm items-start"
                                        >
                                            <span
                                                class="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider shrink-0 mt-[2px]"
                                                :class="{
                                                    'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 border border-green-200 dark:border-green-800/50':
                                                        change.type === 'feat',
                                                    'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400 border border-blue-200 dark:border-blue-800/50':
                                                        change.type === 'fix',
                                                    'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400 border border-purple-200 dark:border-purple-800/50':
                                                        change.type === 'perf',
                                                    'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300 border border-gray-200 dark:border-gray-700/50':
                                                        change.type === 'chore',
                                                }"
                                            >
                                                {{ change.type }}
                                            </span>
                                            <span
                                                class="text-foreground/90 leading-snug"
                                                >{{ change.text }}</span
                                            >
                                        </li>
                                    </ul>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Info Tab -->
                    <div v-if="activeTab === 'info'" class="space-y-6">
                        <div>
                            <h3 class="text-lg font-medium">About QuraMate</h3>
                            <p class="text-sm text-muted-foreground mb-6">
                                System information and credits.
                            </p>

                            <div
                                class="flex flex-col items-center justify-center p-8 border border-border rounded-xl bg-card shadow-sm text-center"
                            >
                                <div
                                    class="w-24 h-24 rounded-2xl mb-6 flex items-center justify-center shadow-lg shadow-black/5 rotate-3 transition-transform hover:rotate-6 overflow-hidden bg-white/5 p-1 border border-border"
                                >
                                    <img
                                        src="../assets/images/new-icon.png"
                                        alt="QuraMate Logo"
                                        class="w-full h-full object-contain"
                                    />
                                </div>

                                <h1
                                    class="text-3xl font-bold tracking-tight mb-2"
                                >
                                    QuraMate
                                </h1>
                                <p
                                    class="text-sm font-mono bg-muted/50 px-3 py-1 rounded-full text-foreground/80 mb-6 border border-border"
                                >
                                    Version {{ appVersion }}
                                </p>

                                <p
                                    class="text-muted-foreground max-w-sm mb-8 text-sm leading-relaxed"
                                >
                                    A modern, lightweight database management
                                    tool designed for developers. Built with Vue
                                    3, Tailwind CSS, and Go.
                                </p>

                                <div class="grid grid-cols-2 gap-4 w-full mb-6">
                                    <div
                                        class="flex flex-col items-center p-3 rounded-lg bg-muted/30 border border-border border-dashed"
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
                                            class="lucide lucide-github mb-2 text-foreground/70"
                                        >
                                            <path
                                                d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36-.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4"
                                            />
                                            <path d="M9 18c-4.51 2-5-2-7-2" />
                                        </svg>
                                        <a
                                            href="https://github.com/RealDewKJ/QuraMate"
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            class="text-sm font-medium hover:underline hover:text-primary"
                                            >Source Code</a
                                        >
                                    </div>
                                    <div
                                        class="flex flex-col items-center p-3 rounded-lg bg-muted/30 border border-border border-dashed"
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
                                            class="lucide lucide-globe mb-2 text-foreground/70"
                                        >
                                            <circle cx="12" cy="12" r="10" />
                                            <path
                                                d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"
                                            />
                                            <path d="M2 12h20" />
                                        </svg>
                                        <a
                                            href="https://quramate.vercel.app"
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            class="text-sm font-medium hover:underline hover:text-primary"
                                            >Website</a
                                        >
                                    </div>
                                </div>

                                <div class="grid grid-cols-2 gap-4 w-full">
                                    <div
                                        class="flex flex-col items-center p-3 rounded-lg bg-muted/30 border border-border border-dashed text-center"
                                    >
                                        <p
                                            class="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-1"
                                        >
                                            License
                                        </p>
                                        <p
                                            class="text-sm text-foreground flex flex-col items-center gap-1"
                                        >
                                            <span>Apache 2.0 License</span>
                                            <a
                                                href="https://github.com/RealDewKJ/QuraMate/blob/main/LICENSE"
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                class="font-medium text-primary hover:underline text-xs"
                                            >
                                                View License
                                            </a>
                                        </p>
                                    </div>

                                    <div
                                        class="flex flex-col items-center p-3 rounded-lg bg-muted/30 border border-border border-dashed text-center"
                                    >
                                        <p
                                            class="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-1"
                                        >
                                            Open Source
                                        </p>
                                        <a
                                            href="https://github.com/RealDewKJ/QuraMate/blob/main/THIRD_PARTY_LICENSES.md"
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            class="inline-flex items-center justify-center rounded-md text-sm font-medium border border-input bg-background hover:bg-accent hover:text-accent-foreground h-9 px-3 py-1 mt-1"
                                        >
                                            View Libraries
                                        </a>
                                    </div>

                                    <div
                                        class="col-span-2 flex flex-col items-center p-3 rounded-lg bg-muted/30 border border-border border-dashed text-center mt-2"
                                    >
                                        <p
                                            class="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-1"
                                        >
                                            Copyright Notice
                                        </p>
                                        <p class="text-sm text-foreground">
                                            © 2026 QuraMate Team. All rights
                                            reserved.
                                        </p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Footer -->
            <div
                class="flex items-center justify-end border-t border-border p-4 bg-muted/20"
            >
                <div class="flex gap-2">
                    <button
                        @click="close"
                        class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2"
                    >
                        Cancel
                    </button>
                    <button
                        @click="save"
                        class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-10 px-4 py-2"
                    >
                        Save Changes
                    </button>
                </div>
            </div>
        </div>
        <Toast ref="toastRef" />
    </div>
</template>

<script setup>
import { ref, reactive, h, onMounted, watch, computed } from "vue";
import {
    GetAppLogs,
    ClearAppLogs,
    SaveSetting,
    LoadSetting,
    GetCurrentVersion,
    SaveAIProviderKey,
    LoadAIProviderKey,
    DeleteAIProviderKey,
} from "../../wailsjs/go/main/App";
import Toast from "./Toast.vue";
import { colorMode } from "../composables/useTheme";
import changelogData from "../data/changelog.json";
import {
    AI_PROVIDER_DEFINITIONS,
    AI_PROVIDER_DEFAULT_CONFIGS,
} from "../lib/ai/config";
import { completeWithProvider } from "../lib/ai/client";
import {
    DEFAULT_GRID_SCREENSHOT_SHORTCUT,
} from "../composables/useResultGridScreenshot";

const props = defineProps({
    isOpen: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["close", "save"]);
const toastRef = ref(null);

// Tabs configuration
const tabs = [
    { id: "general", label: "General", icon: "Settings" },
    { id: "keybindings", label: "Keybindings", icon: "Keyboard" },
    { id: "appearance", label: "Appearance", icon: "Palette" },
    { id: "editor", label: "SQL Editor", icon: "Type" },
    { id: "ai", label: "AI Provider", icon: "Bot" },
    { id: "changelogs", label: "Changelogs", icon: "History" },
    { id: "logs", label: "System Logs", icon: "Terminal" },
    { id: "info", label: "About", icon: "Info" },
];

const activeTab = ref("general");
const showAiKey = ref(false);
const isTestingProvider = ref(false);
const providerTestResult = ref(null);
const appVersion = ref("");
const aiProviders = AI_PROVIDER_DEFINITIONS;
const aiProviderValues = aiProviders.map((provider) => provider.value);
const CUSTOM_MODEL_OPTION_VALUE = "__custom_model__";
const defaultAiApiKeys = Object.freeze(
    aiProviderValues.reduce((acc, providerId) => {
        acc[providerId] = "";
        return acc;
    }, {}),
);
const providerModelOptions = Object.freeze(
    aiProviders.reduce((acc, provider) => {
        acc[provider.value] = provider.modelOptions || [provider.defaultModel];
        return acc;
    }, {}),
);
const defaultAiProviderConfigs = Object.freeze(
    JSON.parse(JSON.stringify(AI_PROVIDER_DEFAULT_CONFIGS)),
);
const providerApiKeys = reactive({ ...defaultAiApiKeys });
const customProviderModels = reactive(
    aiProviderValues.reduce((acc, providerId) => {
        acc[providerId] = "";
        return acc;
    }, {}),
);
const touchedProviderApiKeys = reactive(
    aiProviderValues.reduce((acc, providerId) => {
        acc[providerId] = false;
        return acc;
    }, {}),
);

const appLogs = ref([]);
const changelogs = changelogData;
const isMacPlatform =
    typeof navigator !== "undefined" &&
    /(Mac|iPhone|iPad|iPod)/i.test(navigator.platform || navigator.userAgent);
const platformModifierLabel = isMacPlatform ? "Cmd" : "Ctrl";
const formatShortcutForPlatform = (shortcut) => {
    return String(shortcut || "").replace(/\bCtrl\b/g, platformModifierLabel);
};

const keybindingGroupsBase = [
    {
        id: "query-editor",
        label: "Query Editor & Tabs",
        items: [
            {
                action: "Run query",
                key: "Ctrl+Enter",
                description: "Execute SQL in the active query tab.",
            },
            {
                action: "Save query",
                key: "Ctrl+S",
                description: "Save the current query to file.",
            },
            {
                action: "Save query as",
                key: "Ctrl+Shift+S",
                description: "Save the current query as a new file.",
            },
            {
                action: "Open new query tab",
                key: "Ctrl+N",
                description: "Create a new SQL tab.",
            },
            {
                action: "Close active query tab",
                key: "Ctrl+W",
                description: "Close the current SQL tab.",
            },
            {
                action: "Beautify SQL",
                key: "Alt+Shift+F",
                description: "Format SQL in the active editor.",
            },
            {
                action: "Open table design",
                key: "Ctrl+D",
                description: "Open designer for the current table tab.",
            },
        ],
    },
    {
        id: "refresh",
        label: "Refresh",
        items: [
            {
                action: "Refresh current context",
                key: "F5",
                description: "Reload current data context.",
            },
            {
                action: "Refresh current context",
                key: "Ctrl+R",
                description: "Reload current data context.",
            },
            {
                action: "Refresh database tree",
                key: "Ctrl+Shift+R",
                description: "Reload tables, views, and routines.",
            },
        ],
    },
    {
        id: "result-grid",
        label: "Result Grid",
        items: [
            {
                action: "Screenshot result grid",
                key: DEFAULT_GRID_SCREENSHOT_SHORTCUT,
                description:
                    "Capture result grid as image with table name and timestamp.",
            },
            {
                action: "Copy selected cell/row(s)",
                key: "Ctrl+C",
                description:
                    "Copy selected data while focus is in query results.",
            },
            {
                action: "Paste rows into table",
                key: "Ctrl+V",
                description:
                    "Paste tabular clipboard data into editable table results.",
            },
            {
                action: "Save editing cell",
                key: "Enter",
                description: "Commit inline cell edit.",
            },
            {
                action: "Cancel editing cell",
                key: "Esc",
                description: "Cancel inline cell edit.",
            },
        ],
    },
];

const keybindingGroups = computed(() => {
    return keybindingGroupsBase.map((group) => ({
        ...group,
        items: group.items.map((item) => ({
            ...item,
            key: formatShortcutForPlatform(item.key),
        })),
    }));
});

const fetchLogs = async () => {
    try {
        const logs = await GetAppLogs();
        appLogs.value = logs.reverse(); // Newest first
    } catch (e) {
        console.error("Failed to fetch app logs", e);
    }
};

const clearLogs = async () => {
    try {
        await ClearAppLogs();
        appLogs.value = [];
        toastRef.value?.success("Logs cleared successfully");
    } catch (e) {
        console.error("Failed to clear logs", e);
    }
};

watch(activeTab, (newTab) => {
    if (newTab === "logs") {
        fetchLogs();
    }
});

const expandedVersion = ref(null);

const toggleVersion = (version) => {
    expandedVersion.value = expandedVersion.value === version ? null : version;
};

// Deep copy of settings for the form
const settings = reactive({
    general: {
        language: "en",
        enableSafeMode: true,
        enablePerfLogs: false,
        enableQueryHistory: true,
        queryHistoryRetentionDays: 30,
    },
    appearance: {
        theme: "system",
        appFont: "system-ui, sans-serif",
        usePointerCursors: true,
    },
    editor: {
        fontFamily: "'JetBrains Mono', monospace",
        fontSize: 14,
    },
    ai: {
        provider: "openai",
        providerConfigs: JSON.parse(JSON.stringify(defaultAiProviderConfigs)),
    },
});

const applyAppearancePreferences = (appearance) => {
    const root = document.documentElement;
    const usePointerCursors = appearance?.usePointerCursors !== false;

    root.classList.toggle("pref-pointer-cursors", usePointerCursors);
    root.classList.toggle("pref-disable-pointer-cursors", !usePointerCursors);
};

const extractLegacyAiApiKeys = (parsedSettings) => {
    const legacyApiKeys = { ...defaultAiApiKeys };
    if (!parsedSettings?.ai || typeof parsedSettings.ai !== "object") {
        return legacyApiKeys;
    }

    const legacyProvider =
        typeof parsedSettings.ai.provider === "string"
            ? parsedSettings.ai.provider
            : "openai";

    if (
        typeof parsedSettings.ai.apiKey === "string" &&
        parsedSettings.ai.apiKey.trim() !== ""
    ) {
        if (aiProviderValues.includes(legacyProvider)) {
            legacyApiKeys[legacyProvider] = parsedSettings.ai.apiKey.trim();
        }
    }

    if (
        parsedSettings.ai.apiKeys &&
        typeof parsedSettings.ai.apiKeys === "object"
    ) {
        aiProviderValues.forEach((provider) => {
            const value = parsedSettings.ai.apiKeys[provider];
            if (typeof value === "string" && value.trim() !== "") {
                legacyApiKeys[provider] = value.trim();
            }
        });
    }

    return legacyApiKeys;
};

const extractLegacyAiProviderConfigs = (parsedSettings) => {
    const legacyConfigs = JSON.parse(JSON.stringify(defaultAiProviderConfigs));
    if (!parsedSettings?.ai || typeof parsedSettings.ai !== "object") {
        return legacyConfigs;
    }

    const selectedProvider =
        typeof parsedSettings.ai.provider === "string"
            ? parsedSettings.ai.provider
            : "openai";
    if (aiProviderValues.includes(selectedProvider)) {
        if (
            typeof parsedSettings.ai.baseURL === "string" &&
            parsedSettings.ai.baseURL.trim() !== ""
        ) {
            legacyConfigs[selectedProvider].baseURL =
                parsedSettings.ai.baseURL.trim();
        }
        if (
            typeof parsedSettings.ai.model === "string" &&
            parsedSettings.ai.model.trim() !== ""
        ) {
            legacyConfigs[selectedProvider].model =
                parsedSettings.ai.model.trim();
        }
    }

    if (
        parsedSettings.ai.providerConfigs &&
        typeof parsedSettings.ai.providerConfigs === "object"
    ) {
        aiProviderValues.forEach((providerId) => {
            const providerConfig =
                parsedSettings.ai.providerConfigs[providerId];
            if (!providerConfig || typeof providerConfig !== "object") {
                return;
            }
            if (
                typeof providerConfig.baseURL === "string" &&
                providerConfig.baseURL.trim() !== ""
            ) {
                legacyConfigs[providerId].baseURL =
                    providerConfig.baseURL.trim();
            }
            if (
                typeof providerConfig.model === "string" &&
                providerConfig.model.trim() !== ""
            ) {
                legacyConfigs[providerId].model = providerConfig.model.trim();
            }
        });
    }

    return legacyConfigs;
};

const normalizeAiSettings = () => {
    if (!settings.ai || typeof settings.ai !== "object") {
        settings.ai = {
            provider: "openai",
            providerConfigs: JSON.parse(
                JSON.stringify(defaultAiProviderConfigs),
            ),
        };
    }

    const isValidProvider = aiProviderValues.includes(settings.ai.provider);
    settings.ai.provider = isValidProvider ? settings.ai.provider : "openai";

    const currentProviderConfigs =
        settings.ai.providerConfigs &&
        typeof settings.ai.providerConfigs === "object"
            ? settings.ai.providerConfigs
            : {};
    const mergedProviderConfigs = JSON.parse(
        JSON.stringify(defaultAiProviderConfigs),
    );
    aiProviderValues.forEach((providerId) => {
        const existingConfig = currentProviderConfigs[providerId];
        if (!existingConfig || typeof existingConfig !== "object") {
            return;
        }
        if (
            typeof existingConfig.baseURL === "string" &&
            existingConfig.baseURL.trim() !== ""
        ) {
            mergedProviderConfigs[providerId].baseURL =
                existingConfig.baseURL.trim();
        }
        if (
            typeof existingConfig.model === "string" &&
            existingConfig.model.trim() !== ""
        ) {
            mergedProviderConfigs[providerId].model =
                existingConfig.model.trim();
        }
    });
    settings.ai.providerConfigs = mergedProviderConfigs;

    if ("baseURL" in settings.ai) {
        delete settings.ai.baseURL;
    }
    if ("model" in settings.ai) {
        delete settings.ai.model;
    }
    if ("apiKeys" in settings.ai) {
        delete settings.ai.apiKeys;
    }
    if ("apiKey" in settings.ai) {
        delete settings.ai.apiKey;
    }
};

const currentProviderApiKey = computed({
    get() {
        return providerApiKeys[settings.ai.provider] || "";
    },
    set(value) {
        providerApiKeys[settings.ai.provider] = value || "";
        touchedProviderApiKeys[settings.ai.provider] = true;
    },
});

const currentProviderBaseURL = computed({
    get() {
        const providerConfig =
            settings.ai.providerConfigs?.[settings.ai.provider];
        return providerConfig?.baseURL || "";
    },
    set(value) {
        const trimmed = (value || "").trim();
        if (!settings.ai.providerConfigs[settings.ai.provider]) {
            settings.ai.providerConfigs[settings.ai.provider] = {
                ...defaultAiProviderConfigs[settings.ai.provider],
            };
        }
        settings.ai.providerConfigs[settings.ai.provider].baseURL = trimmed;
    },
});

const currentProviderModel = computed({
    get() {
        const providerConfig =
            settings.ai.providerConfigs?.[settings.ai.provider];
        return providerConfig?.model || "";
    },
    set(value) {
        const trimmed = (value || "").trim();
        if (!settings.ai.providerConfigs[settings.ai.provider]) {
            settings.ai.providerConfigs[settings.ai.provider] = {
                ...defaultAiProviderConfigs[settings.ai.provider],
            };
        }
        settings.ai.providerConfigs[settings.ai.provider].model = trimmed;
    },
});

const currentProviderModelOptions = computed(() => {
    return providerModelOptions[settings.ai.provider] || [];
});

const selectedProviderModelOption = computed({
    get() {
        const currentModel = (currentProviderModel.value || "").trim();
        const options = currentProviderModelOptions.value;
        if (!currentModel) {
            return options[0] || CUSTOM_MODEL_OPTION_VALUE;
        }
        if (options.includes(currentModel)) {
            return currentModel;
        }
        if (!customProviderModels[settings.ai.provider]) {
            customProviderModels[settings.ai.provider] = currentModel;
        }
        return CUSTOM_MODEL_OPTION_VALUE;
    },
    set(value) {
        if (value === CUSTOM_MODEL_OPTION_VALUE) {
            if (!customProviderModels[settings.ai.provider]) {
                customProviderModels[settings.ai.provider] =
                    currentProviderModel.value || "";
            }
            currentProviderModel.value =
                customProviderModels[settings.ai.provider] || "";
            return;
        }
        currentProviderModel.value = value;
    },
});

const currentProviderCustomModel = computed({
    get() {
        return customProviderModels[settings.ai.provider] || "";
    },
    set(value) {
        customProviderModels[settings.ai.provider] = value || "";
        currentProviderModel.value = value || "";
    },
});

const canTestProvider = computed(() => {
    const providerId = settings.ai.provider;
    const providerConfig = settings.ai.providerConfigs?.[providerId];
    const hasBaseURL = !!providerConfig?.baseURL?.trim();
    const hasModel = !!providerConfig?.model?.trim();
    const hasApiKey =
        providerId === "local" ? true : !!providerApiKeys[providerId]?.trim();
    return hasBaseURL && hasModel && hasApiKey && !isTestingProvider.value;
});

const testProviderConnection = async () => {
    if (!canTestProvider.value) {
        toastRef.value?.error(
            "Please fill API key, Base URL, and Model before testing",
        );
        return;
    }

    const providerId = settings.ai.provider;
    const providerConfig = settings.ai.providerConfigs?.[providerId];
    const startedAt = performance.now();
    isTestingProvider.value = true;
    providerTestResult.value = null;

    try {
        const result = await completeWithProvider({
            provider: providerId,
            apiKey: (providerApiKeys[providerId] || "").trim(),
            baseURL: providerConfig.baseURL,
            model: providerConfig.model,
            messages: [
                {
                    role: "user",
                    content: "Reply with exactly: CONNECTION_OK",
                },
            ],
            temperature: 0,
            maxTokens: 12,
        });

        const latencyMs = Math.round(performance.now() - startedAt);
        providerTestResult.value = {
            ok: true,
            latencyMs,
            message: result.text || "Connected successfully",
        };
        toastRef.value?.success(`Provider reachable (${latencyMs}ms)`);
    } catch (e) {
        const latencyMs = Math.round(performance.now() - startedAt);
        const details = e?.details
            ? JSON.stringify(e.details).slice(0, 500)
            : "";
        const message = e?.message || String(e);
        providerTestResult.value = {
            ok: false,
            latencyMs,
            message,
            details,
        };
        toastRef.value?.error(`Provider test failed: ${message}`);
    } finally {
        isTestingProvider.value = false;
    }
};

watch(
    () => settings.ai.provider,
    () => {
        providerTestResult.value = null;
    },
);

const loadProviderApiKeys = async () => {
    await Promise.all(
        aiProviders.map(async (provider) => {
            const key = await LoadAIProviderKey(provider.value);
            providerApiKeys[provider.value] = key || "";
        }),
    );
};

const saveProviderApiKeys = async () => {
    await Promise.all(
        aiProviders.map(async (provider) => {
            if (!touchedProviderApiKeys[provider.value]) {
                return;
            }

            const value = (providerApiKeys[provider.value] || "").trim();
            const result = value
                ? await SaveAIProviderKey(provider.value, value)
                : await DeleteAIProviderKey(provider.value);
            if (result !== "Success") {
                throw new Error(
                    `Failed to save key for ${provider.label}: ${result}`,
                );
            }
            touchedProviderApiKeys[provider.value] = false;
        }),
    );
};

const migrateLegacyAiKeysToKeychain = async (legacyApiKeys) => {
    let hasMigrated = false;
    await Promise.all(
        aiProviders.map(async (provider) => {
            const providerId = provider.value;
            const legacyValue = (legacyApiKeys[providerId] || "").trim();
            if (!legacyValue) {
                return;
            }

            const existingValue = (providerApiKeys[providerId] || "").trim();
            if (existingValue) {
                return;
            }

            const result = await SaveAIProviderKey(providerId, legacyValue);
            if (result !== "Success") {
                throw new Error(
                    `Failed to migrate legacy key for ${provider.label}: ${result}`,
                );
            }
            providerApiKeys[providerId] = legacyValue;
            hasMigrated = true;
        }),
    );
    return hasMigrated;
};

const loadSettings = async () => {
    let parsed = null;
    try {
        const savedSettingsJson = await LoadSetting("user_settings");
        if (savedSettingsJson) {
            parsed = JSON.parse(savedSettingsJson);
            Object.assign(settings, parsed);
        }
    } catch (e) {
        console.error("Failed to load settings from backend:", e);
    }
    delete settings.shortcuts;

    const legacyApiKeys = extractLegacyAiApiKeys(parsed);
    const legacyProviderConfigs = extractLegacyAiProviderConfigs(parsed);
    normalizeAiSettings();
    settings.ai.providerConfigs = legacyProviderConfigs;
    await loadProviderApiKeys();

    try {
        const migrated = await migrateLegacyAiKeysToKeychain(legacyApiKeys);
        if (migrated) {
            await SaveSetting("user_settings", JSON.stringify(settings));
        }
    } catch (e) {
        console.error("Failed to migrate AI API keys to secure storage:", e);
        toastRef.value?.error(
            "Some AI keys could not be migrated to secure storage",
        );
    }

    settings.appearance.theme =
        colorMode.value === "auto" ? "system" : colorMode.value;
    if (!settings.appearance.appFont) {
        settings.appearance.appFont = "system-ui, sans-serif";
    }
    if (settings.appearance.usePointerCursors === undefined) {
        settings.appearance.usePointerCursors = true;
    }

    // Default safe mode to true if undefined
    if (settings.general.enableSafeMode === undefined) {
        settings.general.enableSafeMode = true;
    }
    if (settings.general.enablePerfLogs === undefined) {
        settings.general.enablePerfLogs = false;
    }
    if (settings.general.enableQueryHistory === undefined) {
        settings.general.enableQueryHistory = true;
    }
    const retention = Number(settings.general.queryHistoryRetentionDays);
    if (!Number.isFinite(retention) || retention <= 0) {
        settings.general.queryHistoryRetentionDays = 30;
    } else {
        settings.general.queryHistoryRetentionDays = Math.min(
            3650,
            Math.max(1, Math.trunc(retention)),
        );
    }

    applyAppearancePreferences(settings.appearance);
};

// Load actual theme on mount to show correct active state
onMounted(async () => {
    loadSettings();
    try {
        appVersion.value = await GetCurrentVersion();
    } catch (e) {
        console.error("Failed to fetch app version:", e);
        appVersion.value = "1.1.0"; // Fallback
    }
});

watch(
    () => props.isOpen,
    (newVal) => {
        if (newVal) {
            loadSettings();
        }
    },
);

const close = () => {
    emit("close");
};

const save = async () => {
    // Apply theme immediately on save
    colorMode.value =
        settings.appearance.theme === "system"
            ? "auto"
            : settings.appearance.theme || "auto";

    // Apply font immediately on save
    if (settings.appearance.appFont) {
        document.documentElement.style.fontFamily = settings.appearance.appFont;
    }
    applyAppearancePreferences(settings.appearance);

    try {
        await saveProviderApiKeys();
        await SaveSetting("user_settings", JSON.stringify(settings));
        // Also save preferred language directly if needed by other systems early
        localStorage.setItem("language", settings.general.language); // Keep this just in case for early load fallback
        toastRef.value?.success("Settings saved successfully!");
        emit("save", { ...settings });
    } catch (e) {
        toastRef.value?.error("Failed to save settings: " + e);
    }
};

const setTheme = (theme) => {
    settings.appearance.theme = theme;
};

// Simple Lucide icon renderer for this component
const getIcon = (name) => {
    const icons = {
        Settings: h(
            "svg",
            {
                xmlns: "http://www.w3.org/2000/svg",
                viewBox: "0 0 24 24",
                fill: "none",
                stroke: "currentColor",
                "stroke-width": "2",
                "stroke-linecap": "round",
                "stroke-linejoin": "round",
            },
            [
                h("path", {
                    d: "M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z",
                }),
                h("circle", { cx: "12", cy: "12", r: "3" }),
            ],
        ),
        Palette: h(
            "svg",
            {
                xmlns: "http://www.w3.org/2000/svg",
                viewBox: "0 0 24 24",
                fill: "none",
                stroke: "currentColor",
                "stroke-width": "2",
                "stroke-linecap": "round",
                "stroke-linejoin": "round",
            },
            [
                h("circle", {
                    cx: "13.5",
                    cy: "6.5",
                    r: ".5",
                    fill: "currentColor",
                }),
                h("circle", {
                    cx: "17.5",
                    cy: "10.5",
                    r: ".5",
                    fill: "currentColor",
                }),
                h("circle", {
                    cx: "8.5",
                    cy: "7.5",
                    r: ".5",
                    fill: "currentColor",
                }),
                h("circle", {
                    cx: "6.5",
                    cy: "12.5",
                    r: ".5",
                    fill: "currentColor",
                }),
                h("path", {
                    d: "M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.648-.746 1.648-1.688 0-.437-.18-.835-.437-1.125-.29-.289-.438-.652-.438-1.125a1.64 1.64 0 0 1 1.668-1.668h1.996c3.051 0 5.555-2.503 5.555-5.554C21.965 6.012 17.461 2 12 2z",
                }),
            ],
        ),
        Type: h(
            "svg",
            {
                xmlns: "http://www.w3.org/2000/svg",
                viewBox: "0 0 24 24",
                fill: "none",
                stroke: "currentColor",
                "stroke-width": "2",
                "stroke-linecap": "round",
                "stroke-linejoin": "round",
            },
            [
                h("polyline", { points: "4 7 4 4 20 4 20 7" }),
                h("line", { x1: "9", x2: "15", y1: "20", y2: "20" }),
                h("line", { x1: "12", x2: "12", y1: "4", y2: "20" }),
            ],
        ),
        Bot: h(
            "svg",
            {
                xmlns: "http://www.w3.org/2000/svg",
                viewBox: "0 0 24 24",
                fill: "none",
                stroke: "currentColor",
                "stroke-width": "2",
                "stroke-linecap": "round",
                "stroke-linejoin": "round",
            },
            [
                h("path", { d: "M12 8V4H8" }),
                h("rect", {
                    width: "16",
                    height: "12",
                    x: "4",
                    y: "8",
                    rx: "2",
                }),
                h("path", { d: "M2 14h2" }),
                h("path", { d: "M20 14h2" }),
                h("path", { d: "M15 13v2" }),
                h("path", { d: "M9 13v2" }),
            ],
        ),
        FileText: h(
            "svg",
            {
                xmlns: "http://www.w3.org/2000/svg",
                viewBox: "0 0 24 24",
                fill: "none",
                stroke: "currentColor",
                "stroke-width": "2",
                "stroke-linecap": "round",
                "stroke-linejoin": "round",
            },
            [
                h("path", {
                    d: "M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z",
                }),
                h("path", { d: "M14 2v4a2 2 0 0 0 2 2h4" }),
                h("path", { d: "M10 9H8" }),
                h("path", { d: "M16 13H8" }),
                h("path", { d: "M16 17H8" }),
            ],
        ),
        History: h(
            "svg",
            {
                xmlns: "http://www.w3.org/2000/svg",
                viewBox: "0 0 24 24",
                fill: "none",
                stroke: "currentColor",
                "stroke-width": "2",
                "stroke-linecap": "round",
                "stroke-linejoin": "round",
            },
            [
                h("path", {
                    d: "M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8",
                }),
                h("path", { d: "M3 3v5h5" }),
                h("path", { d: "M12 7v5l4 2" }),
            ],
        ),
        Terminal: h(
            "svg",
            {
                xmlns: "http://www.w3.org/2000/svg",
                viewBox: "0 0 24 24",
                fill: "none",
                stroke: "currentColor",
                "stroke-width": "2",
                "stroke-linecap": "round",
                "stroke-linejoin": "round",
            },
            [
                h("polyline", { points: "4 17 10 11 4 5" }),
                h("line", { x1: "12", x2: "20", y1: "19", y2: "19" }),
            ],
        ),
        Keyboard: h(
            "svg",
            {
                xmlns: "http://www.w3.org/2000/svg",
                viewBox: "0 0 24 24",
                fill: "none",
                stroke: "currentColor",
                "stroke-width": "2",
                "stroke-linecap": "round",
                "stroke-linejoin": "round",
            },
            [
                h("rect", { width: "20", height: "16", x: "2", y: "4", rx: "2" }),
                h("path", { d: "M6 8h.01" }),
                h("path", { d: "M10 8h.01" }),
                h("path", { d: "M14 8h.01" }),
                h("path", { d: "M18 8h.01" }),
                h("path", { d: "M8 12h.01" }),
                h("path", { d: "M12 12h.01" }),
                h("path", { d: "M16 12h.01" }),
                h("path", { d: "M7 16h10" }),
            ],
        ),
        Info: h(
            "svg",
            {
                xmlns: "http://www.w3.org/2000/svg",
                viewBox: "0 0 24 24",
                fill: "none",
                stroke: "currentColor",
                "stroke-width": "2",
                "stroke-linecap": "round",
                "stroke-linejoin": "round",
            },
            [
                h("circle", { cx: "12", cy: "12", r: "10" }),
                h("path", { d: "M12 16v-4" }),
                h("path", { d: "M12 8h.01" }),
            ],
        ),
    };
    return icons[name];
};
</script>
