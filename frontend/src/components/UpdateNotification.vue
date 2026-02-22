<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { CheckForUpdates, OpenDownloadURL, GetCurrentVersion } from '../../wailsjs/go/main/App';
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
            info.releaseNotes ? `${info.releaseNotes.substring(0, 120)}${info.releaseNotes.length > 120 ? '...' : ''}` : 'A new version is available.',
            `🎉 New update available v${info.latestVersion}`,
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
            '✅ Up to date',
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
        await OpenDownloadURL(updateInfo.value.downloadURL);
    }
};

defineExpose({ manualCheck, checking, currentVersion });
</script>

<template>
    <Toast ref="toastRef" />
</template>
