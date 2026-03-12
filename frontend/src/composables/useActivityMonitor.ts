import { computed, reactive, ref, watch } from 'vue';
import type { Ref } from 'vue';

import type { QueryTab } from '../types/dashboard';

export interface ActivityTask {
    id: string;      // maps to sessionId
    tabId: string;   // unused now, but keep for compatibility
    tabName: string; // maps to user/host
    query: string;   // maps to QueryText
    headBlock: string;
    blockedById?: string;
    blockedTaskIds: string[];
    blockingCount: number;
    isBlocker: boolean;
    startedAt: number;
    source: string;  // maps to Database/Command
    status: string;  // maps to Status/State
}

type QueryKind = 'read' | 'write' | 'other';

interface QueryEvent {
    timestamp: number;
    kind: QueryKind;
}

interface MonitorSample {
    timestamp: number;
    activeConnections: number;
    readQps: number;
    writeQps: number;
    cpuUsage: number;
    memoryUsage: number;
    bucket0to1: number;
    bucket1to5: number;
    bucketGt5: number;
}

import { GetServerProcesses, KillServerProcess } from '../../wailsjs/go/main/App';

interface UseActivityMonitorOptions {
    tabs: Ref<QueryTab[]>;
    activeTabId: Ref<string | null>;
    isActivityMonitorOpen: Ref<boolean>;
    connectionId: Ref<string>;
    onCancelError?: (message: string) => void;
    onCancelSuccess?: (message: string) => void;
}

