import type { ResultSet } from '../types/database';

export const DEFAULT_GRID_SCREENSHOT_SHORTCUT = 'Ctrl+Shift+I';

interface ShortcutDescriptor {
    ctrl: boolean;
    shift: boolean;
    alt: boolean;
    key: string;
}

interface BuildGridImageOptions {
    resultSet: ResultSet;
    tableName: string;
    filters?: Record<string, string>;
    sortColumn?: string;
    sortDirection?: 'asc' | 'desc' | null;
    timestamp?: Date;
    maxRows?: number;
}

export interface GridImageExportResult {
    blob: Blob;
    fileName: string;
    totalRows: number;
    renderedRows: number;
    timestampLabel: string;
}

const FONT_FAMILY = "'Segoe UI', 'Inter', Arial, sans-serif";
const SUBTITLE_FONT = `400 13px ${FONT_FAMILY}`;
const CELL_FONT = `400 12px ${FONT_FAMILY}`;
const MAX_CANVAS_WIDTH = 1500;
const MIN_COLUMN_WIDTH = 110;
const MAX_COLUMN_WIDTH = 320;
const KEY_ALIASES: Record<string, string> = {
    control: 'ctrl',
    cmd: 'cmd',
    command: 'cmd',
    meta: 'cmd',
    option: 'alt',
    esc: 'escape',
    return: 'enter',
    ' ': 'space'
};

const sanitizeShortcutKey = (value: string): string => {
    const raw = String(value || '').trim();
    if (!raw) return '';
    const lower = raw.toLowerCase();
    return KEY_ALIASES[lower] || lower;
};

const normalizeSingleKey = (value: string): string => {
    const key = sanitizeShortcutKey(value);
    if (!key) return '';
    if (key.length === 1) return key.toUpperCase();
    if (/^f\d{1,2}$/i.test(key)) return key.toUpperCase();
    if (key === 'space') return 'Space';
    if (key === 'arrowup') return 'Up';
    if (key === 'arrowdown') return 'Down';
    if (key === 'arrowleft') return 'Left';
    if (key === 'arrowright') return 'Right';
    return key.charAt(0).toUpperCase() + key.slice(1);
};

const parseShortcut = (binding: string | null | undefined): ShortcutDescriptor => {
    const parts = String(binding || '')
        .split('+')
        .map((part) => sanitizeShortcutKey(part))
        .filter(Boolean);

    const descriptor: ShortcutDescriptor = {
        ctrl: parts.includes('ctrl') || parts.includes('cmd'),
        shift: parts.includes('shift'),
        alt: parts.includes('alt'),
        key: ''
    };

    const keyPart = parts.find((part) => !['ctrl', 'cmd', 'shift', 'alt'].includes(part));
    if (keyPart) {
        descriptor.key = keyPart;
    }

    return descriptor;
};

export const normalizeShortcutString = (binding: string | null | undefined): string => {
    const parsed = parseShortcut(binding);
    if (!parsed.key) {
        return DEFAULT_GRID_SCREENSHOT_SHORTCUT;
    }

    const parts: string[] = [];
    if (parsed.ctrl) parts.push('Ctrl');
    if (parsed.shift) parts.push('Shift');
    if (parsed.alt) parts.push('Alt');
    parts.push(normalizeSingleKey(parsed.key));

    return parts.join('+');
};

export const keyboardEventToShortcut = (event: KeyboardEvent): string => {
    const normalizedKey = sanitizeShortcutKey(event.key);
    if (!normalizedKey || ['ctrl', 'shift', 'alt', 'cmd'].includes(normalizedKey)) {
        return '';
    }

    const parts: string[] = [];
    if (event.ctrlKey || event.metaKey) parts.push('Ctrl');
    if (event.shiftKey) parts.push('Shift');
    if (event.altKey) parts.push('Alt');
    parts.push(normalizeSingleKey(normalizedKey));

    return parts.join('+');
};

