import { ref, computed } from 'vue';
import { QueryTab } from '../types/dashboard';

export function useTabs() {
    const tabs = ref<QueryTab[]>([]);
    const activeTabId = ref<string | null>(null);
    const tabCounter = ref(0);

    const generateId = () => {
        return Date.now().toString(36) + Math.random().toString(36).substr(2);
    };

    const activeTab = computed(() => tabs.value.find(t => t.id === activeTabId.value));

    const addTab = () => {
        const newId = generateId();
        tabCounter.value++;
        tabs.value.push({
            id: newId,
            name: `Query ${tabCounter.value}`,
            query: '',
            resultSets: [],
            primaryKeys: [],
            filters: {},
            sortColumn: undefined,
            sortDirection: null,
            error: '',
            isLoading: false,
            isExplaining: false,
            explanation: undefined,
            queryExecuted: false,
            activeQueryIds: [],
            resultViewTab: 'data',
            totalRowCount: undefined,
            isPartialStats: false,
            editorHeight: 300,
            columnWidths: {}
        });
        activeTabId.value = newId;
    };

    const closeTab = (id: string) => {
        const index = tabs.value.findIndex(t => t.id === id);
        if (index !== -1) {
            tabs.value.splice(index, 1);
            if (activeTabId.value === id) {
                if (tabs.value.length > 0) {
                    activeTabId.value = tabs.value[tabs.value.length - 1].id;
                } else {
                    activeTabId.value = null;
                }
            }
        }
    };

    return {
        tabs,
        activeTabId,
        tabCounter,
        activeTab,
        addTab,
        closeTab,
        generateId
    };
}
