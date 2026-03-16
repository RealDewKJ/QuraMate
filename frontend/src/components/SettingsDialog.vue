<template>
    <div
        v-if="isOpen"
        class="fixed inset-0 z-50 flex items-center justify-center"
    >
        <!-- Overlay -->
        <div
            class="fixed inset-0 bg-background/80 transition-opacity"
            @click="close"
        ></div>

        <!-- Dialog -->
        <div
            class="relative z-50 flex w-full max-w-[1000px] h-[700px] flex-col rounded-xl border border-border bg-card text-card-foreground shadow-lg overflow-hidden animate-in fade-in zoom-in-95"
        >
            <!-- Header -->
            <div class="flex flex-col space-y-1.5 p-6 border-b border-border">
                <h2 class="text-2xl font-semibold leading-none tracking-tight">
                    {{ t("common.settings.title") }}
                </h2>
                <p class="text-sm text-muted-foreground">
                    {{ t("common.settings.description") }}
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
                    <span class="sr-only">{{ t("common.close") }}</span>
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
                            <h3 class="text-lg font-medium">{{ t("common.settings.general.title") }}</h3>
                            <p class="text-sm text-muted-foreground mb-4">
                                {{ t("common.settings.general.description") }}
                            </p>

                            <div class="space-y-4">
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        {{ t("common.settings.general.languageLabel") }}
                                    </label>
                                    <p class="text-xs text-muted-foreground">
                                        {{ t("common.settings.general.languageDescription") }}
                                    </p>
                                    <div class="relative max-w-sm">
                                        <button
                                            :ref="(element) => setSettingsDropdownButtonRef('language', element)"
                                            id="settings-language"
                                            type="button"
                                            aria-haspopup="listbox"
                                            :aria-expanded="openSettingsDropdownId === 'language' ? 'true' : 'false'"
                                            :class="settingsListboxTriggerClass"
                                            @click="toggleSettingsDropdown('language', localeSelectOptions, settings.general.language)"
                                            @keydown="handleSettingsDropdownTriggerKeydown($event, 'language', localeSelectOptions, settings.general.language)"
                                        >
                                            <span>{{ getSettingsDropdownLabel(localeSelectOptions, settings.general.language) }}</span>
                                            <svg class="h-4 w-4 text-muted-foreground transition-transform duration-150"
                                                :class="openSettingsDropdownId === 'language' ? 'rotate-180' : ''"
                                                xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
                                                aria-hidden="true">
                                                <path fill-rule="evenodd"
                                                    d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                                                    clip-rule="evenodd" />
                                            </svg>
                                        </button>
                                        <div
                                            v-if="openSettingsDropdownId === 'language'"
                                            :ref="(element) => setSettingsDropdownMenuRef('language', element)"
                                            role="listbox"
                                            aria-labelledby="settings-language"
                                            :class="settingsListboxMenuClass"
                                        >
                                            <button
                                                v-for="(option, index) in localeSelectOptions"
                                                :key="option.value"
                                                type="button"
                                                data-settings-option
                                                role="option"
                                                :tabindex="highlightedSettingsDropdownIndex === index ? 0 : -1"
                                                :aria-selected="settings.general.language === option.value ? 'true' : 'false'"
                                                :class="[
                                                    settingsListboxOptionClass,
                                                    highlightedSettingsDropdownIndex === index ? 'bg-accent/50 text-accent-foreground' : '',
                                                    settings.general.language === option.value ? 'bg-accent/70 text-accent-foreground' : '',
                                                ]"
                                                @click="selectSettingsDropdownOption('language', option.value, (value) => { settings.general.language = value; })"
                                                @focus="highlightedSettingsDropdownIndex = index"
                                                @keydown="handleSettingsDropdownOptionKeydown($event, 'language', localeSelectOptions, (value) => { settings.general.language = value; })"
                                            >
                                                <span>{{ option.label }}</span>
                                                <svg
                                                    v-if="settings.general.language === option.value"
                                                    xmlns="http://www.w3.org/2000/svg"
                                                    width="14"
                                                    height="14"
                                                    viewBox="0 0 24 24"
                                                    fill="none"
                                                    stroke="currentColor"
                                                    stroke-width="2"
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    class="lucide lucide-check"
                                                >
                                                    <path d="M20 6 9 17l-5-5" />
                                                </svg>
                                            </button>
                                        </div>
                                    </div>
                                </div>

                                <div
                                    class="grid gap-2 pt-4 border-t border-border"
                                >
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        {{ t("common.settings.general.safeModeLabel") }}
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
                                            >{{ t("common.settings.general.safeModeDescription") }}</span
                                        >
                                    </div>
                                </div>

                                <div class="grid gap-2 pt-4 border-t border-border">
                                    <label class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        {{ t("common.settings.general.queryHistoryLabel") }}
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
                                        <span class="text-sm text-muted-foreground">{{ t("common.settings.general.queryHistoryDescription") }}</span>
                                    </div>
                                </div>
                                <div class="grid gap-2">
                                    <label class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        {{ t("common.settings.general.queryHistoryRetentionLabel") }}
                                    </label>
                                    <p class="text-xs text-muted-foreground">
                                        {{ t("common.settings.general.queryHistoryRetentionDescription") }}
                                    </p>
                                    <div class="relative max-w-sm">
                                        <button
                                            :ref="(element) => setSettingsDropdownButtonRef('query-history-retention', element)"
                                            id="settings-query-history-retention"
                                            type="button"
                                            :disabled="!settings.general.enableQueryHistory"
                                            aria-haspopup="listbox"
                                            :aria-expanded="openSettingsDropdownId === 'query-history-retention' ? 'true' : 'false'"
                                            :class="settingsListboxTriggerClass"
                                            @click="toggleSettingsDropdown('query-history-retention', queryHistoryRetentionOptions, settings.general.queryHistoryRetentionDays, !settings.general.enableQueryHistory)"
                                            @keydown="handleSettingsDropdownTriggerKeydown($event, 'query-history-retention', queryHistoryRetentionOptions, settings.general.queryHistoryRetentionDays, !settings.general.enableQueryHistory)"
                                        >
                                            <span>{{ getSettingsDropdownLabel(queryHistoryRetentionOptions, settings.general.queryHistoryRetentionDays) }}</span>
                                            <svg class="h-4 w-4 text-muted-foreground transition-transform duration-150"
                                                :class="openSettingsDropdownId === 'query-history-retention' ? 'rotate-180' : ''"
                                                xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
                                                aria-hidden="true">
                                                <path fill-rule="evenodd"
                                                    d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                                                    clip-rule="evenodd" />
                                            </svg>
                                        </button>
                                        <div
                                            v-if="openSettingsDropdownId === 'query-history-retention' && settings.general.enableQueryHistory"
                                            :ref="(element) => setSettingsDropdownMenuRef('query-history-retention', element)"
                                            role="listbox"
                                            aria-labelledby="settings-query-history-retention"
                                            :class="settingsListboxMenuClass"
                                        >
                                            <button
                                                v-for="(option, index) in queryHistoryRetentionOptions"
                                                :key="option.value"
                                                type="button"
                                                data-settings-option
                                                role="option"
                                                :tabindex="highlightedSettingsDropdownIndex === index ? 0 : -1"
                                                :aria-selected="settings.general.queryHistoryRetentionDays === option.value ? 'true' : 'false'"
                                                :class="[
                                                    settingsListboxOptionClass,
                                                    highlightedSettingsDropdownIndex === index ? 'bg-accent/50 text-accent-foreground' : '',
                                                    settings.general.queryHistoryRetentionDays === option.value ? 'bg-accent/70 text-accent-foreground' : '',
                                                ]"
                                                @click="selectSettingsDropdownOption('query-history-retention', option.value, (value) => { settings.general.queryHistoryRetentionDays = value; })"
                                                @focus="highlightedSettingsDropdownIndex = index"
                                                @keydown="handleSettingsDropdownOptionKeydown($event, 'query-history-retention', queryHistoryRetentionOptions, (value) => { settings.general.queryHistoryRetentionDays = value; })"
                                            >
                                                <span>{{ option.label }}</span>
                                                <svg
                                                    v-if="settings.general.queryHistoryRetentionDays === option.value"
                                                    xmlns="http://www.w3.org/2000/svg"
                                                    width="14"
                                                    height="14"
                                                    viewBox="0 0 24 24"
                                                    fill="none"
                                                    stroke="currentColor"
                                                    stroke-width="2"
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    class="lucide lucide-check"
                                                >
                                                    <path d="M20 6 9 17l-5-5" />
                                                </svg>
                                            </button>
                                        </div>
                                    </div>
                                </div>

                                <div class="grid gap-2 pt-4 border-t border-border">
                                    <label class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        {{ t("common.settings.general.workspacePersistenceLabel") }}
                                    </label>
                                    <div class="flex items-center space-x-2">
                                        <button
                                            type="button"
                                            role="switch"
                                            :aria-checked="settings.general.persistWorkspaceState"
                                            @click="settings.general.persistWorkspaceState = !settings.general.persistWorkspaceState"
                                            class="peer inline-flex h-[24px] w-[44px] shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50"
                                            :class="settings.general.persistWorkspaceState ? 'bg-primary' : 'bg-input'"
                                        >
                                            <span
                                                class="pointer-events-none block h-5 w-5 rounded-full bg-background shadow-lg ring-0 transition-transform"
                                                :class="settings.general.persistWorkspaceState ? 'translate-x-5' : 'translate-x-0'"
                                            >
                                            </span>
                                        </button>
                                        <span class="text-sm text-muted-foreground">{{ t("common.settings.general.workspacePersistenceDescription") }}</span>
                                    </div>
                                </div>
                                <div class="grid gap-2">
                                    <label class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        {{ t("common.settings.general.encryptLocalDataLabel") }}
                                    </label>
                                    <p class="text-xs text-muted-foreground">
                                        {{ t("common.settings.general.encryptLocalDataDescription") }}
                                    </p>
                                    <div class="flex items-center space-x-2">
                                        <button
                                            type="button"
                                            role="switch"
                                            :aria-checked="settings.general.encryptLocalPersistentData"
                                            @click="settings.general.encryptLocalPersistentData = !settings.general.encryptLocalPersistentData"
                                            class="peer inline-flex h-[24px] w-[44px] shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50"
                                            :class="settings.general.encryptLocalPersistentData ? 'bg-primary' : 'bg-input'"
                                        >
                                            <span
                                                class="pointer-events-none block h-5 w-5 rounded-full bg-background shadow-lg ring-0 transition-transform"
                                                :class="settings.general.encryptLocalPersistentData ? 'translate-x-5' : 'translate-x-0'"
                                            >
                                            </span>
                                        </button>
                                        <span class="text-sm text-muted-foreground">{{ t("common.settings.general.encryptLocalDataHelp") }}</span>
                                    </div>
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

                    <!-- SQL Server Tab -->
                    <div
                        v-if="props.showSqlServerSettings && activeTab === 'sql-server'"
                        class="space-y-6"
                    >
                        <div>
                            <h3 class="text-lg font-medium">
                                {{ t("common.settings.tabs.sqlServer") }}
                            </h3>
                            <p class="text-sm text-muted-foreground mb-4">
                                {{ t("common.settings.sqlServer.description") }}
                            </p>

                            <div class="space-y-4">
                                <div class="grid gap-2">
                                    <label class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        {{ t("common.settings.sqlServer.trustCertificateLabel") }}
                                    </label>
                                    <p class="text-xs text-muted-foreground">
                                        {{ t("common.settings.sqlServer.trustCertificateDescription") }}
                                    </p>
                                    <div class="flex items-center space-x-2">
                                        <button
                                            type="button"
                                            role="switch"
                                            :aria-checked="settings.general.trustSqlServerCertificateByDefault"
                                            @click="settings.general.trustSqlServerCertificateByDefault = !settings.general.trustSqlServerCertificateByDefault"
                                            class="peer inline-flex h-[24px] w-[44px] shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50"
                                            :class="settings.general.trustSqlServerCertificateByDefault ? 'bg-primary' : 'bg-input'"
                                        >
                                            <span
                                                class="pointer-events-none block h-5 w-5 rounded-full bg-background shadow-lg ring-0 transition-transform"
                                                :class="settings.general.trustSqlServerCertificateByDefault ? 'translate-x-5' : 'translate-x-0'"
                                            >
                                            </span>
                                        </button>
                                        <span class="text-sm text-muted-foreground">
                                            {{ t("common.settings.sqlServer.trustCertificateHelp") }}
                                        </span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Keybindings Tab -->
                    <div v-if="activeTab === 'keybindings'" class="space-y-6">
                        <div>
                            <h3 class="text-lg font-medium">{{ t("common.settings.tabs.keybindings") }}</h3>
                            <p class="text-sm text-muted-foreground mb-4">
                                {{ t("common.settings.keybindings.description") }}
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
                            <h3 class="text-lg font-medium">{{ t("common.settings.tabs.appearance") }}</h3>
                            <p class="text-sm text-muted-foreground mb-4">
                                {{ t("common.settings.appearance.description") }}
                            </p>

                            <div class="space-y-4">
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        {{ t("common.settings.appearance.themeLabel") }}
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
                                            <span>{{ t("common.settings.appearance.themeOptions.light") }}</span>
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
                                            <span>{{ t("common.settings.appearance.themeOptions.dark") }}</span>
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
                                            <span>{{ t("common.settings.appearance.themeOptions.system") }}</span>
                                        </button>
                                    </div>
                                </div>
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none"
                                    >
                                        {{ t("common.settings.appearance.usePointerCursorsLabel") }}
                                    </label>
                                    <p class="text-xs text-muted-foreground">
                                        {{ t("common.settings.appearance.usePointerCursorsDescription") }}
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
                                        {{ t("common.settings.appearance.appFontLabel") }}
                                    </label>
                                    <div class="relative max-w-sm">
                                        <button
                                            :ref="(element) => setSettingsDropdownButtonRef('app-font', element)"
                                            id="settings-app-font"
                                            type="button"
                                            aria-haspopup="listbox"
                                            :aria-expanded="openSettingsDropdownId === 'app-font' ? 'true' : 'false'"
                                            :class="settingsListboxTriggerClass"
                                            @click="toggleSettingsDropdown('app-font', appFontOptions, settings.appearance.appFont)"
                                            @keydown="handleSettingsDropdownTriggerKeydown($event, 'app-font', appFontOptions, settings.appearance.appFont)"
                                        >
                                            <span>{{ getSettingsDropdownLabel(appFontOptions, settings.appearance.appFont) }}</span>
                                            <svg class="h-4 w-4 text-muted-foreground transition-transform duration-150"
                                                :class="openSettingsDropdownId === 'app-font' ? 'rotate-180' : ''"
                                                xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
                                                aria-hidden="true">
                                                <path fill-rule="evenodd"
                                                    d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                                                    clip-rule="evenodd" />
                                            </svg>
                                        </button>
                                        <div
                                            v-if="openSettingsDropdownId === 'app-font'"
                                            :ref="(element) => setSettingsDropdownMenuRef('app-font', element)"
                                            role="listbox"
                                            aria-labelledby="settings-app-font"
                                            :class="settingsListboxMenuClass"
                                        >
                                            <button
                                                v-for="(option, index) in appFontOptions"
                                                :key="option.value"
                                                type="button"
                                                data-settings-option
                                                role="option"
                                                :tabindex="highlightedSettingsDropdownIndex === index ? 0 : -1"
                                                :aria-selected="settings.appearance.appFont === option.value ? 'true' : 'false'"
                                                :class="[
                                                    settingsListboxOptionClass,
                                                    highlightedSettingsDropdownIndex === index ? 'bg-accent/50 text-accent-foreground' : '',
                                                    settings.appearance.appFont === option.value ? 'bg-accent/70 text-accent-foreground' : '',
                                                ]"
                                                @click="selectSettingsDropdownOption('app-font', option.value, (value) => { settings.appearance.appFont = value; })"
                                                @focus="highlightedSettingsDropdownIndex = index"
                                                @keydown="handleSettingsDropdownOptionKeydown($event, 'app-font', appFontOptions, (value) => { settings.appearance.appFont = value; })"
                                            >
                                                <span>{{ option.label }}</span>
                                                <svg
                                                    v-if="settings.appearance.appFont === option.value"
                                                    xmlns="http://www.w3.org/2000/svg"
                                                    width="14"
                                                    height="14"
                                                    viewBox="0 0 24 24"
                                                    fill="none"
                                                    stroke="currentColor"
                                                    stroke-width="2"
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    class="lucide lucide-check"
                                                >
                                                    <path d="M20 6 9 17l-5-5" />
                                                </svg>
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- SQL Editor Tab -->
                    <div v-if="activeTab === 'editor'" class="space-y-6">
                        <div>
                            <h3 class="text-lg font-medium">{{ t("common.settings.tabs.editor") }}</h3>
                            <p class="text-sm text-muted-foreground mb-4">
                                {{ t("common.settings.editor.description") }}
                            </p>

                            <div class="space-y-4">
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        {{ t("common.settings.editor.fontFamilyLabel") }}
                                    </label>
                                    <div class="relative max-w-sm">
                                        <button
                                            :ref="(element) => setSettingsDropdownButtonRef('editor-font-family', element)"
                                            id="settings-editor-font-family"
                                            type="button"
                                            aria-haspopup="listbox"
                                            :aria-expanded="openSettingsDropdownId === 'editor-font-family' ? 'true' : 'false'"
                                            :class="settingsListboxTriggerClass"
                                            @click="toggleSettingsDropdown('editor-font-family', editorFontFamilyOptions, settings.editor.fontFamily)"
                                            @keydown="handleSettingsDropdownTriggerKeydown($event, 'editor-font-family', editorFontFamilyOptions, settings.editor.fontFamily)"
                                        >
                                            <span>{{ getSettingsDropdownLabel(editorFontFamilyOptions, settings.editor.fontFamily) }}</span>
                                            <svg class="h-4 w-4 text-muted-foreground transition-transform duration-150"
                                                :class="openSettingsDropdownId === 'editor-font-family' ? 'rotate-180' : ''"
                                                xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
                                                aria-hidden="true">
                                                <path fill-rule="evenodd"
                                                    d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                                                    clip-rule="evenodd" />
                                            </svg>
                                        </button>
                                        <div
                                            v-if="openSettingsDropdownId === 'editor-font-family'"
                                            :ref="(element) => setSettingsDropdownMenuRef('editor-font-family', element)"
                                            role="listbox"
                                            aria-labelledby="settings-editor-font-family"
                                            :class="settingsListboxMenuClass"
                                        >
                                            <button
                                                v-for="(option, index) in editorFontFamilyOptions"
                                                :key="option.value"
                                                type="button"
                                                data-settings-option
                                                role="option"
                                                :tabindex="highlightedSettingsDropdownIndex === index ? 0 : -1"
                                                :aria-selected="settings.editor.fontFamily === option.value ? 'true' : 'false'"
                                                :class="[
                                                    settingsListboxOptionClass,
                                                    highlightedSettingsDropdownIndex === index ? 'bg-accent/50 text-accent-foreground' : '',
                                                    settings.editor.fontFamily === option.value ? 'bg-accent/70 text-accent-foreground' : '',
                                                ]"
                                                @click="selectSettingsDropdownOption('editor-font-family', option.value, (value) => { settings.editor.fontFamily = value; })"
                                                @focus="highlightedSettingsDropdownIndex = index"
                                                @keydown="handleSettingsDropdownOptionKeydown($event, 'editor-font-family', editorFontFamilyOptions, (value) => { settings.editor.fontFamily = value; })"
                                            >
                                                <span>{{ option.label }}</span>
                                                <svg
                                                    v-if="settings.editor.fontFamily === option.value"
                                                    xmlns="http://www.w3.org/2000/svg"
                                                    width="14"
                                                    height="14"
                                                    viewBox="0 0 24 24"
                                                    fill="none"
                                                    stroke="currentColor"
                                                    stroke-width="2"
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    class="lucide lucide-check"
                                                >
                                                    <path d="M20 6 9 17l-5-5" />
                                                </svg>
                                            </button>
                                        </div>
                                    </div>
                                </div>
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        {{ t("common.settings.editor.fontSizeLabel") }}
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
                                        {{ t("common.settings.editor.previewLabel") }}
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
                                {{ t("common.settings.tabs.ai") }}
                            </h3>
                            <p class="text-sm text-muted-foreground mb-4">
                                {{ t("common.settings.ai.description") }}
                            </p>

                            <div class="mb-4 rounded-xl border border-border/70 bg-muted/20 p-4">
                                <div class="flex flex-wrap items-center gap-2">
                                    <span class="text-sm font-semibold text-foreground">{{ t("common.settings.ai.runtimeStatusLabel") }}</span>
                                    <span
                                        class="inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium"
                                        :class="
                                            activeProviderStatus === 'configured'
                                                ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300'
                                                : activeProviderStatus === 'error'
                                                    ? 'bg-red-500/15 text-red-700 dark:text-red-300'
                                                    : 'bg-amber-500/15 text-amber-700 dark:text-amber-300'
                                        "
                                    >
                                        {{
                                            activeProviderStatus === "configured"
                                                ? t("common.settings.ai.statusConfigured")
                                                : activeProviderStatus === "error"
                                                    ? t("common.settings.ai.statusError")
                                                    : t("common.settings.ai.statusNotConfigured")
                                        }}
                                    </span>
                                    <span
                                        class="inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium"
                                        :class="
                                            currentProviderDefinition?.status === 'stable'
                                                ? 'bg-sky-500/15 text-sky-700 dark:text-sky-300'
                                                : 'bg-amber-500/15 text-amber-700 dark:text-amber-300'
                                        "
                                    >
                                        {{
                                            currentProviderDefinition?.status === "stable"
                                                ? t("common.settings.ai.stabilityStable")
                                                : t("common.settings.ai.stabilityExperimental")
                                        }}
                                    </span>
                                    <span
                                        v-if="hasPendingAiChanges"
                                        class="inline-flex items-center rounded-full bg-violet-500/15 px-2 py-0.5 text-[11px] font-medium text-violet-700 dark:text-violet-300"
                                    >
                                        {{ t("common.settings.ai.unsavedChanges") }}
                                    </span>
                                </div>
                                <div class="mt-3 grid gap-3 text-xs text-muted-foreground md:grid-cols-3">
                                    <div class="rounded-lg border border-border/60 bg-background/70 p-3">
                                        <div class="text-[10px] uppercase tracking-wide">{{ t("common.settings.ai.lastTestedLabel") }}</div>
                                        <div class="mt-1 text-sm text-foreground">
                                            {{ currentPersistedProviderState.lastTestedAt || t("common.settings.ai.lastTestedNever") }}
                                        </div>
                                    </div>
                                    <div class="rounded-lg border border-border/60 bg-background/70 p-3">
                                        <div class="text-[10px] uppercase tracking-wide">{{ t("common.settings.ai.effectiveEndpointLabel") }}</div>
                                        <div class="mt-1 break-all text-sm text-foreground">
                                            {{ currentPersistedProviderState.effectiveEndpointOrigin || t("common.settings.ai.effectiveEndpointUnavailable") }}
                                        </div>
                                    </div>
                                    <div class="rounded-lg border border-border/60 bg-background/70 p-3">
                                        <div class="text-[10px] uppercase tracking-wide">{{ t("common.settings.ai.lastResultLabel") }}</div>
                                        <div class="mt-1 text-sm text-foreground">
                                            {{
                                                currentPersistedProviderState.lastTestResult === "success"
                                                    ? t("common.settings.ai.lastResultSuccess")
                                                    : currentPersistedProviderState.lastTestResult === "failure"
                                                        ? currentPersistedProviderState.lastTestMessage || t("common.settings.ai.lastResultFailed")
                                                        : t("common.settings.ai.lastResultNotTested")
                                            }}
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div class="space-y-4">
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        {{ t("common.settings.ai.providerLabel") }}
                                    </label>
                                    <div class="relative max-w-sm">
                                        <button
                                            :ref="(element) => setSettingsDropdownButtonRef('ai-provider', element)"
                                            id="settings-ai-provider"
                                            type="button"
                                            aria-haspopup="listbox"
                                            :aria-expanded="openSettingsDropdownId === 'ai-provider' ? 'true' : 'false'"
                                            :class="settingsListboxTriggerClass"
                                            @click="toggleSettingsDropdown('ai-provider', aiProviderOptions, settings.ai.provider)"
                                            @keydown="handleSettingsDropdownTriggerKeydown($event, 'ai-provider', aiProviderOptions, settings.ai.provider)"
                                        >
                                            <span>{{ getSettingsDropdownLabel(aiProviderOptions, settings.ai.provider) }}</span>
                                            <svg class="h-4 w-4 text-muted-foreground transition-transform duration-150"
                                                :class="openSettingsDropdownId === 'ai-provider' ? 'rotate-180' : ''"
                                                xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
                                                aria-hidden="true">
                                                <path fill-rule="evenodd"
                                                    d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                                                    clip-rule="evenodd" />
                                            </svg>
                                        </button>
                                        <div
                                            v-if="openSettingsDropdownId === 'ai-provider'"
                                            :ref="(element) => setSettingsDropdownMenuRef('ai-provider', element)"
                                            role="listbox"
                                            aria-labelledby="settings-ai-provider"
                                            :class="settingsListboxMenuClass"
                                        >
                                            <button
                                                v-for="(option, index) in aiProviderOptions"
                                                :key="option.value"
                                                type="button"
                                                data-settings-option
                                                role="option"
                                                :tabindex="highlightedSettingsDropdownIndex === index ? 0 : -1"
                                                :aria-selected="settings.ai.provider === option.value ? 'true' : 'false'"
                                                :class="[
                                                    settingsListboxOptionClass,
                                                    highlightedSettingsDropdownIndex === index ? 'bg-accent/50 text-accent-foreground' : '',
                                                    settings.ai.provider === option.value ? 'bg-accent/70 text-accent-foreground' : '',
                                                ]"
                                                @click="selectSettingsDropdownOption('ai-provider', option.value, (value) => { settings.ai.provider = value; })"
                                                @focus="highlightedSettingsDropdownIndex = index"
                                                @keydown="handleSettingsDropdownOptionKeydown($event, 'ai-provider', aiProviderOptions, (value) => { settings.ai.provider = value; })"
                                            >
                                                <span class="flex items-center gap-2">
                                                    <span>{{ option.label }}</span>
                                                    <span
                                                        v-if="AI_PROVIDER_DEFINITION_MAP[option.value]?.status === 'experimental'"
                                                        class="inline-flex items-center rounded-full bg-amber-500/15 px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide text-amber-700 dark:text-amber-300"
                                                    >
                                                        {{ t("common.settings.ai.stabilityExperimental") }}
                                                    </span>
                                                </span>
                                                <svg
                                                    v-if="settings.ai.provider === option.value"
                                                    xmlns="http://www.w3.org/2000/svg"
                                                    width="14"
                                                    height="14"
                                                    viewBox="0 0 24 24"
                                                    fill="none"
                                                    stroke="currentColor"
                                                    stroke-width="2"
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    class="lucide lucide-check"
                                                >
                                                    <path d="M20 6 9 17l-5-5" />
                                                </svg>
                                            </button>
                                        </div>
                                    </div>
                                </div>
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        {{ t("common.settings.ai.apiKeyLabel") }}
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
                                        v-if="false"
                                        class="text-[10px] text-muted-foreground"
                                    >
                                        API key จะถูกเก็บแยกตาม provider ใน OS
                                        Keychain/Credential Vault ของเครื่องนี้
                                        และจะไม่ถูกบันทึกลงไฟล์ settings
                                        database.
                                    </p>
                                    <p class="text-[10px] text-muted-foreground">
                                        {{ t("common.settings.ai.apiKeyHelp") }}
                                    </p>
                                </div>
                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        {{ t("common.settings.ai.baseUrlLabel") }}
                                    </label>
                                    <input
                                        v-model="currentProviderBaseURL"
                                        placeholder="https://api.example.com/v1"
                                        class="flex h-10 w-full max-w-xl rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                    />
                                    <p
                                        v-if="false"
                                        class="text-[10px] text-muted-foreground"
                                    >
                                        ใช้ endpoint เฉพาะ provider นี้เท่านั้น
                                        (แก้ได้แยกกันทุก provider)
                                    </p>
                                    <p class="text-[10px] text-muted-foreground">
                                        {{ t("common.settings.ai.baseUrlHelp") }}
                                    </p>
                                </div>

                                <div class="grid gap-2">
                                    <label
                                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                    >
                                        {{ t("common.settings.ai.modelLabel") }}
                                    </label>
                                    <div class="relative max-w-xl">
                                        <button
                                            :ref="(element) => setSettingsDropdownButtonRef('ai-model', element)"
                                            id="settings-ai-model"
                                            type="button"
                                            aria-haspopup="listbox"
                                            :aria-expanded="openSettingsDropdownId === 'ai-model' ? 'true' : 'false'"
                                            :class="settingsListboxTriggerClass"
                                            @click="toggleSettingsDropdown('ai-model', providerModelSelectOptions, selectedProviderModelOption)"
                                            @keydown="handleSettingsDropdownTriggerKeydown($event, 'ai-model', providerModelSelectOptions, selectedProviderModelOption)"
                                        >
                                            <span>{{ getSettingsDropdownLabel(providerModelSelectOptions, selectedProviderModelOption) }}</span>
                                            <svg class="h-4 w-4 text-muted-foreground transition-transform duration-150"
                                                :class="openSettingsDropdownId === 'ai-model' ? 'rotate-180' : ''"
                                                xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
                                                aria-hidden="true">
                                                <path fill-rule="evenodd"
                                                    d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                                                    clip-rule="evenodd" />
                                            </svg>
                                        </button>
                                        <div
                                            v-if="openSettingsDropdownId === 'ai-model'"
                                            :ref="(element) => setSettingsDropdownMenuRef('ai-model', element)"
                                            role="listbox"
                                            aria-labelledby="settings-ai-model"
                                            :class="settingsListboxMenuClass"
                                        >
                                            <button
                                                v-for="(option, index) in providerModelSelectOptions"
                                                :key="option.value"
                                                type="button"
                                                data-settings-option
                                                role="option"
                                                :tabindex="highlightedSettingsDropdownIndex === index ? 0 : -1"
                                                :aria-selected="selectedProviderModelOption === option.value ? 'true' : 'false'"
                                                :class="[
                                                    settingsListboxOptionClass,
                                                    highlightedSettingsDropdownIndex === index ? 'bg-accent/50 text-accent-foreground' : '',
                                                    selectedProviderModelOption === option.value ? 'bg-accent/70 text-accent-foreground' : '',
                                                ]"
                                                @click="selectSettingsDropdownOption('ai-model', option.value, (value) => { selectedProviderModelOption = value; })"
                                                @focus="highlightedSettingsDropdownIndex = index"
                                                @keydown="handleSettingsDropdownOptionKeydown($event, 'ai-model', providerModelSelectOptions, (value) => { selectedProviderModelOption = value; })"
                                            >
                                                <span>{{ option.label }}</span>
                                                <svg
                                                    v-if="selectedProviderModelOption === option.value"
                                                    xmlns="http://www.w3.org/2000/svg"
                                                    width="14"
                                                    height="14"
                                                    viewBox="0 0 24 24"
                                                    fill="none"
                                                    stroke="currentColor"
                                                    stroke-width="2"
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    class="lucide lucide-check"
                                                >
                                                    <path d="M20 6 9 17l-5-5" />
                                                </svg>
                                            </button>
                                        </div>
                                    </div>
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
                                        v-if="false"
                                        class="text-[10px] text-muted-foreground"
                                    >
                                        รายการ model จะเปลี่ยนตาม provider
                                        ที่เลือก และรองรับ custom model
                                    </p>
                                </div>
                                <p class="text-[10px] text-muted-foreground">
                                    {{ t("common.settings.ai.modelHelp") }}
                                </p>

                                <div class="space-y-3 rounded-lg border border-border/60 bg-muted/20 p-4">
                                    <div class="text-xs text-muted-foreground">
                                        {{ t("common.settings.ai.sharingDisclosure") }}
                                    </div>
                                    <label class="flex items-start gap-3">
                                        <input
                                            v-model="settings.ai.allowCustomBaseURL"
                                            type="checkbox"
                                            class="mt-1 h-4 w-4 rounded border-input text-primary focus:ring-ring"
                                        />
                                        <span class="space-y-1">
                                            <span class="block text-sm font-medium">{{ t("common.settings.ai.allowCustomEndpointLabel") }}</span>
                                            <span class="block text-[11px] text-muted-foreground">{{ t("common.settings.ai.allowCustomEndpointHelp") }}</span>
                                        </span>
                                    </label>
                                    <label class="flex items-start gap-3">
                                        <input
                                            v-model="settings.ai.shareSchemaContext"
                                            type="checkbox"
                                            class="mt-1 h-4 w-4 rounded border-input text-primary focus:ring-ring"
                                        />
                                        <span class="space-y-1">
                                            <span class="block text-sm font-medium">{{ t("common.settings.ai.shareSchemaContextLabel") }}</span>
                                            <span class="block text-[11px] text-muted-foreground">{{ t("common.settings.ai.shareSchemaContextHelp") }}</span>
                                        </span>
                                    </label>
                                    <label class="flex items-start gap-3">
                                        <input
                                            v-model="settings.ai.shareQueryHistory"
                                            type="checkbox"
                                            class="mt-1 h-4 w-4 rounded border-input text-primary focus:ring-ring"
                                        />
                                        <span class="space-y-1">
                                            <span class="block text-sm font-medium">{{ t("common.settings.ai.shareQueryHistoryLabel") }}</span>
                                            <span class="block text-[11px] text-muted-foreground">{{ t("common.settings.ai.shareQueryHistoryHelp") }}</span>
                                        </span>
                                    </label>
                                    <label class="flex items-start gap-3">
                                        <input
                                            v-model="settings.ai.shareResultSample"
                                            type="checkbox"
                                            class="mt-1 h-4 w-4 rounded border-input text-primary focus:ring-ring"
                                        />
                                        <span class="space-y-1">
                                            <span class="block text-sm font-medium">{{ t("common.settings.ai.shareResultSampleLabel") }}</span>
                                            <span class="block text-[11px] text-muted-foreground">{{ t("common.settings.ai.shareResultSampleHelp") }}</span>
                                        </span>
                                    </label>
                                    <label class="flex items-start gap-3">
                                        <input
                                            v-model="settings.ai.shareExecutionPlan"
                                            type="checkbox"
                                            class="mt-1 h-4 w-4 rounded border-input text-primary focus:ring-ring"
                                        />
                                        <span class="space-y-1">
                                            <span class="block text-sm font-medium">{{ t("common.settings.ai.shareExecutionPlanLabel") }}</span>
                                            <span class="block text-[11px] text-muted-foreground">{{ t("common.settings.ai.shareExecutionPlanHelp") }}</span>
                                        </span>
                                    </label>
                                </div>

                                <div class="pt-2 flex flex-col gap-2">
                                    <p class="text-xs text-muted-foreground">
                                        {{ t("common.settings.ai.saveAndTestHelp") }}
                                    </p>
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
                                                    ? t("common.settings.ai.connectionSuccessful")
                                                    : t("common.settings.ai.connectionFailed")
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
                                <h3 class="text-lg font-medium">{{ t("common.settings.tabs.logs") }}</h3>
                                <p class="text-sm text-muted-foreground">
                                    {{ t("common.settings.logs.description") }}
                                </p>
                            </div>
                            <button
                                @click="clearLogs"
                                class="inline-flex items-center justify-center rounded-md text-xs font-medium border border-input bg-background hover:bg-accent hover:text-accent-foreground h-8 px-3"
                            >
                                {{ t("common.settings.logs.clear") }}
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
                                    {{ t("common.settings.logs.empty") }}
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
                            <h3 class="text-lg font-medium">{{ t("common.settings.tabs.changelogs") }}</h3>
                            <p class="text-sm text-muted-foreground">
                                {{ t("common.settings.changelogs.description") }}
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
                            <h3 class="text-lg font-medium">{{ t("common.settings.tabs.info") }}</h3>
                            <p class="text-sm text-muted-foreground mb-6">
                                {{ t("common.settings.info.description") }}
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
                        {{ t("common.cancel") }}
                    </button>
                    <button
                        v-if="activeTab === 'ai'"
                        @click="saveAndTest"
                        :disabled="!canSaveAndTest"
                        class="inline-flex items-center justify-center rounded-md border border-primary/30 bg-primary/10 px-4 py-2 text-sm font-medium text-primary transition-colors hover:bg-primary/15 disabled:pointer-events-none disabled:opacity-50"
                    >
                        {{
                            isTestingProvider
                                ? "Saving and testing..."
                                : "Save and Test"
                        }}
                    </button>
                    <button
                        @click="save"
                        :disabled="isSavingSettings || isTestingProvider"
                        class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-10 px-4 py-2"
                    >
                        {{
                            isSavingSettings && !isTestingProvider
                                ? "Saving..."
                                : t("common.save")
                        }}
                    </button>
                </div>
            </div>
        </div>
        <Toast ref="toastRef" />
    </div>
