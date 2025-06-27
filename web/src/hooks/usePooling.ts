import type { Ref } from 'vue'
import { onUnmounted, ref } from 'vue'

export interface UsePoolingReturn {
  pooling: Ref<boolean>
}

export function usePooling(callback: () => void, interval: number): UsePoolingReturn {
  const pooling = ref(true)

  const timer = setInterval(() => {
    if (!pooling.value) {
      clearInterval(timer)
      return
    }
    callback()
  }, interval)

  onUnmounted(() => {
    clearInterval(timer)
  })

  return {
    pooling,
  }
}
