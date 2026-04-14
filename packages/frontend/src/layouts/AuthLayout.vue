<script setup lang="ts">
const particleOptions = {
  background: { color: { value: 'transparent' } },
  fpsLimit: 60,
  particles: {
    color: { value: '#00f3ff' },
    links: { color: '#00f3ff', distance: 100, enable: true, opacity: 0.2, width: 1 },
    move: { enable: true, speed: 0.8, direction: 'none', random: true, straight: false, outModes: 'out' },
    number: { density: { enable: true, area: 800 }, value: 120 },
    opacity: { value: 0.3, animation: { enable: true, speed: 1, minimumValue: 0.1 } },
    shape: { type: 'circle' },
    size: { value: { min: 0.5, max: 1.5 } }
  },
  detectRetina: true
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center relative overflow-hidden bg-[#02050a]">
    <!-- 动态星空散点背景，跨路由保持不变 -->
    <vue-particles
      id="tsparticles-auth"
      class="absolute inset-0 z-0"
      :options="particleOptions"
    />

    <!-- 子页面内容区（带平滑过渡） -->
    <router-view v-slot="{ Component }">
      <transition name="fade-transform" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </div>
</template>

<style scoped>
.fade-transform-enter-active,
.fade-transform-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.fade-transform-enter-from,
.fade-transform-leave-to {
  opacity: 0;
  transform: translateY(10px) scale(0.98);
}
</style>