</template>

<script setup>
import {
    ref,
    reactive,
    h,
    onMounted,
    onBeforeUnmount,
    watch,
    computed,
    nextTick,
} from "vue";
import { useI18n } from "vue-i18n";
import {
    GetAppLogs,
    ClearAppLogs,
    SaveSetting,
    LoadSetting,
    GetCurrentVersion,
    GetLocalDataEncryptionEnabled,
    SaveAIProviderKey,
    LoadAIProviderKey,
    DeleteAIProviderKey,
    SetLocalDataEncryptionEnabled,
} from "../../wailsjs/go/app/App";
import Toast from "./Toast.vue";
import { colorMode } from "../composables/useTheme";
import changelogData from "../data/changelog.json";
import {
    AI_PROVIDER_DEFINITIONS,
    AI_PROVIDER_DEFINITION_MAP,
    AI_PROVIDER_DEFAULT_CONFIGS,
} from "../lib/ai/config";
import {
    createDefaultAiSettingsSnapshot,
    createDefaultProviderStateMap,
    getEffectiveEndpointOrigin,
    normalizeAiSettings as normalizePersistedAiSettings,
    parsePersistedAiSettings,
    testSavedProviderConnection,
    toProviderErrorTestResult,
} from "../composables/useAiProvider";
import {
    DEFAULT_GRID_SCREENSHOT_SHORTCUT,
} from "../composables/useResultGridScreenshot";
import { localeOptions, setAppLocale } from "../i18n";

