<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { CheckForUpdates, OpenDownloadURL, GetCurrentVersion } from '../../wailsjs/go/main/App';

interface UpdateInfo {
    available: boolean;
    currentVersion: string;
    latestVersion: string;
    releaseNotes: string;
    downloadURL: string;
    publishedAt: string;
}

const show = ref(false);
const updateInfo = ref<UpdateInfo | null>(null);
const checking = ref(false);
const currentVersion = ref('');

onMounted(async () => {
    currentVersion.value = await GetCurrentVersion();
    EventsOn('app:update-available', (info: UpdateInfo) => {
        updateInfo.value = info;
        show.value = true;
    });
});

onUnmounted(() => {
    EventsOff('app:update-available');
});

const manualCheck = async () => {
    checking.value = true;
    try {
        const info = await CheckForUpdates();
        if (info.available) {
            updateInfo.value = info;
            show.value = true;
        } else {
            updateInfo.value = info;
            show.value = true;
            // Show "up to date" message briefly
            setTimeout(() => {
                if (!updateInfo.value?.available) {
                    show.value = false;
                }
            }, 4000);
        }
    } finally {
        checking.value = false;
    }
};

const downloadUpdate = async () => {
    if (updateInfo.value?.downloadURL) {
        await OpenDownloadURL(updateInfo.value.downloadURL);
    }
};

const dismiss = () => {
    show.value = false;
};

defineExpose({ manualCheck, checking, currentVersion });
</script>

<template>
    <!-- Notification Banner -->
    <Transition name="slide-down">
        <div v-if="show && updateInfo" class="update-notification">
            <!-- Update Available -->
            <div v-if="updateInfo.available" class="update-banner update-available">
                <div class="update-content">
                    <div class="update-icon">
                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                            <polyline points="7 10 12 15 17 10" />
                            <line x1="12" y1="15" x2="12" y2="3" />
                        </svg>
                    </div>
                    <div class="update-text">
                        <div class="update-title">
                            🎉 อัพเดตใหม่ <span class="version-badge">v{{ updateInfo.latestVersion }}</span>
                        </div>
                        <div v-if="updateInfo.releaseNotes" class="update-notes">
                            {{ updateInfo.releaseNotes.substring(0, 120) }}{{ updateInfo.releaseNotes.length > 120 ?
                            '...' : '' }}
                        </div>
                    </div>
                    <div class="update-actions">
                        <button class="btn-update" @click="downloadUpdate">
                            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round">
                                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                                <polyline points="7 10 12 15 17 10" />
                                <line x1="12" y1="15" x2="12" y2="3" />
                            </svg>
                            อัพเดตเลย
                        </button>
                        <button class="btn-dismiss" @click="dismiss" title="ปิด">
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round">
                                <line x1="18" y1="6" x2="6" y2="18" />
                                <line x1="6" y1="6" x2="18" y2="18" />
                            </svg>
                        </button>
                    </div>
                </div>
            </div>

            <!-- Already Up-to-Date -->
            <div v-else class="update-banner update-current">
                <div class="update-content">
                    <div class="update-icon text-green-400">
                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
                            <polyline points="22 4 12 14.01 9 11.01" />
                        </svg>
                    </div>
                    <div class="update-text">
                        <div class="update-title">✅ เวอร์ชันล่าสุดแล้ว (v{{ updateInfo.currentVersion }})</div>
                    </div>
                    <button class="btn-dismiss" @click="dismiss" title="ปิด">
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <line x1="18" y1="6" x2="6" y2="18" />
                            <line x1="6" y1="6" x2="18" y2="18" />
                        </svg>
                    </button>
                </div>
            </div>
        </div>
    </Transition>
</template>

<style scoped>
.update-notification {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 9999;
    display: flex;
    justify-content: center;
    pointer-events: none;
}

.update-banner {
    pointer-events: auto;
    margin: 12px;
    max-width: 680px;
    width: 100%;
    border-radius: 12px;
    backdrop-filter: blur(16px);
    box-shadow:
        0 8px 32px rgba(0, 0, 0, 0.3),
        0 0 0 1px rgba(255, 255, 255, 0.08),
        inset 0 1px 0 rgba(255, 255, 255, 0.06);
}

.update-available {
    background: linear-gradient(135deg, rgba(34, 197, 94, 0.15), rgba(59, 130, 246, 0.15));
    border: 1px solid rgba(34, 197, 94, 0.3);
}

.update-current {
    background: linear-gradient(135deg, rgba(34, 197, 94, 0.12), rgba(34, 197, 94, 0.06));
    border: 1px solid rgba(34, 197, 94, 0.25);
}

.update-content {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 14px 18px;
}

.update-icon {
    flex-shrink: 0;
    color: #22c55e;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border-radius: 10px;
    background: rgba(34, 197, 94, 0.1);
}

.update-text {
    flex: 1;
    min-width: 0;
}

.update-title {
    font-size: 14px;
    font-weight: 600;
    color: #e2e8f0;
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
}

.version-badge {
    display: inline-flex;
    align-items: center;
    padding: 2px 8px;
    font-size: 11px;
    font-weight: 700;
    border-radius: 6px;
    background: linear-gradient(135deg, #22c55e, #3b82f6);
    color: white;
    letter-spacing: 0.02em;
}

.update-notes {
    font-size: 12px;
    color: #94a3b8;
    margin-top: 4px;
    line-height: 1.4;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.update-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
}

.btn-update {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 8px 16px;
    font-size: 12px;
    font-weight: 600;
    color: white;
    background: linear-gradient(135deg, #22c55e, #16a34a);
    border: none;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s ease;
    white-space: nowrap;
    box-shadow: 0 2px 8px rgba(34, 197, 94, 0.3);
}

.btn-update:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 16px rgba(34, 197, 94, 0.4);
    background: linear-gradient(135deg, #16a34a, #15803d);
}

.btn-update:active {
    transform: translateY(0);
}

.btn-dismiss {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    color: #94a3b8;
    cursor: pointer;
    transition: all 0.15s ease;
}

.btn-dismiss:hover {
    background: rgba(255, 255, 255, 0.12);
    color: #e2e8f0;
}

/* Slide-down transition */
.slide-down-enter-active {
    transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.slide-down-leave-active {
    transition: all 0.25s ease-in;
}

.slide-down-enter-from {
    opacity: 0;
    transform: translateY(-100%);
}

.slide-down-leave-to {
    opacity: 0;
    transform: translateY(-100%);
}
</style>
