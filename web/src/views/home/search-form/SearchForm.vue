<script setup lang="ts">
import {
  NH1,
  NInput,
  NInputNumber,
  NButton,
  NForm,
  NFormItem,
  useMessage,
  type FormInst,
  useLoadingBar,
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ref, watch } from 'vue'
import { useRequest } from '@/hooks/useRequest.ts'
import { createHashCrackTask } from '@/api/manager/hash-crack-api.ts'
import type { HashCrackTaskInput } from '@/model/hash-crack.ts'
import type { FormRules } from 'naive-ui/es/form/src/interface'

const loadingBar = useLoadingBar()
const message = useMessage()
const { push } = useRouter()
const { t } = useI18n()
const { data, loading, error, apiCall } = useRequest(async () => createHashCrackTask(formValue.value))

const formRef = ref<FormInst | null>(null)
const formValue = ref<HashCrackTaskInput>({
  hash: '',
  maxLength: 1,
})
const rules: FormRules = {
  hash: {
    type: 'string',
    required: true,
    message: t('errorRequiredHash'),
    trigger: 'blur',
  },
  maxLength: {
    type: 'number',
    required: true,
    message: t('errorRequiredMaxLength'),
    trigger: 'blur',
  },
}

const handleClick = async (e: MouseEvent) => {
  e.preventDefault()

  try {
    await formRef.value?.validate()
  } catch (error) {
    // Обработка не требуется, потому что у FormItem появятся сообщения об ошибках
    console.error(`Validation error:`, error)
    return
  }

  await apiCall()

  if (error.value) {
    console.error('API call error:', error.value)
    message.error(error.value.message)
    return
  }

  if (data.value) {
    await push({
      path: `/tasks/${data.value.requestId}`,
    })
  }
}

watch(loading, () => {
  if (loading.value) {
    loadingBar.start()
  } else {
    loadingBar.finish()
  }
})
</script>

<template>
  <n-h1>{{ t('enterData') }}</n-h1>
  <n-form ref="formRef" :model="formValue" :rules="rules">
    <n-form-item class="hash-form-item" :label="t('labelHash')" path="hash">
      <n-input v-model:value="formValue.hash" :placeholder="t('enterHash')" />
    </n-form-item>
    <n-form-item class="max-length-form-item" :label="t('labelMaxLength')" path="maxLength">
      <n-input-number v-model:value="formValue.maxLength" :placeholder="t('enterMaxLength')" min="1" />
    </n-form-item>
    <n-form-item label-placement="left" label-width="0">
      <n-button type="info" :loading="loading" @click="handleClick">{{ t('startBruteForce') }} </n-button>
    </n-form-item>
  </n-form>
</template>

<style scoped>
.n-form {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: center;
  align-items: flex-end;
  width: 100%;
  box-sizing: border-box;
}

.n-form-item {
  margin: 0;
}

.hash-form-item {
  flex: 1 1 auto;
  min-width: 250px;
}

.max-length-form-item {
  flex: 0 0 auto;
  min-width: 150px;
}

@media screen and (max-width: 1000px) {
  .n-form {
    flex-direction: column;
    align-items: stretch;
  }

  .hash-form-item,
  .max-length-form-item {
    width: 100%;
  }
}
.n-form-item.n-form-item--top-labelled {
  grid-template-columns: none;
}
</style>
