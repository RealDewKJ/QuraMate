<template>
    <div class="relative w-full h-full" ref="triggerRef">
        <!-- Trigger Input -->
        <div class="flex items-center w-full h-full px-2 bg-background text-foreground border border-primary focus-within:ring-1 focus-within:ring-primary rounded-sm shadow-sm cursor-pointer group"
            @click.stop="togglePopover">
            <input ref="inputRef" :value="formattedDisplayValue" readonly
                class="flex-1 bg-transparent border-none outline-none text-sm cursor-pointer whitespace-nowrap overflow-hidden"
                placeholder="Select date..." @keydown.enter.prevent="handleConfirm"
                @keydown.esc.prevent="handleCancel" />
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                class="text-muted-foreground group-hover:text-primary transition-colors">
                <rect width="18" height="18" x="3" y="4" rx="2" ry="2" />
                <line x1="16" x2="16" y1="2" y2="6" />
                <line x1="8" x2="8" y1="2" y2="6" />
                <line x1="3" x2="21" y1="10" y2="10" />
            </svg>
        </div>

        <!-- Calendar Popover (Teleported to body to avoid clipping by overflow:hidden) -->
        <Teleport to="body">
            <div v-if="isOpen" v-active-click-outside="closePopover" ref="popoverRef"
                class="fixed z-[9999] p-3 bg-popover text-popover-foreground border border-border rounded-lg shadow-xl animate-in fade-in zoom-in-95 duration-200 min-w-[280px]"
                :style="popoverStyle">
                <div class="flex flex-col gap-3">
                    <!-- Header -->
                    <div class="flex items-center gap-1 px-1">
                        <button @click="changeMonth(-1)"
                            class="p-1 hover:bg-accent rounded-md transition-colors text-foreground">
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round">
                                <path d="m15 18-6-6 6-6" />
                            </svg>
                        </button>

                        <div class="flex-1 flex items-center justify-center gap-1">
                            <select :value="viewDate.getMonth()" @change="handleMonthChange"
                                class="bg-transparent hover:bg-accent rounded px-1 py-0.5 text-sm font-semibold cursor-pointer outline-none focus:ring-1 focus:ring-primary appearance-none text-center min-w-[90px] text-foreground">
                                <option v-for="(name, idx) in monthNames" :key="name" :value="idx" class="bg-popover">{{
                                    name }}</option>
                            </select>

                            <select :value="viewDate.getFullYear()" @change="handleYearChange"
                                class="bg-transparent hover:bg-accent rounded px-1 py-0.5 text-sm font-semibold cursor-pointer outline-none focus:ring-1 focus:ring-primary appearance-none text-center min-w-[60px] text-foreground">
                                <option v-for="year in years" :key="year" :value="year" class="bg-popover">{{ year }}
                                </option>
                            </select>
                        </div>

                        <button @click="changeMonth(1)"
                            class="p-1 hover:bg-accent rounded-md transition-colors text-foreground">
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round">
                                <path d="m9 18 6-6-6-6" />
                            </svg>
                        </button>
                    </div>

                    <!-- Grid -->
                    <div class="grid grid-cols-7 gap-1 text-center">
                        <div v-for="day in weekDays" :key="day"
                            class="text-[10px] font-bold text-muted-foreground uppercase py-1">
                            {{ day }}
                        </div>
                        <div v-for="(day, idx) in calendarDays" :key="idx" @click="selectDate(day)"
                            class="h-8 w-8 flex items-center justify-center text-xs rounded-md cursor-pointer transition-all"
                            :class="[
                                !day.currentMonth ? 'text-muted-foreground/30 hover:bg-accent/50' : 'hover:bg-accent hover:text-accent-foreground',
                                isSameDay(day.date, selectedDate) ? 'bg-primary text-primary-foreground font-bold hover:bg-primary' : '',
                                isToday(day.date) && !isSameDay(day.date, selectedDate) ? 'border border-primary/50 text-primary' : ''
                            ]">
                            {{ day.date.getDate() }}
                        </div>
                    </div>

                    <!-- Time Picker (if type is datetime-local) -->
                    <div v-if="type === 'datetime-local'" class="flex items-center gap-2 pt-2 border-t border-border">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="text-muted-foreground">
                            <circle cx="12" cy="12" r="10" />
                            <polyline points="12 6 12 12 16 14" />
                        </svg>
                        <div class="flex items-center gap-1 flex-1">
                            <input type="number" v-model="hours" min="0" max="23"
                                class="w-10 bg-accent rounded px-1 py-0.5 text-center text-xs focus:ring-1 focus:ring-primary outline-none" />
                            <span class="text-xs text-foreground">:</span>
                            <input type="number" v-model="minutes" min="0" max="59"
                                class="w-10 bg-accent rounded px-1 py-0.5 text-center text-xs focus:ring-1 focus:ring-primary outline-none" />
                            <span class="text-xs text-foreground">:</span>
                            <input type="number" v-model="seconds" min="0" max="59"
                                class="w-10 bg-accent rounded px-1 py-0.5 text-center text-xs focus:ring-1 focus:ring-primary outline-none" />
                        </div>
                    </div>

                    <!-- Footer -->
                    <div class="flex items-center justify-end gap-2 pt-2 border-t border-border">
                        <button @click="handleCancel"
                            class="px-3 py-1 text-xs hover:bg-accent rounded-md transition-colors text-foreground">Cancel</button>
                        <button @click="handleConfirm"
                            class="px-3 py-1 text-xs bg-primary text-primary-foreground font-semibold rounded-md shadow-sm hover:bg-primary/90 transition-colors">Apply</button>
                    </div>
                </div>
            </div>
        </Teleport>
    </div>
