<script lang="ts" setup>
interface ConnectionTab {
  id: string;
  name: string;
}

const props = defineProps<{
  activeTabId: string | null;
  connections: ConnectionTab[];
  currentVersion: string;
  isCheckingUpdates: boolean;
  newConnectionLabel: string;
  checkForUpdatesLabel: string;
}>();

const emit = defineEmits<{
  (e: 'requestUpdateCheck'): void;
  (e: 'openHomeTab'): void;
  (e: 'openConnectionTab', id: string): void;
}>();
</script>

<template>
  <div class="z-10 flex items-center border-b border-border bg-background/80 backdrop-blur-sm">
    <div class="flex flex-1 items-center overflow-x-auto">
      <button
        type="button"
        class="flex items-center border-r border-border px-4 py-3 text-sm font-medium transition-colors hover:bg-muted/50 focus:outline-none"
        :class="{
          'border-b-2 border-b-primary bg-background text-primary': props.activeTabId === null,
          'text-muted-foreground': props.activeTabId !== null,
        }"
        @click="emit('openHomeTab')"
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
          class="lucide lucide-plus mr-2"
        >
          <path d="M5 12h14" />
          <path d="M12 5v14" />
        </svg>
        {{ newConnectionLabel }}
      </button>

      <button
        v-for="connection in props.connections"
        :key="connection.id"
        type="button"
        class="group flex min-w-[150px] max-w-[250px] items-center border-r border-border px-4 py-3 text-sm font-medium transition-colors hover:bg-muted/50"
        :class="{
          'border-b-2 border-b-primary bg-background text-primary': props.activeTabId === connection.id,
          'text-muted-foreground': props.activeTabId !== connection.id,
        }"
        @click="emit('openConnectionTab', connection.id)"
      >
        <div class="mr-2 flex items-center truncate">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="lucide lucide-database mr-2 shrink-0"
          >
            <ellipse cx="12" cy="5" rx="9" ry="3" />
            <path d="M3 5V19A9 3 0 0 0 21 19V5" />
            <path d="M3 12A9 3 0 0 0 21 12" />
          </svg>
          <span class="truncate">{{ connection.name }}</span>
        </div>
      </button>
    </div>

    <div class="flex shrink-0 items-center gap-2 border-l border-border px-3">
      <span v-if="props.currentVersion" class="font-mono text-[11px] text-muted-foreground">
        v{{ props.currentVersion }}
      </span>
      <button
        type="button"
        class="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted/60 hover:text-primary"
        :class="{ 'animate-spin': props.isCheckingUpdates }"
        :title="checkForUpdatesLabel"
        @click="emit('requestUpdateCheck')"
      >
        <svg
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
          <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
          <path d="M3 3v5h5" />
          <path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16" />
          <path d="M21 21v-5h-5" />
        </svg>
      </button>
    </div>
  </div>
</template>
