<script setup lang="ts">
import { NConfigProvider, NGlobalStyle, enUS, dateEnUS, ruRU, dateRuRU, lightTheme, darkTheme } from 'naive-ui'
import { useLocaleStore } from '@/stores/locale/locale.ts'
import { useThemeStore } from '@/stores/theme/theme.ts'

const { locale: localeKey } = useLocaleStore()
const { theme: themeKey } = useThemeStore()

const locale = localeKey === 'ru' ? ruRU : enUS
const dateLocale = localeKey === 'ru' ? dateRuRU : dateEnUS
const theme = themeKey === 'dark' ? darkTheme : lightTheme
</script>

<template>
  <n-config-provider :theme="theme" :locale="locale" :date-locale="dateLocale">
    <n-global-style />
    <slot />
  </n-config-provider>
</template>
