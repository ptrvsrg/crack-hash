import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import i18n from '@/i18n'

type SupportedLocale = 'en' | 'ru'

export const useLocaleStore = defineStore(
  'locale',
  () => {
    const locale = ref<SupportedLocale>('en')

    function getSupportedLocales(): SupportedLocale[] {
      return ['en', 'ru']
    }

    function changeLocale(newLocale: SupportedLocale) {
      locale.value = newLocale
    }

    watch(locale, (newLocale: SupportedLocale) => {
      i18n.global.locale.value = newLocale
    })

    return { locale, changeLocale, getSupportedLocales }
  },
  {
    persist: {},
  },
)
