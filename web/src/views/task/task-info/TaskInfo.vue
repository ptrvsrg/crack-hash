<script setup lang="ts">
import type { TaskInfoProps } from '@/views/task/task-info/props.ts'
import TaskResult from '@/views/task/task-info/task-result/TaskResult.vue'
import SubtasksStatus from '@/views/task/task-info/subtasks-status/SubtasksStatus.vue'
import TaskStatus from '@/views/task/task-info/task-status/TaskStatus.vue'
import TaskProgress from '@/views/task/task-info/task-progress/TaskProgress.vue'

const { task } = defineProps<TaskInfoProps>()
</script>

<template>
  <div class="task-info">
    <TaskResult :data="task.data" />
    <TaskStatus :status="task.status" />
    <TaskProgress :percent="task.percent" :status="task.status" />
    <SubtasksStatus :subtaskStatuses="task.subtasks.map((t) => t.status)" />
  </div>
</template>

<style scoped>
.task-info {
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  height: fit-content;
  width: 80%;
  gap: 20px;
}
</style>
