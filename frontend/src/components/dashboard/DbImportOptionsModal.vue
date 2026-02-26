<script lang="ts" setup>
interface Props {
    isOpen: boolean;
    targetTable: string;
    filePath: string;
    isMssql: boolean;
    enableIdentityInsert: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'confirm'): void;
    (e: 'update:enable-identity-insert', value: boolean): void;
}>();
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
        <div class="bg-card w-full max-w-md rounded-lg shadow-lg border border-border p-6 space-y-4 animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-center justify-between">
                <h3 class="text-lg font-semibold text-foreground">Import Options</h3>
                <button @click="emit('close')" class="text-muted-foreground hover:text-foreground">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-x">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            <div class="space-y-4">
                <div class="flex flex-col gap-1 text-sm">
                    <span class="font-semibold text-muted-foreground">Target Table:</span>
                    <span>{{ targetTable }}</span>
                </div>
                <div class="flex flex-col gap-1 text-sm">
                    <span class="font-semibold text-muted-foreground">File:</span>
                    <span class="truncate" :title="filePath">{{ filePath }}</span>
                </div>

                <div v-if="isMssql" class="flex items-center space-x-2 pt-2">
                    <input
                        id="identityInsert"
                        type="checkbox"
                        :checked="enableIdentityInsert"
                        @change="emit('update:enable-identity-insert', ($event.target as HTMLInputElement).checked)"
                        class="h-4 w-4 rounded border-input bg-background text-primary focus:ring-primary"
                    >
                    <label for="identityInsert"
                        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                        Enable Identity Insert (SET IDENTITY_INSERT ON)
                    </label>
                </div>
            </div>

            <div class="flex justify-end gap-3 pt-4">
                <button @click="emit('close')"
                    class="px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground transition-colors">
                    Cancel
                </button>
                <button @click="emit('confirm')"
                    class="px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm">
                    Import
                </button>
            </div>
        </div>
    </div>
</template>
