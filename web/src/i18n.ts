import { createI18n } from 'vue-i18n'
import en from '@/assets/locales/en.json'
import ru from '@/assets/locales/ru.json'

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  messages: { en, ru },
})

export default i18n
