import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

interface UserInfo {
  id: string
  email: string
  nickname: string
  language: string
  theme: string
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string>(localStorage.getItem('access_token') || '')
  const refreshToken = ref<string>(localStorage.getItem('refresh_token') || '')
  const user = ref<UserInfo | null>(JSON.parse(localStorage.getItem('user') || 'null'))

  const isLoggedIn = computed(() => !!accessToken.value)

  function setTokens(access: string, refresh: string) {
    accessToken.value = access
    refreshToken.value = refresh
    localStorage.setItem('access_token', access)
    localStorage.setItem('refresh_token', refresh)
  }

  function setUser(userInfo: UserInfo) {
    user.value = userInfo
    localStorage.setItem('user', JSON.stringify(userInfo))
  }

  function logout() {
    accessToken.value = ''
    refreshToken.value = ''
    user.value = null
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('user')
  }

  return { accessToken, refreshToken, user, isLoggedIn, setTokens, setUser, logout }
})
