<script lang="ts" setup>
import BootstrapperActionCard from './components/BootstrapperActionCard.vue';
import BootstrapperHeroPanel from './components/BootstrapperHeroPanel.vue';
import { useBootstrapper } from './composables/useBootstrapper';

const { canStart, context, primaryActionLabel, startInstall } = useBootstrapper();
</script>

<template>
  <main class="bootstrapper-shell">
    <div class="bootstrapper-noise"></div>
    <div class="bootstrapper-layout">
      <BootstrapperHeroPanel :version="context.version" />
      <BootstrapperActionCard
        :can-start="canStart"
        :primary-action-label="primaryActionLabel"
        @start="startInstall"
      />
    </div>
  </main>
</template>
