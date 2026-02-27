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

    const getColumns = (tab: QueryTab) => {
        if (tab.resultSets && tab.resultSets.length > 0 && tab.resultSets[0].columns) return tab.resultSets[0].columns;
        return [];
    };

    return {
        activeResultSet,
        getColumns
    };
}