export const shortcutMatchesEvent = (event: KeyboardEvent, binding: string | null | undefined): boolean => {
    const parsed = parseShortcut(binding);
    if (!parsed.key) {
        return false;
    }

    const eventKey = sanitizeShortcutKey(event.key);
    const eventCtrl = event.ctrlKey || event.metaKey;

    if (parsed.ctrl !== eventCtrl) return false;
    if (parsed.shift !== event.shiftKey) return false;
    if (parsed.alt !== event.altKey) return false;

    return parsed.key === eventKey;
};

const formatTimestampLabel = (timestamp: Date): string => {
    const formatter = new Intl.DateTimeFormat(undefined, {
        year: 'numeric',
        month: 'short',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false
    });
    return formatter.format(timestamp);
};

const toFilenameTimestamp = (timestamp: Date): string => {
    const pad = (n: number) => String(n).padStart(2, '0');
    return `${timestamp.getFullYear()}${pad(timestamp.getMonth() + 1)}${pad(timestamp.getDate())}-${pad(timestamp.getHours())}${pad(timestamp.getMinutes())}${pad(timestamp.getSeconds())}`;
};

const sanitizeFileSegment = (value: string): string => {
    return String(value || 'query_result')
        .toLowerCase()
        .replace(/[^a-z0-9_-]+/g, '_')
        .replace(/^_+|_+$/g, '') || 'query_result';
};

const toCellText = (value: unknown): string => {
    if (value === null || value === undefined) return 'NULL';
    if (typeof value === 'string') return value;
    if (typeof value === 'number' || typeof value === 'boolean') return String(value);

    try {
        return JSON.stringify(value);
    } catch {
        return String(value);
    }
};

const truncateText = (value: string, maxLength = 72): string => {
    if (value.length <= maxLength) return value;
    return `${value.slice(0, maxLength - 1)}…`;
};

const applyFiltersAndSort = (
    resultSet: ResultSet,
    filters?: Record<string, string>,
    sortColumn?: string,
    sortDirection?: 'asc' | 'desc' | null
): Record<string, unknown>[] => {
    if (!resultSet?.rows) return [];

    let rows = resultSet.rows as Record<string, unknown>[];

    if (filters) {
        const activeFilters = Object.entries(filters).filter(([, value]) => value !== '' && value !== null && value !== undefined);
        if (activeFilters.length > 0) {
            rows = rows.filter((row) => {
                return activeFilters.every(([column, filterValue]) => {
                    const rowValue = row[column];
                    if (rowValue === null || rowValue === undefined) return false;
                    return String(rowValue).toLowerCase().includes(String(filterValue).toLowerCase());
                });
            });
        }
    }

    if (sortColumn && sortDirection) {
        rows = [...rows].sort((a, b) => {
            const valueA = a[sortColumn];
            const valueB = b[sortColumn];

            if (valueA === valueB) return 0;
            if (valueA === null || valueA === undefined) return 1;
            if (valueB === null || valueB === undefined) return -1;

            if (valueA < valueB) return sortDirection === 'asc' ? -1 : 1;
            if (valueA > valueB) return sortDirection === 'asc' ? 1 : -1;
            return 0;
        });
    }

    return rows;
};

const drawRoundedRect = (
    ctx: CanvasRenderingContext2D,
    x: number,
    y: number,
    width: number,
    height: number,
    radius: number,
    fill: string | CanvasGradient,
    stroke?: string
) => {
    const safeRadius = Math.min(radius, width / 2, height / 2);
    ctx.beginPath();
    ctx.moveTo(x + safeRadius, y);
    ctx.lineTo(x + width - safeRadius, y);
    ctx.quadraticCurveTo(x + width, y, x + width, y + safeRadius);
    ctx.lineTo(x + width, y + height - safeRadius);
    ctx.quadraticCurveTo(x + width, y + height, x + width - safeRadius, y + height);
    ctx.lineTo(x + safeRadius, y + height);
    ctx.quadraticCurveTo(x, y + height, x, y + height - safeRadius);
    ctx.lineTo(x, y + safeRadius);
    ctx.quadraticCurveTo(x, y, x + safeRadius, y);
    ctx.closePath();
    ctx.fillStyle = fill;
    ctx.fill();

    if (stroke) {
        ctx.strokeStyle = stroke;
        ctx.lineWidth = 1;
        ctx.stroke();
    }
};

