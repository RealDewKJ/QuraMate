<script lang="ts" setup>
defineProps<{
  version: string;
}>();
</script>

<template>
  <section class="hero-panel">
    <div class="hero-orb hero-orb-one"></div>
    <div class="hero-orb hero-orb-two"></div>
    <div class="hero-orb hero-orb-three"></div>

    <div class="hero-icon-wrap">
      <img class="hero-appicon" src="/appicon.png" alt="QuraMate app icon" />
    </div>

    <h1 class="hero-wordmark">QuraMate</h1>

    <div class="hero-meta">v{{ version }}</div>
  </section>
</template>