const props = defineProps({
    isOpen: {
        type: Boolean,
        default: false,
    },
    showSqlServerSettings: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["close", "save"]);
const { t } = useI18n({ useScope: "global" });
const toastRef = ref(null);
const loadedLocalDataEncryptionEnabled = ref(false);
const settingsDropdownMenuRefs = new Map();
const settingsDropdownButtonRefs = new Map();
const openSettingsDropdownId = ref(null);
const highlightedSettingsDropdownIndex = ref(-1);

// Tabs configuration
const tabs = computed(() => [
    { id: "general", label: t("common.settings.tabs.general"), icon: "Settings" },
    ...(props.showSqlServerSettings
        ? [
              {
                  id: "sql-server",
                  label: t("common.settings.tabs.sqlServer"),
                  icon: "Database",
              },
          ]
        : []),
    { id: "keybindings", label: t("common.settings.tabs.keybindings"), icon: "Keyboard" },
    { id: "appearance", label: t("common.settings.tabs.appearance"), icon: "Palette" },
    { id: "editor", label: t("common.settings.tabs.editor"), icon: "Type" },
    { id: "ai", label: t("common.settings.tabs.ai"), icon: "Bot" },
    { id: "changelogs", label: t("common.settings.tabs.changelogs"), icon: "History" },
    { id: "logs", label: t("common.settings.tabs.logs"), icon: "Terminal" },
    { id: "info", label: t("common.settings.tabs.info"), icon: "Info" },
]);

const activeTab = ref("general");
const showAiKey = ref(false);
const isTestingProvider = ref(false);
const isSavingSettings = ref(false);
const providerTestResult = ref(null);
const persistedAiSettings = ref(createDefaultAiSettingsSnapshot());
const appVersion = ref("");
const aiProviders = AI_PROVIDER_DEFINITIONS;
const aiProviderValues = aiProviders.map((provider) => provider.id);
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
const defaultAiProviderState = Object.freeze(createDefaultProviderStateMap());
const providerApiKeys = reactive({ ...defaultAiApiKeys });
const persistedProviderApiKeys = reactive({ ...defaultAiApiKeys });
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

const keybindingGroups = computed(() => {
    return [
        {
            id: "query-editor",
            label: t("common.settings.keybindings.groups.queryEditorTabs"),
            items: [
                {
                    action: t("common.settings.keybindings.actions.runQuery.label"),
                    key: "Ctrl+Enter",
                    description: t("common.settings.keybindings.actions.runQuery.description"),
                },
                {
                    action: t("common.settings.keybindings.actions.saveQuery.label"),
                    key: "Ctrl+S",
                    description: t("common.settings.keybindings.actions.saveQuery.description"),
                },
                {
                    action: t("common.settings.keybindings.actions.saveQueryAs.label"),
                    key: "Ctrl+Shift+S",
                    description: t("common.settings.keybindings.actions.saveQueryAs.description"),
                },
                {
                    action: t("common.settings.keybindings.actions.openNewQueryTab.label"),
                    key: "Ctrl+N",
                    description: t("common.settings.keybindings.actions.openNewQueryTab.description"),
                },
                {
                    action: t("common.settings.keybindings.actions.closeActiveQueryTab.label"),
                    key: "Ctrl+W",
                    description: t("common.settings.keybindings.actions.closeActiveQueryTab.description"),
                },
                {
                    action: t("common.settings.keybindings.actions.beautifySql.label"),
                    key: "Alt+Shift+F",
                    description: t("common.settings.keybindings.actions.beautifySql.description"),
                },
                {
                    action: t("common.settings.keybindings.actions.openTableDesign.label"),
                    key: "Ctrl+D",
                    description: t("common.settings.keybindings.actions.openTableDesign.description"),
                },
            ],
        },
        {
            id: "refresh",
            label: t("common.settings.keybindings.groups.refresh"),
            items: [
                {
                    action: t("common.settings.keybindings.actions.refreshCurrentContext.label"),
                    key: "F5",
                    description: t("common.settings.keybindings.actions.refreshCurrentContext.description"),
                },
                {
                    action: t("common.settings.keybindings.actions.refreshCurrentContext.label"),
                    key: "Ctrl+R",
                    description: t("common.settings.keybindings.actions.refreshCurrentContext.description"),
                },
                {
                    action: t("common.settings.keybindings.actions.refreshDatabaseTree.label"),
                    key: "Ctrl+Shift+R",
                    description: t("common.settings.keybindings.actions.refreshDatabaseTree.description"),
                },
            ],
        },
        {
            id: "result-grid",
            label: t("common.settings.keybindings.groups.resultGrid"),
            items: [
                {
                    action: t("common.settings.keybindings.actions.screenshotResultGrid.label"),
                    key: DEFAULT_GRID_SCREENSHOT_SHORTCUT,
                    description: t("common.settings.keybindings.actions.screenshotResultGrid.description"),
                },
                {
                    action: t("common.settings.keybindings.actions.copySelectedCells.label"),
                    key: "Ctrl+C",
                    description: t("common.settings.keybindings.actions.copySelectedCells.description"),
                },
                {
                    action: t("common.settings.keybindings.actions.pasteRowsIntoTable.label"),
                    key: "Ctrl+V",
                    description: t("common.settings.keybindings.actions.pasteRowsIntoTable.description"),
                },
                {
                    action: t("common.settings.keybindings.actions.saveEditingCell.label"),
                    key: "Enter",
                    description: t("common.settings.keybindings.actions.saveEditingCell.description"),
                },
                {
                    action: t("common.settings.keybindings.actions.cancelEditingCell.label"),
                    key: "Esc",
                    description: t("common.settings.keybindings.actions.cancelEditingCell.description"),
                },
            ],
        },
    ].map((group) => ({
        ...group,
        items: group.items.map((item) => ({
            ...item,
            key: formatShortcutForPlatform(item.key),
        })),
    }));
});

const settingsListboxTriggerClass =
    "flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background transition-colors placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 hover:bg-accent/40";
const settingsListboxMenuClass =
    "absolute left-0 top-full z-50 mt-2 w-full overflow-hidden rounded-2xl border border-border/80 bg-popover/95 py-1 text-popover-foreground shadow-xl ring-1 ring-black/5 backdrop-blur animate-in fade-in zoom-in-95 duration-100";
const settingsListboxOptionClass =
    "flex w-full items-center justify-between px-3 py-2 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground";

const localeSelectOptions = computed(() =>
    localeOptions.map((option) => ({
        value: option.value,
        label: t(option.labelKey),
    })),
);

const queryHistoryRetentionOptions = computed(() => [
    { value: 7, label: t("common.settings.general.queryHistoryRetentionDays.7") },
    { value: 30, label: t("common.settings.general.queryHistoryRetentionDays.30") },
    { value: 90, label: t("common.settings.general.queryHistoryRetentionDays.90") },
    { value: 180, label: t("common.settings.general.queryHistoryRetentionDays.180") },
    { value: 365, label: t("common.settings.general.queryHistoryRetentionDays.365") },
]);

const appFontOptions = computed(() => [
    {
        value: "system-ui, sans-serif",
        label: t("common.settings.appearance.appFontSystemDefault"),
    },
    { value: "'Sarabun', sans-serif", label: "Sarabun" },
    { value: "'Inter', sans-serif", label: "Inter" },
    { value: "'Roboto', sans-serif", label: "Roboto" },
    { value: "'Open Sans', sans-serif", label: "Open Sans" },
]);

const editorFontFamilyOptions = computed(() => [
    { value: "'JetBrains Mono', monospace", label: "JetBrains Mono" },
    { value: "'Fira Code', monospace", label: "Fira Code" },
    { value: "'Cascadia Code', monospace", label: "Cascadia Code" },
    { value: "Consolas, monospace", label: "Consolas" },
    { value: "Courier New, monospace", label: "Courier New" },
]);

const aiProviderOptions = computed(() =>
    aiProviders.map((provider) => ({
        value: provider.value,
        label: provider.label,
    })),
);

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

watch(
    () => props.showSqlServerSettings,
    (showSqlServerSettings) => {
        if (!showSqlServerSettings && activeTab.value === "sql-server") {
            activeTab.value = "general";
        }
    },
);

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
        enableQueryHistory: false,
        queryHistoryRetentionDays: 30,
        persistWorkspaceState: false,
        encryptLocalPersistentData: false,
        trustSqlServerCertificateByDefault: true,
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
        providerState: JSON.parse(JSON.stringify(defaultAiProviderState)),
        allowCustomBaseURL: false,
        shareSchemaContext: false,
        shareQueryHistory: false,
        shareResultSample: false,
        shareExecutionPlan: false,
    },
});

