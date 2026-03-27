<script lang="ts" setup>
import { computed } from 'vue';

import type { RecoveryCandidate } from '../../composables/appConnectionSessionShared';

const props = defineProps<{
  candidates: RecoveryCandidate[];
  rememberRecoveryChoice: boolean;
  title: string;
  description: string;
  selectAllLabel: string;
  unselectAllLabel: string;
  rememberChoiceLabel: string;
  skipLabel: string;
  recoverLabel: string;
}>();

const emit = defineEmits<{
  (e: 'restoreSelectedConnections'): void;
  (e: 'selectAllRecoveryConnections'): void;
  (e: 'dismissRecoveryPrompt'): void;
  (e: 'setRecoveryConnectionSelection', payload: { sessionKey: string; selected: boolean }): void;
  (e: 'clearRecoverySelections'): void;
  (e: 'setRememberRecoveryChoice', value: boolean): void;
}>();

const hasAnySelection = computed(() =>
  props.candidates.some((candidate) => candidate.selected),
);

const getCandidateSubtitle = (candidate: RecoveryCandidate) =>
  `${candidate.config.type} - ${candidate.config.database || candidate.config.host}`;

const handleRememberChoiceChange = (event: Event) => {
  const target = event.target as HTMLInputElement;
  emit('setRememberRecoveryChoice', target.checked);
};

const handleCandidateToggle = (sessionKey: string, event: Event) => {
  const target = event.target as HTMLInputElement;
  emit('setRecoveryConnectionSelection', {
    sessionKey,
    selected: target.checked,
  });
};
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-background/80 p-4 backdrop-blur-sm">
    <div class="w-full max-w-xl space-y-4 rounded-xl border border-border bg-card p-6 text-card-foreground shadow-xl">
      <div>
        <h2 class="text-lg font-semibold">{{ title }}</h2>
        <p class="text-sm text-muted-foreground">{{ description }}</p>
      </div>

      <div class="flex items-center justify-end gap-2">
        <button
          type="button"
          class="inline-flex h-8 items-center justify-center rounded-md border border-input bg-background px-3 py-1.5 text-xs font-medium hover:bg-accent"
          @click="emit('selectAllRecoveryConnections')"
        >
          {{ selectAllLabel }}
        </button>
        <button
          type="button"
          class="inline-flex h-8 items-center justify-center rounded-md border border-input bg-background px-3 py-1.5 text-xs font-medium hover:bg-accent"
          @click="emit('clearRecoverySelections')"
        >
          {{ unselectAllLabel }}
        </button>
      </div>

      <div class="max-h-[340px] overflow-y-auto rounded-md border border-border">
        <label
          v-for="candidate in candidates"
          :key="candidate.sessionKey"
          class="flex cursor-pointer items-start gap-3 border-b border-border p-3 last:border-b-0 hover:bg-muted/40"
        >
          <input
            type="checkbox"
            class="mt-1 h-4 w-4"
            :checked="candidate.selected"
            @change="handleCandidateToggle(candidate.sessionKey, $event)"
          />
          <div class="min-w-0">
            <div class="truncate text-sm font-medium">{{ candidate.name }}</div>
            <div class="truncate text-xs text-muted-foreground">
              {{ getCandidateSubtitle(candidate) }}
            </div>
          </div>
        </label>
      </div>

      <label class="flex items-center gap-2 text-sm text-muted-foreground">
        <input
          type="checkbox"
          class="h-4 w-4"
          :checked="rememberRecoveryChoice"
          @change="handleRememberChoiceChange"
        />
        {{ rememberChoiceLabel }}
      </label>

      <div class="flex justify-end gap-2">
        <button
          type="button"
          class="inline-flex h-9 items-center justify-center rounded-md border border-input bg-background px-4 py-2 text-sm font-medium hover:bg-accent"
          @click="emit('dismissRecoveryPrompt')"
        >
          {{ skipLabel }}
        </button>
        <button
          type="button"
          class="inline-flex h-9 items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          :disabled="!hasAnySelection"
          @click="emit('restoreSelectedConnections')"
        >
          {{ recoverLabel }}
        </button>
      </div>
    </div>
  </div>
</template>
