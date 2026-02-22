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

        <div class="w-full max-w-md space-y-4 bg-card text-card-foreground p-6 rounded-xl border shadow-lg">
            <div class="text-center space-y-1">
                <div class="flex items-center justify-center mx-auto mb-2">
                    <img src="../assets/images/new-icon.png" class="w-20 h-20 object-contain" alt="VaultDB Icon" />
                </div>
                <p class="text-muted-foreground text-sm">VaultDB - Connect to your database to start managing data.</p>
            </div>

            <div class="space-y-3">
                <div class="space-y-3 animate-in fade-in slide-in-from-top-4 duration-300">
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
                                <option value="mariadb">MariaDB</option>
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
                        class="space-y-3 animate-in fade-in slide-in-from-top-4 duration-300">
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
                        <div class="grid grid-cols-2 gap-4">
                            <div class="space-y-2">
                                <label class="text-sm font-medium leading-none" for="user">User</label>
                                <input v-model="config.user" id="user" type="text"
                                    class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                            </div>
                            <div class="space-y-2">
                                <label class="text-sm font-medium leading-none" for="password">Password</label>
                                <div class="relative">
                                    <input v-model="config.password" id="password" :type="showPassword ? 'text' : 'password'"
                                        class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 pr-10 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                                    <button type="button" @click="showPassword = !showPassword"
                                        class="absolute inset-y-0 right-0 flex items-center pr-3 text-muted-foreground hover:text-foreground">
                                        <svg v-if="!showPassword" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-eye"><path d="M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0"/><circle cx="12" cy="12" r="3"/></svg>
                                        <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-eye-off"><path d="M9.88 9.88a3 3 0 1 0 4.24 4.24"/><path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68"/><path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61"/><line x1="2" x2="22" y1="2" y2="22"/></svg>
                                    </button>
                                </div>
                            </div>
                        </div>
                        <div class="space-y-2">
                            <label class="text-sm font-medium leading-none" for="database">Database Name</label>
                            <input v-model="config.database" id="database" type="text"
                                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                        </div>
                    </div>

                    <div v-else class="space-y-3 animate-in fade-in slide-in-from-top-4 duration-300">
                        <div class="space-y-2">
                            <label class="text-sm font-medium leading-none" for="filepath">Database File Path</label>
                            <div class="flex items-center space-x-2">
                                <input v-model="config.database" id="filepath" type="text" placeholder="/path/to/db.sqlite"
                                    class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                                <button type="button" @click="handleSelectSqliteFile"
                                    class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2">
                                    Browse...
                                </button>
                            </div>
                        </div>
                    </div>

                    <!-- SSH Tunnel Config -->
                    <div
                        class="space-y-3 animate-in fade-in slide-in-from-top-4 duration-300 border-t pt-3 border-border">
                        <div class="flex items-center space-x-2">
                            <input type="checkbox" id="sshEnabled" v-model="config.sshEnabled"
                                class="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary">
                            <label for="sshEnabled"
                                class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                Use SSH Tunnel
                            </label>
                        </div>

                        <div v-if="config.sshEnabled" class="space-y-3 pl-4 border-l-2 border-border/50 ml-1">
                            <div class="grid grid-cols-2 gap-4">
                                <div class="space-y-2">
                                    <label class="text-sm font-medium leading-none" for="sshHost">SSH Host</label>
                                    <input v-model="config.sshHost" id="sshHost" type="text"
                                        placeholder="ssh.example.com"
                                        class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                                </div>
                                <div class="space-y-2">
                                    <label class="text-sm font-medium leading-none" for="sshPort">SSH Port</label>
                                    <input v-model.number="config.sshPort" id="sshPort" type="number" placeholder="22"
                                        class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                                </div>
                            </div>
                            <div class="space-y-2">
                                <label class="text-sm font-medium leading-none" for="sshUser">SSH User</label>
                                <input v-model="config.sshUser" id="sshUser" type="text"
                                    class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                            </div>
                            <div class="space-y-2">
                                <label class="text-sm font-medium leading-none" for="sshPassword">SSH Password</label>
                                <div class="relative">
                                    <input v-model="config.sshPassword" id="sshPassword" :type="showSshPassword ? 'text' : 'password'"
                                        class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 pr-10 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                                    <button type="button" @click="showSshPassword = !showSshPassword"
                                        class="absolute inset-y-0 right-0 flex items-center pr-3 text-muted-foreground hover:text-foreground">
                                        <svg v-if="!showSshPassword" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-eye"><path d="M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0"/><circle cx="12" cy="12" r="3"/></svg>
                                        <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-eye-off"><path d="M9.88 9.88a3 3 0 1 0 4.24 4.24"/><path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68"/><path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61"/><line x1="2" x2="22" y1="2" y2="22"/></svg>
                                    </button>
                                </div>
                            </div>
                            <div class="space-y-2">
                                <label class="text-sm font-medium leading-none" for="sshKeyFile">SSH Key File
                                    (Optional)</label>
                                <input v-model="config.sshKeyFile" id="sshKeyFile" type="text"
                                    placeholder="/path/to/private_key"
                                    class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                            </div>
                        </div>
                    </div>

                    <div class="flex items-center space-x-2 animate-in fade-in slide-in-from-top-4 duration-300">
                        <input type="checkbox" id="readOnly" v-model="config.readOnly"
                            class="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary">
                        <label for="readOnly"
                            class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">Read
                            Only Mode</label>
                    </div>

                    <div class="flex gap-2 mt-4">
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
                                stroke-linejoin="round" class="lucide lucide-clock">
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
                                        <svg v-else-if="conn.type === 'mysql' || conn.type === 'mariadb'" xmlns="http://www.w3.org/2000/svg"
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
import { ConnectDB, TestConnection, SetReadOnly, SelectSqliteFile } from '../../wailsjs/go/main/App';

