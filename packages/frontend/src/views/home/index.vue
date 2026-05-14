<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Globe, ListTodo, Orbit, PenTool } from 'lucide-vue-next'
import gsap from 'gsap'

import { useAuthStore } from '@/stores/use-auth-store'
import { useThemeStore } from '@/stores/use-theme-store'

const router = useRouter()
const authStore = useAuthStore()
const themeStore = useThemeStore()
const { t } = useI18n()

const heroRef = ref<HTMLElement | null>(null)
const glowRedRef = ref<HTMLElement | null>(null)
const glowGreenRef = ref<HTMLElement | null>(null)
const glowBlueRef = ref<HTMLElement | null>(null)
const glowCoreRef = ref<HTMLElement | null>(null)

let heroContext: gsap.Context | null = null
let glowTicker: ((time: number, deltaTime: number) => void) | null = null

const nextLanguageLabel = computed(() => (themeStore.language === 'zh' ? t('lang.en') : t('lang.zh')))

function toggleLanguage() {
  const nextLanguage = themeStore.language === 'zh' ? 'en' : 'zh'
  themeStore.setLanguage(nextLanguage)
}

function goToTasks() {
  if (authStore.isLoggedIn) {
    router.push({ name: 'tasks' })
    return
  }
  router.push({ name: 'login' })
}

function goToPointCloud() {
  router.push({ name: 'point-cloud' })
}

