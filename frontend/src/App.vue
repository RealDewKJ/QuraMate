<script lang="ts" setup>
import { ref } from 'vue';
import DbConnection from './components/DbConnection.vue';
import DbDashboard from './components/DbDashboard.vue';

const isConnected = ref(false);

const handleConnected = () => {
  isConnected.value = true;
};

const handleDisconnect = () => {
  isConnected.value = false;
};
</script>

<template>
  <DbConnection v-if="!isConnected" @connected="handleConnected" />
  <DbDashboard v-else @disconnect="handleDisconnect" />
</template>

<style>
/* Reset basic styles if necessary, but Tailwind should handle most */
html,
body,
#app {
  height: 100%;
  margin: 0;
}
</style>
