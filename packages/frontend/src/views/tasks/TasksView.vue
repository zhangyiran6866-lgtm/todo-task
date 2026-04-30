<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { useInfiniteScroll } from "@vueuse/core";
import {
  AlertCircle,
  CheckCircle2,
  ChevronDown,
  Circle,
  Globe,
  LayoutGrid,
  ListTodo,
  Palette,
  Plus,
  Rows3,
} from "lucide-vue-next";

import type { Task } from "@/api/task";
import CreateTaskDrawer from "@/components/tasks/CreateTaskDrawer.vue";
import TaskCard from "@/components/tasks/TaskCard.vue";
import { useAuthStore } from "@/stores/use-auth-store";
import { useTaskStore } from "@/stores/use-task-store";
import { useThemeStore } from "@/stores/use-theme-store";

const router = useRouter();
const authStore = useAuthStore();
const taskStore = useTaskStore();
const themeStore = useThemeStore();
const { t } = useI18n();

const scrollContainer = ref<HTMLElement | null>(null);
const isDrawerOpen = ref(false);
const isMobile = ref(window.innerWidth < 768);
const viewMode = ref<"card" | "list">(isMobile.value ? "list" : "card");

const userMenuRef = ref<HTMLElement | null>(null);
const themeMenuRef = ref<HTMLElement | null>(null);
const languageMenuRef = ref<HTMLElement | null>(null);
const isUserMenuOpen = ref(false);
const isThemeMenuOpen = ref(false);
const isLanguageMenuOpen = ref(false);
const completingTaskId = ref<string | null>(null);

const statusOptions = computed(() => [
  { label: t("tasks.statusAll"), value: "", icon: ListTodo },
  { label: t("tasks.statusTodo"), value: "todo", icon: Circle },
  { label: t("tasks.statusExpired"), value: "expired", icon: AlertCircle },
  { label: t("tasks.statusDone"), value: "done", icon: CheckCircle2 },
]);

const priorityOptions = computed(() => [
  { label: t("tasks.priorityAll"), value: "", color: "bg-white/20" },
  { label: t("tasks.priorityCritical"), value: "critical", color: "bg-rose-500" },
  { label: t("tasks.priorityImportant"), value: "important", color: "bg-purple-500" },
  { label: t("tasks.priorityUrgent"), value: "urgent", color: "bg-amber-500" },
  { label: t("tasks.priorityLow"), value: "low", color: "bg-emerald-500" },
  { label: t("tasks.priorityRoutine"), value: "routine", color: "bg-blue-400" },
]);

const themeOptions = computed(() => [
  { value: "cyan", label: t("theme.cyan"), bgClass: "bg-[#00f3ff]" },
  { value: "purple", label: t("theme.purple"), bgClass: "bg-[#bc13fe]" },
  { value: "green", label: t("theme.green"), bgClass: "bg-[#39ff14]" },
  { value: "pink", label: t("theme.pink"), bgClass: "bg-[#ff8a00]" },
] as const);

const languageOptions = computed(() => [
  { value: "zh", label: t("lang.zh") },
  { value: "en", label: t("lang.en") },
] as const);

const activeThemeOption = computed(() => {
  return (
    themeOptions.value.find((option) => option.value === themeStore.theme) ||
    themeOptions.value[0]
  );
});

const activeLanguageOption = computed(() => {
  return (
    languageOptions.value.find((option) => option.value === themeStore.language) ||
    languageOptions.value[0]
  );
});

const displayUserName = computed(
  () => authStore.user?.nickname || authStore.user?.email || t("common.user"),
);
const userInitial = computed(() =>
  displayUserName.value.trim().charAt(0).toUpperCase(),
);