const applyAppearancePreferences = (appearance) => {
    const root = document.documentElement;
    const usePointerCursors = appearance?.usePointerCursors !== false;

    root.classList.toggle("pref-pointer-cursors", usePointerCursors);
    root.classList.toggle("pref-disable-pointer-cursors", !usePointerCursors);
};

const cloneJson = (value) => JSON.parse(JSON.stringify(value));

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

const syncDraftAiSettings = (snapshot) => {
    settings.ai = cloneJson(snapshot);
};

const syncPersistedAiSettings = (snapshot) => {
    persistedAiSettings.value = cloneJson(snapshot);
};

const currentProviderDefinition = computed(
    () => AI_PROVIDER_DEFINITION_MAP[settings.ai.provider],
);

const currentPersistedProviderState = computed(() => {
    return (
        persistedAiSettings.value.providerState?.[settings.ai.provider] ||
        createDefaultProviderStateMap()[settings.ai.provider]
    );
});

const hasPendingAiChanges = computed(() => {
    const providerId = settings.ai.provider;
    const draftAi = normalizePersistedAiSettings(settings.ai);
    const persistedAi = normalizePersistedAiSettings(persistedAiSettings.value);
    const draftKey = (providerApiKeys[providerId] || "").trim();
    const persistedKey = (persistedProviderApiKeys[providerId] || "").trim();
    return (
        JSON.stringify(draftAi) !== JSON.stringify(persistedAi) ||
        draftKey !== persistedKey
    );
});

