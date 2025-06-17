<script setup lang="ts">
import { NDropdown, NIcon, NButton } from 'naive-ui'
import { LanguageOutlined } from '@vicons/material'
import { useLocaleStore } from '@/stores/locale/locale.ts'
import { i18n } from '@/i18n.ts'

const { getSupportedLocales, changeLocale } = useLocaleStore()

const options = getSupportedLocales().map((locale) => {
  return {
    key: locale,
    label: i18n.global.t('currentLanguage', {}, { locale }),
  }
})

const handleSelect = changeLocale
</script>

<template>
  <n-dropdown trigger="click" :options="options" @select="handleSelect">
    <n-button circle quaternary>
      <template #icon>
        <n-icon size="30" color="#000">
          <LanguageOutlined />
        </n-icon>
      </template>
    </n-button>
  </n-dropdown>
</template>
