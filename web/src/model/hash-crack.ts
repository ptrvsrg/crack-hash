export interface HashCrackTaskInput {
  hash: string
  maxLength: number
}

export interface HashCrackTaskIDOutput {
  requestId: string
}

export interface HashCrackTaskStatusParams {
  requestID: string
}

export interface HashCrackTaskStatusOutput {
  status: HashCrackTaskStatus
  data: string[]
  percent: number
  subtasks: HashCrackSubtaskStatusOutput[]
}

export interface HashCrackSubtaskStatusOutput {
  status: HashCrackSubtaskStatus
  data: string[]
  percent: number
}

export interface HashCrackTaskMetadataInput {
  limit: number
  offset: number
}

export interface HashCrackTaskMetadataOutput {
  requestId: string
  createdAt: Date
  hash: string
  maxLength: number
}

export interface HashCrackTaskMetadatasOutput {
  count: number
  tasks: HashCrackTaskMetadataOutput[]
}

export enum HashCrackTaskStatus {
  PENDING = 'PENDING',
  IN_PROGRESS = 'IN_PROGRESS',
  READY = 'READY',
  PARTIAL_READY = 'PARTIAL_READY',
  ERROR = 'ERROR',
  UNKNOWN = 'UNKNOWN',
}

export enum HashCrackSubtaskStatus {
  PENDING = 'PENDING',
  IN_PROGRESS = 'IN_PROGRESS',
  SUCCESS = 'SUCCESS',
  ERROR = 'ERROR',
  UNKNOWN = 'UNKNOWN',
}
