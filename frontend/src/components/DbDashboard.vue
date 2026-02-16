<template>
    <div class="flex h-screen">
        <!-- Sidebar -->
        <div class="w-64 bg-gray-800 text-white flex flex-col">
            <div class="p-4 bg-gray-900 font-bold">Tables</div>
            <div class="flex-1 overflow-y-auto">
                <ul>
                    <li v-for="table in tables" :key="table" class="p-2 hover:bg-gray-700 cursor-pointer"
                        @click="selectTable(table)">
                        {{ table }}
                    </li>
                </ul>
            </div>
            <!-- Disconnect Button -->
            <div class="p-4 border-t border-gray-700">
                <button @click="disconnect"
                    class="w-full bg-red-600 hover:bg-red-700 text-white font-bold py-2 px-4 rounded">
                    Disconnect
                </button>
            </div>
        </div>

        <!-- Main Content -->
        <div class="flex-1 flex flex-col overflow-hidden bg-gray-100">
            <!-- Query Editor -->
            <div class="p-4 bg-white border-b border-gray-200">
                <textarea v-model="query" class="w-full h-32 p-2 border border-gray-300 rounded font-mono"
                    placeholder="SELECT * FROM table..."></textarea>
                <button @click="runQuery"
                    class="mt-2 bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded">
                    Run Query
                </button>
            </div>

            <!-- Results -->
            <div class="flex-1 overflow-auto p-4">
                <div v-if="error" class="text-red-500 bg-red-100 p-4 rounded mb-4">
                    {{ error }}
                </div>

                <div v-if="results.length > 0" class="bg-white shadow rounded overflow-x-auto">
                    <table class="min-w-full leading-normal">
                        <thead>
                            <tr>
                                <th v-for="col in columns" :key="col"
                                    class="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">
                                    {{ col }}
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="(row, idx) in results" :key="idx">
                                <td v-for="col in columns" :key="col"
                                    class="px-5 py-5 border-b border-gray-200 bg-white text-sm">
                                    {{ row[col] }}
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
                <div v-else-if="!error && queryExecuted" class="text-gray-500 text-center mt-10">
                    No results found.
                </div>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed } from 'vue';
import { GetTables, ExecuteQuery, DisconnectDB } from '../../wailsjs/go/main/App';

const emit = defineEmits(['disconnect']);

const tables = ref<string[]>([]);
const query = ref('');
const results = ref<any[]>([]);
const error = ref('');
const queryExecuted = ref(false);

const columns = computed(() => {
    if (results.value.length === 0) return [];
    return Object.keys(results.value[0]);
});

const loadTables = async () => {
    try {
        tables.value = await GetTables();
    } catch (e) {
        console.error("Failed to load tables", e);
    }
};

const selectTable = (tableName: string) => {
    query.value = `SELECT * FROM ${tableName} LIMIT 100`;
    runQuery();
};

const runQuery = async () => {
    error.value = '';
    results.value = [];
    queryExecuted.value = false;

    try {
        const res = await ExecuteQuery(query.value);
        if (res.error) {
            error.value = res.error;
        } else {
            results.value = res.data || [];
        }
        queryExecuted.value = true;
    } catch (e: any) {
        error.value = e.toString();
    }
};

const disconnect = async () => {
    try {
        await DisconnectDB();
        emit('disconnect');
    } catch (e) {
        console.error("Failed to disconnect", e);
        // Still emit disconnect to change UI
        emit('disconnect');
    }
};

onMounted(() => {
    loadTables();
});
</script>
