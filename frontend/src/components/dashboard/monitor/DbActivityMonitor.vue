<script lang="ts" setup>
import { computed, ref } from 'vue';

interface ActivityTask {
    id: string;
    tabId: string;
    tabName: string;
    query: string;
    headBlock: string;
    blockedById?: string;
    blockedTaskIds: string[];
    blockingCount: number;
    isBlocker: boolean;
    startedAt: number;
    source: string;
    status: string;
}

interface MonitorSample {
    activeConnections: number;
    readQps: number;
    writeQps: number;
    cpuUsage: number;
    memoryUsage: number;
    bucket0to1: number;
    bucket1to5: number;
    bucketGt5: number;
}

interface Props {
    activityTaskCount: number;
    monitorRefreshRate: 3 | 5;
    latestMonitorSample: MonitorSample;
    connectionChartPoints: string;
    readQpsChartPoints: string;
    writeQpsChartPoints: string;
    longRunningTotal: number;
    activityTasksList: ActivityTask[];
    formatActivityTime: (time: number) => string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
    close: [];
    'kill-all': [];
    'focus-task': [task: ActivityTask];
    'kill-task': [taskId: string];
    'update:monitor-refresh-rate': [rate: 3 | 5];
}>();

const showKillAllConfirm = ref(false);

const confirmKillAll = () => {
    emit('kill-all');
    showKillAllConfirm.value = false;
};

const showQueryModal = ref(false);
const selectedQuery = ref('');

const openQueryModal = (query: string) => {
    selectedQuery.value = query;
    showQueryModal.value = true;
};

const showKillConfirm = ref(false);
const taskToKill = ref<ActivityTask | null>(null);
const inspectorRows = computed(() =>
    props.activityTasksList
        .filter((task) => task.blockedById || task.blockingCount > 0)
        .map((task) => ({
            ...task,
            blockedLabel: task.blockedById || '-',
            impactedLabel: task.blockingCount > 0 ? `${task.blockingCount} waiting` : 'No dependents',
        }))
);

const requestKillTask = (task: ActivityTask) => {
    taskToKill.value = task;
    showKillConfirm.value = true;
};

const confirmKillTask = () => {
    if (taskToKill.value) {
        emit('kill-task', taskToKill.value.id);
        showKillConfirm.value = false;
        taskToKill.value = null;
    }
};
</script>

