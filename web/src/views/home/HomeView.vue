<script setup lang="ts">
import SearchForm from '@/views/home/search-form/SearchForm.vue'
import RequestHistory from '@/views/home/request-history/RequestHistory.vue'
</script>

<template>
  <main>
    <SearchForm />
    <RequestHistory />
  </main>
</template>

<style scoped>
main {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: fit-content;
  width: 100%;
  padding: 20px;
  box-sizing: border-box;
}
</style>
