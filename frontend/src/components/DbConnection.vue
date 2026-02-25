<template>
  <div class="flex flex-col items-center justify-center min-h-screen p-4 transition-colors duration-300">
    <div class="absolute top-4 right-4">
      <button @click="toggleSettings(true)"
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
          class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 border border-destructive text-destructive hover:bg-destructive/10 h-10 px-4 py-2">
          Cancel
        </button>
      </div>

      <div v-else class="space-y-3">
        <div class="space-y-3 animate-in fade-in slide-in-from-top-4 duration-300">
          <div class="space-y-2">
            <label class="text-sm font-medium leading-none" for="connName">Connection Name (Optional)</label>
            <input v-model="config.name" id="connName" type="text" placeholder="My Database"
              class="flex min-h-[44px] h-auto w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50" />
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
              for="dbType">Database Type</label>
            <div class="relative">
              <select v-model="config.type" id="dbType"
                class="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 appearance-none">
                <option value="postgres">PostgreSQL</option>
                <option value="mysql">MySQL</option>
                <option value="mariadb">MariaDB</option>
                <option value="mssql">MSSQL</option>
                <option value="sqlite">SQLite</option>
                <option value="duckdb">DuckDB</option>
                <option value="greenplum">Greenplum</option>
                <option value="redshift">Redshift</option>
                <option value="cockroachdb">CockroachDB</option>
                <option value="databend">Databend</option>
                <option value="libsql">LibSQL</option>
              </select>
              <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-input-foreground">
                <svg class="h-4 w-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                  <path fill-rule="evenodd"
                    d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                    clip-rule="evenodd" />
                </svg>
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
                  class="flex min-h-[44px] h-auto w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50" />
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium leading-none" for="port">Port</label>
                <input v-model.number="config.port" id="port" type="number"
                  class="flex min-h-[44px] h-auto w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <label class="text-sm font-medium leading-none" for="user">User</label>
                <input v-model="config.user" id="user" type="text"
                  class="flex min-h-[44px] h-auto w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50" />
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium leading-none" for="password">Password</label>
                <div class="relative">
                  <input v-model="config.password" id="password" :type="showPassword ? 'text' : 'password'"
                    class="flex h-11 w-full rounded-md border border-input bg-background px-3 py-2 pr-10 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50" />
                  <button type="button" @click="togglePassword()"
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
                class="flex min-h-[44px] h-auto w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50" />
            </div>
          </div>

          <div v-else class="space-y-3 animate-in fade-in slide-in-from-top-4 duration-300">
            <div class="space-y-2">
              <label class="text-sm font-medium leading-none" for="filepath">Database File Path</label>
              <div class="flex items-center space-x-2">
                <input v-model="config.database" id="filepath" type="text" placeholder="/path/to/db.sqlite"
                  class="flex min-h-[44px] h-auto w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50" />
                <button type="button" @click="handleSelectSqliteFile"
                  class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2">
                  Browse...
                </button>
              </div>
            </div>
          </div>

          <div class="space-y-4 animate-in fade-in slide-in-from-top-4 duration-300"></div>

          <!-- SSH Tunnel Config -->
          <div class="space-y-3 animate-in fade-in slide-in-from-top-4 duration-300 border-t pt-3 border-border">
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
                    class="flex min-h-[44px] h-auto w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50" />
                </div>
                <div class="space-y-2">
                  <label class="text-sm font-medium leading-none" for="sshPort">SSH Port</label>
                  <input v-model.number="config.sshPort" id="sshPort" type="number" placeholder="22"
                    class="flex min-h-[44px] h-auto w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50" />
                </div>
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium leading-none" for="sshUser">SSH User</label>
                <input v-model="config.sshUser" id="sshUser" type="text"
                  class="flex min-h-[44px] h-auto w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50" />
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium leading-none" for="sshPassword">SSH Password</label>
                <div class="relative">
                  <input v-model="config.sshPassword" id="sshPassword" :type="showSshPassword ? 'text' : 'password'"
                    class="flex h-11 w-full rounded-md border border-input bg-background px-3 py-2 pr-10 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50" />
                  <button type="button" @click="toggleSshPassword()"
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
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium leading-none" for="sshKeyFile">SSH Key File (Optional)</label>
                <input v-model="config.sshKeyFile" id="sshKeyFile" type="text" placeholder="/path/to/private_key"
                  class="flex min-h-[44px] h-auto w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50" />
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
            <button v-if="!isTesting" @click="connect" :class="{ 'opacity-50 cursor-not-allowed': isLoading }"
              :disabled="isLoading"
              class="flex-1 inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-10 px-4 py-2">
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
            <button v-if="!isLoading" @click="testConnection" :class="{ 'opacity-50 cursor-not-allowed': isTesting }"
              :disabled="isLoading || isTesting"
              class="flex-1 inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2">
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
              class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 border border-destructive text-destructive hover:bg-destructive/10 h-10 px-4 py-2">
              Cancel
            </button>
            <button v-if="!isLoading && !isTesting" @click="toggleSavedModal(true)"
              class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 w-10">
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

      <div v-if="showSavedModal"
        class="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm transition-all duration-100 animate-in fade-in">
        <div ref="savedModalRef"
          class="fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-6 shadow-lg duration-200 sm:rounded-lg md:w-full animate-in fade-in zoom-in-95 slide-in-from-left-1/2 slide-in-from-top-48">
          <div class="flex flex-col space-y-1.5 text-center sm:text-left">
            <h2 class="text-lg font-semibold leading-none tracking-tight">
              Saved Connections
            </h2>
            <p class="text-sm text-muted-foreground">
              Select a connection to load its details.
            </p>
          </div>
          <div class="grid gap-4 py-4 max-h-[60vh] overflow-y-auto">
            <div v-if="savedConnections.length === 0" class="text-center text-muted-foreground text-sm py-8">
              No saved connections found.
            </div>
            <div v-else class="space-y-2">
              <div v-for="(conn, index) in savedConnections" :key="index"
                class="flex items-center justify-between p-3 rounded-lg border bg-card hover:bg-accent hover:text-accent-foreground transition-colors cursor-pointer group"
                @click="selectConnection(conn)">
                <div class="flex items-center gap-3 overflow-hidden">
                  <div class="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center text-primary">
                    <svg v-if="
                      conn.type === 'postgres' ||
                      conn.type === 'greenplum' ||
                      conn.type === 'redshift' ||
                      conn.type === 'cockroachdb'
                    " xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                      stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                      class="lucide lucide-database">
                      <ellipse cx="12" cy="5" rx="9" ry="3" />
                      <path d="M3 5V19A9 3 0 0 0 21 19V5" />
                      <path d="M3 12A9 3 0 0 0 21 12" />
                    </svg>
                    <svg v-else-if="
                      conn.type === 'mysql' ||
                      conn.type === 'mariadb' ||
                      conn.type === 'databend'
                    " xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                      stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                      class="lucide lucide-database">
                      <ellipse cx="12" cy="5" rx="9" ry="3" />
                      <path d="M3 5V19A9 3 0 0 0 21 19V5" />
                      <path d="M3 12A9 3 0 0 0 21 12" />
                    </svg>
                    <svg v-else-if="
                      conn.type === 'sqlite' ||
                      conn.type === 'duckdb' ||
                      conn.type === 'libsql'
                    " xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                      stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                      class="lucide lucide-file-code">
                      <path d="M10 12.5 8 15l2 2.5" />
                      <path d="m14 12.5 2 2.5-2 2.5" />
                      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                      <path d="M14 2v6h6" />
                    </svg>
                    <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                      fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                      class="lucide lucide-database">
                      <ellipse cx="12" cy="5" rx="9" ry="3" />
                      <path d="M3 5V19A9 3 0 0 0 21 19V5" />
                      <path d="M3 12A9 3 0 0 0 21 12" />
                    </svg>
                  </div>
                  <div class="flex flex-col truncate text-left">
                    <span class="text-sm font-medium truncate">{{
                      getConnectionLabel(conn)
                    }}</span>
                    <span class="text-xs text-muted-foreground truncate">{{ conn.host }}:{{ conn.port }}</span>
                  </div>
                </div>
                <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button @click.stop="editConnection(conn)"
                    class="p-2 rounded-md hover:bg-accent hover:text-accent-foreground transition-colors"
                    title="Edit Connection">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                      stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                      class="lucide lucide-pencil">
                      <path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z" />
                      <path d="m15 5 4 4" />
                    </svg>
                  </button>
                  <button @click.stop="removeConnection(index)"
                    class="p-2 rounded-md hover:bg-destructive hover:text-destructive-foreground transition-colors"
                    title="Delete Connection">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                      stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
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
            </div>
          </div>
          <div class="flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2">
            <button @click="toggleSavedModal(false)"
              class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2">
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
    <Toast ref="toastRef" />
    <SettingsDialog :isOpen="showSettings" @close="toggleSettings(false)" @save="handleSettingsSave" />
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, watch, onMounted, computed, shallowRef } from "vue";
import {
  useLocalStorage,
  useToggle,
  onClickOutside,
  onKeyStroke,
  watchImmediate,
} from "@vueuse/core";
import {
  ConnectDB,
  TestConnection,
  SetReadOnly,
  SelectSqliteFile,
  SaveCredential,
  DeleteCredential,
} from "../../wailsjs/go/main/App";
import SettingsDialog from "./SettingsDialog.vue";
import Toast from "./Toast.vue";

