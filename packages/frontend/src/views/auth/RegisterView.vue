<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { useAuthStore } from "@/stores/use-auth-store";

const router = useRouter();
const authStore = useAuthStore();
const { t } = useI18n();

const email = ref("");
const password = ref("");
const nickname = ref("");
const errorMsg = ref("");
const successMsg = ref("");
const loading = ref(false);
const showPassword = ref(false);

function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message || t("auth.errRegisterFailed");
  }
  return t("auth.errRegisterFailed");
}

const handleRegister = async () => {
  errorMsg.value = "";
  successMsg.value = "";

  if (!email.value || !password.value || !nickname.value) {
    errorMsg.value = t("auth.errRegisterIncomplete");
    return;
  }

  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(email.value)) {
    errorMsg.value = t("auth.errEmailInvalid");
    return;
  }

  if (password.value.length < 8) {
    errorMsg.value = t("auth.errPasswordLength");
    return;
  }

  try {
    loading.value = true;
    await authStore.register({
      email: email.value,
      password: password.value,
      nickname: nickname.value,
    });

    // 模拟一段额外的 loading 效果以符合“一定的loading等待效果”的要求
    await new Promise((resolve) => setTimeout(resolve, 800));

    // 显示霓虹风格的成功提示而不是原生弹窗
    successMsg.value = t("auth.registerSuccess");

    // 等待一小段时间后跳转至登录页
    setTimeout(() => {
      router.push("/login");
    }, 1500);
  } catch (error: unknown) {
    errorMsg.value = getErrorMessage(error);
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <div
    class="w-full max-w-md p-8 rounded-2xl bg-[rgba(5,10,15,0.4)] backdrop-blur-xl border border-[rgba(0,243,255,0.15)] z-10 transition-all"
  >
    <router-link
      to="/"
      class="mb-6 inline-flex items-center gap-2 text-sm font-medium text-[var(--text-secondary)] transition-colors hover:text-neon"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        class="h-4 w-4"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M3 12l9-9 9 9M5 10v10h14V10"
        />
      </svg>
      {{ t("auth.backHome") }}
    </router-link>

    <h2 class="text-3xl font-bold text-center text-neon drop-shadow-neon mb-8">
      {{ t("auth.registerTitle") }}
    </h2>

    <form
      class="space-y-6"
      novalidate
      @submit.prevent="handleRegister"
    >
      <div>
        <label
          class="block text-sm font-medium text-[var(--text-secondary)] mb-2"
        >{{ t("auth.nickname") }}</label>
        <div class="relative">
          <input
            v-model="nickname"
            type="text"
            class="w-full px-4 py-3 bg-[rgba(255,255,255,0.05)] border border-[rgba(255,255,255,0.1)] rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-[var(--neon)] transition-colors pr-10"
            :placeholder="t('auth.nicknamePlaceholder')"
          >
          <button
            v-if="nickname"
            type="button"
            class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-white transition-colors focus:outline-none"
            @click="nickname = ''"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="h-5 w-5"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fill-rule="evenodd"
                d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                clip-rule="evenodd"
              />
            </svg>
          </button>
        </div>
      </div>

      <div>
        <label
          class="block text-sm font-medium text-[var(--text-secondary)] mb-2"
        >{{ t("auth.email") }}</label>
        <div class="relative">
          <input
            v-model="email"
            type="email"
            class="w-full px-4 py-3 bg-[rgba(255,255,255,0.05)] border border-[rgba(255,255,255,0.1)] rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-[var(--neon)] transition-colors pr-10"
            :placeholder="t('auth.registerEmailPlaceholder')"
          >
          <button
            v-if="email"
            type="button"
            class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-white transition-colors focus:outline-none"
            @click="email = ''"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="h-5 w-5"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fill-rule="evenodd"
                d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                clip-rule="evenodd"
              />
            </svg>
          </button>
        </div>
      </div>

      <div>
        <label
          class="block text-sm font-medium text-[var(--text-secondary)] mb-2"
        >{{ t("auth.password") }}</label>
        <div class="relative">
          <input
            v-model="password"
            :type="showPassword ? 'text' : 'password'"
            class="w-full px-4 py-3 bg-[rgba(255,255,255,0.05)] border border-[rgba(255,255,255,0.1)] rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-[var(--neon)] transition-colors pr-10"
            :placeholder="t('auth.registerPasswordPlaceholder')"
          >
          <button
            type="button"
            class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-[var(--neon)] transition-colors focus:outline-none"
            @mousedown="showPassword = true"
            @mouseup="showPassword = false"
            @mouseleave="showPassword = false"
            @touchstart.prevent="showPassword = true"
            @touchend.prevent="showPassword = false"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
              />
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
              />
            </svg>
          </button>
        </div>
      </div>

      <div
        v-if="errorMsg"
        class="text-[#ff4444] text-sm mt-2 p-3 bg-[rgba(255,68,68,0.1)] border border-[#ff4444] rounded flex items-center shadow-[0_0_10px_rgba(255,68,68,0.3)] animate-pulse"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="h-4 w-4 mr-2"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
        {{ errorMsg }}
      </div>

      <div
        v-if="successMsg"
        class="text-[#00ffcc] text-sm mt-2 p-3 bg-[rgba(0,255,204,0.1)] border border-[#00ffcc] rounded flex items-center shadow-[0_0_10px_rgba(0,255,204,0.3)] animate-pulse"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="h-4 w-4 mr-2"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        {{ successMsg }}
      </div>

      <button
        type="submit"
        :disabled="loading"
        class="w-full py-3 px-4 bg-[var(--neon)] text-black font-bold rounded-lg hover:brightness-110 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {{ loading ? t("auth.registering") : t("auth.register") }}
      </button>
    </form>

    <div class="mt-6 text-center text-sm text-[var(--text-secondary)]">
      {{ t("auth.hasAccount") }}
      <router-link
        to="/login"
        class="text-neon hover:underline"
      >
        {{ t("auth.toLogin") }}
      </router-link>
    </div>
  </div>
</template>
