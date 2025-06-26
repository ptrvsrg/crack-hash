<script setup lang="ts">
import { HashCrackSubtaskStatus } from '@/model/hash-crack.ts'
import { Doughnut } from 'vue-chartjs'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { useI18n } from 'vue-i18n'
import { computed, ref, watch } from 'vue'
import type { StatusDoughnutChartProps } from '@/views/task/task-info/subtasks-status/status-doughnut-chart/props.ts'

const props = defineProps<StatusDoughnutChartProps>()
const { t, locale } = useI18n()

ChartJS.register(ArcElement, Tooltip, Legend)

const getStatusLabel = (status: HashCrackSubtaskStatus) => {
  const labels = {
    [HashCrackSubtaskStatus.SUCCESS]: t('ready'),
    [HashCrackSubtaskStatus.ERROR]: t('error'),
    [HashCrackSubtaskStatus.IN_PROGRESS]: t('inProgress'),
    [HashCrackSubtaskStatus.UNKNOWN]: t('unknown'),
    [HashCrackSubtaskStatus.PENDING]: t('pending'),
  }
  return labels[status] || t('unknown')
}

const getStatusColor = (status: HashCrackSubtaskStatus) => {
  const colors = {
    [HashCrackSubtaskStatus.SUCCESS]: '#4CAF50',
    [HashCrackSubtaskStatus.ERROR]: '#E53935',
    [HashCrackSubtaskStatus.IN_PROGRESS]: '#1E88E5',
    [HashCrackSubtaskStatus.UNKNOWN]: '#9E9E9E',
    [HashCrackSubtaskStatus.PENDING]: '#FFB300',
  }
  return colors[status] || '#9E9E9E'
}

const chartData = computed(() => {
  const statuses = Object.values(HashCrackSubtaskStatus)
  return {
    labels: statuses.map(getStatusLabel),
    datasets: [
      {
        backgroundColor: statuses.map(getStatusColor),
        data: statuses.map((status) => props.subtaskStatuses.filter((s) => s === status).length),
      },
    ],
  }
})

const chartKey = ref(0)
const options = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'right' as const,
      align: 'start' as const,
    },
  },
}

watch(locale, () => {
  chartKey.value++ // force chart update
})
</script>

<template>
  <div class="chart-wrapper">
    <Doughnut :key="chartKey" :data="chartData" :options="options" />
  </div>
</template>

<style scoped>
.chart-wrapper {
  height: 100%;
  width: auto;
}
</style>
