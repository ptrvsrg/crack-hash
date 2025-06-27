import { useAxios } from '../hooks/useAxios.ts'

const url = '/assets/config/appConfig.json'

export interface AppConfig {
  managerUrl: string
}

async function loadConfigJson(): Promise<AppConfig> {
  try {
    const response = await useAxios().get<AppConfig>(url)
    return response.data
  } catch (e) {
    console.error(e)
    throw e
  }
}

const appConfig = await loadConfigJson()

export default appConfig
