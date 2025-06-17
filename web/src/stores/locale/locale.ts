import { createI18n } from 'vue-i18n'
import en from '@/assets/locales/en.json'
import ru from '@/assets/locales/ru.json'
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const i18n = createI18n({
  legacy: false,
  locale: getLocaleInStorage(),
  fallbackLocale: 'en',
  messages: { en, ru },
})

export const useLocaleStore = defineStore(
  'locale',
  () => {
    const theme = ref('en')

    function getSupportedLocales(): string[] {
      return ['en', 'ru']
    }

    function changeLocale(locale: string): void {
      i18n.global.locale.value = locale
      localStorage.setItem(localeKey, locale)
    }

    return { theme, changeTheme }
  },
  {
    persist: {},
  },
)
