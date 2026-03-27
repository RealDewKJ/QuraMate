<script lang="ts" setup>
defineProps<{
  isLoading: boolean;
  isTesting: boolean;
  testSuccess: string;
}>();

defineEmits<{
  (e: 'cancelConnection'): void;
  (e: 'connect'): void;
  (e: 'openSavedConnections'): void;
  (e: 'testConnection'): void;
}>();
</script>

<template>
  <div class="space-y-3">
    <div class="mt-4 flex gap-2">
      <button
        v-if="!isTesting"
        type="button"
        :disabled="isLoading"
        class="inline-flex h-11 flex-1 items-center justify-center whitespace-nowrap rounded-full bg-primary px-4 text-sm font-semibold text-primary-foreground shadow-sm transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
        @click="$emit('connect')"
      >
        <span v-if="isLoading" class="mr-2">
          <svg class="h-4 w-4 animate-spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </span>
        {{ isLoading ? 'Connecting...' : 'Connect' }}
      </button>

      <button
        v-if="!isLoading"
        type="button"
        :disabled="isLoading || isTesting"
        class="inline-flex h-11 flex-1 items-center justify-center whitespace-nowrap rounded-full border border-border/80 bg-background/95 text-sm font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
        @click="$emit('testConnection')"
      >
        <span v-if="isTesting" class="mr-2">
          <svg class="h-4 w-4 animate-spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </span>
        {{ isTesting ? 'Testing...' : 'Test Connection' }}
      </button>

      <button
        v-if="isLoading || isTesting"
        type="button"
        class="inline-flex h-11 items-center justify-center whitespace-nowrap rounded-full border border-destructive/60 bg-background/95 px-4 text-sm font-medium text-destructive shadow-sm transition-colors hover:bg-destructive/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
        @click="$emit('cancelConnection')"
      >
        Cancel
      </button>

      <button
        v-if="!isLoading && !isTesting"
        type="button"
        aria-label="Open saved connections"
        class="inline-flex h-11 w-11 items-center justify-center whitespace-nowrap rounded-full border border-border/80 bg-background/95 text-sm font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
        @click="$emit('openSavedConnections')"
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
        >
          <circle cx="12" cy="12" r="10" />
          <polyline points="12 6 12 12 16 14" />
        </svg>
      </button>
    </div>

    <div
      v-if="testSuccess"
      class="flex items-center gap-2 rounded-md bg-green-500/15 p-3 text-sm text-green-600 animate-in fade-in zoom-in duration-300"
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
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
        <polyline points="22 4 12 14.01 9 11.01" />
      </svg>
      {{ testSuccess }}
    </div>
  </div>
</template>
