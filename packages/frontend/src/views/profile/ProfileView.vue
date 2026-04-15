<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import {
  ArrowLeft,
  Eye,
  KeyRound,
  LoaderCircle,
  LogOut,
  Palette,
  ShieldCheck,
  TriangleAlert,
  UserRound,
} from "lucide-vue-next";

import { userApi } from "@/api/user";
import { useAuthStore } from "@/stores/use-auth-store";
import { useThemeStore } from "@/stores/use-theme-store";

const router = useRouter();
const authStore = useAuthStore();
const themeStore = useThemeStore();
const { t } = useI18n();

const profileForm = reactive({
  nickname: authStore.user?.nickname || "",
  language: themeStore.language,
  theme: themeStore.theme,
});
const profileNotice = ref("");
const isSavingProfile = ref(false);

const passwordForm = reactive({
  oldPassword: "",
  newPassword: "",
  confirmPassword: "",
});
const isSubmittingPassword = ref(false);
const passwordErrorMessage = ref("");
const showConfirmDialog = ref(false);
const showOldPassword = ref(false);
const showNewPassword = ref(false);
const showConfirmPassword = ref(false);

const canSubmitPassword = computed(() => {
  return Boolean(
    passwordForm.oldPassword &&
      passwordForm.newPassword &&
      passwordForm.confirmPassword &&
      !isSubmittingPassword.value,
  );
});

const themeOptions = computed(() => [
  { value: "cyan", label: t("theme.cyan"), swatch: "bg-[#00f3ff]" },
  { value: "purple", label: t("theme.purple"), swatch: "bg-[#bc13fe]" },
  { value: "green", label: t("theme.green"), swatch: "bg-[#39ff14]" },
  { value: "pink", label: t("theme.pink"), swatch: "bg-[#ff8a00]" },
] as const);

const languageOptions = computed(() => [
  { value: "zh", label: t("lang.zh") },
  { value: "en", label: t("lang.en") },
] as const);

function validatePasswordForm() {
  if (
    !passwordForm.oldPassword ||
    !passwordForm.newPassword ||
    !passwordForm.confirmPassword
  ) {
    return t("profile.validationRequired");
  }
  if (passwordForm.newPassword.length < 8) {
    return t("profile.validationLength");
  }
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    return t("profile.validationMismatch");
  }
  if (passwordForm.oldPassword === passwordForm.newPassword) {
    return t("profile.validationSame");
  }
  return "";
}

function getErrorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return t("profile.changeFailed");
}

function resetPasswordForm() {
  passwordForm.oldPassword = "";
  passwordForm.newPassword = "";
  passwordForm.confirmPassword = "";
  showOldPassword.value = false;
  showNewPassword.value = false;
  showConfirmPassword.value = false;
}

function handleBack() {
  router.push("/tasks");
}

async function handleSaveProfile() {
  profileNotice.value = "";
  isSavingProfile.value = true;

  try {
    themeStore.setLanguage(profileForm.language);
    themeStore.setTheme(profileForm.theme);

    if (authStore.user) {
      authStore.setUser({
        ...authStore.user,
        nickname: profileForm.nickname,
        language: profileForm.language,
        theme: profileForm.theme,
      });
    }

    try {
      const updated = await userApi.updateMe({
        nickname: profileForm.nickname,
        language: profileForm.language,
        theme: profileForm.theme,
      });
      authStore.setUser(updated);
    } catch {
      // Some backend envs may not expose profile update endpoint yet.
    }

    profileNotice.value = t("profile.saveSuccess");
  } catch {
    profileNotice.value = t("profile.saveFailed");
  } finally {
    isSavingProfile.value = false;
  }
}

async function handleChangePassword() {
  passwordErrorMessage.value = "";
  const validationError = validatePasswordForm();
  if (validationError) {
    passwordErrorMessage.value = validationError;
    return;
  }
  showConfirmDialog.value = true;
}

function handleCancelConfirm() {
  if (isSubmittingPassword.value) return;
  showConfirmDialog.value = false;
}

async function handleConfirmChangePassword() {
  passwordErrorMessage.value = "";
  showConfirmDialog.value = false;
  isSubmittingPassword.value = true;

  try {
    await userApi.changePassword({
      old_password: passwordForm.oldPassword,
      new_password: passwordForm.newPassword,
    });
    resetPasswordForm();
    await authStore.logout();
    await router.push("/login");
  } catch (error) {
    passwordErrorMessage.value = getErrorMessage(error);
  } finally {
    isSubmittingPassword.value = false;
  }
}
</script>

