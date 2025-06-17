<script setup lang="ts">
import { NButton, NIcon, NDropdown } from 'naive-ui'
import { DarkTheme20Filled } from '@vicons/fluent'
import { useI18n } from 'vue-i18n'
import { useThemeStore } from '@/stores/theme/theme.ts'

const { changeTheme } = useThemeStore()

const { t } = useI18n()

const options = [
  {
    key: 'dark',
    label: t('darkTheme'),
  },
  {
    key: 'light',
    label: t('lightTheme'),
  },
]

const handleSelect = changeTheme
</script>

<template>
  <n-dropdown trigger="click" :options="options" @select="handleSelect">
    <n-button circle quaternary>
      <template #icon>
        <n-icon size="30" color="#000">
          <DarkTheme20Filled />
        </n-icon>
      </template>
    </n-button>
  </n-dropdown>
</template>