const activeProviderStatus = computed(() => {
    const providerId = settings.ai.provider;
    const providerConfig = persistedAiSettings.value.providerConfigs?.[providerId];
    const providerState = currentPersistedProviderState.value;
    const definition = currentProviderDefinition.value;
    const hasApiKey =
        definition?.requiresApiKey === false
            ? true
            : !!persistedProviderApiKeys[providerId]?.trim();
    const hasBaseURL = !!providerConfig?.baseURL?.trim();
    const hasModel = !!providerConfig?.model?.trim();

    if (!hasBaseURL || !hasModel || !hasApiKey) {
        return "not_configured";
    }
    if (providerState?.lastTestResult === "failure") {
        return "error";
    }
    return "configured";
});

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

const providerModelSelectOptions = computed(() => [
    ...currentProviderModelOptions.value.map((model) => ({
        value: model,
        label: model,
    })),
    {
        value: CUSTOM_MODEL_OPTION_VALUE,
        label: t("common.settings.ai.customModelOption"),
    },
]);

const getSettingsDropdownLabel = (options, value) => {
    return (
        options.find((option) => option.value === value)?.label ??
        options[0]?.label ??
        ""
    );
};

const setSettingsDropdownMenuRef = (dropdownId, element) => {
    if (element) {
        settingsDropdownMenuRefs.set(dropdownId, element);
        return;
    }
    settingsDropdownMenuRefs.delete(dropdownId);
};

