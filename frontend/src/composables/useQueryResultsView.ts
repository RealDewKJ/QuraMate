import { computed } from 'vue';
import { useVirtualList } from '@vueuse/core';
import type { Ref } from 'vue';

import type { QueryTab } from '../types/dashboard';

export function useQueryResultsView(activeTab: Ref<QueryTab | undefined>) {
    const activeResultSet = computed(() => {
        if (!activeTab.value || !activeTab.value.resultSets || activeTab.value.resultSets.length === 0) return null;
        return activeTab.value.resultSets[0];
    });

    const filteredResults = computed(() => {
        if (!activeResultSet.value) return [];
        let data = activeResultSet.value.rows || [];

        const tab = activeTab.value;
        if (!tab) return data;

        const filters = tab.filters;
        if (filters && Object.keys(filters).length > 0) {
            data = data.filter((row) => {
                for (const [col, filterText] of Object.entries(filters)) {
                    if (!filterText) continue;
                    const val = row[col];
                    const strVal = val === null ? 'NULL' : String(val).toLowerCase();
                    if (!strVal.includes(filterText.toLowerCase())) return false;
                }
                return true;
            });
        }

        if (tab.sortColumn && tab.sortDirection) {
            const col = tab.sortColumn;
            const dir = tab.sortDirection;

            data = [...data].sort((a, b) => {
                const valA = a[col];
                const valB = b[col];

                if (valA === valB) return 0;
                if (valA === null) return 1;
                if (valB === null) return -1;

                if (valA < valB) return dir === 'asc' ? -1 : 1;
                if (valA > valB) return dir === 'asc' ? 1 : -1;
                return 0;
            });
        }

        return data;
    });

    const { list: virtualList, containerProps, wrapperProps } = useVirtualList(filteredResults, {
        itemHeight: 37,
        overscan: 10,
    });

    const padTop = computed(() => {
        if (virtualList.value.length === 0) return 0;
        const start = virtualList.value[0].index;
        return start * 37;
    });

    const padBottom = computed(() => {
        if (virtualList.value.length === 0) return 0;
        const end = virtualList.value[virtualList.value.length - 1].index;
        const total = filteredResults.value.length;
        return (total - end - 1) * 37;
    });

    const getColumns = (tab: QueryTab) => {
        if (tab.resultSets && tab.resultSets.length > 0 && tab.resultSets[0].columns) return tab.resultSets[0].columns;
        return [];
    };

    const toggleSort = (col: string) => {
        if (!activeTab.value) return;

        if (activeTab.value.sortColumn === col) {
            if (activeTab.value.sortDirection === 'asc') {
                activeTab.value.sortDirection = 'desc';
            } else if (activeTab.value.sortDirection === 'desc') {
                activeTab.value.sortDirection = null;
                activeTab.value.sortColumn = undefined;
            } else {
                activeTab.value.sortDirection = 'asc';
            }
        } else {
            activeTab.value.sortColumn = col;
            activeTab.value.sortDirection = 'asc';
        }
    };

    return {
        activeResultSet,
        filteredResults,
        virtualList,
        containerProps,
        wrapperProps,
        padTop,
        padBottom,
        getColumns,
        toggleSort,
    };
}
