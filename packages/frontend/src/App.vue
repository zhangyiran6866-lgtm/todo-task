<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useThemeStore } from '@/stores/use-theme-store'

const themeStore = useThemeStore()
const { locale } = useI18n()

onMounted(() => {
  themeStore.initialize()
  locale.value = themeStore.language
})

watch(
  () => themeStore.language,
  (nextLanguage) => {
    locale.value = nextLanguage
  }
)
</script>

<template>
  <RouterView />
</template>
