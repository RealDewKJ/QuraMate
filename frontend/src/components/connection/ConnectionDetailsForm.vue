<script lang="ts" setup>
import type { PropType } from 'vue';

import type { ConnectionConfig, ConnectionFieldErrors, ConnectionInputMode } from '../../composables/useConnectionForm';
import type { DatabaseTypeOption, DatabaseTypeUiController } from '../../composables/useDbConnectionUi';
import DatabaseTypeOptionIcon from './DatabaseTypeOptionIcon.vue';

defineProps({
  config: {
    type: Object as PropType<ConnectionConfig>,
    required: true,
  },
  fieldErrors: {
    type: Object as PropType<ConnectionFieldErrors>,
    required: true,
  },
  connectionInputMode: {
    type: String as PropType<ConnectionInputMode>,
    required: true,
  },
  supportsConnectionString: {
    type: Boolean,
    required: true,
  },
  connectionString: {
    type: String,
    required: true,
  },
  connectionStringPlaceholder: {
    type: String,
    required: true,
  },
  databaseTypeOptions: {
    type: Array as PropType<ReadonlyArray<DatabaseTypeOption>>,
    required: true,
  },
  databaseTypeUi: {
    type: Object as PropType<DatabaseTypeUiController>,
    required: true,
  },
  inputModeOptions: {
    type: Array as PropType<ReadonlyArray<{ value: ConnectionInputMode; label: string }>>,
    required: true,
  },
  selectedDatabaseTypeLabel: {
    type: String,
    required: true,
  },
  showPassword: {
    type: Boolean,
    required: true,
  },
});

const emit = defineEmits<{
  (e: 'selectDatabaseType', type: ConnectionConfig['type']): void;
  (e: 'selectSqliteFile'): void;
  (e: 'togglePassword'): void;
  (e: 'updateConfig', patch: Partial<ConnectionConfig>): void;
  (e: 'updateConnectionInputMode', mode: ConnectionInputMode): void;
  (e: 'updateConnectionString', value: string): void;
}>();

const handleTextInput = (key: keyof ConnectionConfig, event: Event) => {
  emit('updateConfig', { [key]: (event.target as HTMLInputElement).value } as Partial<ConnectionConfig>);
};

const handleNumberInput = (key: keyof ConnectionConfig, event: Event) => {
  emit('updateConfig', { [key]: Number((event.target as HTMLInputElement).value || 0) } as Partial<ConnectionConfig>);
};
</script>

