<script lang="ts" setup>
import AICopilotModal from '../AICopilotModal.vue';
import type { AiCopilotMode, AiCopilotModeOption } from '../../types/aiCopilot';

defineProps<{
    isOpen: boolean;
    mode: AiCopilotMode;
    modeOptions: AiCopilotModeOption[];
    prompt: string;
    backendLanguage: string;
    isLoading: boolean;
    result: string;
    error: string;
    latencyMs: number;
    suggestedSql: string;
}>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'run'): void;
    (e: 'apply-sql'): void;
    (e: 'update:mode', value: AiCopilotMode): void;
    (e: 'update:prompt', value: string): void;
    (e: 'update:backend-language', value: string): void;
}>();
</script>

<template>
    <AICopilotModal
        :is-open="isOpen"
        :mode="mode"
        :mode-options="modeOptions"
        :prompt="prompt"
        :backend-language="backendLanguage"
        :is-loading="isLoading"
        :result="result"
        :error="error"
        :latency-ms="latencyMs"
        :suggested-sql="suggestedSql"
        @close="emit('close')"
        @run="emit('run')"
        @apply-sql="emit('apply-sql')"
        @update:mode="emit('update:mode', $event)"
        @update:prompt="emit('update:prompt', $event)"
        @update:backend-language="emit('update:backend-language', $event)"
    />
</template>
