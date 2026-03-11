<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { EventsOff, EventsOn } from '../../wailsjs/runtime/runtime';
import { CheckForUpdates, GetCurrentVersion, OpenDownloadURL, PerformUpdate } from '../../wailsjs/go/main/App';
import Toast from './Toast.vue';

interface UpdateInfo {
    available: boolean;
    currentVersion: string;
    latestVersion: string;
    releaseNotes: string;
    downloadURL: string;
    publishedAt: string;
}

interface UpdateProgress {
    stage: 'idle' | 'preparing' | 'downloading' | 'installing' | 'finalizing';
    percent: number;
    message: string;
}

const updateInfo = ref<UpdateInfo | null>(null);
const checking = ref(false);
const currentVersion = ref('');
const toastRef = ref<InstanceType<typeof Toast> | null>(null);
const isUpdating = ref(false);
const updateProgress = ref<UpdateProgress>({
    stage: 'idle',
    percent: 0,
    message: ''
});

const stageLabel = computed(() => {
    switch (updateProgress.value.stage) {
        case 'preparing':
            return 'Preparing';
        case 'downloading':
            return 'Downloading';
        case 'installing':
            return 'Installing';
        case 'finalizing':
            return 'Finalizing';
        default:
            return 'Updating';
    }
});

onMounted(async () => {
    currentVersion.value = await GetCurrentVersion();

    EventsOn('app:update-available', (info: UpdateInfo) => {
        updateInfo.value = info;
        showUpdateToast(info);
    });

    EventsOn('app:update-progress', (progress: UpdateProgress) => {
        isUpdating.value = true;
        updateProgress.value = {
            stage: progress.stage || 'downloading',
            percent: Number.isFinite(progress.percent) ? Math.max(0, Math.min(100, progress.percent)) : 0,
            message: progress.message || 'Updating application...'
        };
    });
});

onUnmounted(() => {
    EventsOff('app:update-available');
    EventsOff('app:update-progress');
});

const showUpdateToast = (info: UpdateInfo) => {
    if (!toastRef.value) return;

    if (info.available) {
        toastRef.value.info(
            `New version v${info.latestVersion} is ready.`,
            'Update available',
            15000,
            {
                label: 'Update Now',
                onClick: (id) => {
                    downloadUpdate();
                    toastRef.value?.remove(id);
                }
            }
        );
    } else {
        toastRef.value.success(
            `You are on the latest version v${info.currentVersion}.`,
            'Up to date',
            4000
        );
    }
};

const manualCheck = async () => {
    checking.value = true;
    try {
        const info = await CheckForUpdates();
        updateInfo.value = info;
        showUpdateToast(info);
    } finally {
        checking.value = false;
    }
};

const downloadUpdate = async () => {
    if (!updateInfo.value?.downloadURL) return;

    const url = updateInfo.value.downloadURL;
    isUpdating.value = true;
    updateProgress.value = {
        stage: 'preparing',
        percent: 3,
        message: 'Preparing update package...'
    };

    try {
        await PerformUpdate(url);
    } catch (err: unknown) {
        isUpdating.value = false;
        updateProgress.value = {
            stage: 'idle',
            percent: 0,
            message: ''
        };

        if (toastRef.value) {
            const message = err instanceof Error ? err.message : 'Failed to apply the update';
            toastRef.value.error(
                message,
                'Update failed',
                10000,
                {
                    label: 'Download manually',
                    onClick: () => {
                        OpenDownloadURL(url);
                    }
                }
            );
        }
    }
};

defineExpose({ manualCheck, checking, currentVersion });
</script>

<template>
    <Toast ref="toastRef" />

    <Teleport to="body">
        <Transition name="update-overlay">
            <div v-if="isUpdating" class="fixed inset-0 z-[120] flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
                <div
                    class="w-full max-w-xl rounded-2xl border border-border/70 bg-card shadow-2xl overflow-hidden update-shell">
                    <div class="h-1.5 bg-secondary/80 overflow-hidden">
                        <div class="h-full update-indeterminate"></div>
                    </div>

                    <div class="p-6 sm:p-8">
                        <div class="flex items-start gap-4">
                            <div class="update-spinner mt-1"></div>
                            <div class="min-w-0 flex-1">
                                <p class="text-xs uppercase tracking-[0.18em] text-muted-foreground">System Update</p>
                                <h3 class="mt-1 text-xl sm:text-2xl font-semibold text-foreground">
                                    {{ stageLabel }} QuraMate
                                </h3>
                                <p class="mt-2 text-sm text-muted-foreground">
                                    {{ updateProgress.message || 'Updating application...' }}
                                </p>
                            </div>
                        </div>

                        <div class="mt-6">
                            <div class="flex items-center justify-between text-xs text-muted-foreground">
                                <span>Progress</span>
                                <span>{{ updateProgress.percent }}%</span>
                            </div>
                            <div class="mt-2 h-2 rounded-full bg-secondary overflow-hidden">
                                <div class="h-full rounded-full update-progress-fill"
                                    :style="{ width: `${updateProgress.percent}%` }"></div>
                            </div>
                        </div>

                        <p class="mt-5 text-xs text-muted-foreground/90">
                            Please keep this window open while the update is being installed.
                        </p>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped>
.update-shell {
    background-image:
        radial-gradient(100% 120% at 90% -10%, hsl(var(--primary) / 0.17) 0%, transparent 55%),
        radial-gradient(80% 100% at -10% 110%, hsl(var(--primary) / 0.1) 0%, transparent 60%);
}

.update-spinner {
    width: 1.4rem;
    height: 1.4rem;
    border-radius: 9999px;
    border: 2px solid hsl(var(--primary) / 0.25);
    border-top-color: hsl(var(--primary));
    animation: spin 0.9s linear infinite;
}

.update-progress-fill {
    background: linear-gradient(90deg, hsl(var(--primary) / 0.75), hsl(var(--primary)));
    transition: width 280ms ease;
}

.update-indeterminate {
    width: 35%;
    background: linear-gradient(90deg, transparent, hsl(var(--primary) / 0.65), transparent);
    animation: shimmer 1.5s ease-in-out infinite;
}

.update-overlay-enter-active,
.update-overlay-leave-active {
    transition: opacity 0.24s ease;
}

.update-overlay-enter-from,
.update-overlay-leave-to {
    opacity: 0;
}

@keyframes spin {
    to {
        transform: rotate(360deg);
    }
}

@keyframes shimmer {
    0% {
        transform: translateX(-180%);
    }

    100% {
        transform: translateX(360%);
    }
}
</style>
