<script lang="ts" setup>
defineProps<{
  canStart: boolean;
  primaryActionLabel: string;
}>();

const emit = defineEmits<{
  start: [];
}>();
</script>

<template>
  <section class="action-card">
    <button class="primary-button" type="button" :disabled="!canStart" @click="emit('start')">{{ primaryActionLabel }}</button>
  </section>
</template>
