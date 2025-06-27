<script setup lang="ts">
import { NButton, NIcon, useThemeVars } from 'naive-ui'
import { Github } from '@vicons/fa'

const themeVars = useThemeVars()

const onClick = () => {
  window.open('https://github.com/ptrvsrg/crack-hash', '_blank')
}
</script>

<template>
  <n-button circle quaternary @click="onClick">
    <template #icon>
      <n-icon size="28" :color="themeVars.textColorBase">
        <Github />
      </n-icon>
    </template>
  </n-button>
</template>
