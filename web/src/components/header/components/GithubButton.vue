<script setup lang="ts">
import { NButton, NIcon } from 'naive-ui'
import { Github } from '@vicons/fa'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const onClick = () => {
  window.open('https://github.com/ptrvsrg/crack-hash', '_blank')
}
</script>

<template>
  <n-button circle quaternary @click="onClick">
    <template #icon>
      <n-icon size="28" color="#000">
        <Github />
      </n-icon>
    </template>
  </n-button>
</template>