</template>

<script lang="ts" setup>
import { ref, computed, watch, onMounted, onUnmounted, CSSProperties } from 'vue';

const props = defineProps<{
    modelValue: string | null | undefined;
    type: 'date' | 'datetime-local';
}>();

const emit = defineEmits(['update:modelValue', 'confirm', 'cancel']);

const isOpen = ref(false);
const inputRef = ref<HTMLInputElement | null>(null);
const triggerRef = ref<HTMLElement | null>(null);
const popoverRef = ref<HTMLElement | null>(null);

// Positioning
const popoverStyle = ref<CSSProperties>({});

const updatePosition = () => {
    if (!isOpen.value || !triggerRef.value) return;

    const rect = triggerRef.value.getBoundingClientRect();
    const spaceBelow = window.innerHeight - rect.bottom;
    const popoverHeight = 350; // Estimated

    const style: CSSProperties = {
        left: `${rect.left}px`,
    };

    if (spaceBelow < popoverHeight && rect.top > popoverHeight) {
        // Show above if not enough space below
        style.bottom = `${window.innerHeight - rect.top + 4}px`;
    } else {
        style.top = `${rect.bottom + 4}px`;
    }

    popoverStyle.value = style;
};

// Date State
const selectedDate = ref<Date>(new Date());
const viewDate = ref<Date>(new Date());

// Time State
const hours = ref(0);
const minutes = ref(0);
const seconds = ref(0);

// Consts
const monthNames = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];
const weekDays = ["Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"];

const years = computed(() => {
    const currentYear = new Date().getFullYear();
    const start = 1900;
    const end = currentYear + 100;
    const result = [];
    for (let i = start; i <= end; i++) {
        result.push(i);
    }
    return result;
});

const handleMonthChange = (e: Event) => {
    const month = parseInt((e.target as HTMLSelectElement).value);
    viewDate.value = new Date(viewDate.value.getFullYear(), month, 1);
};

const handleYearChange = (e: Event) => {
    const year = parseInt((e.target as HTMLSelectElement).value);
    viewDate.value = new Date(year, viewDate.value.getMonth(), 1);
};

// Initialize from prop
const initFromProp = () => {
    if (!props.modelValue) {
        const now = new Date();
        selectedDate.value = now;
        viewDate.value = new Date(now.getFullYear(), now.getMonth(), 1);
        hours.value = now.getHours();
        minutes.value = now.getMinutes();
        seconds.value = now.getSeconds();
        return;
    }

    try {
        const d = new Date(props.modelValue);
        if (!isNaN(d.getTime())) {
            selectedDate.value = d;
            viewDate.value = new Date(d.getFullYear(), d.getMonth(), 1);
            hours.value = d.getHours();
            minutes.value = d.getMinutes();
            seconds.value = d.getSeconds();
        }
    } catch (e) {
        console.warn("Invalid date passed to DatePicker", props.modelValue);
    }
};

onMounted(() => {
    initFromProp();
    if (inputRef.value) inputRef.value.focus();
    window.addEventListener('scroll', updatePosition, true);
    window.addEventListener('resize', updatePosition);
});

onUnmounted(() => {
    window.removeEventListener('scroll', updatePosition, true);
    window.removeEventListener('resize', updatePosition);
});

watch(() => props.modelValue, initFromProp);

