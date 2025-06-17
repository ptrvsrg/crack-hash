<script setup lang="ts">
import { NResult, NButton } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import type { NotFoundProps } from '@/components/common/not-found/properties.ts'
import { useRouter } from 'vue-router'

const { push } = useRouter()
const { t } = useI18n()

const onClick = () => push('/')
const props = defineProps<NotFoundProps>()
</script>

<template>
  <n-result :title="props.title ?? t('notFound')" :description="props.description ?? t('notFoundDescription')">
    <template #icon />
    <template #footer>
      <n-button @click="onClick">{{ t('goToHome') }}</n-button>
    </template>
  </n-result>
</template>
