<script lang="ts" setup>
import type { PropType } from 'vue';

import { app } from '../../../wailsjs/go/models';
import type { ConnectionConfig, SSHTrustAuditEntry } from '../../composables/useConnectionForm';

defineProps({
  canTrustCurrentSshHost: {
    type: Boolean,
    required: true,
  },
  config: {
    type: Object as PropType<ConnectionConfig>,
    required: true,
  },
  error: {
    type: String,
    required: true,
  },
  expectedSshFingerprint: {
    type: String,
    required: true,
  },
  filteredSshTrustAudit: {
    type: Array as PropType<SSHTrustAuditEntry[]>,
    required: true,
  },
  isFingerprintMatch: {
    type: Boolean,
    required: true,
  },
  isFingerprintMismatch: {
    type: Boolean,
    required: true,
  },
  isLoadingHostKeyInfo: {
    type: Boolean,
    required: true,
  },
  isPinnedFingerprintMismatch: {
    type: Boolean,
    required: true,
  },
  isTrustingHost: {
    type: Boolean,
    required: true,
  },
  pinnedSshFingerprint: {
    type: String,
    required: true,
  },
  sshHostKeyInfo: {
    type: Object as PropType<app.SSHHostKeyInfo | null>,
    required: true,
  },
  sshRotationConfirmText: {
    type: String,
    required: true,
  },
  sshRotationReason: {
    type: String,
    required: true,
  },
  sshTrustAudit: {
    type: Array as PropType<SSHTrustAuditEntry[]>,
    required: true,
  },
  sshTrustAuditSearch: {
    type: String,
    required: true,
  },
});

const emit = defineEmits<{
  (e: 'acceptPinnedFingerprintRotation'): void;
  (e: 'clearSshTrustAudit'): void;
  (e: 'copyCurrentSshFingerprint'): void;
  (e: 'exportSshTrustAudit'): void;
  (e: 'importSshTrustAudit'): void;
  (e: 'loadCurrentSshHostKeyInfo'): void;
  (e: 'trustCurrentSshHost'): void;
  (e: 'updateConfig', patch: Partial<ConnectionConfig>): void;
  (e: 'updateExpectedSshFingerprint', value: string): void;
  (e: 'updateSshRotationConfirmText', value: string): void;
  (e: 'updateSshRotationReason', value: string): void;
  (e: 'updateSshTrustAuditSearch', value: string): void;
}>();
</script>

