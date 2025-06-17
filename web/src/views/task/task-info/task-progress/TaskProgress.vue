<script setup lang="ts">
import { NDescriptions, NDescriptionsItem, NH3, NTag, NText } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import type { TaskStatusProps } from '@/views/task/task-info/task-status/props.ts'
import { HashCrackTaskStatus } from '@/model/hash-crack.ts'

const { status } = defineProps<TaskStatusProps>()

const { t } = useI18n()

const statusTagTypeFunc = (status: HashCrackTaskStatus) => {
  switch (status) {
    case HashCrackTaskStatus.ERROR:
      return 'error'
    case HashCrackTaskStatus.READY:
      return 'success'
    case HashCrackTaskStatus.PARTIAL_READY:
      return 'success'
    case HashCrackTaskStatus.IN_PROGRESS:
      return 'info'
    case HashCrackTaskStatus.PENDING:
      return 'warning'
    default:
      return 'default'
  }
}

const statusTextFunc = (status: HashCrackTaskStatus) => {
  switch (status) {
    case HashCrackTaskStatus.ERROR:
      return t('error')
    case HashCrackTaskStatus.READY:
      return t('ready')
    case HashCrackTaskStatus.PARTIAL_READY:
      return t('partialReady')
    case HashCrackTaskStatus.IN_PROGRESS:
      return t('inProgress')
    case HashCrackTaskStatus.PENDING:
      return t('pending')
    default:
      return t('unknown')
  }
}

const statusTagType = statusTagTypeFunc(status)
const statusText = statusTextFunc(status)
</script>

<template>
  <n-descriptions size="large" label-placement="left" separator="">
    <n-descriptions-item>
      <template #label>
        <n-h3 class="inline-text margin-right">
          <n-text> {{ t('taskStatus') }}:</n-text>
        </n-h3>
      </template>
      <div class="found-word-list">
        <n-tag :type="statusTagType" round :bordered="false">
          {{ statusText }}
        </n-tag>
      </div>
    </n-descriptions-item>
  </n-descriptions>
</template>

<style scoped>
.inline-text {
  display: inline;
}

.margin-right {
  margin-right: 10px;
}

.found-word-list {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}
</style>
