import { reactive, ref } from 'vue';
import type { Ref } from 'vue';

import type { QueryTab } from '../types/dashboard';

export function useResultSetLayout(activeTab: Ref<QueryTab | undefined>) {
    const isResizing = ref(false);
    const isColumnResizing = ref(false);
    const resizingColumn = ref<string | null>(null);
    const startX = ref(0);
    const startWidth = ref(0);
    const resultSetHeights = reactive<Record<string, number>>({});
    const isResultSetResizing = ref(false);
    const resizingResultSetKey = ref<string | null>(null);
    const resultSetResizeStartY = ref(0);
    const resultSetResizeStartHeight = ref(0);

    const clamp = (val: number, min: number, max: number) => Math.max(min, Math.min(max, val));

    const stopResizing = () => {
        isResizing.value = false;
        document.removeEventListener('mousemove', doResize);
        document.removeEventListener('mouseup', stopResizing);
        document.body.style.cursor = '';
    };

    const doResize = (e: MouseEvent) => {
        if (!isResizing.value || !activeTab.value) return;

        const queryArea = document.querySelector('.query-area-container');
        if (queryArea) {
            const rect = queryArea.getBoundingClientRect();
            const newHeight = e.clientY - rect.top;

            if (newHeight >= 120 && newHeight <= window.innerHeight - 150) {
                activeTab.value.editorHeight = newHeight;
            }
        }
    };

    const startResizing = () => {
        isResizing.value = true;
        document.addEventListener('mousemove', doResize);
        document.addEventListener('mouseup', stopResizing);
        document.body.style.cursor = 'row-resize';
    };

    const stopColumnResize = () => {
        isColumnResizing.value = false;
        resizingColumn.value = null;
        document.removeEventListener('mousemove', doColumnResize);
        document.removeEventListener('mouseup', stopColumnResize);
        document.body.style.cursor = '';
    };

    const doColumnResize = (e: MouseEvent) => {
        if (!isColumnResizing.value || !activeTab.value || !resizingColumn.value) return;

        const deltaX = e.clientX - startX.value;
        const newWidth = Math.max(50, startWidth.value + deltaX);

        activeTab.value.columnWidths[resizingColumn.value] = newWidth;
    };

    const startColumnResize = (e: MouseEvent, col: string) => {
        if (!activeTab.value) return;

        isColumnResizing.value = true;
        resizingColumn.value = col;
        startX.value = e.clientX;

        if (!activeTab.value.columnWidths[col]) {
            const th = (e.target as HTMLElement).closest('th');
            startWidth.value = th ? th.offsetWidth : 150;
        } else {
            startWidth.value = activeTab.value.columnWidths[col];
        }

        document.addEventListener('mousemove', doColumnResize);
        document.addEventListener('mouseup', stopColumnResize);
        document.body.style.cursor = 'col-resize';
        e.preventDefault();
        e.stopPropagation();
    };

    const getResultSetHeightKey = (resultSetIndex: number) => {
        if (!activeTab.value) return '';
        return `${activeTab.value.id}:${resultSetIndex}`;
    };

    const getAutoResultSetHeight = (resultSet: any, resultSetIndex: number) => {
        const rowCount = Array.isArray(resultSet?.rows) ? resultSet.rows.length : 0;
        const colCount = Array.isArray(resultSet?.columns) ? resultSet.columns.length : 0;

        const score = Math.max(1, Math.ceil(rowCount / 40) + Math.ceil(colCount / 8));
        const indexBoost = resultSetIndex === 0 ? 1 : 0;
        const targetHeight = 220 + Math.min(6, score + indexBoost) * 52;

        return Math.min(620, targetHeight);
    };

    const getResultSetCardStyle = (resultSet: any, resultSetIndex: number) => {
        const key = getResultSetHeightKey(resultSetIndex);
        const height = (key && resultSetHeights[key]) ? resultSetHeights[key] : getAutoResultSetHeight(resultSet, resultSetIndex);

        return {
            height: `${height}px`,
            maxHeight: '72vh',
        };
    };

    const stopResultSetResize = () => {
        isResultSetResizing.value = false;
        resizingResultSetKey.value = null;
        document.removeEventListener('mousemove', doResultSetResize);
        document.removeEventListener('mouseup', stopResultSetResize);
        document.body.style.cursor = '';
    };

    const doResultSetResize = (e: MouseEvent) => {
        if (!isResultSetResizing.value || !resizingResultSetKey.value) return;

        const deltaY = e.clientY - resultSetResizeStartY.value;
        const maxHeight = Math.max(280, Math.floor(window.innerHeight * 0.8));
        const newHeight = clamp(resultSetResizeStartHeight.value + deltaY, 180, maxHeight);

        resultSetHeights[resizingResultSetKey.value] = newHeight;
    };

    const startResultSetResize = (e: MouseEvent, resultSetIndex: number, resultSet: any) => {
        const key = getResultSetHeightKey(resultSetIndex);
        if (!key) return;

        isResultSetResizing.value = true;
        resizingResultSetKey.value = key;
        resultSetResizeStartY.value = e.clientY;
        resultSetResizeStartHeight.value = resultSetHeights[key] ?? getAutoResultSetHeight(resultSet, resultSetIndex);

        document.addEventListener('mousemove', doResultSetResize);
        document.addEventListener('mouseup', stopResultSetResize);
        document.body.style.cursor = 'row-resize';
        e.preventDefault();
        e.stopPropagation();
    };

    const stopAllResizing = () => {
        stopResizing();
        stopColumnResize();
        stopResultSetResize();
    };

    return {
        startResizing,
        startColumnResize,
        startResultSetResize,
        getResultSetCardStyle,
        stopAllResizing,
    };
}
