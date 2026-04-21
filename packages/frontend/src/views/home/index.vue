<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import gsap from 'gsap'
import { useAuthStore } from '@/stores/use-auth-store'

const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()
const heroRef = ref<HTMLElement | null>(null)
const glowARef = ref<HTMLElement | null>(null)
const glowBRef = ref<HTMLElement | null>(null)
const glowMergeRef = ref<HTMLElement | null>(null)
let heroContext: gsap.Context | null = null
let glowTicker: ((time: number, deltaTime: number) => void) | null = null

type GlowPresetKey = 'soft' | 'standard' | 'vivid'

interface GlowPreset {
  velocity: number
  moveFactor: number
  pulseScale: number
  pulseOpacity: number
  pulseDuration: number
  glowSize: string
  mergeSize: string
  glowBlur: string
  mergeBlur: string
  opacityABase: number
  opacityAWave: number
  opacityAByMerge: number
  opacityBBase: number
  opacityBWave: number
  opacityBByMerge: number
  opacityMergeBase: number
  opacityMergeByRatio: number
  colorACore: string
  colorAEdge: string
  colorBCore: string
  colorBEdge: string
  colorMergeCore: string
  colorMergeMid: string
  colorMergeEdge: string
}

// Hero 光晕开关：改这一行即可切换档位（'soft' | 'standard' | 'vivid'）
const HERO_GLOW_PRESET: GlowPresetKey = 'soft'

const GLOW_PRESETS: Record<GlowPresetKey, GlowPreset> = {
  soft: {
    velocity: 1,
    moveFactor: 8,
    pulseScale: 1.28,
    pulseOpacity: 0.56,
    pulseDuration: 1.4,
    glowSize: 'min(54vw, 680px)',
    mergeSize: 'min(60vw, 760px)',
    glowBlur: '40px',
    mergeBlur: '36px',
    opacityABase: 0.36,
    opacityAWave: 0.14,
    opacityAByMerge: 0.08,
    opacityBBase: 0.34,
    opacityBWave: 0.14,
    opacityBByMerge: 0.08,
    opacityMergeBase: 0.02,
    opacityMergeByRatio: 0.44,
    colorACore: 'rgba(0, 170, 188, 0.44)',
    colorAEdge: 'rgba(0, 132, 148, 0.04)',
    colorBCore: 'rgba(150, 30, 190, 0.42)',
    colorBEdge: 'rgba(126, 26, 160, 0.04)',
    colorMergeCore: 'rgba(104, 116, 222, 0.5)',
    colorMergeMid: 'rgba(62, 172, 204, 0.12)',
    colorMergeEdge: 'rgba(130, 54, 182, 0.015)'
  },
  standard: {
    velocity: 1.18,
    moveFactor: 9,
    pulseScale: 1.35,
    pulseOpacity: 0.62,
    pulseDuration: 1.25,
    glowSize: 'min(56vw, 720px)',
    mergeSize: 'min(62vw, 800px)',
    glowBlur: '42px',
    mergeBlur: '38px',
    opacityABase: 0.42,
    opacityAWave: 0.16,
    opacityAByMerge: 0.1,
    opacityBBase: 0.4,
    opacityBWave: 0.16,
    opacityBByMerge: 0.1,
    opacityMergeBase: 0.03,
    opacityMergeByRatio: 0.5,
    colorACore: 'rgba(0, 196, 222, 0.52)',
    colorAEdge: 'rgba(0, 152, 172, 0.05)',
    colorBCore: 'rgba(168, 38, 218, 0.5)',
    colorBEdge: 'rgba(136, 30, 176, 0.05)',
    colorMergeCore: 'rgba(120, 132, 236, 0.58)',
    colorMergeMid: 'rgba(70, 190, 230, 0.14)',
    colorMergeEdge: 'rgba(148, 62, 198, 0.02)'
  },
  vivid: {
    velocity: 1.35,
    moveFactor: 10,
    pulseScale: 1.42,
    pulseOpacity: 0.68,
    pulseDuration: 1.1,
    glowSize: 'min(58vw, 760px)',
    mergeSize: 'min(64vw, 860px)',
    glowBlur: '44px',
    mergeBlur: '40px',
    opacityABase: 0.48,
    opacityAWave: 0.18,
    opacityAByMerge: 0.12,
    opacityBBase: 0.46,
    opacityBWave: 0.18,
    opacityBByMerge: 0.12,
    opacityMergeBase: 0.04,
    opacityMergeByRatio: 0.56,
    colorACore: 'rgba(0, 218, 245, 0.62)',
    colorAEdge: 'rgba(0, 170, 192, 0.06)',
    colorBCore: 'rgba(188, 46, 240, 0.6)',
    colorBEdge: 'rgba(150, 36, 194, 0.06)',
    colorMergeCore: 'rgba(136, 148, 248, 0.68)',
    colorMergeMid: 'rgba(78, 208, 244, 0.16)',
    colorMergeEdge: 'rgba(168, 70, 216, 0.025)'
  }
}

