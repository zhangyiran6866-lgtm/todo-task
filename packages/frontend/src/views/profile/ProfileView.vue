<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import {
  ArrowLeft,
  Eye,
  KeyRound,
  LoaderCircle,
  LogOut,
  ShieldCheck,
  TriangleAlert,
} from "lucide-vue-next";

import { userApi } from "@/api/user";
import { useAuthStore } from "@/stores/use-auth-store";

const router = useRouter();
const authStore = useAuthStore();

const form = reactive({
  oldPassword: "",
  newPassword: "",
  confirmPassword: "",
});
const isSubmitting = ref(false);
const errorMessage = ref("");
const showConfirmDialog = ref(false);
const showOldPassword = ref(false);
const showNewPassword = ref(false);
const showConfirmPassword = ref(false);

const canSubmit = computed(() => {
  return Boolean(
    form.oldPassword &&
    form.newPassword &&
    form.confirmPassword &&
    !isSubmitting.value,
  );
});

function validateForm() {
  if (!form.oldPassword || !form.newPassword || !form.confirmPassword) {
    return "请完整填写旧密码、新密码和确认密码";
  }
  if (form.newPassword.length < 8) {
    return "新密码至少需要 8 位";
  }
  if (form.newPassword !== form.confirmPassword) {
    return "两次输入的新密码不一致";
  }
  if (form.oldPassword === form.newPassword) {
    return "新密码不能与旧密码相同";
  }
  return "";
}

function getErrorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return "修改密码失败，请稍后重试";
}

function resetForm() {
  form.oldPassword = "";
  form.newPassword = "";
  form.confirmPassword = "";
  showOldPassword.value = false;
  showNewPassword.value = false;
  showConfirmPassword.value = false;
}

function handleBack() {
  router.push("/tasks");
}

async function handleChangePassword() {
  errorMessage.value = "";

  const validationError = validateForm();
  if (validationError) {
    errorMessage.value = validationError;
    return;
  }

  showConfirmDialog.value = true;
}

function handleCancelConfirm() {
  if (isSubmitting.value) return;
  showConfirmDialog.value = false;
}