const formattedDisplayValue = computed(() => {
    if (!props.modelValue) return "";
    const d = new Date(props.modelValue);
    if (isNaN(d.getTime())) return props.modelValue;

    if (props.type === 'date') {
        return d.toISOString().split('T')[0];
    } else {
        // Format as YYYY-MM-DD HH:mm:ss for display
        const date = d.toISOString().split('T')[0];
        const time = [
            d.getHours().toString().padStart(2, '0'),
            d.getMinutes().toString().padStart(2, '0'),
            d.getSeconds().toString().padStart(2, '0')
        ].join(':');
        return `${date} ${time}`;
    }
});

const calendarDays = computed(() => {
    const days = [];
    const year = viewDate.value.getFullYear();
    const month = viewDate.value.getMonth();

    const firstDayOfMonth = new Date(year, month, 1).getDay();
    const daysInMonth = new Date(year, month + 1, 0).getDate();
    const daysInPrevMonth = new Date(year, month, 0).getDate();

    // Prev Month Days
    for (let i = firstDayOfMonth - 1; i >= 0; i--) {
        days.push({
            date: new Date(year, month - 1, daysInPrevMonth - i),
            currentMonth: false
        });
    }

    // Current Month Days
    for (let i = 1; i <= daysInMonth; i++) {
        days.push({
            date: new Date(year, month, i),
            currentMonth: true
        });
    }

    // Next Month Days
    const remainingSlots = 42 - days.length;
    for (let i = 1; i <= remainingSlots; i++) {
        days.push({
            date: new Date(year, month + 1, i),
            currentMonth: false
        });
    }

    return days;
});

const togglePopover = () => {
    isOpen.value = !isOpen.value;
    if (isOpen.value) {
        initFromProp();
        // Use nextTick to ensure popover is rendered before calculating position
        setTimeout(updatePosition, 0);
    }
};

const closePopover = () => {
    isOpen.value = false;
};

const changeMonth = (delta: number) => {
    viewDate.value = new Date(viewDate.value.getFullYear(), viewDate.value.getMonth() + delta, 1);
};

const selectDate = (day: { date: Date, currentMonth: boolean }) => {
    selectedDate.value = new Date(day.date);
    if (!day.currentMonth) {
        viewDate.value = new Date(day.date.getFullYear(), day.date.getMonth(), 1);
    }
};

const isSameDay = (d1: Date, d2: Date) => {
    return d1.getFullYear() === d2.getFullYear() &&
        d1.getMonth() === d2.getMonth() &&
        d1.getDate() === d2.getDate();
};

const isToday = (d: Date) => {
    return isSameDay(d, new Date());
};

const handleConfirm = () => {
    const finalDate = new Date(selectedDate.value);
    if (props.type === 'datetime-local') {
        finalDate.setHours(hours.value);
        finalDate.setMinutes(minutes.value);
        finalDate.setSeconds(seconds.value);
    }

    let output = "";
    if (props.type === 'date') {
        output = finalDate.toISOString().split('T')[0];
    } else {
        // We want to return something consistent. 
        // Usually DBs expect YYYY-MM-DD HH:mm:ss or ISO
        // Let's return YYYY-MM-DDTHH:mm:ss for internal state consistency with input type=datetime-local
        const date = finalDate.getFullYear() + '-' +
            String(finalDate.getMonth() + 1).padStart(2, '0') + '-' +
            String(finalDate.getDate()).padStart(2, '0');
        const time = String(finalDate.getHours()).padStart(2, '0') + ':' +
            String(finalDate.getMinutes()).padStart(2, '0') + ':' +
            String(finalDate.getSeconds()).padStart(2, '0');
        output = `${date}T${time}`;
    }

    emit('update:modelValue', output);
    emit('confirm');
    isOpen.value = false;
};

const handleCancel = () => {
    emit('cancel');
    isOpen.value = false;
};

// Custom type for element with clickOutsideEvent
interface ClickOutsideElement extends HTMLElement {
    clickOutsideEvent?: (event: Event) => void;
}

// Directive for click outside
const vActiveClickOutside = {
    mounted(el: ClickOutsideElement, binding: any) {
        el.clickOutsideEvent = (event: Event) => {
            if (!(el === event.target || el.contains(event.target as Node))) {
                binding.value();
            }
        };
        document.addEventListener("mousedown", el.clickOutsideEvent);
    },
    unmounted(el: ClickOutsideElement) {
        if (el.clickOutsideEvent) {
            document.removeEventListener("mousedown", el.clickOutsideEvent);
        }
    },
};
</script>

<style scoped>
/* Ensure the input looks like a normal cell input when inactive */
input:focus {
    outline: none;
}
</style>
