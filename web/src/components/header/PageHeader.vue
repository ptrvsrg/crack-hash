<script setup lang="ts">
import { NPageHeader, NAvatar, NH2 } from 'naive-ui'
import logo from '@/assets/logo.svg'
import LanguageSwitcher from '@/components/common/header/components/LanguageSwitcher.vue'
import GithubButton from '@/components/common/header/components/GithubButton.vue'
import ThemeSwitcher from '@/components/common/header/components/ThemeSwitcher.vue'
</script>

<template>
  <n-page-header>
    <template #title>
      <RouterLink to="/" class="logo-link">
        <n-h2>Crack HASH</n-h2>
      </RouterLink>
    </template>
    <template #avatar>
      <RouterLink to="/" class="logo-link">
        <n-avatar :src="logo" color="#fff0" href="/" />
      </RouterLink>
    </template>
    <template #extra>
      <div class="extra-wrapper">
        <ThemeSwitcher />
        <LanguageSwitcher />
        <GithubButton />
      </div>
    </template>
  </n-page-header>
</template>

<style scoped>
.logo-link {
  text-decoration: none;
  display: flex;
}

.n-h2 {
  margin: 0;
}

a {
  text-decoration: none;
  margin: 0;
  padding: 0;
}

.extra-wrapper {
  display: flex;
  gap: 5px;
  align-items: center;
  justify-content: center;
  flex-direction: row;
}
</style>