async function handleConfirmChangePassword() {
  errorMessage.value = "";
  showConfirmDialog.value = false;
  isSubmitting.value = true;

  try {
    await userApi.changePassword({
      old_password: form.oldPassword,
      new_password: form.newPassword,
    });
    resetForm();
    await authStore.logout();
    await router.push("/login");
  } catch (error) {
    errorMessage.value = getErrorMessage(error);
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <main class="min-h-screen bg-[#050a0f] text-[var(--text-primary)]">
    <header class="border-b border-white/10 bg-black/20 backdrop-blur-xl">
      <div
        class="mx-auto flex max-w-5xl items-center justify-between px-6 py-5"
      >
        <button
          class="inline-flex h-10 w-10 items-center justify-center rounded-md border border-white/10 text-white/65 transition-colors hover:border-neon hover:text-neon"
          type="button"
          aria-label="返回任务列表"
          @click="handleBack"
        >
          <ArrowLeft class="h-10 w-10" />
        </button>

        <div class="text-right">
          <h1 class="text-2xl font-semibold text-white">
            修改密码
          </h1>
        </div>
      </div>
    </header>

    <section class="mx-auto max-w-xl px-6 py-8">
      <section
        class="rounded-2xl border border-[rgba(0,243,255,0.15)] bg-[rgba(5,10,15,0.4)] p-8 shadow-[0_0_30px_rgba(0,0,0,0.25)] backdrop-blur-xl"
      >
        <div class="flex items-center gap-3">
          <div
            class="flex h-10 w-10 items-center justify-center rounded-md border border-neon/70 bg-neon/10 text-neon"
          >
            <KeyRound class="h-5 w-5" />
          </div>
          <div>
            <h2 class="text-xl font-semibold text-white">
              修改密码
            </h2>
            <p class="mt-1 text-sm text-white/45">
              验证旧密码后更新账户凭据
            </p>
          </div>
        </div>

        <form
          class="mt-10 space-y-10"
          @submit.prevent="handleChangePassword"
        >
          <div class="mb-2 mt-4">
            <label
              class="mb-2 block text-sm font-medium text-[var(--text-secondary)]"
            >旧密码</label>
            <div class="relative">
              <input
                v-model.trim="form.oldPassword"
                :type="showOldPassword ? 'text' : 'password'"
                class="mb-1 w-full rounded-lg border border-[rgba(255,255,255,0.1)] bg-[rgba(255,255,255,0.05)] px-4 py-3 pr-10 text-white placeholder-gray-500 transition-colors focus:border-[var(--neon)] focus:outline-none"
                autocomplete="current-password"
                placeholder="请输入当前密码"
              >
              <button
                type="button"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 transition-colors hover:text-[var(--neon)] focus:outline-none"
                @mousedown="showOldPassword = true"
                @mouseup="showOldPassword = false"
                @mouseleave="showOldPassword = false"
                @touchstart.prevent="showOldPassword = true"
                @touchend.prevent="showOldPassword = false"
              >
                <Eye class="h-5 w-5" />
              </button>
            </div>
          </div>

          <div class="mb-2">
            <label
              class="mb-2 block text-sm font-medium text-[var(--text-secondary)]"
            >新密码</label>
            <div class="relative">
              <input
                v-model.trim="form.newPassword"
                :type="showNewPassword ? 'text' : 'password'"
                class="mb-1 w-full rounded-lg border border-[rgba(255,255,255,0.1)] bg-[rgba(255,255,255,0.05)] px-4 py-3 pr-10 text-white placeholder-gray-500 transition-colors focus:border-[var(--neon)] focus:outline-none"
                autocomplete="new-password"
                placeholder="请输入至少 8 位新密码"
              >
              <button
                type="button"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 transition-colors hover:text-[var(--neon)] focus:outline-none"
                @mousedown="showNewPassword = true"
                @mouseup="showNewPassword = false"
                @mouseleave="showNewPassword = false"
                @touchstart.prevent="showNewPassword = true"
                @touchend.prevent="showNewPassword = false"
              >
                <Eye class="h-5 w-5" />
              </button>
            </div>
          </div>

          <div class="mb-2">
            <label
              class="mb-2 block text-sm font-medium text-[var(--text-secondary)]"
            >确认新密码</label>
            <div class="relative">
              <input
                v-model.trim="form.confirmPassword"
                :type="showConfirmPassword ? 'text' : 'password'"
                class="mb-1 w-full rounded-lg border border-[rgba(255,255,255,0.1)] bg-[rgba(255,255,255,0.05)] px-4 py-3 pr-10 text-white placeholder-gray-500 transition-colors focus:border-[var(--neon)] focus:outline-none"
                autocomplete="new-password"
                placeholder="请再次输入新密码"
              >
              <button
                type="button"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 transition-colors hover:text-[var(--neon)] focus:outline-none"
                @mousedown="showConfirmPassword = true"
                @mouseup="showConfirmPassword = false"
                @mouseleave="showConfirmPassword = false"
                @touchstart.prevent="showConfirmPassword = true"
                @touchend.prevent="showConfirmPassword = false"
              >
                <Eye class="h-5 w-5" />
              </button>
            </div>
          </div>

          <p
            v-if="errorMessage"
            class="rounded-md border border-rose-400/20 bg-rose-500/10 px-4 py-3 text-sm text-rose-200"
          >
            {{ errorMessage }}
          </p>
          <div class="grid grid-cols-2 gap-4 pt-4">
            <button
              class="flex h-14 w-full items-center justify-center gap-2 rounded-lg bg-[var(--neon)] px-4 text-center font-bold leading-none text-black transition-all hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-50"
              type="submit"
              :disabled="!canSubmit"
            >
              <LoaderCircle
                v-if="isSubmitting"
                class="h-4 w-4 animate-spin"
              />
              <ShieldCheck
                v-else
                class="h-4 w-4"
              />
              保存新密码
            </button>
            <button
              class="flex h-14 w-full items-center justify-center rounded-lg border border-[rgba(255,255,255,0.1)] px-4 text-center font-medium leading-none text-[var(--text-secondary)] transition-colors hover:border-[var(--neon)] hover:text-neon"
              type="button"
              @click="resetForm"
            >
              清空
            </button>
          </div>
        </form>
      </section>
    </section>

    <div
      v-if="showConfirmDialog"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/65 px-5 backdrop-blur-sm"
    >
      <div
        class="w-full max-w-md rounded-2xl border border-[rgba(0,243,255,0.2)] bg-[rgba(5,10,15,0.92)] p-6 shadow-[0_0_30px_rgba(0,243,255,0.15)]"
      >
        <div class="flex items-start gap-3">
          <div
            class="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-amber-300/40 bg-amber-400/10 text-amber-300"
          >
            <TriangleAlert class="h-5 w-5" />
          </div>
          <div class="min-w-0">
            <h3 class="text-lg font-semibold text-white">
              确认修改密码
            </h3>
            <p class="mt-2 text-sm leading-6 text-white/65">
              修改密码后将退出当前账号，并跳转到登录页重新登录。确认继续吗？
            </p>
          </div>
        </div>

        <div class="mt-6 grid grid-cols-2 gap-3">
          <button
            class="flex h-12 w-full items-center justify-center rounded-lg border border-white/15 px-4 text-sm font-medium text-white/75 transition-colors hover:border-white/35 hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
            type="button"
            :disabled="isSubmitting"
            @click="handleCancelConfirm"
          >
            取消
          </button>
          <button
            class="flex h-12 w-full items-center justify-center gap-2 rounded-lg bg-[var(--neon)] px-4 text-sm font-bold text-black transition-all hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
            type="button"
            :disabled="isSubmitting"
            @click="handleConfirmChangePassword"
          >
            <LoaderCircle
              v-if="isSubmitting"
              class="h-4 w-4 animate-spin"
            />
            <LogOut
              v-else
              class="h-4 w-4"
            />
            确认修改
          </button>
        </div>
      </div>
    </div>
  </main>
</template>