const glowPreset = GLOW_PRESETS[HERO_GLOW_PRESET]
const heroGlowVars = computed(() => ({
  '--hero-glow-size': glowPreset.glowSize,
  '--hero-glow-merge-size': glowPreset.mergeSize,
  '--hero-glow-blur': glowPreset.glowBlur,
  '--hero-glow-merge-blur': glowPreset.mergeBlur,
  '--hero-glow-a-core': glowPreset.colorACore,
  '--hero-glow-a-edge': glowPreset.colorAEdge,
  '--hero-glow-b-core': glowPreset.colorBCore,
  '--hero-glow-b-edge': glowPreset.colorBEdge,
  '--hero-glow-merge-core': glowPreset.colorMergeCore,
  '--hero-glow-merge-mid': glowPreset.colorMergeMid,
  '--hero-glow-merge-edge': glowPreset.colorMergeEdge
}))

function handleStart() {
  router.push({ name: 'tasks' })
}

function handleToLogin() {
  router.push({ name: 'login' })
}

onMounted(() => {
  if (!heroRef.value) return

  heroContext = gsap.context(() => {
    const timeline = gsap.timeline({
      defaults: {
        ease: 'power4.out'
      }
    })

    timeline
      .from('.hero-grid', { opacity: 0, scale: 1.06, duration: 1.1 })
      .from('.hero-glow', { opacity: 0, scale: 0.65, stagger: 0.12, duration: 0.9 }, '-=0.8')
      .from('.hero-title-line', { y: 42, opacity: 0, stagger: 0.12, duration: 1.05 }, '-=0.65')
      .from('.hero-subtitle', { y: 24, opacity: 0, duration: 0.85 }, '-=0.65')
      .from('.hero-cta-btn', { y: 20, opacity: 0, stagger: 0.1, duration: 0.75 }, '-=0.5')
      .from('.hero-metric', { y: 16, opacity: 0, stagger: 0.08, duration: 0.65 }, '-=0.45')

    const glowA = glowARef.value
    const glowB = glowBRef.value
    const glowMerge = glowMergeRef.value
    if (!glowA || !glowB || !glowMerge) return

    const stateA = { x: 280, y: 220, vx: 0.11 * glowPreset.velocity, vy: 0.09 * glowPreset.velocity, phase: 0.2 }
    const stateB = { x: 760, y: 520, vx: -0.1 * glowPreset.velocity, vy: -0.08 * glowPreset.velocity, phase: 1.1 }
    const nearDistance = 240
    const farDistance = 560
    const margin = 90

    const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value))

    const getBounds = () => {
      const root = heroRef.value?.closest('main')
      const rect = root?.getBoundingClientRect()
      return {
        width: rect?.width ?? window.innerWidth,
        height: rect?.height ?? window.innerHeight
      }
    }

    const initialBounds = getBounds()
    stateA.x = initialBounds.width * 0.28
    stateA.y = initialBounds.height * 0.32
    stateB.x = initialBounds.width * 0.72
    stateB.y = initialBounds.height * 0.68

    gsap.set([glowA, glowB, glowMerge], { xPercent: -50, yPercent: -50 })

    glowTicker = (_time, deltaTime) => {
      const step = Math.min(deltaTime, 3)
      const bounds = getBounds()

      stateA.x += stateA.vx * step * glowPreset.moveFactor
      stateA.y += stateA.vy * step * glowPreset.moveFactor
      stateB.x += stateB.vx * step * glowPreset.moveFactor
      stateB.y += stateB.vy * step * glowPreset.moveFactor

      if (stateA.x < margin || stateA.x > bounds.width - margin) stateA.vx *= -1
      if (stateA.y < margin || stateA.y > bounds.height - margin) stateA.vy *= -1
      if (stateB.x < margin || stateB.x > bounds.width - margin) stateB.vx *= -1
      if (stateB.y < margin || stateB.y > bounds.height - margin) stateB.vy *= -1

      stateA.x = clamp(stateA.x, margin, bounds.width - margin)
      stateA.y = clamp(stateA.y, margin, bounds.height - margin)
      stateB.x = clamp(stateB.x, margin, bounds.width - margin)
      stateB.y = clamp(stateB.y, margin, bounds.height - margin)

      const now = performance.now() / 1000
      const breatheA = 0.92 + 0.22 * Math.sin(now * 2.35 + stateA.phase)
      const breatheB = 0.9 + 0.25 * Math.sin(now * 2.65 + stateB.phase)

      const dx = stateA.x - stateB.x
      const dy = stateA.y - stateB.y
      const distance = Math.hypot(dx, dy)
      const mergeRatio = clamp((farDistance - distance) / (farDistance - nearDistance), 0, 1)

      const middleX = (stateA.x + stateB.x) / 2
      const middleY = (stateA.y + stateB.y) / 2
      const mergeBreathe = 1 + 0.28 * Math.sin(now * 2.9)

      gsap.set(glowA, {
        x: stateA.x,
        y: stateA.y,
        scale: breatheA + mergeRatio * 0.22,
        opacity: glowPreset.opacityABase + glowPreset.opacityAWave * Math.sin(now * 2.2 + 0.4) + mergeRatio * glowPreset.opacityAByMerge
      })

      gsap.set(glowB, {
        x: stateB.x,
        y: stateB.y,
        scale: breatheB + mergeRatio * 0.22,
        opacity: glowPreset.opacityBBase + glowPreset.opacityBWave * Math.sin(now * 2.45 + 1) + mergeRatio * glowPreset.opacityBByMerge
      })

      gsap.set(glowMerge, {
        x: middleX,
        y: middleY,
        scale: (0.62 + mergeRatio * 0.86) * mergeBreathe,
        opacity: glowPreset.opacityMergeBase + mergeRatio * glowPreset.opacityMergeByRatio
      })
    }

    gsap.ticker.add(glowTicker)
  }, heroRef)
})