function goToImageAnnotator() {
  router.push({ name: 'image-annotator' })
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
      .from('.hero-glow', { opacity: 0, scale: 0.58, stagger: 0.1, duration: 0.95 })
      .from('.lang-toggle', { y: -14, opacity: 0, duration: 0.65 }, '-=0.75')
      .from('.hero-title-line', { y: 42, opacity: 0, duration: 1.0 }, '-=0.55')
      .from('.hero-subtitle', { y: 20, opacity: 0, duration: 0.8 }, '-=0.62')
      .from('.hero-action-btn', { y: 16, opacity: 0, stagger: 0.08, duration: 0.7 }, '-=0.45')

    const redGlow = glowRedRef.value
    const greenGlow = glowGreenRef.value
    const blueGlow = glowBlueRef.value
    const coreGlow = glowCoreRef.value
    if (!redGlow || !greenGlow || !blueGlow || !coreGlow) return

    const states = [
      { x: 0, y: 0, vx: 0.12, vy: 0.09, phase: 0.1 },
      { x: 0, y: 0, vx: -0.11, vy: 0.08, phase: 1.6 },
      { x: 0, y: 0, vx: 0.09, vy: -0.1, phase: 2.4 }
    ]

    const margin = 100
    const speedFactor = 9
    const pairNear = 220
    const pairFar = 560

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
    states[0].x = initialBounds.width * 0.25
    states[0].y = initialBounds.height * 0.34
    states[1].x = initialBounds.width * 0.74
    states[1].y = initialBounds.height * 0.32
    states[2].x = initialBounds.width * 0.52
    states[2].y = initialBounds.height * 0.72

    gsap.set([redGlow, greenGlow, blueGlow, coreGlow], { xPercent: -50, yPercent: -50 })

    glowTicker = (_time, deltaTime) => {
      const step = Math.min(deltaTime, 3)
      const bounds = getBounds()

      for (const state of states) {
        state.x += state.vx * step * speedFactor
        state.y += state.vy * step * speedFactor

        if (state.x < margin || state.x > bounds.width - margin) state.vx *= -1
        if (state.y < margin || state.y > bounds.height - margin) state.vy *= -1

        state.x = clamp(state.x, margin, bounds.width - margin)
        state.y = clamp(state.y, margin, bounds.height - margin)
      }

      const now = performance.now() / 1000
      const [red, green, blue] = states

      const setGlow = (
        el: HTMLElement,
        state: { x: number; y: number; phase: number },
        baseOpacity: number,
        waveOpacity: number,
        baseScale: number,
        waveScale: number,
        pairBoost: number
      ) => {
        const breathe = Math.sin(now * 2.1 + state.phase)
        gsap.set(el, {
          x: state.x,
          y: state.y,
          scale: baseScale + waveScale * breathe + pairBoost,
          opacity: baseOpacity + waveOpacity * Math.sin(now * 2.45 + state.phase)
        })
      }

      const distRG = Math.hypot(red.x - green.x, red.y - green.y)
      const distRB = Math.hypot(red.x - blue.x, red.y - blue.y)
      const distGB = Math.hypot(green.x - blue.x, green.y - blue.y)

      const pairMix = (distance: number) => clamp((pairFar - distance) / (pairFar - pairNear), 0, 1)
      const rgMix = pairMix(distRG)
      const rbMix = pairMix(distRB)
      const gbMix = pairMix(distGB)

      setGlow(redGlow, red, 0.34, 0.12, 0.92, 0.18, (rgMix + rbMix) * 0.14)
      setGlow(greenGlow, green, 0.34, 0.12, 0.9, 0.18, (rgMix + gbMix) * 0.14)
      setGlow(blueGlow, blue, 0.36, 0.12, 0.92, 0.2, (rbMix + gbMix) * 0.14)

      const centroidX = (red.x + green.x + blue.x) / 3
      const centroidY = (red.y + green.y + blue.y) / 3
      const spread =
        (Math.hypot(red.x - centroidX, red.y - centroidY) +
          Math.hypot(green.x - centroidX, green.y - centroidY) +
          Math.hypot(blue.x - centroidX, blue.y - centroidY)) /
        3
      const triadMix = clamp((260 - spread) / 160, 0, 1)
      const coreStrength = clamp(triadMix * 0.72 + (rgMix + rbMix + gbMix) / 3 * 0.3, 0, 0.85)

      gsap.set(coreGlow, {
        x: centroidX,
        y: centroidY,
        scale: 0.64 + coreStrength * 0.88,
        opacity: 0.08 + coreStrength * 0.48
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
  <main class="relative min-h-screen overflow-hidden bg-[var(--bg-dark)]">
    <button
      class="lang-toggle absolute right-4 top-4 z-30 inline-flex items-center gap-2 rounded-full border border-white/15 bg-[#0b1219]/80 px-3 py-2 text-xs font-medium text-white/85 backdrop-blur-md transition-all duration-300 hover:-translate-y-0.5 hover:border-neon hover:text-neon md:right-7 md:top-7"
      :aria-label="t('home.languageToggleAria')"
      @click="toggleLanguage"
    >
      <Globe class="h-3.5 w-3.5" />
      <span>{{ nextLanguageLabel }}</span>
    </button>

    <div
      ref="glowRedRef"
      class="hero-glow hero-glow--red absolute left-0 top-0"
    />
    <div
      ref="glowGreenRef"
      class="hero-glow hero-glow--green absolute left-0 top-0"
    />
    <div
      ref="glowBlueRef"
      class="hero-glow hero-glow--blue absolute left-0 top-0"
    />
    <div
      ref="glowCoreRef"
      class="hero-glow hero-glow--core absolute left-0 top-0"
    />

    <section
      ref="heroRef"
      class="relative z-10 mx-auto flex min-h-screen w-full max-w-5xl items-center justify-center px-6 py-16"
    >
      <div class="hero-panel flex w-full max-w-2xl min-h-[20.5rem] flex-col justify-center rounded-2xl border border-[#0f1a27]/65 bg-[#071019]/28 px-7 py-9 text-center backdrop-blur-lg md:min-h-[24.5rem] md:px-10 md:py-11">
        <h1 class="mb-5 text-4xl font-bold leading-tight text-neon md:text-6xl">
          <span class="hero-title-line block">{{ t('home.title') }}</span>
        </h1>

        <p class="hero-subtitle mx-auto max-w-2xl text-base text-[var(--text-secondary)] md:text-lg">
          {{ t('home.subtitle') }}
        </p>

        <div class="mt-9 grid grid-cols-1 gap-3 sm:grid-cols-3">
          <button
            class="hero-action-btn inline-flex items-center justify-center gap-2 rounded-full border border-[#37b4ff]/55 bg-[#092336]/56 px-4 py-3 text-sm font-semibold text-[#8ad9ff] transition-colors duration-300 hover:bg-[#11344e]/72"
            @click="goToTasks"
          >
            <ListTodo class="h-4 w-4" />
            <span>{{ t('home.entryTasks') }}</span>
          </button>

          <button
            class="hero-action-btn inline-flex items-center justify-center gap-2 rounded-full border border-[#3fd6b8]/50 bg-[#0a2a27]/54 px-4 py-3 text-sm font-semibold text-[#97f3df] transition-colors duration-300 hover:bg-[#123b37]/72"
            @click="goToPointCloud"
          >
            <Orbit class="h-4 w-4" />
            <span>{{ t('home.entryPointCloud') }}</span>
          </button>

          <button
            class="hero-action-btn inline-flex items-center justify-center gap-2 rounded-full border border-[#8d7bff]/50 bg-[#191734]/56 px-4 py-3 text-sm font-semibold text-[#c6bbff] transition-colors duration-300 hover:bg-[#282253]/74"
            @click="goToImageAnnotator"
          >
            <PenTool class="h-4 w-4" />
            <span>{{ t('home.entryAnnotator') }}</span>
          </button>
        </div>
      </div>
    </section>
  </main>
</template>

<style lang="less" scoped>
.hero-glow {
  pointer-events: none;
  width: min(62vw, 820px);
  height: min(62vw, 820px);
  border-radius: 999px;
  filter: blur(56px) saturate(140%);
  mix-blend-mode: screen;
}

.hero-glow--red {
  background: radial-gradient(circle, rgba(255, 78, 140, 0.42) 0%, rgba(190, 44, 94, 0.04) 74%);
}

.hero-glow--green {
  background: radial-gradient(circle, rgba(92, 255, 170, 0.38) 0%, rgba(44, 176, 126, 0.035) 74%);
}

.hero-glow--blue {
  background: radial-gradient(circle, rgba(78, 176, 255, 0.44) 0%, rgba(36, 106, 200, 0.04) 74%);
}

.hero-glow--core {
  width: min(68vw, 920px);
  height: min(68vw, 920px);
  background: radial-gradient(
    circle,
    rgba(102, 124, 255, 0.56) 0%,
    rgba(58, 124, 236, 0.22) 45%,
    rgba(58, 66, 168, 0.03) 78%
  );
  filter: blur(52px) saturate(156%);
}

.hero-panel {
  box-shadow:
    0 28px 72px rgba(0, 0, 0, 0.5),
    inset 0 1px 0 rgba(14, 24, 36, 0.26);
}
</style>
