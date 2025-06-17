<script setup lang="ts">
import StatusDoughnutChart from '@/views/task/task-info/subtasks-status/status-doughnut-chart/StatusDoughnutChart.vue'
import type { SubtaskStatusProps } from '@/views/task/task-info/subtasks-status/props.ts'
import { NCollapse, NCollapseItem, NH3, NText } from 'naive-ui'
import { useI18n } from 'vue-i18n'

const props = defineProps<SubtaskStatusProps>()

const { t } = useI18n()
</script>

<template>
  <n-h3 style="margin: 0">
    <n-text>{{ t('subtaskStatuses') }}:</n-text>
  </n-h3>
  <n-collapse default-expanded-names="0">
    <n-collapse-item title="Doughnut" name="1">
      <StatusDoughnutChart :subtasks="props.subtasks" />
    </n-collapse-item>
    <n-collapse-item title="Progress bars" name="2"> </n-collapse-item>
  </n-collapse>
</template>

<style scoped>
</style>