const setSettingsDropdownButtonRef = (dropdownId, element) => {
    if (element) {
        settingsDropdownButtonRefs.set(dropdownId, element);
        return;
    }
    settingsDropdownButtonRefs.delete(dropdownId);
};

const focusHighlightedSettingsOption = async (dropdownId) => {
    await nextTick();
    const menuElement = settingsDropdownMenuRefs.get(dropdownId);
    const optionElements = menuElement?.querySelectorAll("[data-settings-option]");
    const optionElement = optionElements?.[highlightedSettingsDropdownIndex.value];
    optionElement?.focus();
};

const openSettingsDropdown = async (dropdownId, options, currentValue) => {
    openSettingsDropdownId.value = dropdownId;
    const selectedIndex = options.findIndex(
        (option) => option.value === currentValue,
    );
    highlightedSettingsDropdownIndex.value =
        selectedIndex >= 0 ? selectedIndex : 0;
    await focusHighlightedSettingsOption(dropdownId);
};

const closeSettingsDropdown = () => {
    openSettingsDropdownId.value = null;
    highlightedSettingsDropdownIndex.value = -1;
};

const toggleSettingsDropdown = async (
    dropdownId,
    options,
    currentValue,
    disabled = false,
) => {
    if (disabled) {
        return;
    }
    if (openSettingsDropdownId.value === dropdownId) {
        closeSettingsDropdown();
        return;
    }
    await openSettingsDropdown(dropdownId, options, currentValue);
};

