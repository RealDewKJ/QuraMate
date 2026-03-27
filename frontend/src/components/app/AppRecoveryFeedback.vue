<script lang="ts" setup>
defineProps<{
  errorTitle: string;
  errors: string[];
  recoveringLabel: string;
  isRecoveringSession: boolean;
}>();
</script>

<template>
  <div
    v-if="isRecoveringSession"
    class="fixed bottom-4 right-4 z-40 rounded-md border border-border bg-card px-4 py-2 text-sm shadow"
  >
    {{ recoveringLabel }}
  </div>

  <div
    v-if="errors.length > 0"
    class="fixed bottom-4 left-4 z-40 max-w-lg rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive"
  >
    <div class="mb-1 font-medium">{{ errorTitle }}</div>
    <ul class="list-disc pl-5">
      <li v-for="item in errors" :key="item">{{ item }}</li>
    </ul>
  </div>
</template>