<template>
  <main class="min-h-screen bg-[#050a0f] text-[var(--text-primary)]">
    <header class="border-b border-white/10 bg-black/20 backdrop-blur-xl">
      <div class="mx-auto flex max-w-5xl items-center justify-between px-4 md:px-6 py-5">
        <button
          class="inline-flex h-10 w-10 items-center justify-center rounded-md border border-white/10 text-white/65 transition-colors hover:border-neon hover:text-neon"
          type="button"
          :aria-label="t('tasks.backToList')"
          @click="handleBack"
        >
          <ArrowLeft class="h-5 w-5" />
        </button>
        <h1 class="text-xl md:text-2xl font-semibold text-white">
          {{ t("profile.title") }}
        </h1>
      </div>
    </header>

    <section class="mx-auto max-w-5xl px-4 md:px-6 py-6 md:py-8 grid gap-6 md:grid-cols-2">
      <section
        class="rounded-2xl border border-[rgba(0,243,255,0.15)] bg-[rgba(5,10,15,0.4)] p-6 shadow-[0_0_30px_rgba(0,0,0,0.25)] backdrop-blur-xl"
      >
        <div class="flex items-center gap-3">
          <div
            class="flex h-10 w-10 items-center justify-center rounded-md border border-neon/70 bg-neon/10 text-neon"
          >
            <UserRound class="h-5 w-5" />
          </div>
          <h2 class="text-lg font-semibold text-white">
            {{ t("profile.accountInfo") }}
          </h2>
        </div>

        <div class="mt-6 space-y-6">
          <div>
            <label class="mb-2 block text-sm font-medium text-[var(--text-secondary)]">
              {{ t("profile.email") }}
            </label>
            <input
              :value="authStore.user?.email || ''"
              disabled
              class="w-full rounded-lg border border-white/10 bg-white/5 px-4 py-3 text-white/70"
            >
          </div>

          <div>
            <label class="mb-2 block text-sm font-medium text-[var(--text-secondary)]">
              {{ t("profile.nickname") }}
            </label>
            <input
              v-model="profileForm.nickname"
              :placeholder="t('profile.nicknamePlaceholder')"
              class="w-full rounded-lg border border-white/10 bg-white/5 px-4 py-3 text-white placeholder-white/35 focus:border-neon focus:outline-none"
            >
          </div>

          <div>
            <label class="mb-2 block text-sm font-medium text-[var(--text-secondary)]">
              {{ t("profile.language") }}
            </label>
            <select
              v-model="profileForm.language"
              class="w-full rounded-lg border border-white/10 bg-white/5 px-4 py-3 text-white focus:border-neon focus:outline-none"
            >
              <option
                v-for="option in languageOptions"
                :key="option.value"
                :value="option.value"
                class="bg-[#0b1219]"
              >
                {{ option.label }}
              </option>
            </select>
          </div>

          <div>
            <label class="mb-3 block text-sm font-medium text-[var(--text-secondary)]">
              {{ t("profile.theme") }}
            </label>
            <div class="grid grid-cols-2 gap-3">
              <button
                v-for="option in themeOptions"
                :key="option.value"
                type="button"
                class="h-11 rounded-lg border px-3 text-sm transition-colors flex items-center justify-center gap-2"
                :class="
                  profileForm.theme === option.value
                    ? 'border-neon bg-neon/10 text-neon'
                    : 'border-white/10 bg-white/5 text-white/70 hover:border-white/25'
                "
                @click="profileForm.theme = option.value"
              >
                <Palette class="h-4 w-4" />
                <span>{{ option.label }}</span>
                <span :class="['h-2.5 w-2.5 rounded-full', option.swatch]" />
              </button>
            </div>
          </div>

          <p
            v-if="profileNotice"
            class="rounded-md border border-emerald-400/20 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-200"
          >
            {{ profileNotice }}
          </p>

          <button
            class="flex h-12 w-full items-center justify-center gap-2 rounded-lg bg-[var(--neon)] px-4 font-bold text-black transition-all hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-55"
            type="button"
            :disabled="isSavingProfile"
            @click="handleSaveProfile"
          >
            <LoaderCircle
              v-if="isSavingProfile"
              class="h-4 w-4 animate-spin"
            />
            <ShieldCheck
              v-else
              class="h-4 w-4"
            />
            {{ t("profile.saveProfile") }}
          </button>
        </div>
      </section>

      <section
        class="rounded-2xl border border-[rgba(0,243,255,0.15)] bg-[rgba(5,10,15,0.4)] p-6 shadow-[0_0_30px_rgba(0,0,0,0.25)] backdrop-blur-xl"
      >
        <div class="flex items-center gap-3">
          <div
            class="flex h-10 w-10 items-center justify-center rounded-md border border-neon/70 bg-neon/10 text-neon"
          >
            <KeyRound class="h-5 w-5" />
          </div>
          <div>
            <h2 class="text-lg font-semibold text-white">
              {{ t("profile.security") }}
            </h2>
            <p class="mt-1 text-sm text-white/45">
              {{ t("profile.securityDesc") }}
            </p>
          </div>
        </div>

        <form
          class="mt-8 space-y-7"
          @submit.prevent="handleChangePassword"
        >
          <div>
            <label class="mb-2 block text-sm font-medium text-[var(--text-secondary)]">
              {{ t("profile.oldPassword") }}
            </label>
            <div class="relative">
              <input
                v-model.trim="passwordForm.oldPassword"
                :type="showOldPassword ? 'text' : 'password'"
                :placeholder="t('profile.oldPasswordPlaceholder')"
                class="w-full rounded-lg border border-[rgba(255,255,255,0.1)] bg-[rgba(255,255,255,0.05)] px-4 py-3 pr-10 text-white placeholder-gray-500 transition-colors focus:border-[var(--neon)] focus:outline-none"
                autocomplete="current-password"
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

          <div>
            <label class="mb-2 block text-sm font-medium text-[var(--text-secondary)]">
              {{ t("profile.newPassword") }}
            </label>
            <div class="relative">
              <input
                v-model.trim="passwordForm.newPassword"
                :type="showNewPassword ? 'text' : 'password'"
                :placeholder="t('profile.newPasswordPlaceholder')"
                class="w-full rounded-lg border border-[rgba(255,255,255,0.1)] bg-[rgba(255,255,255,0.05)] px-4 py-3 pr-10 text-white placeholder-gray-500 transition-colors focus:border-[var(--neon)] focus:outline-none"
                autocomplete="new-password"
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

          <div>
            <label class="mb-2 block text-sm font-medium text-[var(--text-secondary)]">
              {{ t("profile.confirmPassword") }}
            </label>
            <div class="relative">
              <input
                v-model.trim="passwordForm.confirmPassword"
                :type="showConfirmPassword ? 'text' : 'password'"
                :placeholder="t('profile.confirmPasswordPlaceholder')"
                class="w-full rounded-lg border border-[rgba(255,255,255,0.1)] bg-[rgba(255,255,255,0.05)] px-4 py-3 pr-10 text-white placeholder-gray-500 transition-colors focus:border-[var(--neon)] focus:outline-none"
                autocomplete="new-password"
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
            v-if="passwordErrorMessage"
            class="rounded-md border border-rose-400/20 bg-rose-500/10 px-4 py-3 text-sm text-rose-200"
          >
            {{ passwordErrorMessage }}
          </p>

          <div class="grid grid-cols-2 gap-4 pt-2">
            <button
              class="flex h-12 w-full items-center justify-center gap-2 rounded-lg bg-[var(--neon)] px-4 text-center font-bold leading-none text-black transition-all hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-50"
              type="submit"
              :disabled="!canSubmitPassword"
            >
              <LoaderCircle
                v-if="isSubmittingPassword"
                class="h-4 w-4 animate-spin"
              />
              <ShieldCheck
                v-else
                class="h-4 w-4"
              />
              {{ t("profile.savePassword") }}
            </button>
            <button
              class="flex h-12 w-full items-center justify-center rounded-lg border border-[rgba(255,255,255,0.1)] px-4 text-center font-medium leading-none text-[var(--text-secondary)] transition-colors hover:border-[var(--neon)] hover:text-neon"
              type="button"
              @click="resetPasswordForm"
            >
              {{ t("common.clear") }}
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
              {{ t("profile.confirmDialogTitle") }}
            </h3>
            <p class="mt-2 text-sm leading-6 text-white/65">
              {{ t("profile.confirmDialogDesc") }}
            </p>
          </div>
        </div>

        <div class="mt-6 grid grid-cols-2 gap-3">
          <button
            class="flex h-12 w-full items-center justify-center rounded-lg border border-white/15 px-4 text-sm font-medium text-white/75 transition-colors hover:border-white/35 hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
            type="button"
            :disabled="isSubmittingPassword"
            @click="handleCancelConfirm"
          >
            {{ t("common.cancel") }}
          </button>
          <button
            class="flex h-12 w-full items-center justify-center gap-2 rounded-lg bg-[var(--neon)] px-4 text-sm font-bold text-black transition-all hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
            type="button"
            :disabled="isSubmittingPassword"
            @click="handleConfirmChangePassword"
          >
            <LoaderCircle
              v-if="isSubmittingPassword"
              class="h-4 w-4 animate-spin"
            />
            <LogOut
              v-else
              class="h-4 w-4"
            />
            {{ t("profile.confirmChange") }}
          </button>
        </div>
      </div>
    </div>
  </main>
</template>