const drawCellText = (
    ctx: CanvasRenderingContext2D,
    text: string,
    x: number,
    y: number,
    width: number,
    height: number,
    color = '#111827',
    font = CELL_FONT
) => {
    const truncated = truncateText(text);
    ctx.save();
    ctx.beginPath();
    ctx.rect(x, y, width, height);
    ctx.clip();
    ctx.fillStyle = color;
    ctx.font = font;
    ctx.textAlign = 'left';
    ctx.textBaseline = 'middle';
    ctx.fillText(truncated, x + 10, y + height / 2);
    ctx.restore();
};

const drawChip = (
    ctx: CanvasRenderingContext2D,
    text: string,
    x: number,
    y: number,
    options?: { fill?: string; textColor?: string; font?: string; paddingX?: number; height?: number; stroke?: string }
): number => {
    const fill = options?.fill || '#f1f5f9';
    const textColor = options?.textColor || '#334155';
    const font = options?.font || SUBTITLE_FONT;
    const paddingX = options?.paddingX ?? 12;
    const height = options?.height ?? 28;
    const stroke = options?.stroke;

    ctx.font = font;
    const width = Math.ceil(ctx.measureText(text).width) + paddingX * 2;
    drawRoundedRect(ctx, x, y, width, height, height / 2, fill, stroke);

    ctx.fillStyle = textColor;
    ctx.font = font;
    ctx.textAlign = 'left';
    ctx.textBaseline = 'middle';
    ctx.fillText(text, x + paddingX, y + height / 2);

    return width;
};

