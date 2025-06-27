import type { HashCrackTaskStatus } from '@/model/hash-crack.ts'

export type TaskProgressProps = {
  percent: number
  status: HashCrackTaskStatus
}
