<script setup lang="ts">
import StatusDoughnutChart from '@/views/task/task-info/subtasks-status/status-doughnut-chart/StatusDoughnutChart.vue'
import type { SubtasksStatusProps } from '@/views/task/task-info/subtasks-status/props.ts'
import { NH3, NText } from 'naive-ui'
import { useI18n } from 'vue-i18n'

const { subtaskStatuses } = defineProps<SubtasksStatusProps>()

const { t } = useI18n()
</script>

<template>
  <n-h3 style="margin: 0">
    <n-text>{{ t('subtaskStatuses') }}:</n-text>
  </n-h3>
  <StatusDoughnutChart :subtaskStatuses="subtaskStatuses" />
</template>

<style scoped></style>
