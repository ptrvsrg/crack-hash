<script setup lang="ts">
import { NSpin } from 'naive-ui'
</script>

<template>
  <div class="spin-wrapper">
    <n-spin size="large" />
  </div>
</template>

<style scoped>
.spin-wrapper {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  justify-content: center;
  align-items: center;
}
</style>
