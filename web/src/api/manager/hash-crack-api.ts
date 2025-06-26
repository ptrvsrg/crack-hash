import type {
  HashCrackTaskIDOutput,
  HashCrackTaskInput,
  HashCrackTaskMetadataInput,
  HashCrackTaskMetadatasOutput,
  HashCrackTaskStatusOutput,
  HashCrackTaskStatusParams,
} from '@/model/hash-crack.ts'
import { createRequest } from '@/api/createRequest.ts'
import appConfig from '@/config/appConfig.ts'

const baseURL = appConfig.managerUrl

export async function createHashCrackTask(data: HashCrackTaskInput): Promise<HashCrackTaskIDOutput> {
  return await createRequest<HashCrackTaskIDOutput>({
    method: 'POST',
    baseURL,
    url: `/v1/hash/crack`,
    data,
  })
}

export async function getHashCrackTaskMetadatas(
  params: HashCrackTaskMetadataInput,
): Promise<HashCrackTaskMetadatasOutput> {
  return await createRequest<HashCrackTaskMetadatasOutput>({
    method: 'GET',
    baseURL,
    url: `/v1/hash/crack/metadatas`,
    params,
  })
}

export async function getHashCrackTaskStatus(params: HashCrackTaskStatusParams): Promise<HashCrackTaskStatusOutput> {
  return await createRequest<HashCrackTaskStatusOutput>({
    method: 'GET',
    baseURL,
    url: `/v1/hash/crack/status`,
    params,
  })
}
