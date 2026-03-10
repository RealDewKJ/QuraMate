<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { CheckForUpdates, OpenDownloadURL, GetCurrentVersion, PerformUpdate } from '../../wailsjs/go/main/App';
import Toast from './Toast.vue';

interface UpdateInfo {
    available: boolean;
    currentVersion: string;
    latestVersion: string;
    releaseNotes: string;
    downloadURL: string;
    publishedAt: string;
}

const updateInfo = ref<UpdateInfo | null>(null);
const checking = ref(false);
const currentVersion = ref('');
const toastRef = ref<InstanceType<typeof Toast> | null>(null);

onMounted(async () => {
    currentVersion.value = await GetCurrentVersion();
    EventsOn('app:update-available', (info: UpdateInfo) => {
        updateInfo.value = info;
        showUpdateToast(info);
    });
});

onUnmounted(() => {
    EventsOff('app:update-available');
});

const showUpdateToast = (info: UpdateInfo) => {
    if (!toastRef.value) return;

    if (info.available) {
        toastRef.value.info(
            'Click "Update Now" to download and install the latest version.',
            `New update available v${info.latestVersion}`,
            15000, // Show longer for updates
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
            '\u2705 Up to date',
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
    if (updateInfo.value?.downloadURL) {
        const url = updateInfo.value.downloadURL;
        if (!toastRef.value) return;

        const loadingId = toastRef.value.info(
            'Downloading and installing the update...',
            '\uD83D\uDD04 Updating...',
            0 // keep open
        );

        try {
            await PerformUpdate(url);
            toastRef.value.remove(loadingId);
            toastRef.value.success(
                'The update has been installed successfully. Please restart the application to apply changes.',
                '\u2705 Update Complete',
                0 // Keep open until app restarts
            );
        } catch (err: any) {
            toastRef.value.remove(loadingId);
            toastRef.value.error(
                err.toString() || 'Failed to apply the update',
                '\u274C Update Failed',
                10000
            );
            // Fallback to opening browser
            await OpenDownloadURL(url);
        }
    }
};

defineExpose({ manualCheck, checking, currentVersion });
</script>

<template>
    <Toast ref="toastRef" />
</template>
