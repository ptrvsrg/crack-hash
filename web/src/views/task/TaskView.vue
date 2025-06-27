<script setup lang="ts">
import { NH1, NText, NPopover, useLoadingBar, useMessage } from 'naive-ui'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useRequest } from '@/hooks/useRequest.ts'
import { getHashCrackTaskStatus } from '@/api/manager/hash-crack-api.ts'
import NotFound from '@/components/not-found/NotFound.vue'
import TaskInfo from '@/views/task/task-info/TaskInfo.vue'
import { onBeforeUnmount, onMounted, watch } from 'vue'
import { usePooling } from '@/hooks/usePooling.ts'
import { HashCrackTaskStatus } from '@/model/hash-crack.ts'
import PageSpinner from '@/components/spinner/PageSpinner.vue'
import axios from 'axios';

const {
  params: { id: taskId },
} = useRoute()
const loadingBar = useLoadingBar()
const message = useMessage()
const { t } = useI18n()
const { apiCall, data, loading, error } = useRequest(async () => getHashCrackTaskStatus({ requestID: taskId as string }))

const { pooling } = usePooling(() => {
  apiCall().then(() => {
    // check error
    if (error.value) {
      console.error('API call error:', error.value)
      message.error(error.value.message)
      pooling.value = false
      return
    }

    // check status
    if (data.value?.status !== HashCrackTaskStatus.IN_PROGRESS && data.value?.status !== HashCrackTaskStatus.PENDING) {
      pooling.value = false
    }
  })
}, 30000)

onMounted(() => {
  apiCall().then(() => {
    if (error.value) {
      console.error('API call error:', error.value)
      message.error(error.value.message)
    }
  })
})

onBeforeUnmount(() => {
  pooling.value = false
})

watch(loading, () => {
  if (loading.value) {
    loadingBar.start()
  } else {
    loadingBar.finish()
  }
})
</script>

<template>
  <main v-if="loading">
    <PageSpinner />
  </main>
  <main v-else-if="error">
    <NotFound
      v-if="axios.isAxiosError(error) && error.response?.status === 404"
      :title="t('taskNotFound', { id: taskId })"
      :description="t('taskNotFoundDescription')"
    />
    <NotFound
      v-else
      :title="t('taskError', { id: taskId })"
      :description="t('taskErrorDescription', { error: error.message })"
    />
  </main>
  <main v-else-if="!data">
    <NotFound
      :title="t('taskNotFound', { id: taskId })"
      :description="t('taskNotFoundDescription')"
    />
  </main>
  <main v-else>
    <n-h1 class="title">
      <n-text class="task-title-text">{{ t('task') }}</n-text>
      <n-popover trigger="hover" :delay="500" :duration="500">
        <template #trigger>
          <n-text italic code class="code task-id-text">{{ taskId }}</n-text>
        </template>
        <span>{{ taskId }}</span>
      </n-popover>
    </n-h1>
    <TaskInfo :task="data" />
  </main>
</template>

<style scoped>
main {
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  align-items: center;
  height: 100%;
  width: 100%;
  padding: 20px;
  box-sizing: border-box;
}

.title {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  white-space: nowrap;
}

.task-title-text {
  flex-shrink: 0;
}

.task-id-text {
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
  margin-left: 8px;
}
</style>
