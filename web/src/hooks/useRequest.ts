import { type Ref, ref } from 'vue'

export interface UseRequestReturn<T> {
  data: Ref<T | null>
  loading: Ref<boolean>
  error: Ref<Error | null>
  apiCall: () => Promise<void>
}

export function useRequest<T>(apiCall: () => Promise<T | null>): UseRequestReturn<T> {
  const data = ref<T | null>(null) as Ref<T | null>
  const loading = ref(false)
  const error = ref<Error | null>(null)

  const fetch = async () => {
    loading.value = true
    error.value = null

    try {
      data.value = await apiCall()
    } catch (e) {
      error.value = e instanceof Error ? e : new Error(String(e))
    } finally {
      loading.value = false
    }
  }

  return {
    data,
    loading,
    error,
    apiCall: fetch,
  }
}
