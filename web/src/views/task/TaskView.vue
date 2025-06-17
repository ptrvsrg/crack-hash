<script setup lang="ts">
import {
  NH1,
  NInput,
  NInputNumber,
  NButton,
  NForm,
  NFormItem,
  useMessage,
  type FormInst
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ref } from 'vue'
import { useRequest } from '@/hooks/useRequest.ts'
import { createHashCrackTask } from '@/api/manager/hash-crack-api.ts'
import type { HashCrackTaskInput, HashCrackTaskIDOutput } from '@/model/hash-crack.ts'
import type { FormRules } from 'naive-ui/es/form/src/interface'

const message = useMessage()
const { push } = useRouter()
const { t } = useI18n()
const {
  data,
  loading,
  error,
  fetch
} = useRequest<HashCrackTaskIDOutput>(() => createHashCrackTask(formValue.value))

const formRef = ref<FormInst | null>(null)
const formValue = ref<HashCrackTaskInput>({
  hash: '',
  maxLength: 1
})
const rules: FormRules = {
  hash: {
    type: 'string',
    required: true,
    message: t('errorRequiredHash'),
    trigger: 'blur'
  },
  maxLength: {
    type: 'number',
    required: true,
    message: t('errorRequiredMaxLength'),
    trigger: 'blur'
  }
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

  await fetch()

  if (error.value) {
    console.error('API call error:', error.value)
    message.error(error.value.message)
    return
  }

  if (data.value) {
    await push({
      path: `/task/${data.value.requestId}`
    })
  }
}
</script>

<template>
  <main>
    <n-h1>{{ t('enterData') }}</n-h1>
    <n-form ref="formRef" :model="formValue" :rules="rules">
      <n-form-item class="hash-form-item" :label="t('labelHash')" path="hash">
        <n-input v-model:value="formValue.hash" :placeholder="t('enterHash')" />
      </n-form-item>
      <n-form-item class="max-length-form-item" :label="t('labelMaxLength')" path="maxLength">
        <n-input-number v-model:value="formValue.maxLength" :placeholder="t('enterMaxLength')" min="1" />
      </n-form-item>
      <n-form-item>
        <n-button type="info" :loading="loading" @click="handleClick">{{ t('startBruteForce') }} </n-button>
      </n-form-item>
    </n-form>
  </main>
</template>

<style scoped>
main {
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  align-items: center;
  height: fit-content;
  width: 100%;
  padding: 20px;
  box-sizing: border-box;
}

.n-form {
  display: inline-flex;
  gap: 10px;
  justify-content: center;
  align-items: flex-start;
  width: 800px;
  box-sizing: border-box;
}

.n-form-item {
  margin: 0;
}

.hash-form-item {
  width: 100%;
}

.max-length-form-item {
  width: fit-content;
}

.n-form-item.n-form-item--top-labelled {
  grid-template-columns: none;
}
</style>