const toastRef = ref<InstanceType<typeof Toast> | null>(null);

const handleSettingsSave = () => {
  // Handled internally by SettingsDialog
};

const props = defineProps<{
  activeConnections: any[];
  pendingSqlFile?: { path: string; name: string; content: string } | null;
}>();

const emit = defineEmits([
  "connected",
  "connection-exists",
  "connection-updated",
]);

const config = reactive({
  id: "",
  name: "",
  type: "postgres",
  host: "localhost",
  port: 5432,
  user: "postgres",
  password: "",
  database: "postgres",
  readOnly: false,
  sshEnabled: false,
  sshHost: "",
  sshPort: 22,
  sshUser: "",
  sshPassword: "",
  sshKeyFile: "",
  encoding: "",
});

const resetConfig = () => {
  Object.assign(config, {
    id: "",
    name: "",
    type: "postgres",
    host: "localhost",
    port: 5432,
    user: "postgres",
    password: "",
    database: "postgres",
    readOnly: false,
    sshEnabled: false,
    sshHost: "",
    sshPort: 22,
    sshUser: "",
    sshPassword: "",
    sshKeyFile: "",
    encoding: "",
  });
};

const error = ref("");
const testSuccess = ref("");
const connectionToken = ref(0);
const [isLoading, toggleLoading] = useToggle(false);
const [isTesting, toggleTesting] = useToggle(false);
const [isQuickConnecting, toggleQuickConnecting] = useToggle(false);
const [showSettings, toggleSettings] = useToggle(false);
const [showSavedModal, toggleSavedModal] = useToggle(false);
const savedConnections = useLocalStorage<any[]>("savedConnections", []);
const [showPassword, togglePassword] = useToggle(false);
const [showSshPassword, toggleSshPassword] = useToggle(false);

