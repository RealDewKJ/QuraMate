<script lang="ts" setup>
defineProps<{
    imageUrl: string | null;
}>();

const emit = defineEmits<{
    (e: 'close'): void;
}>();
</script>

<template>
    <div v-if="imageUrl" class="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 p-10" @click="emit('close')">
        <div class="relative max-w-full max-h-full flex items-center justify-center animate-in zoom-in-95 duration-200" @click.stop>
            <img :src="imageUrl"
                class="max-w-full max-h-[85vh] rounded-lg shadow-2xl object-contain border-4 border-background bg-card">
            <button @click="emit('close')"
                class="absolute -top-4 -right-4 h-10 w-10 rounded-full bg-background border border-border shadow-lg flex items-center justify-center hover:bg-muted transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-x">
                    <path d="M18 6 6 18" />
                    <path d="m6 6 12 12" />
                </svg>
            </button>
            <div class="absolute -bottom-10 left-0 right-0 text-center">
                <a :href="imageUrl" target="_blank"
                    class="text-white hover:text-primary transition-colors text-sm font-medium underline underline-offset-4">Open in browser</a>
            </div>
        </div>
    </div>
</template>
