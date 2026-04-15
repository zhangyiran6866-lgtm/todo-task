import { defineStore } from "pinia";
import { ref, computed } from "vue";
import request from "@/utils/request";
import { userApi, type UserInfo } from "@/api/user";

interface LoginPayload {
  email: string;
  password: string;
}

interface RegisterPayload extends LoginPayload {
  nickname: string;
}

interface TokenResponse {
  access_token: string;
  refresh_token: string;
}

export const useAuthStore = defineStore("auth", () => {
  const accessToken = ref<string>(localStorage.getItem("access_token") || "");
  const refreshToken = ref<string>(localStorage.getItem("refresh_token") || "");
  const user = ref<UserInfo | null>(
    JSON.parse(localStorage.getItem("user") || "null"),
  );

  const isLoggedIn = computed(() => !!accessToken.value);

  function setToken(access: string, refresh: string) {
    accessToken.value = access;
    refreshToken.value = refresh;
    localStorage.setItem("access_token", access);
    localStorage.setItem("refresh_token", refresh);
  }

  function setUser(userInfo: UserInfo) {
    user.value = userInfo;
    localStorage.setItem("user", JSON.stringify(userInfo));
  }

  function logoutSync() {
    accessToken.value = "";
    refreshToken.value = "";
    user.value = null;
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    localStorage.removeItem("user");
  }

  async function login(payload: LoginPayload) {
    const data = (await request.post(
      "/auth/login",
      payload,
    )) as unknown as TokenResponse;
    setToken(data.access_token, data.refresh_token);
    await fetchUser();
  }

  async function register(payload: RegisterPayload) {
    await request.post("/auth/register", payload);
    // 注册成功后不自动登录，返回交由外部跳转至登录页
  }

  async function fetchUser() {
    const data = await userApi.getMe();
    setUser(data);
  }

  async function logout() {
    try {
      if (refreshToken.value) {
        await request.post("/auth/logout", {
          refresh_token: refreshToken.value,
        });
      }
    } catch {
      // Ignore
    } finally {
      logoutSync();
    }
  }

  return {
    accessToken,
    refreshToken,
    user,
    isLoggedIn,
    setToken,
    setUser,
    login,
    register,
    fetchUser,
    logout,
    logoutSync,
  };
});
