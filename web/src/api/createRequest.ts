import axios, { type AxiosError } from 'axios'
import type { AxiosRequestConfig } from 'axios'
import type { ErrorOutput } from '@/model/common.ts'

const axiosInstance = axios.create({
  timeout: 100_000,
  headers: {
    'Content-Type': 'application/json',
  },
})

export async function createRequest<T>(config: AxiosRequestConfig): Promise<T> {
  try {
    const { data } = await axiosInstance.request(config)
    return data
  } catch (e: unknown) {
    const error = e as AxiosError<ErrorOutput, T>

    if (axios.isAxiosError(error)) {
      error.message = error.response?.data?.message ?? error.message
    }

    throw error
  }
}
