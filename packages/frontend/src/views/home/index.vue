<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import gsap from 'gsap'
import { useAuthStore } from '@/stores/use-auth-store'

const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()
const heroRef = ref<HTMLElement | null>(null)

function handleStart() {
  if (authStore.isLoggedIn) {
    router.push({ name: 'tasks' })
  } else {
    router.push({ name: 'login' })
  }
}

onMounted(() => {
  if (!heroRef.value) return
  const timeline = gsap.timeline({
    defaults: {
      ease: 'power4.out'
    }
  })
  timeline
    .from('.hero-title', { y: 32, opacity: 0, duration: 1 })
    .from('.hero-subtitle', { y: 24, opacity: 0, duration: 0.95 }, '-=0.55')
    .from('.hero-action', { y: 20, opacity: 0, duration: 0.85 }, '-=0.45')
})
</script>

<template>
  <main class="min-h-screen flex flex-col items-center justify-center bg-[var(--bg-dark)]">
    <div
      ref="heroRef"
      class="text-center space-y-6 px-6"
    >
      <h1 class="hero-title text-5xl md:text-6xl font-bold text-neon tracking-widest shadow-neon">
        {{ t('common.appName') }}
      </h1>
      <p class="hero-subtitle text-[var(--text-secondary)] text-base md:text-lg">
        {{ t('home.subtitle') }}
      </p>
      <button
        id="btn-start"
        class="hero-action mt-8 px-8 py-3 bg-glass border border-neon text-neon rounded-lg shadow-neon
               hover:bg-[var(--neon-glow)] transition-all duration-300 cursor-pointer text-base font-medium"
        @click="handleStart"
      >
        {{ t('home.start') }}
      </button>
    </div>
  </main>
</template>