export function useActivityMonitor(options: UseActivityMonitorOptions) {
    const activityTasksList = ref<ActivityTask[]>([]);
    const activityTaskCount = computed(() => activityTasksList.value.length);

    const monitorRefreshRate = ref<3 | 5>(3);
    const queryEvents = ref<QueryEvent[]>([]);
    const monitorHistory = ref<MonitorSample[]>([]);
    let monitorTimer: number | null = null;

    const getQueryKind = (query: string): QueryKind => {
        const keyword = (query || '').trim().split(/\s+/)[0]?.toUpperCase();
        if (keyword === 'SELECT' || keyword === 'WITH' || keyword === 'SHOW' || keyword === 'DESCRIBE' || keyword === 'EXPLAIN') {
            return 'read';
        }
        if (keyword === 'INSERT' || keyword === 'UPDATE' || keyword === 'DELETE' || keyword === 'MERGE' || keyword === 'REPLACE' || keyword === 'TRUNCATE' || keyword === 'DROP' || keyword === 'ALTER' || keyword === 'CREATE') {
            return 'write';
        }
        return 'other';
    };

    const clamp = (val: number, min: number, max: number) => Math.max(min, Math.min(max, val));

    const makeEmptySample = (): MonitorSample => ({
        timestamp: Date.now(),
        activeConnections: 0,
        readQps: 0,
        writeQps: 0,
        cpuUsage: 0,
        memoryUsage: 0,
        bucket0to1: 0,
        bucket1to5: 0,
        bucketGt5: 0
    });

    const latestMonitorSample = computed(() =>
        monitorHistory.value.length > 0 ? monitorHistory.value[monitorHistory.value.length - 1] : makeEmptySample()
    );

    const isConnectionNotFoundError = (err: unknown): boolean => {
        const message = err && typeof err === 'object' && 'message' in err
            ? String((err as { message?: unknown }).message ?? '')
            : String(err ?? '');
        return message.toLowerCase().includes('connection not found');
    };



    const fetchServerProcesses = async () => {
        if (!options.connectionId.value) return;
        try {
            const processes: any = await GetServerProcesses(options.connectionId.value);
            if (processes && Array.isArray(processes)) {
                // Update query events based on current processes
                const now = Date.now();
                
                const newTasks: ActivityTask[] = processes.map((p: any) => {
                    const queryText = p.queryText || '';
                    queryEvents.value.push({
                        timestamp: now,
                        kind: getQueryKind(queryText)
                    });

                    const blockedById = p.headBlock ? String(p.headBlock) : '';
                    return {
                        id: String(p.sessionId),
                        tabId: '',
                        tabName: `${p.user} @ ${p.host}`,
                        query: queryText,
                        headBlock: blockedById,
                        blockedById: blockedById || undefined,
                        blockedTaskIds: [],
                        blockingCount: 0,
                        isBlocker: false,
                        startedAt: now - (p.elapsedTime || 0), // Estimate start time
                        source: `${p.database} [${p.command}]`,
                        status: p.status || p.state || 'running'
                    };
                });

                const taskMap = new Map(newTasks.map((task) => [task.id, task]));
                for (const task of newTasks) {
                    if (task.blockedById && taskMap.has(task.blockedById)) {
                        const blocker = taskMap.get(task.blockedById)!;
                        blocker.blockedTaskIds.push(task.id);
                    }
                }
                for (const task of newTasks) {
                    task.blockingCount = task.blockedTaskIds.length;
                    task.isBlocker = task.blockingCount > 0;
                }

                activityTasksList.value = newTasks.sort((a, b) => {
                    if (a.isBlocker !== b.isBlocker) {
                        return a.isBlocker ? -1 : 1;
                    }
                    return b.startedAt - a.startedAt;
                });
            } else {
                activityTasksList.value = [];
            }
        } catch (e) {
            // Avoid noisy logs during disconnect/reconnect windows.
            if (isConnectionNotFoundError(e)) {
                activityTasksList.value = [];
                return;
            }
            console.error('Failed to fetch server processes', e);
        }
    };

    const formatActivityTime = (time: number) => {
        return new Date(time).toLocaleTimeString();
    };

    const collectMonitorSample = async () => {
        await fetchServerProcesses();
        
        const now = Date.now();
        const windowMs = monitorRefreshRate.value * 1000;

        queryEvents.value = queryEvents.value.filter((ev) => now - ev.timestamp <= 10 * 60 * 1000);
        const recentEvents = queryEvents.value.filter((ev) => now - ev.timestamp <= windowMs);

        const readCount = recentEvents.filter((ev) => ev.kind === 'read').length;
        const writeCount = recentEvents.filter((ev) => ev.kind === 'write').length;

        const readQps = Number((readCount / monitorRefreshRate.value).toFixed(2));
        const writeQps = Number((writeCount / monitorRefreshRate.value).toFixed(2));

        const durationsSec = activityTasksList.value.map((task) => (now - task.startedAt) / 1000);
        const bucket0to1 = durationsSec.filter((sec) => sec <= 1).length;
        const bucket1to5 = durationsSec.filter((sec) => sec > 1 && sec <= 5).length;
        const bucketGt5 = durationsSec.filter((sec) => sec > 5).length;

        const cpuUsage = clamp(Math.round(activityTaskCount.value * 18 + (readQps + writeQps) * 22 + (bucketGt5 > 0 ? 15 : 0)), 0, 100);

        let memoryUsage = clamp(Math.round(30 + activityTaskCount.value * 8 + bucketGt5 * 5), 0, 100);
        const perfAny = performance as any;
        if (perfAny && perfAny.memory && perfAny.memory.jsHeapSizeLimit > 0) {
            memoryUsage = clamp(
                Math.round((perfAny.memory.usedJSHeapSize / perfAny.memory.jsHeapSizeLimit) * 100),
                0,
                100
            );
        }

        const sample: MonitorSample = {
            timestamp: now,
            activeConnections: activityTaskCount.value,
            readQps,
            writeQps,
            cpuUsage,
            memoryUsage,
            bucket0to1,
            bucket1to5,
            bucketGt5
        };

        monitorHistory.value.push(sample);
        if (monitorHistory.value.length > 60) {
            monitorHistory.value.shift();
        }
    };

    const startMonitorTimer = () => {
        if (monitorTimer) {
            window.clearInterval(monitorTimer);
        }
        collectMonitorSample();
        monitorTimer = window.setInterval(async () => {
            await collectMonitorSample();
        }, monitorRefreshRate.value * 1000);
    };

    const stopMonitorTimer = () => {
        if (monitorTimer) {
            window.clearInterval(monitorTimer);
            monitorTimer = null;
        }
    };

    const toPolylinePoints = (values: number[], width: number, height: number, maxValue: number) => {
        if (!values.length) return '';
        const safeMax = Math.max(maxValue, 1);
        const step = values.length > 1 ? width / (values.length - 1) : 0;
        return values
            .map((value, index) => {
                const x = Math.round(index * step);
                const y = Math.round(height - (value / safeMax) * height);
                return `${x},${y}`;
            })
            .join(' ');
    };

    const connectionSeries = computed(() => monitorHistory.value.map((s) => s.activeConnections));
    const readQpsSeries = computed(() => monitorHistory.value.map((s) => s.readQps));
    const writeQpsSeries = computed(() => monitorHistory.value.map((s) => s.writeQps));

    const connectionChartMax = computed(() => Math.max(1, ...connectionSeries.value));
    const qpsChartMax = computed(() => Math.max(1, ...readQpsSeries.value, ...writeQpsSeries.value));

    const connectionChartPoints = computed(() => toPolylinePoints(connectionSeries.value, 360, 110, connectionChartMax.value));
    const readQpsChartPoints = computed(() => toPolylinePoints(readQpsSeries.value, 360, 110, qpsChartMax.value));
    const writeQpsChartPoints = computed(() => toPolylinePoints(writeQpsSeries.value, 360, 110, qpsChartMax.value));
    const longRunningTotal = computed(() =>
        Math.max(1, latestMonitorSample.value.bucket0to1 + latestMonitorSample.value.bucket1to5 + latestMonitorSample.value.bucketGt5)
    );

    const focusActivityTask = (task: ActivityTask) => {
        const targetTab = options.tabs.value.find((t) => t.id === task.tabId);
        if (targetTab) {
            options.activeTabId.value = targetTab.id;
        }
        options.isActivityMonitorOpen.value = false;
    };

    const killActivityTask = async (taskId: string) => {
        if (!options.connectionId.value) return;
        
        const task = activityTasksList.value.find(t => t.id === taskId);
        if (task) task.status = 'canceling...';

        try {
            await KillServerProcess(options.connectionId.value, taskId);
            // Refresh list rapidly after kill
            await fetchServerProcesses();
            options.onCancelSuccess?.(`Session ${taskId} killed successfully.`);
        } catch (e: any) {
            console.error('Failed to kill server process', e);
            options.onCancelError?.(`Failed to kill session ${taskId}: ${e.message || e}`);
            if (task) task.status = 'failed to cancel';
        }
    };

    const killAllActivityTasks = async () => {
        if (!options.connectionId.value) return;
        const ids = activityTasksList.value.map((task) => task.id);
        
        // Don't actually kill all server processes normally, this is dangerous for a real DB server.
        // We will just show an error not supported for safety, or implement just the UI side of it.
        options.onCancelError?.('Kill All is disabled for Server Processes for safety reasons.');
    };

    watch(monitorRefreshRate, () => {
        startMonitorTimer();
    });

    return {
        activityTasksList,
        activityTaskCount,
        monitorRefreshRate,
        latestMonitorSample,
        connectionChartPoints,
        readQpsChartPoints,
        writeQpsChartPoints,
        longRunningTotal,
        formatActivityTime,
        focusActivityTask,
        killActivityTask,
        killAllActivityTasks,
        startMonitorTimer,
        stopMonitorTimer,
    };
}