const moveSettingsDropdownHighlight = async (dropdownId, options, direction) => {
    if (!options.length) {
        return;
    }
    if (highlightedSettingsDropdownIndex.value < 0) {
        highlightedSettingsDropdownIndex.value = 0;
    } else {
        highlightedSettingsDropdownIndex.value =
            (highlightedSettingsDropdownIndex.value + direction + options.length) %
            options.length;
    }
    await focusHighlightedSettingsOption(dropdownId);
};

const selectSettingsDropdownOption = (
    dropdownId,
    value,
    assignValue,
) => {
    assignValue(value);
    closeSettingsDropdown();
    nextTick(() => {
        settingsDropdownButtonRefs.get(dropdownId)?.focus();
    });
};

const handleSettingsDropdownTriggerKeydown = async (
    event,
    dropdownId,
    options,
    currentValue,
    disabled = false,
) => {
    if (disabled) {
        return;
    }

    if (event.key === "ArrowDown" || event.key === "ArrowUp") {
        event.preventDefault();
        if (openSettingsDropdownId.value !== dropdownId) {
            await openSettingsDropdown(dropdownId, options, currentValue);
            if (event.key === "ArrowUp" && options.length > 0) {
                highlightedSettingsDropdownIndex.value = options.length - 1;
                await focusHighlightedSettingsOption(dropdownId);
            }
            return;
        }
        await moveSettingsDropdownHighlight(
            dropdownId,
            options,
            event.key === "ArrowDown" ? 1 : -1,
        );
        return;
    }

    if (event.key === "Enter" || event.key === " ") {
        event.preventDefault();
        await toggleSettingsDropdown(dropdownId, options, currentValue, disabled);
        return;
    }

    if (event.key === "Escape" && openSettingsDropdownId.value === dropdownId) {
        event.preventDefault();
        closeSettingsDropdown();
    }
};

const handleSettingsDropdownOptionKeydown = async (
    event,
    dropdownId,
    options,
    assignValue,
) => {
    if (event.key === "ArrowDown" || event.key === "ArrowUp") {
        event.preventDefault();
        await moveSettingsDropdownHighlight(
            dropdownId,
            options,
            event.key === "ArrowDown" ? 1 : -1,
        );
        return;
    }

    if (event.key === "Enter" || event.key === " ") {
        event.preventDefault();
        const highlightedOption = options[highlightedSettingsDropdownIndex.value];
        if (highlightedOption) {
            selectSettingsDropdownOption(
                dropdownId,
                highlightedOption.value,
                assignValue,
            );
        }
        return;
    }

    if (event.key === "Escape") {
        event.preventDefault();
        closeSettingsDropdown();
        nextTick(() => {
            settingsDropdownButtonRefs.get(dropdownId)?.focus();
        });
    }
};

const handleSettingsDropdownDocumentClick = (event) => {
    if (!openSettingsDropdownId.value) {
        return;
    }

    const menuElement = settingsDropdownMenuRefs.get(openSettingsDropdownId.value);
    const buttonElement = settingsDropdownButtonRefs.get(openSettingsDropdownId.value);
    const target = event.target;

    if (menuElement?.contains(target) || buttonElement?.contains(target)) {
        return;
    }

    closeSettingsDropdown();
};

const currentProviderCustomModel = computed({
    get() {
        return customProviderModels[settings.ai.provider] || "";
    },
    set(value) {
        customProviderModels[settings.ai.provider] = value || "";
        currentProviderModel.value = value || "";
    },
});

const canSaveAndTest = computed(() => {
    return !isSavingSettings.value && !isTestingProvider.value;
});

watch(
    () => settings.ai.provider,
    () => {
        const providerState =
            persistedAiSettings.value.providerState?.[settings.ai.provider];
        providerTestResult.value = providerState?.lastTestResult
            ? {
                  ok: providerState.lastTestResult === "success",
                  latencyMs: 0,
                  message: providerState.lastTestMessage || "",
                  details: "",
              }
            : null;
    },
);

const loadProviderApiKeys = async () => {
    await Promise.all(
        aiProviders.map(async (provider) => {
            const key = await LoadAIProviderKey(provider.id);
            providerApiKeys[provider.id] = key || "";
            persistedProviderApiKeys[provider.id] = key || "";
        }),
    );
};

const saveProviderApiKeys = async () => {
    await Promise.all(
        aiProviders.map(async (provider) => {
            if (!touchedProviderApiKeys[provider.id]) {
                return;
            }

            const value = (providerApiKeys[provider.id] || "").trim();
            const result = value
                ? await SaveAIProviderKey(provider.id, value)
                : await DeleteAIProviderKey(provider.id);
            if (result !== "Success") {
                throw new Error(
                    `Failed to save key for ${provider.label}: ${result}`,
                );
            }
            persistedProviderApiKeys[provider.id] = value;
            touchedProviderApiKeys[provider.id] = false;
        }),
    );
};

const migrateLegacyAiKeysToKeychain = async (legacyApiKeys) => {
    let hasMigrated = false;
    await Promise.all(
        aiProviders.map(async (provider) => {
            const providerId = provider.id;
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
            persistedProviderApiKeys[providerId] = legacyValue;
            hasMigrated = true;
        }),
    );
    return hasMigrated;
};