export const buildResultGridImage = async (options: BuildGridImageOptions): Promise<GridImageExportResult> => {
    const timestamp = options.timestamp || new Date();
    const columns = options.resultSet?.columns || [];
    const timestampLabel = formatTimestampLabel(timestamp);

    if (columns.length === 0) {
        throw new Error('No columns found in result set.');
    }

    const rows = applyFiltersAndSort(options.resultSet, options.filters, options.sortColumn, options.sortDirection);
    const maxRows = Math.max(1, options.maxRows || 16);
    const renderedRows = rows.slice(0, maxRows);

    const measuringCanvas = document.createElement('canvas');
    const measuringContext = measuringCanvas.getContext('2d');
    if (!measuringContext) {
        throw new Error('Canvas context is not available.');
    }

    measuringContext.font = CELL_FONT;
    const widths = columns.map((column) => {
        const headerWidth = measuringContext.measureText(String(column)).width + 24;
        const contentWidth = renderedRows.reduce((max, row) => {
            const text = truncateText(toCellText(row[column]));
            const measured = measuringContext.measureText(text).width + 24;
            return Math.max(max, measured);
        }, 120);
        return Math.min(MAX_COLUMN_WIDTH, Math.max(MIN_COLUMN_WIDTH, Math.ceil(Math.max(headerWidth, contentWidth))));
    });

    const indexWidth = 56;
    const baseTableWidth = indexWidth + widths.reduce((sum, width) => sum + width, 0);
    const maxContentWidth = MAX_CANVAS_WIDTH - 64;

    if (baseTableWidth > maxContentWidth) {
        const scalable = widths.reduce((sum, width) => sum + Math.max(0, width - MIN_COLUMN_WIDTH), 0);
        const neededReduction = baseTableWidth - maxContentWidth;

        if (scalable > 0 && neededReduction > 0) {
            columns.forEach((_, index) => {
                const reducible = Math.max(0, widths[index] - MIN_COLUMN_WIDTH);
                const reduction = Math.min(reducible, Math.ceil((reducible / scalable) * neededReduction));
                widths[index] = widths[index] - reduction;
            });
        }
    }

    const tableWidth = indexWidth + widths.reduce((sum, width) => sum + width, 0);
    const padding = 28;
    const outerRadius = 20;
    const heroHeight = 118;
    const chipAreaHeight = 44;
    const tablePanelTopGap = 16;
    const headerHeight = 42;
    const rowHeight = 34;
    const footerHeight = renderedRows.length < rows.length ? 42 : 0;
    const tablePanelPadding = 14;
    const tablePanelHeight = tablePanelPadding + headerHeight + renderedRows.length * rowHeight + footerHeight + tablePanelPadding;
    const canvasWidth = Math.min(MAX_CANVAS_WIDTH, tableWidth + padding * 2 + 24);
    const contentWidth = canvasWidth - padding * 2;
    const tableX = padding + Math.max(0, (contentWidth - tableWidth - tablePanelPadding * 2) / 2) + tablePanelPadding;
    const canvasHeight = padding * 2 + heroHeight + chipAreaHeight + tablePanelTopGap + tablePanelHeight + 12;

    const canvas = document.createElement('canvas');
    const scale = 2;
    canvas.width = Math.round(canvasWidth * scale);
    canvas.height = Math.round(canvasHeight * scale);

    const ctx = canvas.getContext('2d');
    if (!ctx) {
        throw new Error('Failed to create image context.');
    }
    ctx.scale(scale, scale);
    ctx.imageSmoothingEnabled = true;
    ctx.imageSmoothingQuality = 'high';

    ctx.fillStyle = '#e7edf5';
    ctx.fillRect(0, 0, canvasWidth, canvasHeight);

    const gradient = ctx.createLinearGradient(0, 0, canvasWidth, canvasHeight);
    gradient.addColorStop(0, '#1f3b62');
    gradient.addColorStop(0.45, '#2e5b8f');
    gradient.addColorStop(1, '#14b8a6');
    drawRoundedRect(ctx, 10, 10, canvasWidth - 20, canvasHeight - 20, outerRadius, '#ffffff');
    drawRoundedRect(ctx, 10, 10, canvasWidth - 20, heroHeight + 56, outerRadius, gradient);
    drawRoundedRect(ctx, 10, heroHeight + 34, canvasWidth - 20, canvasHeight - (heroHeight + 44), outerRadius, '#f8fafc');

    const contentX = padding;
    let cursorY = padding;

    ctx.fillStyle = '#ffffff';
    ctx.font = `700 30px ${FONT_FAMILY}`;
    ctx.textBaseline = 'middle';
    ctx.fillText(options.tableName || 'Query Result', contentX, cursorY + 50);

    ctx.fillStyle = '#dbeafe';
    ctx.font = `500 14px ${FONT_FAMILY}`;
    ctx.fillText(`Captured ${timestampLabel}`, contentX, cursorY + 78);
    cursorY += heroHeight;

    let chipX = contentX;
    chipX += drawChip(ctx, `${rows.length} rows`, chipX, cursorY, {
        fill: '#e0ecff',
        textColor: '#1d4ed8',
        font: `600 12px ${FONT_FAMILY}`,
        stroke: '#bfdbfe'
    }) + 8;
    drawChip(ctx, `${columns.length} columns`, chipX, cursorY, {
        fill: '#dcfce7',
        textColor: '#166534',
        font: `600 12px ${FONT_FAMILY}`,
        stroke: '#bbf7d0'
    });

    cursorY += chipAreaHeight + tablePanelTopGap;

    const tablePanelX = tableX - tablePanelPadding;
    const tablePanelY = cursorY - tablePanelPadding;
    const tablePanelWidth = tableWidth + tablePanelPadding * 2;
    const tablePanelRadius = 16;

    ctx.save();
    ctx.shadowColor = 'rgba(15, 23, 42, 0.12)';
    ctx.shadowBlur = 18;
    ctx.shadowOffsetY = 10;
    drawRoundedRect(ctx, tablePanelX, tablePanelY, tablePanelWidth, tablePanelHeight, tablePanelRadius, '#ffffff');
    ctx.restore();
    drawRoundedRect(ctx, tablePanelX, tablePanelY, tablePanelWidth, tablePanelHeight, tablePanelRadius, '#ffffff', '#dbe3ee');

    const headerGradient = ctx.createLinearGradient(tableX, cursorY, tableX + tableWidth, cursorY);
    headerGradient.addColorStop(0, '#0f172a');
    headerGradient.addColorStop(1, '#1e3a8a');
    drawRoundedRect(ctx, tableX, cursorY, tableWidth, headerHeight, 12, headerGradient);

    ctx.fillStyle = '#cbd5e1';
    ctx.font = `600 12px ${FONT_FAMILY}`;
    ctx.fillText('#', tableX + 14, cursorY + headerHeight / 2 + 0.5);

    let x = tableX + indexWidth;
    columns.forEach((column, index) => {
        drawCellText(ctx, String(column), x, cursorY, widths[index], headerHeight, '#e2e8f0', `600 12px ${FONT_FAMILY}`);
        x += widths[index];
    });

    cursorY += headerHeight;

    renderedRows.forEach((row, rowIndex) => {
        const isAlt = rowIndex % 2 === 1;
        ctx.fillStyle = isAlt ? '#f7faff' : '#ffffff';
        ctx.fillRect(tableX, cursorY, tableWidth, rowHeight);

        ctx.strokeStyle = '#e9eef6';
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(tableX, cursorY + rowHeight);
        ctx.lineTo(tableX + tableWidth, cursorY + rowHeight);
        ctx.stroke();

        ctx.fillStyle = '#64748b';
        ctx.font = `500 12px ${FONT_FAMILY}`;
        ctx.textAlign = 'left';
        ctx.textBaseline = 'middle';
        ctx.fillText(String(rowIndex + 1), tableX + 14, cursorY + rowHeight / 2);

        let cellX = tableX + indexWidth;
        columns.forEach((column, columnIndex) => {
            const cellText = toCellText(row[column]);
            drawCellText(ctx, cellText, cellX, cursorY, widths[columnIndex], rowHeight, '#0f172a', `500 12px ${FONT_FAMILY}`);
            cellX += widths[columnIndex];
        });

        cursorY += rowHeight;
    });

    if (renderedRows.length < rows.length) {
        ctx.fillStyle = '#475569';
        ctx.font = `500 12px ${FONT_FAMILY}`;
        ctx.textAlign = 'left';
        ctx.textBaseline = 'middle';
        ctx.fillText(`Showing ${renderedRows.length} of ${rows.length} rows for sharing`, tableX, cursorY + 20);
    }

    const blob = await new Promise<Blob>((resolve, reject) => {
        canvas.toBlob((value) => {
            if (!value) {
                reject(new Error('Failed to encode image blob.'));
                return;
            }
            resolve(value);
        }, 'image/png');
    });

    const fileName = `${sanitizeFileSegment(options.tableName)}_${toFilenameTimestamp(timestamp)}.png`;

    return {
        blob,
        fileName,
        totalRows: rows.length,
        renderedRows: renderedRows.length,
        timestampLabel
    };
};

export const downloadBlobAsFile = (blob: Blob, fileName: string) => {
    const objectUrl = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = objectUrl;
    anchor.download = fileName;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);

    setTimeout(() => URL.revokeObjectURL(objectUrl), 1000);
};

export const copyImageBlobToClipboard = async (blob: Blob): Promise<boolean> => {
    const clipboard = navigator.clipboard as Clipboard | undefined;
    const ClipboardItemCtor = (window as unknown as { ClipboardItem?: typeof ClipboardItem }).ClipboardItem;
    if (!clipboard || !ClipboardItemCtor || !clipboard.write) {
        return false;
    }

    try {
        const item = new ClipboardItemCtor({ 'image/png': blob });
        await clipboard.write([item]);
        return true;
    } catch {
        return false;
    }
};