<template>
    <div class="flex-1 overflow-hidden p-4 bg-background">
        <div class="h-full border border-border rounded-lg bg-card shadow-sm flex flex-col overflow-hidden">
            <div class="px-4 py-3 border-b border-border flex items-center justify-between">
                <h3 class="text-base font-semibold flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-activity text-primary">
                        <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
                    </svg>
                    Activity Monitor
                    <span class="text-sm text-muted-foreground font-normal">({{ activityTaskCount }} running)</span>
                </h3>
                <div class="flex items-center gap-2">
                    <label class="text-xs text-muted-foreground">Refresh</label>
                    <select :value="monitorRefreshRate"
                        @change="emit('update:monitor-refresh-rate', Number(($event.target as HTMLSelectElement).value) as 3 | 5)"
                        class="h-8 rounded-md border border-input bg-background px-2 text-xs">
                        <option :value="3">3s</option>
                        <option :value="5">5s</option>
                    </select>
                    <button v-if="activityTaskCount > 1" @click="showKillAllConfirm = true"
                        class="px-3 py-1.5 text-xs font-medium rounded-md border border-destructive text-destructive hover:bg-destructive/10 transition-colors">
                        Kill All
                    </button>
                    <button @click="emit('close')"
                        class="px-3 py-1.5 text-xs font-medium rounded-md border border-input bg-background hover:bg-accent transition-colors">
                        Back to Query
                    </button>
                </div>
            </div>

            <div class="p-4 border-b border-border bg-muted/10">
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
                    <div class="rounded-md border border-border bg-background p-3">
                        <div class="flex items-center justify-between mb-2">
                            <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Connection
                                Count (Real-time)</p>
                            <span class="text-sm font-semibold">{{ latestMonitorSample.activeConnections }}</span>
                        </div>
                        <svg width="100%" height="120" viewBox="0 0 360 110" preserveAspectRatio="none">
                            <polyline points="0,110 360,110" fill="none" stroke="hsl(var(--border))" stroke-width="1" />
                            <polyline :points="connectionChartPoints" fill="none" stroke="hsl(var(--primary))"
                                stroke-width="2.5" />
                        </svg>
                    </div>

                    <div class="rounded-md border border-border bg-background p-3">
                        <div class="flex items-center justify-between mb-2">
                            <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Query per
                                Second (QPS)</p>
                            <span class="text-xs text-muted-foreground">Read {{ latestMonitorSample.readQps }} / Write
                                {{ latestMonitorSample.writeQps }}</span>
                        </div>
                        <svg width="100%" height="120" viewBox="0 0 360 110" preserveAspectRatio="none">
                            <polyline points="0,110 360,110" fill="none" stroke="hsl(var(--border))" stroke-width="1" />
                            <polyline :points="readQpsChartPoints" fill="none" stroke="#3b82f6" stroke-width="2.5" />
                            <polyline :points="writeQpsChartPoints" fill="none" stroke="#ef4444" stroke-width="2.5" />
                        </svg>
                        <div class="mt-2 flex items-center gap-4 text-[11px] text-muted-foreground">
                            <span class="inline-flex items-center gap-1"><span
                                    class="h-2 w-2 rounded-full bg-blue-500"></span>Read</span>
                            <span class="inline-flex items-center gap-1"><span
                                    class="h-2 w-2 rounded-full bg-red-500"></span>Write</span>
                        </div>
                    </div>

                    <div class="rounded-md border border-border bg-background p-3">
                        <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-3">CPU & Memory
                            Usage</p>
                        <div class="space-y-3">
                            <div>
                                <div class="flex justify-between text-xs mb-1">
                                    <span>CPU</span>
                                    <span>{{ latestMonitorSample.cpuUsage }}%</span>
                                </div>
                                <div class="h-2 rounded-full bg-muted overflow-hidden">
                                    <div class="h-full bg-orange-500 transition-all duration-500"
                                        :style="{ width: `${latestMonitorSample.cpuUsage}%` }"></div>
                                </div>
                            </div>
                            <div>
                                <div class="flex justify-between text-xs mb-1">
                                    <span>Memory</span>
                                    <span>{{ latestMonitorSample.memoryUsage }}%</span>
                                </div>
                                <div class="h-2 rounded-full bg-muted overflow-hidden">
                                    <div class="h-full bg-emerald-500 transition-all duration-500"
                                        :style="{ width: `${latestMonitorSample.memoryUsage}%` }"></div>
                                </div>
                            </div>
                            <p class="text-[10px] text-muted-foreground">Estimated from active query load and local
                                runtime telemetry.</p>
                        </div>
                    </div>

                    <div class="rounded-md border border-border bg-background p-3">
                        <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-3">Long-running
                            Queries Distribution</p>
                        <div class="space-y-2">
                            <div>
                                <div class="flex justify-between text-xs mb-1"><span>0-1s</span><span>{{
                                    latestMonitorSample.bucket0to1 }}</span></div>
                                <div class="h-2 rounded-full bg-muted overflow-hidden">
                                    <div class="h-full bg-sky-500 transition-all duration-500"
                                        :style="{ width: `${(latestMonitorSample.bucket0to1 / longRunningTotal) * 100}%` }">
                                    </div>
                                </div>
                            </div>
                            <div>
                                <div class="flex justify-between text-xs mb-1"><span>1-5s</span><span>{{
                                    latestMonitorSample.bucket1to5 }}</span></div>
                                <div class="h-2 rounded-full bg-muted overflow-hidden">
                                    <div class="h-full bg-amber-500 transition-all duration-500"
                                        :style="{ width: `${(latestMonitorSample.bucket1to5 / longRunningTotal) * 100}%` }">
                                    </div>
                                </div>
                            </div>
                            <div>
                                <div class="flex justify-between text-xs mb-1"><span>>5s</span><span>{{
                                    latestMonitorSample.bucketGt5 }}</span></div>
                                <div class="h-2 rounded-full bg-muted overflow-hidden">
                                    <div class="h-full bg-red-500 transition-all duration-500"
                                        :style="{ width: `${(latestMonitorSample.bucketGt5 / longRunningTotal) * 100}%` }">
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="rounded-md border border-border bg-background p-3 lg:col-span-2">
                        <div class="flex items-center justify-between mb-3">
                            <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Lock / Blocking Inspector</p>
                            <span class="text-xs text-muted-foreground">{{ inspectorRows.length }} chain(s)</span>
                        </div>
                        <div v-if="inspectorRows.length > 0" class="space-y-2">
                            <div v-for="row in inspectorRows" :key="`block-${row.id}`" class="rounded-md border border-border px-3 py-2 text-xs">
                                <div class="flex items-center justify-between gap-3">
                                    <div class="font-mono text-foreground">{{ row.id }}</div>
                                    <div class="text-muted-foreground">{{ row.impactedLabel }}</div>
                                </div>
                                <div class="mt-1 flex items-center gap-2 text-muted-foreground">
                                    <span>Blocked by: <span class="font-mono text-foreground">{{ row.blockedLabel }}</span></span>
                                    <span v-if="row.isBlocker" class="rounded-full bg-destructive/10 px-2 py-0.5 text-[10px] font-semibold text-destructive">BLOCKER</span>
                                </div>
                            </div>
                        </div>
                        <div v-else class="text-xs text-muted-foreground">
                            No active blocking chains detected from server process metadata.
                        </div>
                    </div>
                </div>
            </div>

            <div class="flex-1 overflow-auto">
                <table class="w-full text-sm">
                    <thead class="bg-muted/40 text-muted-foreground sticky top-0">
                        <tr>
                            <th class="text-left px-3 py-2 font-medium">Task</th>
                            <th class="text-left px-3 py-2 font-medium">Tab</th>
                            <th class="text-left px-3 py-2 font-medium">Query</th>
                            <th class="text-left px-3 py-2 font-medium">Blocked By</th>
                            <th class="text-left px-3 py-2 font-medium">Started</th>
                            <th class="text-left px-3 py-2 font-medium">Status</th>
                            <th class="text-right px-3 py-2 font-medium">Action</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="task in activityTasksList" :key="task.id"
                            class="border-t border-border hover:bg-muted/30 cursor-pointer"
                            @click="emit('focus-task', task)">
                            <td class="px-3 py-2 align-top">
                                <div class="font-mono text-xs">{{ task.id.slice(0, 8) }}</div>
                                <div class="text-[11px] text-muted-foreground">{{ task.source }}</div>
                            </td>
                            <td class="px-3 py-2 align-top">{{ task.tabName }}</td>
                            <td class="px-3 py-2 align-top">
                                <div class="max-w-xs">
                                    <pre class="whitespace-pre-wrap break-words text-[11px] leading-relaxed font-mono text-foreground line-clamp-3 overflow-hidden"
                                        :title="task.query">{{ task.query || '(empty query)' }}</pre>
                                </div>
                                <button v-if="task.query" @click.stop="openQueryModal(task.query)"
                                    class="mt-1 text-[10px] text-primary hover:underline whitespace-nowrap">View Full
                                    SQL</button>
                            </td>
                            <td class="px-3 py-2 align-top text-xs font-mono"
                                :class="task.headBlock ? 'text-destructive font-bold' : 'text-muted-foreground'">{{
                                    task.headBlock || '-' }}</td>
                            <td class="px-3 py-2 align-top text-xs text-muted-foreground">{{
                                formatActivityTime(task.startedAt) }}</td>
                            <td class="px-3 py-2 align-top text-xs">
                                <span class="px-2 py-1 rounded-full text-[10px] font-semibold"
                                    :class="task.status === 'canceling...' ? 'bg-amber-500/10 text-amber-500' : 'bg-emerald-500/10 text-emerald-500'">
                                    {{ task.status }}
                                </span>
                            </td>
                            <td class="px-3 py-2 align-top text-right">
                                <button @click.stop="requestKillTask(task)"
                                    class="inline-flex items-center justify-center rounded-md text-xs font-medium px-2.5 py-1.5 border border-destructive text-destructive hover:bg-destructive/10 transition-colors"
                                    :disabled="task.status === 'canceling...'">
                                    {{ task.isBlocker ? 'Kill Blocker' : 'Kill' }}
                                </button>
                            </td>
                        </tr>
                    </tbody>
                </table>
                <div v-if="activityTasksList.length === 0"
                    class="py-10 text-center text-sm text-muted-foreground border-t border-border">
                    No active queries.
                </div>
            </div>
        </div>

        <!-- Kill All Confirmation Modal -->
        <div v-if="showKillAllConfirm"
            class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
            <div
                class="bg-card w-full max-w-md rounded-lg shadow-lg border border-border p-6 animate-in zoom-in-95 duration-200">
                <h3 class="text-lg font-semibold text-foreground mb-2 flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="text-destructive">
                        <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" />
                        <path d="M12 9v4" />
                        <path d="M12 17h.01" />
                    </svg>
                    Confirm Kill All
                </h3>
                <p class="text-sm text-muted-foreground mb-6">
                    Are you sure you want to kill <span class="font-bold text-foreground">{{ activityTaskCount }}</span>
                    active server processes? This action cannot be undone and may disrupt other users or application
                    services currently connected to the database.
                </p>
                <div class="flex justify-end gap-3">
                    <button @click="showKillAllConfirm = false"
                        class="px-4 py-2 text-sm font-medium rounded-md bg-muted hover:bg-accent text-foreground transition-colors border border-border">
                        Cancel
                    </button>
                    <button @click="confirmKillAll"
                        class="px-4 py-2 text-sm font-medium rounded-md bg-destructive hover:bg-destructive/90 text-destructive-foreground transition-colors shadow-sm">
                        Yes, Kill All Processes
                    </button>
                </div>
            </div>
        </div>

        <!-- Individual Kill Confirmation Modal -->
        <div v-if="showKillConfirm"
            class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
            @click.self="showKillConfirm = false">
            <div
                class="bg-card w-full max-w-md rounded-lg shadow-lg border border-border p-6 animate-in zoom-in-95 duration-200">
                <h3 class="text-lg font-semibold text-foreground mb-2 flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="text-destructive">
                        <circle cx="12" cy="12" r="10" />
                        <line x1="12" x2="12" y1="8" y2="12" />
                        <line x1="12" x2="12.01" y1="16" y2="16" />
                    </svg>
                    Confirm Kill Process
                </h3>
                <div class="text-sm text-muted-foreground mb-6 space-y-3">
                    <p>Are you sure you want to kill this database process?</p>
                    <div v-if="taskToKill"
                        class="bg-muted/50 p-3 rounded-md font-mono text-[11px] border border-border">
                        <div class="flex justify-between mb-1">
                            <span class="text-muted-foreground">Session ID:</span>
                            <span class="text-foreground text-right ml-2 truncate">{{ taskToKill.id }}</span>
                        </div>
                        <div class="flex justify-between mb-1">
                            <span class="text-muted-foreground">Source:</span>
                            <span class="text-foreground text-right ml-2 truncate">{{ taskToKill.source }}</span>
                        </div>
                        <div class="flex justify-between mb-1">
                            <span class="text-muted-foreground">Blocked by:</span>
                            <span class="text-foreground text-right ml-2 truncate">{{ taskToKill.blockedById || '-' }}</span>
                        </div>
                        <div class="flex justify-between mb-1">
                            <span class="text-muted-foreground">Blocking:</span>
                            <span class="text-foreground text-right ml-2 truncate">{{ taskToKill.blockingCount }} session(s)</span>
                        </div>
                        <div class="line-clamp-2 mt-2 pt-2 border-t border-border/50 text-foreground italic">
                            {{ taskToKill.query || '(empty query)' }}
                        </div>
                    </div>
                    <p class="text-xs text-destructive">
                        {{ taskToKill?.isBlocker ? 'This session is blocking others. Killing it may release waiting queries but can also abort in-flight work.' : 'This action will immediately terminate the selected session.' }}
                    </p>
                </div>
                <div class="flex justify-end gap-3">
                    <button @click="showKillConfirm = false"
                        class="px-4 py-2 text-sm font-medium rounded-md bg-muted hover:bg-accent text-foreground transition-colors border border-border">
                        Cancel
                    </button>
                    <button @click="confirmKillTask"
                        class="px-4 py-2 text-sm font-medium rounded-md bg-destructive hover:bg-destructive/90 text-destructive-foreground transition-colors shadow-sm">
                        Kill Process
                    </button>
                </div>
            </div>
        </div>

        <!-- Query Viewer Modal -->
        <div v-if="showQueryModal"
            class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
            @click.self="showQueryModal = false">
            <div
                class="bg-card w-full max-w-3xl rounded-lg shadow-lg border border-border p-6 flex flex-col max-h-[85vh] animate-in zoom-in-95 duration-200">
                <div class="flex justify-between items-center mb-4">
                    <h3 class="text-lg font-semibold text-foreground flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="text-primary">
                            <polyline points="4 7 4 4 20 4 20 7" />
                            <line x1="9" x2="15" y1="20" y2="20" />
                            <line x1="12" x2="12" y1="4" y2="20" />
                        </svg>
                        Full Query Text
                    </h3>
                    <button @click="showQueryModal = false" class="text-muted-foreground hover:text-foreground">
                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M18 6 6 18" />
                            <path d="m6 6 12 12" />
                        </svg>
                    </button>
                </div>
                <div class="flex-1 overflow-auto bg-muted/30 border border-border rounded-md p-4 mb-4 select-text">
                    <pre
                        class="whitespace-pre-wrap break-words text-[12px] font-mono text-foreground">{{ selectedQuery }}</pre>
                </div>
                <div class="flex justify-end">
                    <button @click="showQueryModal = false"
                        class="px-4 py-2 text-sm font-medium rounded-md bg-primary hover:bg-primary/90 text-primary-foreground transition-colors shadow-sm">
                        Close
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>
