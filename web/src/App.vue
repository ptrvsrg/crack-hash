<script setup lang="ts">
import { RouterView } from 'vue-router'
import Header from '@/components/header/PageHeader.vue'
import PageFooter from '@/components/footer/PageFooter.vue'
import ConfigProvider from '@/providers/config/ConfigProvider.vue'
import MessageProvider from '@/providers/message/MessageProvider.vue'
import LoadingProvider from '@/providers/loading/LoadingProvider.vue'
</script>

<template>
  <ConfigProvider>
    <MessageProvider>
      <LoadingProvider>
        <div class="wrapper">
          <div class="header-cell">
            <Header />
          </div>
          <div class="content-cell">
            <RouterView />
          </div>
          <div class="footer-cell">
            <PageFooter />
          </div>
        </div>
      </LoadingProvider>
    </MessageProvider>
  </ConfigProvider>
</template>

<style scoped>
.wrapper {
  display: grid;
  height: 100%;
  grid-template-rows: auto 1fr auto;
  grid-template-columns: 1fr;
  padding: 10px;
  grid-row-gap: 10px;
  grid-column-gap: 10px;
  box-sizing: border-box;
  width: 100%;
}

.header-cell {
  grid-area: 1 / 1 / 2 / 2;
}

.content-cell {
  grid-area: 2 / 1 / 3 / 2;
  overflow-y: auto;
  margin: auto;
  width: 1000px;
  height: 100%;

  @media screen and (max-width: 1000px) {
    width: 100%;
  }
}

.footer-cell {
  grid-area: 3 / 1 / 4 / 2;
}
</style>
