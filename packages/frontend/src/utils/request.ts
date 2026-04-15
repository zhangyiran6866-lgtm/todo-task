import axios from "axios";
import type {
  AxiosInstance,
  InternalAxiosRequestConfig,
  AxiosResponse,
} from "axios";
import { useAuthStore } from "@/stores/use-auth-store";
import router from "@/router";

// 应对刷新冲突
let isRefreshing = false;
let requestsQueue: Array<(token: string) => void> = [];

function getErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const message = (error.response?.data as { message?: string } | undefined)
      ?.message;
    return message || error.message || "请求失败，请稍后重试";
  }
  if (error instanceof Error) {
    return error.message || "请求失败，请稍后重试";
  }
  return "请求失败，请稍后重试";
}

const request: AxiosInstance = axios.create({
  baseURL: "/api/v1",
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
});

// 请求拦截器
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const authStore = useAuthStore();
    if (authStore.accessToken && config.headers) {
      config.headers.Authorization = `Bearer ${authStore.accessToken}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  },
);

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    // 我们的后端统一结构: { code: 0, message: "成功", data: {} }
    const res = response.data;
    if (res.code !== 0) {
      return Promise.reject(new Error(res.message || "请求失败"));
    }
    return res.data;
  },
  async (error) => {
    const authStore = useAuthStore();
    const originalRequest = error.config as
      | (InternalAxiosRequestConfig & { _retry?: boolean })
      | undefined;
    const requestUrl = originalRequest?.url || "";
    const backendMessage = getErrorMessage(error);

    // 401: Token 失效或未登录
    if (
      error.response &&
      error.response.status === 401 &&
      originalRequest &&
      !originalRequest._retry
    ) {
      // 登录/注册失败时直接返回后端提示，不触发刷新 token
      if (requestUrl === "/auth/login" || requestUrl === "/auth/register") {
        return Promise.reject(new Error(backendMessage));
      }

      // 防止如果是刷新 token 接口 401 时死循环
      if (requestUrl === "/auth/refresh") {
        authStore.logoutSync();
        router.push("/login");
        return Promise.reject(new Error(backendMessage));
      }

      // 开始无感刷新 Token
      if (isRefreshing) {
        return new Promise((resolve) => {
          requestsQueue.push((token: string) => {
            if (originalRequest.headers) {
              originalRequest.headers.Authorization = `Bearer ${token}`;
            }
            resolve(request(originalRequest));
          });
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        if (!authStore.refreshToken) {
          throw new Error("登录状态已失效，请重新登录");
        }

        // 调用刷新接口
        const res = await axios.post("/api/v1/auth/refresh", {
          refresh_token: authStore.refreshToken,
        });

        if (res.data.code === 0) {
          const newToken = res.data.data.access_token;
          const newRefresh = res.data.data.refresh_token;
          authStore.setToken(newToken, newRefresh);

          // 重新执行队列中的请求
          requestsQueue.forEach((cb) => cb(newToken));
          requestsQueue = [];

          // 执行原请求
          originalRequest.headers.Authorization = `Bearer ${newToken}`;
          return request(originalRequest);
        } else {
          throw new Error("刷新登录状态失败，请重新登录");
        }
      } catch (refreshError) {
        requestsQueue = [];
        authStore.logoutSync();
        router.push("/login");
        return Promise.reject(new Error(getErrorMessage(refreshError)));
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(new Error(backendMessage));
  },
);

export default request;
