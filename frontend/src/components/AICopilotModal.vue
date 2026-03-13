<template>
    <div v-if="isOpen" class="fixed inset-0 z-[90] flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="$emit('close')"></div>
        <div
            class="relative z-[91] w-full max-w-5xl max-h-[86vh] bg-background border border-border rounded-xl shadow-2xl flex flex-col overflow-hidden">
            <div class="flex items-center justify-between px-5 py-4 border-b border-border bg-muted/20">
                <div>
                    <h3 class="text-base font-semibold">{{ t("common.aiCopilot.title") }}</h3>
                    <p class="text-xs text-muted-foreground">{{ t("common.aiCopilot.description") }}</p>
                </div>
                <button @click="$emit('close')"
                    class="h-8 w-8 inline-flex items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-foreground">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 p-4 overflow-auto">
                <div class="space-y-3">
                    <div class="grid gap-2">
                        <label class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Task</label>
                        <select v-model="selectedMode"
                            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
                            <option v-for="option in modeOptions" :key="option.value" :value="option.value">
                                {{ option.label }}
                            </option>
                        </select>
                    </div>
                    <div v-if="mode === 'backend_code'" class="grid gap-2">
                        <label class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Backend Language</label>
                        <select v-model="selectedBackendLanguage"
                            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
                            <option>Node.js</option>
                            <option>Python</option>
                            <option>Java</option>
                            <option>Go</option>
                            <option>C#</option>
                        </select>
                    </div>
                    <div class="grid gap-2">
                        <label class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Instruction</label>
                        <textarea v-model="instructionText" rows="8" placeholder="Describe what you want..."
                            class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"></textarea>
                    </div>
                    <div class="flex items-center gap-2">
                        <button @click="$emit('run')" :disabled="isLoading"
                            class="inline-flex items-center justify-center rounded-md text-sm font-medium bg-primary text-primary-foreground hover:bg-primary/90 h-9 px-4 py-2 disabled:pointer-events-none disabled:opacity-50">
                            <svg v-if="isLoading" class="animate-spin mr-2 h-4 w-4" xmlns="http://www.w3.org/2000/svg"
                                fill="none" viewBox="0 0 24 24">
                                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                                <path class="opacity-75" fill="currentColor" d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z" />
                            </svg>
                            {{ isLoading ? 'Running...' : 'Run Copilot' }}
                        </button>
                        <button @click="$emit('apply-sql')" :disabled="!hasSuggestedSql"
                            class="inline-flex items-center justify-center rounded-md text-sm font-medium border border-input bg-background hover:bg-accent h-9 px-4 py-2 disabled:pointer-events-none disabled:opacity-50">
                            Apply SQL to Editor
                        </button>
                    </div>
                    <p v-if="error" class="text-xs text-red-500 break-all">{{ error }}</p>
                    <p v-if="latencyMs > 0" class="text-xs text-muted-foreground">Latency: {{ latencyMs }} ms</p>
                </div>
                <div class="border border-border rounded-md bg-card min-h-[360px] flex flex-col">
                    <div class="px-3 py-2 border-b border-border text-xs font-semibold uppercase text-muted-foreground">AI Output</div>
                    <pre class="p-3 text-xs whitespace-pre-wrap break-words overflow-auto flex-1">{{ result || 'No output yet.' }}</pre>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import type { AiCopilotMode, AiCopilotModeOption } from '../types/aiCopilot';

const props = defineProps<{
    isOpen: boolean;
    mode: AiCopilotMode;
    modeOptions: AiCopilotModeOption[];
    prompt: string;
    backendLanguage: string;
    isLoading: boolean;
    result: string;
    error: string;
    latencyMs: number;
    hasSuggestedSql: boolean;
}>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'run'): void;
    (e: 'apply-sql'): void;
    (e: 'update:mode', value: AiCopilotMode): void;
    (e: 'update:prompt', value: string): void;
    (e: 'update:backend-language', value: string): void;
}>();
const { t } = useI18n({ useScope: 'global' });

const selectedMode = computed({
    get: () => props.mode,
    set: (value: AiCopilotMode) => emit('update:mode', value)
});

const instructionText = computed({
    get: () => props.prompt,
    set: (value: string) => emit('update:prompt', value)
});

const selectedBackendLanguage = computed({
    get: () => props.backendLanguage,
    set: (value: string) => emit('update:backend-language', value)
});

const handleEscapeKeydown = (event: KeyboardEvent) => {
    if (event.key === 'Escape' && props.isOpen) {
        emit('close');
    }
};

onMounted(() => {
    window.addEventListener('keydown', handleEscapeKeydown);
});

onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleEscapeKeydown);
});
</script>
