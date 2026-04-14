import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'

import App from './App.vue'
import router from './router'
import './style.css'

import Particles from "@tsparticles/vue3"
import { loadSlim } from "@tsparticles/slim"

const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('language') || 'zh',
  messages: {}
})

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(i18n)
app.use(Particles, {
  init: async (engine) => {
    await loadSlim(engine)
  }
})
app.mount('#app')
