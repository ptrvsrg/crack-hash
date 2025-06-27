<script setup lang="ts">
import { NH3, NText, NProgress, useThemeVars } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import type { TaskProgressProps } from '@/views/task/task-info/task-progress/props.ts'
import { HashCrackTaskStatus } from '@/model/hash-crack.ts'

const { percent, status } = defineProps<TaskProgressProps>()

const themeVars = useThemeVars()
const getProgressColor = (status: HashCrackTaskStatus) => {
  switch (status) {
    case HashCrackTaskStatus.ERROR:
      return themeVars.value.successColor
    case HashCrackTaskStatus.READY:
      return themeVars.value.successColor
    case HashCrackTaskStatus.PARTIAL_READY:
      return themeVars.value.successColor
    case HashCrackTaskStatus.IN_PROGRESS:
      return themeVars.value.infoColor
    case HashCrackTaskStatus.PENDING:
      return themeVars.value.warningColor
    default:
      return themeVars.value.avatarColor
  }
}

const { t } = useI18n()

const progressColor = getProgressColor(status)
</script>

<template>
  <n-h3 style="margin: 0">
    <n-text>{{ t('taskProgress') }}:</n-text>
  </n-h3>
  <n-progress
    type="line"
    :height="15"
    :color="progressColor"
    :percentage="Math.round(percent * 1000) / 1000"
    :processing="status === HashCrackTaskStatus.IN_PROGRESS"
  />
</template>