const connectionLabel = computed(() => getConnectionLabel(config));

const savedModalRef = ref(null);
onClickOutside(savedModalRef, () => (showSavedModal.value = false));
onKeyStroke("Escape", () => {
  showSavedModal.value = false;
  showSettings.value = false;
});

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

watchImmediate(
  () => config.type,
  (newType) => {
    if (
      newType === "postgres" ||
      newType === "greenplum" ||
      newType === "redshift"
    )
      config.port = 5432;
    else if (newType === "cockroachdb") config.port = 26257;
    else if (newType === "mysql" || newType === "mariadb") config.port = 3306;
    else if (newType === "databend") config.port = 3307;
    else if (newType === "mssql") config.port = 1433;
  },
);

const isConfigEqual = (c1: any, c2: any) => {
  if (!c1 || !c2) return false;

  const type1 = (c1.type || "").toLowerCase();
  const type2 = (c2.type || "").toLowerCase();
  if (type1 !== type2) return false;

  const db1 = c1.database || "";
  const db2 = c2.database || "";
  const name1 = c1.name || "";
  const name2 = c2.name || "";

  if (type1 === "sqlite" || type1 === "duckdb" || type1 === "libsql") {
    return db1 === db2 && name1 === name2;
  }

  const host1 = (c1.host || "").toLowerCase();
  const host2 = (c2.host || "").toLowerCase();
  const port1 = parseInt(String(c1.port), 10);
  const port2 = parseInt(String(c2.port), 10);
  const user1 = c1.user || "";
  const user2 = c2.user || "";

  if (
    host1 !== host2 ||
    port1 !== port2 ||
    user1 !== user2 ||
    db1 !== db2 ||
    name1 !== name2
  ) {
    return false;
  }

  const sshEnabled1 = !!c1.sshEnabled;
  const sshEnabled2 = !!c2.sshEnabled;
  if (sshEnabled1 !== sshEnabled2) return false;

  if (sshEnabled1) {
    const sshHost1 = (c1.sshHost || "").toLowerCase();
    const sshHost2 = (c2.sshHost || "").toLowerCase();
    const sshPort1 = parseInt(String(c1.sshPort || 22), 10);
    const sshPort2 = parseInt(String(c2.sshPort || 22), 10);
    const sshUser1 = c1.sshUser || "";
    const sshUser2 = c2.sshUser || "";
    const sshKeyFile1 = c1.sshKeyFile || "";
    const sshKeyFile2 = c2.sshKeyFile || "";

    return (
      sshHost1 === sshHost2 &&
      sshPort1 === sshPort2 &&
      sshUser1 === sshUser2 &&
      sshKeyFile1 === sshKeyFile2
    );
  }

  return true;
};