const updateDraftProviderState = (providerId, result, statusOverride) => {
    if (!settings.ai.providerState?.[providerId]) {
        settings.ai.providerState[providerId] = {
            ...createDefaultProviderStateMap()[providerId],
        };
    }
    settings.ai.providerState[providerId].status =
        statusOverride ||
        (result.ok ? "configured" : "error");
    settings.ai.providerState[providerId].lastTestResult = result.ok
        ? "success"
        : "failure";
    settings.ai.providerState[providerId].lastTestMessage = result.message || "";
    settings.ai.providerState[providerId].lastTestedAt = result.testedAt || null;
    settings.ai.providerState[providerId].effectiveEndpointOrigin =
        result.endpointOrigin ||
        getEffectiveEndpointOrigin(
            settings.ai.providerConfigs?.[providerId]?.baseURL || "",
        );
};

const loadSettings = async () => {
    let parsed = null;
    let savedSettingsJson = "";
    try {
        loadedLocalDataEncryptionEnabled.value =
            await GetLocalDataEncryptionEnabled();
        savedSettingsJson = await LoadSetting("user_settings");
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
    await loadProviderApiKeys();

    const normalizedAiSettings = normalizePersistedAiSettings({
        ...parsePersistedAiSettings(savedSettingsJson),
        ...(parsed?.ai || {}),
        providerConfigs: legacyProviderConfigs,
    });
    syncPersistedAiSettings(normalizedAiSettings);
    syncDraftAiSettings(normalizedAiSettings);
    providerTestResult.value =
        normalizedAiSettings.providerState?.[normalizedAiSettings.provider]
            ?.lastTestResult
            ? {
                  ok:
                      normalizedAiSettings.providerState[
                          normalizedAiSettings.provider
                      ].lastTestResult === "success",
                  latencyMs: 0,
                  message:
                      normalizedAiSettings.providerState[
                          normalizedAiSettings.provider
                      ].lastTestMessage || "",
                  details: "",
              }
            : null;

    try {
        const migrated = await migrateLegacyAiKeysToKeychain(legacyApiKeys);
        if (migrated) {
            const aiSnapshot = normalizePersistedAiSettings(settings.ai);
            syncPersistedAiSettings(aiSnapshot);
            syncDraftAiSettings(aiSnapshot);
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
        settings.general.enableQueryHistory = false;
    }
    if (settings.general.persistWorkspaceState === undefined) {
        settings.general.persistWorkspaceState = false;
    }
    if (settings.general.trustSqlServerCertificateByDefault === undefined) {
        settings.general.trustSqlServerCertificateByDefault = true;
    }
    settings.general.encryptLocalPersistentData =
        loadedLocalDataEncryptionEnabled.value;
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
    document.addEventListener("click", handleSettingsDropdownDocumentClick);
    window.addEventListener("keydown", handleEscapeKeydown);
});

onBeforeUnmount(() => {
    document.removeEventListener("click", handleSettingsDropdownDocumentClick);
    window.removeEventListener("keydown", handleEscapeKeydown);
});

watch(
    () => props.isOpen,
    (newVal) => {
        if (newVal) {
            loadSettings();
            return;
        }
        closeSettingsDropdown();
    },
);

watch(
    () => settings.general.enableQueryHistory,
    (isEnabled) => {
        if (!isEnabled && openSettingsDropdownId.value === "query-history-retention") {
            closeSettingsDropdown();
        }
    },
);

const close = () => {
    emit("close");
};

const persistSettings = async () => {
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

    await saveProviderApiKeys();
    if (
        settings.general.encryptLocalPersistentData !==
        loadedLocalDataEncryptionEnabled.value
    ) {
        const encryptionResult = await SetLocalDataEncryptionEnabled(
            settings.general.encryptLocalPersistentData,
        );
        if (encryptionResult !== "Success") {
            throw new Error(encryptionResult);
        }
        loadedLocalDataEncryptionEnabled.value =
            settings.general.encryptLocalPersistentData;
    }

    const normalizedAiSettings = normalizePersistedAiSettings(settings.ai);
    normalizedAiSettings.providerState = cloneJson(
        settings.ai.providerState || normalizedAiSettings.providerState,
    );
    aiProviderValues.forEach((providerId) => {
        const definition = AI_PROVIDER_DEFINITION_MAP[providerId];
        const providerConfig = normalizedAiSettings.providerConfigs[providerId];
        const hasApiKey =
            definition?.requiresApiKey === false
                ? true
                : !!(providerApiKeys[providerId] || "").trim();
        const hasBaseURL = !!providerConfig?.baseURL?.trim();
        const hasModel = !!providerConfig?.model?.trim();
        normalizedAiSettings.providerState[providerId].effectiveEndpointOrigin =
            getEffectiveEndpointOrigin(providerConfig?.baseURL || "");
        if (!hasApiKey || !hasBaseURL || !hasModel) {
            normalizedAiSettings.providerState[providerId].status =
                "not_configured";
        } else if (
            normalizedAiSettings.providerState[providerId].lastTestResult !==
            "failure"
        ) {
            normalizedAiSettings.providerState[providerId].status =
                "configured";
        }
    });
    settings.ai = cloneJson(normalizedAiSettings);

    const saveResult = await SaveSetting("user_settings", JSON.stringify(settings));
    if (saveResult !== "Success") {
        throw new Error(saveResult);
    }

    syncPersistedAiSettings(normalizedAiSettings);
    setAppLocale(settings.general.language);
    emit("save", cloneJson(settings));
    return normalizedAiSettings;
};

const save = async () => {
    isSavingSettings.value = true;
    try {
        await persistSettings();
        toastRef.value?.success(t("common.settings.saveSuccess"));
    } catch (e) {
        toastRef.value?.error(
            t("common.settings.saveFailure", { error: String(e) }),
        );
    } finally {
        isSavingSettings.value = false;
    }
};

const saveAndTest = async () => {
    const providerId = settings.ai.provider;
    isSavingSettings.value = true;
    isTestingProvider.value = true;
    providerTestResult.value = null;
    const startedAt = performance.now();

    try {
        const normalizedAiSettings = await persistSettings();
        syncPersistedAiSettings(normalizedAiSettings);
        const result = await testSavedProviderConnection(providerId);
        providerTestResult.value = result;
        updateDraftProviderState(providerId, result, "configured");
        syncPersistedAiSettings(normalizePersistedAiSettings(settings.ai));
        const saveResult = await SaveSetting("user_settings", JSON.stringify(settings));
        if (saveResult !== "Success") {
            throw new Error(saveResult);
        }
        toastRef.value?.success(
            t("common.settings.ai.providerReachable", {
                latency: result.latencyMs,
            }),
        );
    } catch (e) {
        const endpointOrigin = getEffectiveEndpointOrigin(
            settings.ai.providerConfigs?.[providerId]?.baseURL || "",
        );
        const result = toProviderErrorTestResult(
            providerId,
            e,
            Math.round(performance.now() - startedAt),
            endpointOrigin,
        );
        providerTestResult.value = result;
        updateDraftProviderState(providerId, result, "error");
        const normalizedAiSettings = normalizePersistedAiSettings(settings.ai);
        syncPersistedAiSettings(normalizedAiSettings);
        const saveResult = await SaveSetting("user_settings", JSON.stringify(settings));
        if (saveResult !== "Success") {
            console.error("Failed to persist provider test result", saveResult);
        }
        toastRef.value?.error(`Provider test failed: ${result.message}`);
    } finally {
        isSavingSettings.value = false;
        isTestingProvider.value = false;
    }
};

const handleEscapeKeydown = (event) => {
    if (event.key !== "Escape" || !props.isOpen) {
        return;
    }

    if (openSettingsDropdownId.value) {
        closeSettingsDropdown();
        return;
    }

    emit("close");
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
        Database: h(
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
                h("ellipse", { cx: "12", cy: "5", rx: "9", ry: "3" }),
                h("path", { d: "M3 5v14a9 3 0 0 0 18 0V5" }),
                h("path", { d: "M3 12a9 3 0 0 0 18 0" }),
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
