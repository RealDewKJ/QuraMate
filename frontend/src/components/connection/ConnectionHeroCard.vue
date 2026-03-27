<script lang="ts" setup>
defineProps<{
  connectionLabel: string;
  error: string;
  isLoading: boolean;
  isQuickConnecting: boolean;
  pendingSqlFile?: { path: string; name: string; content: string } | null;
}>();

defineEmits<{
  (e: 'cancelConnection'): void;
}>();
</script>

<template>
  <div class="space-y-4">
    <div
      v-if="pendingSqlFile"
      class="w-full rounded-lg border border-primary/30 bg-primary/10 p-3 animate-in fade-in slide-in-from-top-4 duration-300"
    >
      <div class="flex items-center gap-3">
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
          class="shrink-0 text-primary"
        >
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
          <path d="M14 2v6h6" />
          <path d="M10 12.5 8 15l2 2.5" />
          <path d="m14 12.5 2 2.5-2 2.5" />
        </svg>
        <div class="min-w-0">
          <p class="truncate text-sm font-medium text-primary">{{ pendingSqlFile.name }}</p>
          <p class="text-xs text-muted-foreground">Connect to a database to open this SQL file</p>
        </div>
      </div>
    </div>

    <div
      v-if="isLoading && isQuickConnecting && !error"
      class="py-12 animate-in fade-in duration-500"
    >
      <div class="flex flex-col items-center justify-center space-y-4">
        <div class="relative h-16 w-16">
          <div class="absolute inset-0 rounded-full border-4 border-primary/20"></div>
          <div class="absolute inset-0 rounded-full border-4 border-primary border-t-transparent animate-spin"></div>
        </div>
        <div class="text-center">
          <p class="text-lg font-medium">Connecting to {{ connectionLabel }}...</p>
          <p class="text-sm text-muted-foreground">Please wait while we establish a secure connection.</p>
        </div>
        <button
          type="button"
          class="inline-flex h-11 items-center justify-center whitespace-nowrap rounded-full border border-destructive/60 bg-background/95 px-4 text-sm font-medium text-destructive shadow-sm transition-colors hover:bg-destructive/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          @click="$emit('cancelConnection')"
        >
          Cancel
        </button>
      </div>
    </div>

    <div v-else class="space-y-1 text-center">
      <div class="mb-2 flex items-center justify-center">
        <img src="../../assets/images/new-icon.png" class="h-20 w-20 object-contain" alt="QuraMate Icon" />
      </div>
      <p class="text-sm text-muted-foreground">
        QuraMate - Connect to your database to start managing data.
      </p>
    </div>
  </div>
</template>