const connect = async () => {
  performConnect();
};

const cancelConnection = () => {
  connectionToken.value++;
  isLoading.value = false;
  isTesting.value = false;
  isQuickConnecting.value = false;
  error.value = "Connection cancelled by user.";
  testSuccess.value = "";
};

const performConnect = async () => {
  error.value = "";
  testSuccess.value = "";
  isLoading.value = true;
  const currentToken = ++connectionToken.value;

  // Check for existing connection
  const existing = props.activeConnections.find((c) =>
    isConfigEqual(c.config, config),
  );

  if (existing) {
    if (currentToken !== connectionToken.value) return;
    // Check if ReadOnly status has changed
    if (!!existing.config.readOnly !== !!config.readOnly) {
      try {
        await SetReadOnly(existing.id, !!config.readOnly);
        emit("connection-updated", {
          id: existing.id,
          config: { ...existing.config, readOnly: !!config.readOnly },
        });
      } catch (e: any) {
        console.error("Failed to update ReadOnly status:", e);
      }
    }

    emit("connection-exists", existing.id);
    resetConfig();
    isLoading.value = false;
    return;
  }

  try {
    const result = await ConnectDB(config);
    if (currentToken !== connectionToken.value) return;

    if (result.id) {
      await saveConnection({ ...config });
      if (currentToken !== connectionToken.value) return;

      emit("connected", {
        id: result.id,
        name: config.name || getConnectionLabel(config),
        config: { ...config },
      });
      resetConfig();
    } else {
      error.value = result.error || "Unknown error";
    }
  } catch (e: any) {
    if (currentToken !== connectionToken.value) return;
    error.value = e.toString();
  } finally {
    if (currentToken === connectionToken.value) {
      isLoading.value = false;
      isQuickConnecting.value = false;
    }
  }
};

