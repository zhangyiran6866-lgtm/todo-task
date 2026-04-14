import axios from 'axios'
import type { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from 'axios'
import { useAuthStore } from '@/stores/use-auth-store'
import router from '@/router'

// 应对刷新冲突
let isRefreshing = false
let requestsQueue: Array<(token: string) => void> = []

const request: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const authStore = useAuthStore()
    if (authStore.accessToken && config.headers) {
      config.headers.Authorization = `Bearer ${authStore.accessToken}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    // 我们的后端统一结构: { code: 0, message: "success", data: {} }
    const res = response.data
    if (res.code !== 0) {
      // 可以在这统一抛出业务错误提示
      return Promise.reject(new Error(res.message || 'Error'))
    }
    return res.data
  },
  async (error) => {
    const authStore = useAuthStore()
    const originalRequest = error.config

    // 401: Token 失效或未登录
    if (error.response && error.response.status === 401 && !originalRequest._retry) {
      // 防止如果是刷新 token 接口 401 时死循环
      if (originalRequest.url === '/auth/refresh') {
        authStore.logoutSync()
        router.push('/login')
        return Promise.reject(error)
      }

      // 开始无感刷新 Token
      if (isRefreshing) {
        return new Promise((resolve) => {
          requestsQueue.push((token: string) => {
            originalRequest.headers.Authorization = `Bearer ${token}`
            resolve(request(originalRequest))
          })
        })
      }

      originalRequest._retry = true
      isRefreshing = true

      try {
        if (!authStore.refreshToken) {
          throw new Error('No refresh token available')
        }
        
        // 调用刷新接口
        const res = await axios.post('/api/v1/auth/refresh', {
          refresh_token: authStore.refreshToken
        })
        
        if (res.data.code === 0) {
          const newToken = res.data.data.access_token
          const newRefresh = res.data.data.refresh_token
          authStore.setToken(newToken, newRefresh)
          
          // 重新执行队列中的请求
          requestsQueue.forEach((cb) => cb(newToken))
          requestsQueue = []
          
          // 执行原请求
          originalRequest.headers.Authorization = `Bearer ${newToken}`
          return request(originalRequest)
        } else {
          throw new Error('Refresh failed')
        }
      } catch (refreshError) {
        requestsQueue = []
        authStore.logoutSync()
        router.push('/login')
        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }

    // 这里可以处理别的 HTTP StatusCode, 例如 403, 404, 500
    // 对于 409 (Conflict 等业务返回错误), 也可以统一抛出
    const msg = error.response?.data?.message || error.message
    return Promise.reject(new Error(msg))
  }
)

export default request