const props = defineProps<{
    activeConnections: any[]
}>();

const emit = defineEmits(['connected', 'connection-exists', 'connection-updated']);

const config = reactive({
    name: '',
    type: 'postgres',
    host: 'localhost',
    port: 5432,
    user: 'postgres',
    password: '',
    database: 'postgres',
    readOnly: false,
    sshEnabled: false,
    sshHost: '',
    sshPort: 22,
    sshUser: '',
    sshPassword: '',
    sshKeyFile: ''
});

const error = ref('');
const testSuccess = ref('');
const isLoading = ref(false);
const isTesting = ref(false);
const isDark = ref(false);
const showSavedModal = ref(false);
const savedConnections = ref<any[]>([]);
const showPassword = ref(false);
const showSshPassword = ref(false);

const toggleTheme = () => {
    isDark.value = !isDark.value;
    if (isDark.value) {
        document.documentElement.classList.add('dark');
    } else {
        document.documentElement.classList.remove('dark');
    }
};

const handleSelectSqliteFile = async () => {
    try {
        const filePath = await SelectSqliteFile();
        if (filePath) {
            config.database = filePath;
        }
    } catch (e) {
        console.error("Failed to select SQLite file", e);
    }
};

watch(() => config.type, (newType) => {
    if (newType === 'postgres') config.port = 5432;
    else if (newType === 'mysql' || newType === 'mariadb') config.port = 3306;
    else if (newType === 'mssql') config.port = 1433;
});

const isConfigEqual = (c1: any, c2: any) => {
    const type1 = (c1.type || '').toLowerCase();
    const type2 = (c2.type || '').toLowerCase();

    const host1 = (c1.host || '').toLowerCase();
    const host2 = (c2.host || '').toLowerCase();

    const port1 = parseInt(String(c1.port), 10);
    const port2 = parseInt(String(c2.port), 10);

    const user1 = c1.user || '';
    const user2 = c2.user || '';

    const db1 = c1.database || '';
    const db2 = c2.database || '';

    // ReadOnly is explicitly excluded from this comparison for the purpose of finding an existing connection
    // where only the readOnly status might change.

    const name1 = c1.name || '';
    const name2 = c2.name || '';

    // Port is irrelevant for SQLite
    if (type1 === 'sqlite') {
        return type1 === type2 &&
            db1 === db2 && // For sqlite, database is the filepath
            name1 === name2;
    }

    const sshEnabled1 = !!c1.sshEnabled;
    const sshEnabled2 = !!c2.sshEnabled;

    // If SSH is disabled for both, we don't care about other SSH fields matching
    if (!sshEnabled1 && !sshEnabled2) {
        return type1 === type2 &&
            host1 === host2 &&
            port1 === port2 &&
            user1 === user2 &&
            db1 === db2 &&
            name1 === name2;
    }

    const sshHost1 = (c1.sshHost || '').toLowerCase();
    const sshHost2 = (c2.sshHost || '').toLowerCase();
    const sshPort1 = parseInt(String(c1.sshPort || 22), 10);
    const sshPort2 = parseInt(String(c2.sshPort || 22), 10);
    const sshUser1 = c1.sshUser || '';
    const sshUser2 = c2.sshUser || '';
    const sshKeyFile1 = c1.sshKeyFile || '';
    const sshKeyFile2 = c2.sshKeyFile || '';

    return type1 === type2 &&
        host1 === host2 &&
        port1 === port2 &&
        user1 === user2 &&
        db1 === db2 &&
        name1 === name2 &&
        sshEnabled1 === sshEnabled2 &&
        sshHost1 === sshHost2 &&
        sshPort1 === sshPort2 &&
        sshUser1 === sshUser2 &&
        sshKeyFile1 === sshKeyFile2;
};

const connect = async () => {
    error.value = '';
    testSuccess.value = '';
    isLoading.value = true;

    // Check for existing connection
    const existing = props.activeConnections.find(c => isConfigEqual(c.config, config));

    if (existing) {
        // Check if ReadOnly status has changed
        const existingReadOnly = !!existing.config.readOnly;
        const newReadOnly = !!config.readOnly;

        if (existingReadOnly !== newReadOnly) {
            try {
                // Update backend
                await SetReadOnly(existing.id, newReadOnly);

                // Update frontend state
                emit('connection-updated', {
                    id: existing.id,
                    config: { ...existing.config, readOnly: newReadOnly }
                });

                // Update local storage if this was a saved connection? 
                // Logic for saved connections update is separate, but we want the ACTIVE session to update.
            } catch (e: any) {
                console.error("Failed to update ReadOnly status:", e);
                // We could show error, but we'll try to switch anyway
            }
        }

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
    const existsIndex = savedConnections.value.findIndex(c => isConfigEqual(c, newConfig));

    if (existsIndex === -1) {
        savedConnections.value.push(newConfig);
    } else {
        // Optional: Update the existing one (e.g. to move to top or update timestamp if we had one)
        // For now, just ensure we don't duplicate. 
        // We could also remove and push to end to show it as "most recently used" if the list is MRU.
        // Let's do that - move to end (bottom of list seems to be default order).
        // Actually, let's keep it simple. If it exists, do nothing or update details?
        // Since we compare EVERYTHING, it's identical.
        // So just do nothing.
    }
    localStorage.setItem('savedConnections', JSON.stringify(savedConnections.value));
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