const testConnection = async () => {
  error.value = "";
  testSuccess.value = "";
  isTesting.value = true;
  const currentToken = ++connectionToken.value;

  try {
    const testConfig = { ...config };
    if (!testConfig.id) {
      testConfig.id = "";
    }
    const result = await TestConnection(testConfig);
    if (currentToken !== connectionToken.value) return;

    if (result === "Success") {
      testSuccess.value = "Connection successful!";
    } else {
      error.value = result;
    }
  } catch (e: any) {
    if (currentToken !== connectionToken.value) return;
    error.value = e.toString();
  } finally {
    if (currentToken === connectionToken.value) {
      isTesting.value = false;
    }
  }
};

const saveConnection = async (newConfig: any) => {
  if (!newConfig.id) {
    const existing = savedConnections.value.find((c) =>
      isConfigEqual(c, newConfig),
    );
    if (existing && existing.id) {
      newConfig.id = existing.id;
    } else {
      newConfig.id = crypto.randomUUID();
    }
  }

  config.id = newConfig.id;

  if (newConfig.password || newConfig.sshPassword) {
    try {
      await SaveCredential(
        newConfig.id,
        newConfig.password || "",
        newConfig.sshPassword || "",
      );
    } catch (e) {
      console.error("Failed to save credentials to keyring", e);
    }
  }

  const storageConfig = { ...newConfig, password: "", sshPassword: "" };

  const existsIndex = savedConnections.value.findIndex(
    (c) => c.id === storageConfig.id || isConfigEqual(c, storageConfig),
  );

  if (existsIndex === -1) {
    savedConnections.value.push(storageConfig);
  } else {
    savedConnections.value[existsIndex] = storageConfig;
  }
};

const removeConnection = async (index: number) => {
  const conn = savedConnections.value[index];
  if (conn && conn.id) {
    try {
      await DeleteCredential(conn.id);
    } catch (e) {
      console.error("Failed to delete credentials from keyring", e);
    }
  }
  savedConnections.value.splice(index, 1);
};

const selectConnection = (conn: any) => {
  config.name = conn.name || "";
  Object.assign(config, conn);
  // Explicitly clear password from the form view when loading from history
  config.password = "";
  config.sshPassword = "";

  showSavedModal.value = false;
  isQuickConnecting.value = true;
  connect();
};

const editConnection = (conn: any) => {
  config.name = conn.name || "";
  Object.assign(config, conn);
  // Explicitly clear password from the form view when loading from history
  config.password = "";
  config.sshPassword = "";

  showSavedModal.value = false;
};

const getConnectionLabel = (conn: any) => {
  if (conn.name) return conn.name;
  if (
    conn.type === "sqlite" ||
    conn.type === "duckdb" ||
    conn.type === "libsql"
  )
    return `${conn.type === "sqlite" ? "SQLite" : conn.type === "duckdb" ? "DuckDB" : "LibSQL"}: ${conn.database}`;
  if (conn.type === "greenplum")
    return `${conn.user}@${conn.host}:${conn.port}/${conn.database} (Greenplum)`;
  if (conn.type === "redshift")
    return `${conn.user}@${conn.host}:${conn.port}/${conn.database} (Redshift)`;
  if (conn.type === "cockroachdb")
    return `${conn.user}@${conn.host}:${conn.port}/${conn.database} (CockroachDB)`;
  if (conn.type === "databend")
    return `${conn.user}@${conn.host}:${conn.port}/${conn.database} (Databend)`;
  return `${conn.user}@${conn.host}:${conn.port}/${conn.database} (${conn.type})`;
};

onMounted(async () => {
  let needsSave = false;
  const connections = [...savedConnections.value];

  for (const conn of connections) {
    let modified = false;
    if (!conn.id) {
      conn.id = crypto.randomUUID();
      modified = true;
    }
    if (conn.password || conn.sshPassword) {
      try {
        await SaveCredential(
          conn.id,
          conn.password || "",
          conn.sshPassword || "",
        );
        conn.password = "";
        conn.sshPassword = "";
        modified = true;
      } catch (e) {
        console.error("Failed to migrate credentials to keyring", e);
      }
    }
    if (modified) needsSave = true;
  }

  if (needsSave) {
    savedConnections.value = connections;
  }
});
</script>

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
