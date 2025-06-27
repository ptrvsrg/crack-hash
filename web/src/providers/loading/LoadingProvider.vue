<script setup lang="ts">
import { NLoadingBarProvider } from 'naive-ui'
</script>

<template>
  <n-loading-bar-provider>
    <slot />
  </n-loading-bar-provider>
</template>
