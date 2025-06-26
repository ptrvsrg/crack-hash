<script setup lang="ts">
import { NMessageProvider } from 'naive-ui'
</script>

<template>
  <n-message-provider placement="top-right" :duration="5000" keepAliveOnHover closable :max="10">
    <slot />
  </n-message-provider>
</template>