<template>
  <div class="space-y-4 animate-in fade-in slide-in-from-top-4 duration-300">
    <div class="flex items-center space-x-2">
      <input
        id="readOnly"
        type="checkbox"
        :checked="config.readOnly"
        class="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary"
        @change="emit('updateConfig', { readOnly: ($event.target as HTMLInputElement).checked })"
      />
      <label
        for="readOnly"
        class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
      >
        Read Only Mode
      </label>
    </div>

    <div
      v-if="error"
      class="flex items-start gap-2 rounded-md bg-destructive/15 p-3 text-sm text-destructive animate-in fade-in zoom-in duration-300"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        class="mt-0.5 shrink-0"
      >
        <circle cx="12" cy="12" r="10" />
        <line x1="12" x2="12" y1="8" y2="12" />
        <line x1="12" x2="12.01" y1="16" y2="16" />
      </svg>

      <div class="flex-1">
        <p>{{ error }}</p>

        <div v-if="canTrustCurrentSshHost" class="mt-2 space-y-2">
          <button
            v-if="!sshHostKeyInfo"
            type="button"
            :disabled="isLoadingHostKeyInfo"
            class="inline-flex h-8 items-center justify-center rounded-full border border-destructive/60 bg-background/95 px-3 text-xs font-medium text-destructive transition-colors hover:bg-destructive/10 disabled:cursor-not-allowed"
            @click="emit('loadCurrentSshHostKeyInfo')"
          >
            {{ isLoadingHostKeyInfo ? 'Loading...' : 'Show SSH Fingerprint' }}
          </button>

          <div v-else class="rounded-md border border-destructive/30 bg-destructive/5 p-2 text-xs">
            <p><span class="font-medium">Host:</span> {{ sshHostKeyInfo.pattern }}</p>
            <p><span class="font-medium">Type:</span> {{ sshHostKeyInfo.keyType }}</p>
            <p class="break-all"><span class="font-medium">Fingerprint:</span> {{ sshHostKeyInfo.fingerprint }}</p>

            <button
              type="button"
              class="mt-2 inline-flex h-8 items-center justify-center rounded-full border border-destructive/60 bg-background/95 px-3 text-xs font-medium text-destructive transition-colors hover:bg-destructive/10"
              @click="emit('copyCurrentSshFingerprint')"
            >
              Copy Fingerprint
            </button>

            <div class="mt-2 space-y-1">
              <label for="expectedSshFingerprint" class="font-medium">Expected Fingerprint (Optional)</label>
              <input
                id="expectedSshFingerprint"
                type="text"
                placeholder="SHA256:..."
                :value="expectedSshFingerprint"
                class="w-full rounded-md border border-destructive/40 bg-background px-2 py-1 text-xs focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1"
                @input="emit('updateExpectedSshFingerprint', ($event.target as HTMLInputElement).value)"
              />
              <p
                v-if="expectedSshFingerprint.trim()"
                :class="isFingerprintMatch ? 'text-green-600' : 'text-destructive'"
              >
                {{ isFingerprintMatch ? 'Fingerprint matches expected value.' : 'Fingerprint does not match expected value.' }}
              </p>
              <p v-if="pinnedSshFingerprint" class="break-all text-muted-foreground">
                Pinned fingerprint: {{ pinnedSshFingerprint }}
              </p>
              <p v-if="isPinnedFingerprintMismatch" class="text-destructive">
                Fingerprint differs from previously trusted host key for this host.
              </p>

              <div
                v-if="isPinnedFingerprintMismatch"
                class="mt-2 space-y-1 rounded-md border border-destructive/30 bg-background/70 p-2"
              >
                <label for="sshRotationReason" class="font-medium">Rotation Reason</label>
                <input
                  id="sshRotationReason"
                  type="text"
                  placeholder="Explain why this host key changed"
                  :value="sshRotationReason"
                  class="w-full rounded-md border border-destructive/30 bg-background px-2 py-1 text-xs focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1"
                  @input="emit('updateSshRotationReason', ($event.target as HTMLInputElement).value)"
                />
                <label for="sshRotationConfirmText" class="font-medium">Type ROTATE to confirm</label>
                <input
                  id="sshRotationConfirmText"
                  type="text"
                  placeholder="ROTATE"
                  :value="sshRotationConfirmText"
                  class="w-full rounded-md border border-destructive/30 bg-background px-2 py-1 text-xs focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1"
                  @input="emit('updateSshRotationConfirmText', ($event.target as HTMLInputElement).value)"
                />
                <button
                  type="button"
                  class="inline-flex h-8 items-center justify-center rounded-full border border-destructive/60 bg-background/95 px-3 text-xs font-medium text-destructive transition-colors hover:bg-destructive/10"
                  @click="emit('acceptPinnedFingerprintRotation')"
                >
                  Accept Rotation
                </button>
              </div>
            </div>

            <button
              type="button"
              :disabled="isTrustingHost || isFingerprintMismatch || isPinnedFingerprintMismatch"
              class="mt-2 inline-flex h-8 items-center justify-center rounded-full border border-destructive/60 bg-background/95 px-3 text-xs font-medium text-destructive transition-colors hover:bg-destructive/10 disabled:cursor-not-allowed"
              @click="emit('trustCurrentSshHost')"
            >
              {{ isTrustingHost ? 'Trusting...' : 'Trust SSH Host Key' }}
            </button>
          </div>

          <div v-if="sshTrustAudit.length" class="mt-3 rounded-md border border-destructive/30 bg-background/70 p-2">
            <div class="mb-1 flex items-center justify-between gap-2">
              <p class="font-medium">Recent Trusted Hosts</p>
              <div class="flex items-center gap-1">
                <button
                  type="button"
                  class="inline-flex h-7 items-center justify-center rounded-full border border-destructive/60 bg-background/95 px-2 text-[11px] font-medium text-destructive transition-colors hover:bg-destructive/10"
                  @click="emit('importSshTrustAudit')"
                >
                  Import
                </button>
                <button
                  type="button"
                  class="inline-flex h-7 items-center justify-center rounded-full border border-destructive/60 bg-background/95 px-2 text-[11px] font-medium text-destructive transition-colors hover:bg-destructive/10"
                  @click="emit('exportSshTrustAudit')"
                >
                  Export
                </button>
                <button
                  type="button"
                  class="inline-flex h-7 items-center justify-center rounded-full border border-destructive/60 bg-background/95 px-2 text-[11px] font-medium text-destructive transition-colors hover:bg-destructive/10"
                  @click="emit('clearSshTrustAudit')"
                >
                  Clear
                </button>
              </div>
            </div>

            <input
              type="text"
              placeholder="Search host or fingerprint"
              :value="sshTrustAuditSearch"
              class="mb-1 w-full rounded-md border border-destructive/30 bg-background px-2 py-1 text-[11px] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1"
              @input="emit('updateSshTrustAuditSearch', ($event.target as HTMLInputElement).value)"
            />

            <div
              v-for="entry in filteredSshTrustAudit.slice(0, 3)"
              :key="`${entry.pattern}-${entry.trustedAt}`"
              class="border-t border-destructive/20 py-1 first:border-t-0"
            >
              <p class="break-all">{{ entry.pattern }}</p>
              <p class="break-all text-[11px]">{{ entry.fingerprint }}</p>
              <p class="text-[11px] text-muted-foreground">{{ new Date(entry.trustedAt).toLocaleString() }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