onUnmounted(() => {
  if (glowTicker) {
    gsap.ticker.remove(glowTicker)
    glowTicker = null
  }
  heroContext?.revert()
  heroContext = null
})
</script>

<template>
  <main class="relative min-h-screen overflow-hidden bg-[var(--bg-dark)]" :style="heroGlowVars">
    <div class="hero-grid absolute inset-0 pointer-events-none"></div>
    <div ref="glowARef" class="hero-glow hero-glow--a absolute left-0 top-0"></div>
    <div ref="glowBRef" class="hero-glow hero-glow--b absolute left-0 top-0"></div>
    <div ref="glowMergeRef" class="hero-glow hero-glow--merge absolute left-0 top-0"></div>

    <section
      ref="heroRef"
      class="relative z-10 mx-auto flex min-h-screen w-full max-w-5xl items-center justify-center px-6 py-16"
    >
      <div class="hero-panel w-full max-w-3xl rounded-2xl border border-white/10 bg-glass p-8 text-center md:p-12">
        <h1 class="mb-5 text-4xl font-bold leading-tight text-neon md:text-6xl">
          <span class="hero-title-line block">{{ t('common.appName') }}</span>
          <span class="hero-title-line block text-2xl tracking-[0.14em] text-white/90 md:text-3xl">
            {{ t('home.heroTitleSuffix') }}
          </span>
        </h1>

        <p class="hero-subtitle mx-auto max-w-xl text-base text-[var(--text-secondary)] md:text-lg">
          {{ t('home.subtitle') }}
        </p>

        <div class="mt-9 flex flex-col items-center justify-center gap-3 sm:flex-row">
          <button
            v-if="authStore.isLoggedIn"
            id="btn-start"
            class="hero-cta-btn w-full rounded-lg border border-neon bg-[var(--neon)]/10 px-8 py-3 text-base font-semibold text-neon shadow-neon transition-all duration-300 hover:-translate-y-0.5 hover:bg-[var(--neon-glow)] sm:w-auto"
            @click="handleStart"
          >
            {{ t('home.start') }}
          </button>
          <button
            v-else
            id="btn-login"
            class="hero-cta-btn w-full rounded-lg border border-white/20 px-8 py-3 text-base font-medium text-white/85 transition-all duration-300 hover:-translate-y-0.5 hover:border-neon hover:text-neon sm:w-auto"
            @click="handleToLogin"
          >
            {{ t('home.toLogin') }}
          </button>
        </div>

        <div class="mt-8 flex flex-wrap items-center justify-center gap-3 text-xs text-white/70 md:text-sm">
          <span class="hero-metric rounded-full border border-white/10 px-3 py-1">Focus</span>
          <span class="hero-metric rounded-full border border-white/10 px-3 py-1">Plan</span>
          <span class="hero-metric rounded-full border border-white/10 px-3 py-1">Execute</span>
        </div>
      </div>
    </section>
  </main>
</template>

<style lang="less" scoped>
.hero-grid {
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.026) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.026) 1px, transparent 1px),
    radial-gradient(circle at 24% 22%, rgba(0, 243, 255, 0.12), transparent 38%),
    radial-gradient(circle at 78% 70%, rgba(222, 0, 255, 0.1), transparent 40%);
  background-size: 52px 52px, 52px 52px, auto, auto;
  mask-image: radial-gradient(circle at center, black 42%, transparent 90%);
}

.hero-glow {
  pointer-events: none;
  width: var(--hero-glow-size);
  height: var(--hero-glow-size);
  border-radius: 999px;
  filter: blur(var(--hero-glow-blur));
  mix-blend-mode: screen;
}

.hero-glow--a {
  background: radial-gradient(circle, var(--hero-glow-a-core) 0%, var(--hero-glow-a-edge) 72%);
}

.hero-glow--b {
  background: radial-gradient(circle, var(--hero-glow-b-core) 0%, var(--hero-glow-b-edge) 72%);
}

.hero-glow--merge {
  width: var(--hero-glow-merge-size);
  height: var(--hero-glow-merge-size);
  background: radial-gradient(
    circle,
    var(--hero-glow-merge-core) 0%,
    var(--hero-glow-merge-mid) 58%,
    var(--hero-glow-merge-edge) 80%
  );
  filter: blur(var(--hero-glow-merge-blur));
}
</style>
