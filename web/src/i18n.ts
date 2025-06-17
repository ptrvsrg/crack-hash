import { createI18n } from 'vue-i18n'
import en from '@/assets/locales/en.json'
import ru from '@/assets/locales/ru.json'
import { useLocaleStore } from '@/stores/locale/locale.ts'
import { watch } from 'vue'
import { pinia } from '@/stores/pinia.ts'

export const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  messages: { en, ru },
})


const { locale } = useLocaleStore(pinia)
watch(locale, (newLocale) => {
  i18n.global.locale.value = newLocale
})
