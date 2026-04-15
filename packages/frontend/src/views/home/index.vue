<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/use-auth-store'

const router = useRouter()
const authStore = useAuthStore()

function handleStart() {
  if (authStore.isLoggedIn) {
    router.push({ name: 'tasks' })
  } else {
    router.push({ name: 'login' })
  }
}
</script>

<template>
  <main class="min-h-screen flex flex-col items-center justify-center bg-[var(--bg-dark)]">
    <div class="text-center space-y-6">
      <h1 class="text-6xl font-bold text-neon tracking-widest shadow-neon">
        TodoTask
      </h1>
      <p class="text-[var(--text-secondary)] text-lg">
        沉浸式个人任务管理系统
      </p>
      <button
        id="btn-start"
        class="mt-8 px-8 py-3 bg-glass border border-neon text-neon rounded-lg shadow-neon
               hover:bg-[var(--neon-glow)] transition-all duration-300 cursor-pointer text-base font-medium"
        @click="handleStart"
      >
        立即开始
      </button>
    </div>
  </main>
</template>