<template>
  <div class="space-y-3 animate-in fade-in slide-in-from-top-4 duration-300">
    <div class="space-y-2">
      <label class="text-sm font-medium leading-none" for="connName">Connection Name (Optional)</label>
      <input
        id="connName"
        type="text"
        placeholder="My Database"
        :value="config.name"
        class="flex h-auto min-h-[44px] w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm text-foreground ring-offset-background transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
        @input="handleTextInput('name', $event)"
      />
    </div>

    <div class="space-y-2">
      <label class="text-sm font-medium leading-none" for="dbType">Database Type</label>
      <div :ref="databaseTypeUi.setMenuRef" class="relative" data-db-type-menu>
        <button
          id="dbType"
          type="button"
          :ref="databaseTypeUi.setButtonRef"
          :aria-expanded="databaseTypeUi.isOpen ? 'true' : 'false'"
          aria-haspopup="listbox"
          class="flex min-h-[44px] h-auto w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2.5 text-sm text-foreground ring-offset-background transition-colors hover:bg-accent/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          @click="databaseTypeUi.toggleMenu"
          @keydown="databaseTypeUi.handleTriggerKeydown"
        >
          <span class="flex items-center gap-2">
            <DatabaseTypeOptionIcon :type="config.type" :label="selectedDatabaseTypeLabel" />
            <span>{{ selectedDatabaseTypeLabel }}</span>
          </span>
          <svg
            class="h-4 w-4 text-muted-foreground transition-transform duration-150"
            :class="databaseTypeUi.isOpen ? 'rotate-180' : ''"
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 20 20"
            fill="currentColor"
            aria-hidden="true"
          >
            <path
              fill-rule="evenodd"
              d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
              clip-rule="evenodd"
            />
          </svg>
        </button>

        <div
          v-if="databaseTypeUi.isOpen"
          role="listbox"
          aria-labelledby="dbType"
          class="absolute left-0 top-full z-50 mt-2 w-full overflow-hidden rounded-2xl border border-border/80 bg-popover/95 py-1 text-popover-foreground shadow-xl ring-1 ring-black/5 backdrop-blur animate-in fade-in zoom-in-95 duration-100"
        >
          <button
            v-for="(option, index) in databaseTypeOptions"
            :key="option.value"
            type="button"
            data-db-type-option
            role="option"
            :tabindex="databaseTypeUi.highlightedIndex === index ? 0 : -1"
            :aria-selected="config.type === option.value ? 'true' : 'false'"
            :disabled="option.disabled"
            class="flex w-full items-center justify-between px-3 py-2 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground"
            :class="[
              option.disabled ? 'cursor-not-allowed opacity-60 hover:bg-transparent hover:text-inherit' : '',
              databaseTypeUi.highlightedIndex === index ? 'bg-accent/50 text-accent-foreground' : '',
              config.type === option.value ? 'bg-accent/70 text-accent-foreground' : '',
            ]"
            @click="emit('selectDatabaseType', option.value)"
            @focus="databaseTypeUi.setHighlightedIndex(index)"
            @keydown="databaseTypeUi.handleOptionKeydown"
          >
            <div class="flex items-center gap-2">
              <DatabaseTypeOptionIcon :type="option.value" :label="option.label" />
              <span>{{ option.label }}</span>
              <span
                v-if="option.disabled"
                class="rounded-full border border-border/70 px-2 py-0.5 text-[10px] uppercase tracking-wide text-muted-foreground"
              >
                Soon
              </span>
            </div>

            <svg
              v-if="config.type === option.value"
              xmlns="http://www.w3.org/2000/svg"
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M20 6 9 17l-5-5" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <div v-if="supportsConnectionString" class="space-y-2">
      <label class="text-sm font-medium leading-none">Connection Input</label>
      <div class="grid grid-cols-2 gap-2 rounded-xl border border-border/70 bg-muted/30 p-1">
        <button
          v-for="mode in inputModeOptions"
          :key="mode.value"
          type="button"
          class="rounded-lg px-3 py-2 text-sm font-medium transition-colors"
          :class="connectionInputMode === mode.value ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:bg-background/70 hover:text-foreground'"
          @click="emit('updateConnectionInputMode', mode.value)"
        >
          {{ mode.label }}
        </button>
      </div>
    </div>

    <div
      v-if="connectionInputMode === 'fields' && !['sqlite', 'duckdb', 'libsql'].includes(config.type)"
      class="space-y-3 animate-in fade-in slide-in-from-top-4 duration-300"
    >
      <div
        v-if="config.type === 'supabase'"
        class="rounded-lg border border-primary/20 bg-primary/5 px-3 py-2 text-xs text-muted-foreground"
      >
        Use your Supabase direct Postgres connection details here, for example host
        <span class="font-mono text-foreground">db.&lt;project-ref&gt;.supabase.co</span>,
        port <span class="font-mono text-foreground">5432</span>, database
        <span class="font-mono text-foreground">postgres</span>.
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div class="space-y-2">
          <label class="text-sm font-medium leading-none" for="host">Host</label>
          <input
            id="host"
            type="text"
            placeholder="localhost"
            :value="config.host"
            class="flex h-auto min-h-[44px] w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm text-foreground ring-offset-background transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            @input="handleTextInput('host', $event)"
          />
          <p v-if="fieldErrors.host" class="text-xs text-destructive">{{ fieldErrors.host }}</p>
        </div>

        <div class="space-y-2">
          <label class="text-sm font-medium leading-none" for="port">Port</label>
          <input
            id="port"
            type="number"
            :value="config.port"
            class="flex h-auto min-h-[44px] w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm text-foreground ring-offset-background transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            @input="handleNumberInput('port', $event)"
          />
          <p v-if="fieldErrors.port" class="text-xs text-destructive">{{ fieldErrors.port }}</p>
        </div>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div class="space-y-2">
          <label class="text-sm font-medium leading-none" for="user">User</label>
          <input
            id="user"
            type="text"
            :value="config.user"
            class="flex h-auto min-h-[44px] w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm text-foreground ring-offset-background transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            @input="handleTextInput('user', $event)"
          />
          <p v-if="fieldErrors.user" class="text-xs text-destructive">{{ fieldErrors.user }}</p>
        </div>

        <div class="space-y-2">
          <label class="text-sm font-medium leading-none" for="password">Password</label>
          <div class="relative">
            <input
              id="password"
              :type="showPassword ? 'text' : 'password'"
              :value="config.password"
              class="flex h-11 min-h-[44px] w-full rounded-md border border-input bg-background px-3 py-2.5 pr-10 text-sm text-foreground ring-offset-background transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              @input="handleTextInput('password', $event)"
            />
            <button
              type="button"
              aria-label="Toggle password visibility"
              class="absolute inset-y-0 right-0 flex items-center pr-3 text-muted-foreground hover:text-foreground"
              @click="emit('togglePassword')"
            >
              <svg
                v-if="!showPassword"
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
                <path d="M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0" />
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
              >
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
        <input
          id="database"
          type="text"
          :value="config.database"
          class="flex h-auto min-h-[44px] w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm text-foreground ring-offset-background transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          @input="handleTextInput('database', $event)"
        />
        <p v-if="fieldErrors.database" class="text-xs text-destructive">{{ fieldErrors.database }}</p>
      </div>
    </div>

    <div
      v-else-if="connectionInputMode === 'fields'"
      class="space-y-3 animate-in fade-in slide-in-from-top-4 duration-300"
    >
      <div class="space-y-2">
        <label class="text-sm font-medium leading-none" for="filepath">Database File Path</label>
        <div class="flex items-center space-x-2">
          <input
            id="filepath"
            type="text"
            placeholder="/path/to/db.sqlite"
            :value="config.database"
            class="flex h-auto min-h-[44px] w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm text-foreground ring-offset-background transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            @input="handleTextInput('database', $event)"
          />
          <button
            type="button"
            class="inline-flex h-11 items-center justify-center whitespace-nowrap rounded-full border border-border/80 bg-background/95 px-4 text-sm font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            @click="emit('selectSqliteFile')"
          >
            Browse...
          </button>
        </div>
        <p v-if="fieldErrors.database" class="text-xs text-destructive">{{ fieldErrors.database }}</p>
      </div>
    </div>

    <div v-else class="space-y-3 animate-in fade-in slide-in-from-top-4 duration-300">
      <div class="space-y-2">
        <label class="text-sm font-medium leading-none" for="connectionString">Direct Connection String</label>
        <textarea
          id="connectionString"
          :value="connectionString"
          :placeholder="connectionStringPlaceholder"
          rows="4"
          class="min-h-[110px] w-full resize-y rounded-md border border-input bg-background px-3 py-2.5 font-mono text-xs leading-5 text-foreground ring-offset-background transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          @input="emit('updateConnectionString', ($event.target as HTMLTextAreaElement).value)"
        />
        <p class="text-xs text-muted-foreground">
          Paste a full connection string. QuraMate will parse it into the fields used for connect and test.
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
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
