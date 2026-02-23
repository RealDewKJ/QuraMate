import { ref, computed } from 'vue';
import { GetTables, GetViews, GetStoredProcedures, GetFunctions } from '../../wailsjs/go/main/App';

export function useSidebar(connectionId: string) {
    const tableSearch = ref('');
    const tables = ref<string[]>([]);
    const views = ref<string[]>([]);
    const storedProcedures = ref<string[]>([]);
    const functions = ref<string[]>([]);
    const openFolders = ref(['Tables']);

    const toggleFolder = (folder: string) => {
        if (openFolders.value.includes(folder)) {
            openFolders.value = openFolders.value.filter(f => f !== folder);
        } else {
            openFolders.value.push(folder);
        }
    };

    const filteredItems = (items: string[]) => {
        if (!tableSearch.value) return items;
        const search = tableSearch.value.toLowerCase();
        return items.filter(v => v.toLowerCase().includes(search));
    };

    const filteredTables = computed(() => filteredItems(tables.value));
    const filteredViews = computed(() => filteredItems(views.value));
    const filteredStoredProcedures = computed(() => filteredItems(storedProcedures.value));
    const filteredFunctions = computed(() => filteredItems(functions.value));

    const refreshTables = async () => {
        if (!connectionId) return;
        try {
            const result = await GetTables(connectionId);
            tables.value = (result || []).sort((a, b) => a.localeCompare(b));
        } catch (e) {
            console.error("Failed to load tables", e);
        }
    };

    const refreshViews = async () => {
        if (!connectionId) return;
        try {
            const result = await GetViews(connectionId);
            views.value = (result || []).sort((a, b) => a.localeCompare(b));
        } catch (e) {
            console.error("Failed to load views", e);
        }
    };

    const refreshStoredProcedures = async () => {
        if (!connectionId) return;
        try {
            const result = await GetStoredProcedures(connectionId);
            storedProcedures.value = (result || []).sort((a, b) => a.localeCompare(b));
        } catch (e) {
            console.error("Failed to load stored procedures", e);
        }
    };

    const refreshFunctions = async () => {
        if (!connectionId) return;
        try {
            const result = await GetFunctions(connectionId);
            functions.value = (result || []).sort((a, b) => a.localeCompare(b));
        } catch (e) {
            console.error("Failed to load functions", e);
        }
    };

    const loadTables = async () => {
        await Promise.all([
            refreshTables(),
            refreshViews(),
            refreshStoredProcedures(),
            refreshFunctions()
        ]);
    };

    return {
        tableSearch,
        tables,
        views,
        storedProcedures,
        functions,
        openFolders,
        toggleFolder,
        filteredTables,
        filteredViews,
        filteredStoredProcedures,
        filteredFunctions,
        loadTables,
        refreshTables,
        refreshViews,
        refreshStoredProcedures,
        refreshFunctions
    };
}
