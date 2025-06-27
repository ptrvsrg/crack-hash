<script setup lang="ts">
import { h, onMounted, reactive, watch } from 'vue'
import { NButton, NDataTable, NH2, type PaginationProps, useLoadingBar, useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useRequest } from '@/hooks/useRequest.ts'
import { getHashCrackTaskMetadatas } from '@/api/manager/hash-crack-api.ts'
import type { TableColumns } from 'naive-ui/es/data-table/src/interface'
import type { HashCrackTaskMetadataOutput } from '@/model/hash-crack.ts'
import { useRouter } from 'vue-router'

const loadingBar = useLoadingBar()
const message = useMessage()
const { t } = useI18n()
const { push, currentRoute } = useRouter()

const columns: TableColumns<HashCrackTaskMetadataOutput> = [
  {
    title: () => t('labelHash'),
    key: 'hash',
    resizable: true,
    width: '50%'
  },
  {
    title: () => t('labelMaxLength'),
    key: 'maxLength',
    resizable: true,
    width: '40%'
  },
  {
    key: 'click',
    resizable: true,
    render(row) {
      return h(
        NButton,
        {
          onClick: () => {
            push({
              path: `/tasks/${row.requestId}`
            })
          }
        },
        { default: () => t('more') }
      )
    }
  }
]


const page = currentRoute.value.query.historyPage && Number(currentRoute.value.query.historyPage) > 0 ? Number(currentRoute.value.query.historyPage) : 1

const pagination = reactive<PaginationProps>({
  simple: true,
  showSizePicker: true,
  page,
  pageSize: 5,
  pageCount: 1,
  pageSlot: 5
})

const { data, loading, error, apiCall } = useRequest(async () =>
  getHashCrackTaskMetadatas({
    limit: pagination.pageSize ?? 5,
    offset: ((pagination.page ?? 1) - 1) * (pagination.pageSize ?? 5)
  })
)

const handlePageChange = (currentPage: number) => {
  pagination.page = currentPage

  apiCall().then(() => {
    if (error.value) {
      console.error('API call error:', error.value)
      message.error(error.value.message)
      return
    }

    if (data.value) {
      pagination.pageCount = Math.ceil(data.value.count / (pagination.pageSize ?? 5));
      pagination.itemCount = data.value.count
    }
  })
}

const handlePageSizeChange = (pageSize: number) => {
  if (!pagination) return

  pagination.pageSize = pageSize
  pagination.page = 1

  apiCall().then(() => {
    if (error.value) {
      console.error('API call error:', error.value)
      message.error(error.value.message)
      return
    }

    if (data.value) {
      pagination.pageCount = Math.ceil(data.value.count / (pagination.pageSize ?? 5));
      pagination.itemCount = data.value.count
    }
  })
}

onMounted(() => {
  apiCall().then(() => {
    if (error.value) {
      console.error('API call error:', error.value)
      message.error(error.value.message)
      return
    }

    if (data.value && pagination) {
      pagination.pageCount = Math.ceil(data.value.count / (pagination.pageSize ?? 5));
      pagination.itemCount = data.value.count
    }
  })
})

watch(pagination, () => {
  push({
    query: {
      historyPage: pagination.page
    }
  })
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
  <n-h2>{{ t('requestHistory') }}</n-h2>
  <n-data-table
    remote
    :loading="loading"
    :columns="columns"
    :data="data?.tasks ?? []"
    :pagination="pagination"
    @update:page="handlePageChange"
    @update:pageSize="handlePageSizeChange"
  />
</template>

<style scoped></style>