const sortedTasks = computed(() => {
  const getRank = (task: Task) => {
    const isDone = task.status === "done";
    const isExpired =
      !isDone && Boolean(task.due_at) && new Date(task.due_at as string).getTime() < Date.now();
    if (!isDone && !isExpired) return 0; // 未完成
    if (isExpired) return 1; // 已过期
    return 2; // 已完成
  };

  return [...taskStore.tasks].sort((a, b) => {
    const rankDiff = getRank(a) - getRank(b);
    if (rankDiff !== 0) return rankDiff;

    const aDue = a.due_at ? new Date(a.due_at).getTime() : Number.POSITIVE_INFINITY;
    const bDue = b.due_at ? new Date(b.due_at).getTime() : Number.POSITIVE_INFINITY;
    if (aDue !== bDue) return aDue - bDue;

    return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
  });
});

onMounted(async () => {
  window.addEventListener("resize", handleResize);
  if (!authStore.user && authStore.isLoggedIn) {
    try {
      await authStore.fetchUser();
    } catch {
      // Ignore: user info loading fallback
    }
  }

  if (!taskStore.tasks.length) {
    try {
      await taskStore.fetchTasks();
    } catch {
      // Ignore: list loading handled by state
    }
  }

  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  window.removeEventListener("resize", handleResize);
  document.removeEventListener("click", handleClickOutside);
});

useInfiniteScroll(
  scrollContainer,
  () => {
    if (taskStore.nextCursor && !taskStore.isLoading) {
      taskStore.fetchTasks(true);
    }
  },
  { distance: 100 },
);

function setFilter(status: string, priority: string) {
  taskStore.applyFilters(status, priority);
}

function handleResize() {
  isMobile.value = window.innerWidth < 768;
  if (isMobile.value) {
    viewMode.value = "list";
  }
}

function handleUpdateStatus(id: string, status: "todo" | "in_progress" | "done") {
  taskStore.updateTask(id, { status });
}

function handleRequestComplete(id: string) {
  if (completingTaskId.value) return;
  completingTaskId.value = id;
  window.setTimeout(async () => {
    try {
      await taskStore.updateTask(id, { status: "done" });
    } finally {
      completingTaskId.value = null;
    }
  }, 320);
}

function handleContextMenu(_event: MouseEvent, _task: Task) {
  // Reserved for later context menu actions
}

function setTheme(nextTheme: "cyan" | "purple" | "green" | "pink") {
  themeStore.setTheme(nextTheme);
  isThemeMenuOpen.value = false;
}

function setLanguage(nextLanguage: "zh" | "en") {
  themeStore.setLanguage(nextLanguage);
  isLanguageMenuOpen.value = false;
}

function toggleThemeMenu() {
  isThemeMenuOpen.value = !isThemeMenuOpen.value;
  if (isThemeMenuOpen.value) {
    isUserMenuOpen.value = false;
    isLanguageMenuOpen.value = false;
  }
}

function toggleLanguageMenu() {
  isLanguageMenuOpen.value = !isLanguageMenuOpen.value;
  if (isLanguageMenuOpen.value) {
    isUserMenuOpen.value = false;
    isThemeMenuOpen.value = false;
  }
}

function toggleUserMenu() {
  isUserMenuOpen.value = !isUserMenuOpen.value;
  if (isUserMenuOpen.value) {
    isThemeMenuOpen.value = false;
    isLanguageMenuOpen.value = false;
  }
}

function handleClickOutside(event: MouseEvent) {
  if (!(event.target instanceof Node)) {
    return;
  }
  if (userMenuRef.value && !userMenuRef.value.contains(event.target)) {
    isUserMenuOpen.value = false;
  }
  if (themeMenuRef.value && !themeMenuRef.value.contains(event.target)) {
    isThemeMenuOpen.value = false;
  }
  if (languageMenuRef.value && !languageMenuRef.value.contains(event.target)) {
    isLanguageMenuOpen.value = false;
  }
}

function goToProfile() {
  isUserMenuOpen.value = false;
  router.push("/profile");
}

function goToLogs() {
  isUserMenuOpen.value = false;
  router.push("/logs");
}

function goToPointCloud() {
  router.push("/point-cloud");
}

function handleLogout() {
  isUserMenuOpen.value = false;
  authStore.logoutSync();
  router.push("/login");
}
</script>

