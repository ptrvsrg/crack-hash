<script setup lang="ts">
import { NResult, NButton } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import router from '@/router'

const { t } = useI18n()
const onClick = () => {
  router.replace('/')
}
</script>

<template>
  <n-result :title="t('notFound')" :description="t('notFoundDescription')">
    <template #icon />
    <template #footer>
      <n-button @click="onClick">{{ t('goToHome') }}</n-button>
    </template>
  </n-result>
</template>
