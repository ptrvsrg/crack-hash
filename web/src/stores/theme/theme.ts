import { ref } from 'vue'
import { defineStore } from 'pinia'

export const useThemeStore = defineStore(
  'theme',
  () => {
    const theme = ref<'light' | 'dark'>('light')

    function changeTheme(t: 'dark' | 'light'): void {
      theme.value = t
      window.location.reload()
    }

    return { theme, changeTheme }
  },
  {
    persist: {},
  },
)
