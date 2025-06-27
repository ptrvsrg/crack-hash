import './assets/main.css'
import 'vfonts/Roboto.css'

import { createApp } from 'vue'
import App from '@/App.vue'
import pinia from '@/stores/pinia'
import i18n from '@/i18n'
import router from '@/router'

createApp(App).use(pinia).use(i18n).use(router).mount('#app')