<template>
  <div class="min-h-screen bg-[#050a0f] flex flex-col overflow-hidden">
    <header
      class="h-16 border-b border-white/5 flex items-center justify-between px-4 md:px-6 bg-glass sticky top-0 z-20"
    >
      <div class="flex items-center gap-3">
        <div
          class="w-8 h-8 rounded-lg bg-neon/20 border border-neon flex items-center justify-center"
        >
          <span class="text-neon font-bold text-lg leading-none">T</span>
        </div>
        <h1 class="text-lg md:text-xl font-medium tracking-wide text-white">
          {{ t("common.appName") }}
        </h1>
      </div>

      <div class="flex items-center gap-2 md:gap-3">
        <button
          class="rounded-full border border-white/10 px-3 py-1.5 text-sm text-white/80 transition-colors hover:border-neon hover:text-neon"
          @click="goToPointCloud"
        >
          {{ t("tasks.pointCloudBoard") }}
        </button>

        <div
          ref="languageMenuRef"
          class="relative"
        >
          <button
            class="flex items-center gap-2 px-2.5 py-1.5 rounded-full border border-white/10 hover:border-white/25 transition-colors"
            @click="toggleLanguageMenu"
          >
            <Globe class="w-4 h-4 text-white/75" />
            <span class="text-sm text-white/80 hidden sm:block">{{ activeLanguageOption.label }}</span>
            <ChevronDown class="w-4 h-4 text-white/45" />
          </button>
          <div
            v-if="isLanguageMenuOpen"
            class="absolute right-0 top-12 w-32 bg-[#0b1219]/95 border border-white/10 rounded-xl p-1.5 shadow-[0_14px_40px_rgba(0,0,0,0.35)]"
          >
            <button
              v-for="lang in languageOptions"
              :key="lang.value"
              class="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors"
              :class="
                themeStore.language === lang.value
                  ? 'bg-white/10 text-white'
                  : 'text-white/75 hover:bg-white/8'
              "
              @click="setLanguage(lang.value)"
            >
              {{ lang.label }}
            </button>
          </div>
        </div>

        <div
          ref="themeMenuRef"
          class="relative"
        >
          <button
            class="flex items-center gap-2 px-2.5 py-1.5 rounded-full border border-white/10 hover:border-white/25 transition-colors"
            @click="toggleThemeMenu"
          >
            <Palette class="w-4 h-4 text-white/75" />
            <div
              class="w-2 h-2 rounded-full"
              :class="activeThemeOption.bgClass"
            />
            <span class="text-sm text-white/80 hidden sm:block">{{ activeThemeOption.label }}</span>
            <ChevronDown class="w-4 h-4 text-white/45" />
          </button>
          <div
            v-if="isThemeMenuOpen"
            class="absolute right-0 top-12 w-40 bg-[#0b1219]/95 border border-white/10 rounded-xl p-1.5 shadow-[0_14px_40px_rgba(0,0,0,0.35)]"
          >
            <button
              v-for="themeOption in themeOptions"
              :key="themeOption.value"
              class="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors"
              :class="
                themeStore.theme === themeOption.value
                  ? 'bg-white/10 text-white'
                  : 'text-white/75 hover:bg-white/8'
              "
              @click="setTheme(themeOption.value)"
            >
              <div
                class="w-2 h-2 rounded-full"
                :class="themeOption.bgClass"
              />
              <span>{{ themeOption.label }}</span>
            </button>
          </div>
        </div>

        <div
          ref="userMenuRef"
          class="relative"
        >
          <button
            class="flex items-center gap-2 px-2 py-1 rounded-full border border-white/10 hover:border-white/25 transition-colors"
            @click="toggleUserMenu"
          >
            <div
              class="w-8 h-8 rounded-full bg-neon/20 border border-neon flex items-center justify-center text-neon text-xs font-semibold"
            >
              {{ userInitial }}
            </div>
            <span class="text-sm text-white/80 max-w-24 truncate hidden sm:block">
              {{ displayUserName }}
            </span>
            <ChevronDown class="w-4 h-4 text-white/45" />
          </button>

          <div
            v-if="isUserMenuOpen"
            class="absolute right-0 top-12 w-40 bg-[#0b1219]/95 border border-white/10 rounded-xl p-1.5 shadow-[0_14px_40px_rgba(0,0,0,0.35)]"
          >
            <button
              class="w-full text-left px-3 py-2 text-sm text-white/75 rounded-lg hover:bg-white/8 transition-colors"
              @click="goToLogs"
            >
              {{ t("tasks.logCenter") }}
            </button>
            <button
              class="w-full text-left px-3 py-2 text-sm text-white/75 rounded-lg hover:bg-white/8 transition-colors"
              @click="goToProfile"
            >
              {{ t("tasks.profile") }}
            </button>
            <button
              class="w-full text-left px-3 py-2 text-sm text-rose-300 rounded-lg hover:bg-rose-500/10 transition-colors"
              @click="handleLogout"
            >
              {{ t("common.logout") }}
            </button>
          </div>
        </div>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden relative">
      <aside
        class="hidden md:flex w-64 border-r border-white/5 bg-[#0a1118]/80 p-6 flex-col gap-8 flex-shrink-0 z-10"
      >
        <div class="space-y-3">
          <h3 class="text-xs font-semibold text-white/40 uppercase tracking-wider">
            {{ t("tasks.statusFilter") }}
          </h3>
          <div class="space-y-1">
            <button
              v-for="status in statusOptions"
              :key="status.value"
              class="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-all duration-200"
              :class="
                taskStore.filterStatus === status.value
                  ? 'bg-neon/10 text-neon'
                  : 'text-white/60 hover:bg-white/5'
              "
              @click="setFilter(status.value, taskStore.filterPriority)"
            >
              <component
                :is="status.icon"
                class="w-4 h-4"
              />
              <span>{{ status.label }}</span>
            </button>
          </div>
        </div>

        <div class="space-y-3">
          <h3 class="text-xs font-semibold text-white/40 uppercase tracking-wider">
            {{ t("tasks.priorityFilter") }}
          </h3>
          <div class="space-y-1">
            <button
              v-for="p in priorityOptions"
              :key="p.value"
              class="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-all duration-200"
              :class="
                taskStore.filterPriority === p.value
                  ? 'bg-neon/10 text-neon'
                  : 'text-white/60 hover:bg-white/5'
              "
              @click="setFilter(taskStore.filterStatus, p.value)"
            >
              <div
                class="w-2 h-2 rounded-full"
                :class="p.color"
              />
              <span>{{ p.label }}</span>
            </button>
          </div>
        </div>
      </aside>

      <main
        ref="scrollContainer"
        class="flex-1 p-4 md:p-6 overflow-y-auto w-full relative"
      >
        <div class="max-w-5xl mx-auto pb-24">
          <div class="md:hidden mb-5 grid grid-cols-2 gap-2">
            <div class="rounded-xl border border-white/10 bg-white/5 p-2.5 backdrop-blur-md">
              <p class="mb-2 text-xs font-medium tracking-wide text-white/55">
                {{ t("tasks.statusFilter") }}
              </p>
              <div class="relative">
                <select
                  class="w-full appearance-none rounded-lg border border-white/10 bg-[#0b1219]/80 px-2.5 py-2 pr-8 text-sm text-white outline-none transition-colors focus:border-neon"
                  :value="taskStore.filterStatus"
                  @change="setFilter(($event.target as HTMLSelectElement).value, taskStore.filterPriority)"
                >
                  <option
                    v-for="status in statusOptions"
                    :key="status.value"
                    :value="status.value"
                    class="bg-[#0b1219]"
                  >
                    {{ status.label }}
                  </option>
                </select>
                <ChevronDown class="pointer-events-none absolute right-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-white/40" />
              </div>
            </div>

            <div class="rounded-xl border border-white/10 bg-white/5 p-2.5 backdrop-blur-md">
              <p class="mb-2 text-xs font-medium tracking-wide text-white/55">
                {{ t("tasks.priorityFilter") }}
              </p>
              <div class="relative">
                <select
                  class="w-full appearance-none rounded-lg border border-white/10 bg-[#0b1219]/80 px-2.5 py-2 pr-8 text-sm text-white outline-none transition-colors focus:border-neon"
                  :value="taskStore.filterPriority"
                  @change="setFilter(taskStore.filterStatus, ($event.target as HTMLSelectElement).value)"
                >
                  <option
                    v-for="p in priorityOptions"
                    :key="p.value"
                    :value="p.value"
                    class="bg-[#0b1219]"
                  >
                    {{ p.label }}
                  </option>
                </select>
                <ChevronDown class="pointer-events-none absolute right-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-white/40" />
              </div>
            </div>
          </div>

          <div class="mb-6 md:mb-8 flex items-center justify-between">
            <h2 class="text-xl md:text-2xl font-light tracking-wide">
              {{ t("tasks.title") }}
              <span class="text-white/30 text-sm md:text-base ml-2">({{ taskStore.tasks.length }})</span>
            </h2>
            <div class="hidden md:flex items-center gap-1 rounded-lg border border-white/10 bg-white/5 p-1">
              <button
                class="flex h-9 w-9 items-center justify-center rounded-md transition-colors"
                :class="viewMode === 'card' ? 'bg-neon/20 text-neon' : 'text-white/50 hover:text-white/80'"
                :aria-label="t('tasks.viewCard')"
                @click="viewMode = 'card'"
              >
                <LayoutGrid class="h-4 w-4" />
              </button>
              <button
                class="flex h-9 w-9 items-center justify-center rounded-md transition-colors"
                :class="viewMode === 'list' ? 'bg-neon/20 text-neon' : 'text-white/50 hover:text-white/80'"
                :aria-label="t('tasks.viewList')"
                @click="viewMode = 'list'"
              >
                <Rows3 class="h-4 w-4" />
              </button>
            </div>
          </div>

          <TransitionGroup
            v-if="taskStore.tasks.length > 0"
            name="task-sort"
            tag="div"
            class="grid"
            :class="
              viewMode === 'card'
                ? 'grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4'
                : 'grid-cols-1 gap-2.5'
            "
          >
            <TaskCard
              v-for="task in sortedTasks"
              :key="task.id"
              :task="task"
              :view-mode="viewMode"
              :completing-task-id="completingTaskId"
              @update-status="handleUpdateStatus"
              @request-complete="handleRequestComplete"
              @contextmenu="handleContextMenu"
            />
          </TransitionGroup>

          <div
            v-else-if="!taskStore.isLoading"
            class="h-64 flex flex-col items-center justify-center text-white/30"
          >
            <ListTodo class="w-16 h-16 mb-4 opacity-50" />
            <p>{{ t("tasks.empty") }}</p>
          </div>

          <div
            v-if="taskStore.isLoading"
            class="py-8 flex justify-center"
          >
            <div
              class="w-6 h-6 border-2 border-neon border-t-transparent rounded-full animate-spin"
            />
          </div>
        </div>
      </main>

      <button
        class="fixed md:absolute bottom-6 right-4 md:bottom-8 md:right-8 w-14 h-14 bg-neon rounded-full flex items-center justify-center text-[#050a0f] hover:scale-110 hover:shadow-[0_0_20px_var(--neon-glow)] transition-all duration-300 z-40"
        :aria-label="t('tasks.createTask')"
        @click="isDrawerOpen = true"
      >
        <Plus
          class="w-6 h-6"
          stroke-width="2.5"
        />
      </button>
    </div>

    <CreateTaskDrawer v-model="isDrawerOpen" />
  </div>
</template>

<style scoped>
.task-sort-move {
  transition: transform 220ms ease;
}
</style>
