<script setup lang="ts">
import type { TaskStatusProps } from '@/views/task/task-info/task-result/props.ts'
import { NDescriptions, NDescriptionsItem, NH3, NTag, NText } from 'naive-ui'
import { useI18n } from 'vue-i18n'

const { task } = defineProps<TaskStatusProps>()

const { t } = useI18n()

</script>

<template>
  <n-descriptions size="large" label-placement="left" separator="">
    <n-descriptions-item>
      <template #label>
        <n-h3 class="inline-text margin-right">
          <n-text>
            {{ t('wordsFound') }}:
          </n-text>
        </n-h3>
      </template>
      <div class="found-word-list">
        <n-tag v-for="result in task.data" :key="result" type="success" round :bordered="false">
          {{ result }}
        </n-tag>
        <n-text class="secondary-text" v-if="task.data.length === 0" italic>
          {{ t('wordsNotFound') }}
        </n-text>
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

.secondary-text {
  color: darkgrey;
}
</style>
