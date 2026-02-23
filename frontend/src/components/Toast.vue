<template>
    <Teleport to="body">
        <TransitionGroup name="toast" tag="div"
            class="fixed top-4 right-4 z-[100] flex flex-col gap-2 pointer-events-none max-w-sm w-full">
            <div v-for="toast in toasts" :key="toast.id"
                class="pointer-events-auto rounded-lg border shadow-lg p-4 flex items-start gap-3 animate-in slide-in-from-top-2 duration-300"
                :class="toastClass(toast.type)">
                <!-- Icon -->
                <div class="flex-shrink-0 mt-0.5">
                    <!-- Success -->
                    <svg v-if="toast.type === 'success'" xmlns="http://www.w3.org/2000/svg" width="18" height="18"
                        viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                        stroke-linejoin="round" class="text-emerald-500">
                        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
                        <path d="m9 11 3 3L22 4" />
                    </svg>
                    <!-- Error -->
                    <svg v-else-if="toast.type === 'error'" xmlns="http://www.w3.org/2000/svg" width="18" height="18"
                        viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                        stroke-linejoin="round" class="text-red-500">
                        <circle cx="12" cy="12" r="10" />
                        <path d="m15 9-6 6" />
                        <path d="m9 9 6 6" />
                    </svg>
                    <!-- Info -->
                    <svg v-else xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"
                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                        stroke-linejoin="round" class="text-blue-500">
                        <circle cx="12" cy="12" r="10" />
                        <path d="M12 16v-4" />
                        <path d="M12 8h.01" />
                    </svg>
                </div>
                <!-- Content -->
                <div class="flex-1 min-w-0">
                    <p v-if="toast.title" class="text-sm font-semibold text-foreground">{{ toast.title }}</p>
                    <p class="text-sm text-muted-foreground" :class="{ 'mt-0.5': toast.title }">{{ toast.message }}</p>
                    <button v-if="toast.action" @click="toast.action.onClick(toast.id)"
                        class="mt-2 text-xs font-semibold px-3 py-1.5 bg-primary/10 hover:bg-primary/20 text-primary rounded-md transition-colors w-fit cursor-pointer border border-primary/20">
                        {{ toast.action.label }}
                    </button>
                </div>
                <!-- Close -->
                <button @click="removeToast(toast.id)"
                    class="flex-shrink-0 text-muted-foreground hover:text-foreground transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>
        </TransitionGroup>
    </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue';

interface Toast {
    id: number;
    type: 'success' | 'error' | 'info';
    title?: string;
    message: string;
    action?: { label: string; onClick: (id: number) => void };
}

const toasts = ref<Toast[]>([]);
let nextId = 0;

const toastClass = (type: string) => {
    switch (type) {
        case 'success':
            return 'bg-card border-emerald-500/30';
        case 'error':
            return 'bg-card border-red-500/30';
        default:
            return 'bg-card border-blue-500/30';
    }
};

const addToast = (type: Toast['type'], message: string, title?: string, duration = 4000, action?: { label: string; onClick: (id: number) => void }) => {
    const id = nextId++;
    toasts.value.push({ id, type, title, message, action });
    if (duration > 0) {
        setTimeout(() => removeToast(id), duration);
    }
    return id;
};

const removeToast = (id: number) => {
    const idx = toasts.value.findIndex(t => t.id === id);
    if (idx !== -1) toasts.value.splice(idx, 1);
};

// Expose methods for parent components
defineExpose({
    success: (message: string, title?: string, duration?: number, action?: { label: string; onClick: (id: number) => void }) => addToast('success', message, title, duration ?? 4000, action),
    error: (message: string, title?: string, duration?: number, action?: { label: string; onClick: (id: number) => void }) => addToast('error', message, title, duration ?? 6000, action),
    info: (message: string, title?: string, duration?: number, action?: { label: string; onClick: (id: number) => void }) => addToast('info', message, title, duration ?? 4000, action),
    remove: removeToast
});
</script>

<style scoped>
.toast-enter-active {
    transition: all 0.3s ease-out;
}

.toast-leave-active {
    transition: all 0.2s ease-in;
}

.toast-enter-from {
    opacity: 0;
    transform: translateX(100%);
}

.toast-leave-to {
    opacity: 0;
    transform: translateX(100%);
}

.toast-move {
    transition: transform 0.3s ease;
}
</style>
