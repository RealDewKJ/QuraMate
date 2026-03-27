<script lang="ts" setup>
import type { PropType } from 'vue';

import type { ConnectionConfig, ConnectionFieldErrors } from '../../composables/useConnectionForm';

defineProps({
  config: {
    type: Object as PropType<ConnectionConfig>,
    required: true,
  },
  fieldErrors: {
    type: Object as PropType<ConnectionFieldErrors>,
    required: true,
  },
  showSshPassword: {
    type: Boolean,
    required: true,
  },
  supportsSsh: {
    type: Boolean,
    required: true,
  },
});

const emit = defineEmits<{
  (e: 'toggleSshPassword'): void;
  (e: 'updateConfig', patch: Partial<ConnectionConfig>): void;
}>();

const handleTextInput = (key: keyof ConnectionConfig, event: Event) => {
  emit('updateConfig', { [key]: (event.target as HTMLInputElement).value } as Partial<ConnectionConfig>);
};

const handleNumberInput = (key: keyof ConnectionConfig, event: Event) => {
  emit('updateConfig', { [key]: Number((event.target as HTMLInputElement).value || 0) } as Partial<ConnectionConfig>);
};

const handleCheckboxInput = (key: keyof ConnectionConfig, event: Event) => {
  emit('updateConfig', { [key]: (event.target as HTMLInputElement).checked } as Partial<ConnectionConfig>);
};
</script>

<template>
  <div
    v-if="supportsSsh"
    class="space-y-3 border-t border-border pt-3 animate-in fade-in slide-in-from-top-4 duration-300"
  >
    <div class="flex items-center space-x-2">
      <input
        id="sshEnabled"
        type="checkbox"
        :checked="config.sshEnabled"
        class="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary"
        @change="handleCheckboxInput('sshEnabled', $event)"
      />
      <label
        for="sshEnabled"
        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
      >
        Use SSH Tunnel
      </label>
    </div>

    <div v-if="config.sshEnabled" class="ml-1 space-y-3 border-l-2 border-border/50 pl-4">
      <div class="grid grid-cols-2 gap-4">
        <div class="space-y-2">
          <label class="text-sm font-medium leading-none" for="sshHost">SSH Host</label>
          <input
            id="sshHost"
            type="text"
            placeholder="ssh.example.com"
            :value="config.sshHost"
            class="flex h-auto min-h-[44px] w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm text-foreground ring-offset-background transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            @input="handleTextInput('sshHost', $event)"
          />
          <p v-if="fieldErrors.sshHost" class="text-xs text-destructive">{{ fieldErrors.sshHost }}</p>
        </div>

        <div class="space-y-2">
          <label class="text-sm font-medium leading-none" for="sshPort">SSH Port</label>
          <input
            id="sshPort"
            type="number"
            placeholder="22"
            :value="config.sshPort"
            class="flex h-auto min-h-[44px] w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm text-foreground ring-offset-background transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            @input="handleNumberInput('sshPort', $event)"
          />
          <p v-if="fieldErrors.sshPort" class="text-xs text-destructive">{{ fieldErrors.sshPort }}</p>
        </div>
      </div>

      <div class="space-y-2">
        <label class="text-sm font-medium leading-none" for="sshUser">SSH User</label>
        <input
          id="sshUser"
          type="text"
          :value="config.sshUser"
          class="flex h-auto min-h-[44px] w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm text-foreground ring-offset-background transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          @input="handleTextInput('sshUser', $event)"
        />
        <p v-if="fieldErrors.sshUser" class="text-xs text-destructive">{{ fieldErrors.sshUser }}</p>
      </div>

      <div class="space-y-2">
        <label class="text-sm font-medium leading-none" for="sshPassword">SSH Password</label>
        <div class="relative">
          <input
            id="sshPassword"
            :type="showSshPassword ? 'text' : 'password'"
            :value="config.sshPassword"
            class="flex h-11 min-h-[44px] w-full rounded-md border border-input bg-background px-3 py-2.5 pr-10 text-sm text-foreground ring-offset-background transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            @input="handleTextInput('sshPassword', $event)"
          />
          <button
            type="button"
            aria-label="Toggle SSH password visibility"
            class="absolute inset-y-0 right-0 flex items-center pr-3 text-muted-foreground hover:text-foreground"
            @click="emit('toggleSshPassword')"
          >
            <svg
              v-if="!showSshPassword"
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
        <p v-if="fieldErrors.sshAuth" class="text-xs text-destructive">{{ fieldErrors.sshAuth }}</p>
      </div>

      <div class="space-y-2">
        <label class="text-sm font-medium leading-none" for="sshKeyFile">SSH Key File (Optional)</label>
        <input
          id="sshKeyFile"
          type="text"
          placeholder="/path/to/private_key"
          :value="config.sshKeyFile"
          class="flex h-auto min-h-[44px] w-full rounded-md border border-input bg-background px-3 py-2.5 text-sm text-foreground ring-offset-background transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          @input="handleTextInput('sshKeyFile', $event)"
        />
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
