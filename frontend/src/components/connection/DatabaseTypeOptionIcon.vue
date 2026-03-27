<script lang="ts" setup>
import { computed } from 'vue';

import { getDatabaseTypeIconKind, getDatabaseTypeLogo } from '../../composables/useDbConnectionUi';

const props = defineProps<{
  type: string;
  label: string;
}>();

const iconKind = computed(() => getDatabaseTypeIconKind(props.type));
const logoUrl = computed(() => getDatabaseTypeLogo(props.type));
</script>

<template>
  <span
    class="inline-flex h-6 w-6 items-center justify-center rounded-full border border-border/70 bg-background/80 text-muted-foreground"
  >
    <img
      v-if="logoUrl"
      :src="logoUrl"
      :alt="`${label} logo`"
      class="h-4 w-4 object-contain"
    />
    <svg
      v-else-if="iconKind === 'postgres'"
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
      <ellipse cx="12" cy="5" rx="9" ry="3" />
      <path d="M3 5V19A9 3 0 0 0 21 19V5" />
      <path d="M3 12A9 3 0 0 0 21 12" />
    </svg>
    <svg
      v-else-if="iconKind === 'mysql'"
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
      <path d="M6 12h12" />
      <path d="M9 18h6" />
      <path d="M8 6h8" />
      <path d="M4 9h16v6H4z" />
    </svg>
    <svg
      v-else-if="iconKind === 'mssql'"
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
      <path d="M4 20V8l8-4 8 4v12" />
      <path d="M12 4v16" />
      <path d="M4 12h16" />
    </svg>
    <svg
      v-else-if="iconKind === 'file'"
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
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
      <path d="M14 2v6h6" />
    </svg>
    <svg
      v-else-if="iconKind === 'cloud'"
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
      <path d="M17.5 19a4.5 4.5 0 1 0-.9-8.91A6 6 0 1 0 6 17.5" />
      <path d="M12 12v9" />
      <path d="m8.5 17.5 3.5 3.5 3.5-3.5" />
    </svg>
    <svg
      v-else
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
      <ellipse cx="12" cy="5" rx="9" ry="3" />
      <path d="M3 5V19A9 3 0 0 0 21 19V5" />
      <path d="M3 12A9 3 0 0 0 21 12" />
    </svg>
  </span>
</template>
