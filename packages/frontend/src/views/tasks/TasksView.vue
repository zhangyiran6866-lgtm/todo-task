<template>
  <div class="min-h-screen bg-[#050a0f] flex flex-col overflow-hidden">
    <!-- Navbar -->
    <header
      class="h-16 border-b border-white/5 flex items-center justify-between px-6 bg-glass sticky top-0 z-20"
    >
      <div class="flex items-center gap-3">
        <div
          class="w-8 h-8 rounded-lg bg-neon/20 border border-neon flex items-center justify-center"
        >
          <span class="text-neon font-bold text-lg leading-none">T</span>
        </div>
        <h1 class="text-xl font-medium tracking-wide text-white">
          TodoTask
        </h1>
      </div>

      <div class="flex items-center gap-4">
        <div
          ref="themeMenuRef"
          class="relative"
        >
          <button
            class="flex items-center gap-2.5 px-3 py-1.5 rounded-full border border-white/10 hover:border-white/25 transition-colors"
            @click="toggleThemeMenu"
          >
            <div
              class="w-2 h-2 rounded-full"
              :class="activeThemeOption.bgClass"
            />
            <span class="text-sm text-white/80">{{
              activeThemeOption.label
            }}</span>
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
                theme === themeOption.value
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
            <span class="text-sm text-white/80 max-w-24 truncate">{{
              displayUserName
            }}</span>
            <ChevronDown class="w-4 h-4 text-white/45" />
          </button>

          <div
            v-if="isUserMenuOpen"
            class="absolute right-0 top-12 w-40 bg-[#0b1219]/95 border border-white/10 rounded-xl p-1.5 shadow-[0_14px_40px_rgba(0,0,0,0.35)]"
          >
            <button
              class="w-full text-left px-3 py-2 text-sm text-white/75 rounded-lg hover:bg-white/8 transition-colors"
              @click="goToProfile"
            >
              修改密码
            </button>
            <button
              class="w-full text-left px-3 py-2 text-sm text-rose-300 rounded-lg hover:bg-rose-500/10 transition-colors"
              @click="handleLogout"
            >
              退出登录
            </button>
          </div>
        </div>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden relative">
      <!-- Sidebar / Filters -->
      <aside
        class="w-64 border-r border-white/5 bg-[#0a1118]/80 p-6 flex flex-col gap-8 flex-shrink-0 z-10"
      >
        <!-- Status Filter -->
        <div class="space-y-3">
          <h3
            class="text-xs font-semibold text-white/40 uppercase tracking-wider"
          >
            任务状态
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

        <!-- Priority Filter -->
        <div class="space-y-3">
          <h3
            class="text-xs font-semibold text-white/40 uppercase tracking-wider"
          >
            标签状态
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

      <!-- Main Content area with scrolling -->
      <main
        ref="scrollContainer"
        class="flex-1 p-6 overflow-y-auto w-full relative"
      >
        <div class="max-w-5xl mx-auto pb-24">
          <!-- Optional Header -->
          <div class="mb-8 flex items-center justify-between">
            <h2 class="text-2xl font-light tracking-wide">
              任务列表
              <span class="text-white/30 text-base ml-2">({{ taskStore.tasks.length }})</span>
            </h2>
            <div class="flex items-center gap-1 rounded-lg border border-white/10 bg-white/5 p-1">
              <button
                class="flex h-9 w-9 items-center justify-center rounded-md transition-colors"
                :class="viewMode === 'card' ? 'bg-neon/20 text-neon' : 'text-white/50 hover:text-white/80'"
                aria-label="卡片视图"
                @click="viewMode = 'card'"
              >
                <LayoutGrid class="h-4 w-4" />
              </button>
              <button
                class="flex h-9 w-9 items-center justify-center rounded-md transition-colors"
                :class="viewMode === 'list' ? 'bg-neon/20 text-neon' : 'text-white/50 hover:text-white/80'"
                aria-label="列表视图"
                @click="viewMode = 'list'"
              >
                <Rows3 class="h-4 w-4" />
              </button>
            </div>
          </div>

          <!-- Grid layout -->
          <div
            v-if="taskStore.tasks.length > 0"
            class="grid gap-4"
            :class="
              viewMode === 'card'
                ? 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3'
                : 'grid-cols-1'
            "
          >
            <TaskCard
              v-for="task in taskStore.tasks"
              :key="task.id"
              :task="task"
              :view-mode="viewMode"
              @update-status="handleUpdateStatus"
              @contextmenu="handleContextMenu"
            />
          </div>

          <!-- Empty State -->
          <div
            v-else-if="!taskStore.isLoading"
            class="h-64 flex flex-col items-center justify-center text-white/30"
          >
            <ListTodo class="w-16 h-16 mb-4 opacity-50" />
            <p>暂无任务</p>
          </div>

          <!-- Loading indicator -->
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

      <!-- Floating Action Button -->
      <button
        class="absolute bottom-8 right-8 w-14 h-14 bg-neon rounded-full flex items-center justify-center text-[#050a0f] hover:scale-110 hover:shadow-[0_0_20px_var(--neon-glow)] transition-all duration-300 z-30"
        @click="isDrawerOpen = true"
      >
        <Plus
          class="w-6 h-6"
          stroke-width="2.5"
        />
      </button>
    </div>

    <!-- Create Drawer -->
    <CreateTaskDrawer v-model="isDrawerOpen" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, watch } from "vue";
import { useRouter } from "vue-router";
import { useInfiniteScroll } from "@vueuse/core";
import { useTaskStore } from "@/stores/use-task-store";
import { useAuthStore } from "@/stores/use-auth-store";
import TaskCard from "@/components/tasks/TaskCard.vue";
import CreateTaskDrawer from "@/components/tasks/CreateTaskDrawer.vue";
import {
  Plus,
  ListTodo,
  Circle,
  PlayCircle,
  CheckCircle2,
  AlertCircle,
  ChevronDown,
  LayoutGrid,
  Rows3,
} from "lucide-vue-next";
import type { Task } from "@/api/task";

const router = useRouter();
const authStore = useAuthStore();
const taskStore = useTaskStore();

const scrollContainer = ref<HTMLElement | null>(null);
const isDrawerOpen = ref(false);
const viewMode = ref<"card" | "list">("card");
const userMenuRef = ref<HTMLElement | null>(null);
const themeMenuRef = ref<HTMLElement | null>(null);
const isUserMenuOpen = ref(false);
const isThemeMenuOpen = ref(false);

// Theme hack (PRD asks for 4 themes using data-theme on html)
const theme = ref(localStorage.getItem("theme") || "cyan");
watch(theme, (newVal) => {
  document.documentElement.setAttribute("data-theme", newVal);
  localStorage.setItem("theme", newVal);
});
const themeOptions = [
  { value: "cyan", label: "青蓝", bgClass: "bg-[#00f3ff]" },
  { value: "purple", label: "霓紫", bgClass: "bg-[#930ec8]" },
  { value: "green", label: "荧绿", bgClass: "bg-[#39ff14]" },
  { value: "pink", label: "电粉", bgClass: "bg-[#dc86d4]" },
] as const;

const statusOptions = [
  { label: "全部任务", value: "", icon: ListTodo },
  { label: "待处理", value: "todo", icon: Circle },
  { label: "进行中", value: "in_progress", icon: PlayCircle },
  { label: "已过期", value: "expired", icon: AlertCircle },
  { label: "已完成", value: "done", icon: CheckCircle2 },
];

const priorityOptions = [
  { label: "所有类型", value: "", color: "bg-white/20" },
  { label: "重要紧急", value: "critical", color: "bg-rose-500" },
  { label: "重要不紧急", value: "important", color: "bg-purple-500" },
  { label: "紧急不重要", value: "urgent", color: "bg-amber-500" },
  { label: "不重要也不紧急", value: "low", color: "bg-emerald-500" },
  { label: "日常任务", value: "routine", color: "bg-blue-400" },
];

onMounted(async () => {
  document.documentElement.setAttribute("data-theme", theme.value);

  if (!authStore.user && authStore.isLoggedIn) {
    try {
      await authStore.fetchUser();
    } catch (error) {
      console.error("Failed to fetch user profile in tasks view", error);
    }
  }

  if (!taskStore.tasks.length) {
    try {
      await taskStore.fetchTasks();
    } catch (error) {
      console.error("Failed to fetch tasks in tasks view", error);
    }
  }

  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
});

useInfiniteScroll(
  scrollContainer,
  () => {
    // 只有当有 nextCursor 并且未在加载时触发
    if (taskStore.nextCursor && !taskStore.isLoading) {
      taskStore.fetchTasks(true);
    }
  },
  { distance: 100 },
);

function setFilter(status: string, priority: string) {
  taskStore.applyFilters(status, priority);
}

function handleUpdateStatus(id: string, status: string) {
  taskStore.updateTask(id, { status });
}

function handleContextMenu(_event: MouseEvent, task: Task) {
  // To be expanded in Phase 4. For now, native logging.
  console.log("Context menu initiated for", task.id);
}

const displayUserName = computed(
  () => authStore.user?.nickname || authStore.user?.email || "用户",
);
const userInitial = computed(() =>
  displayUserName.value.trim().charAt(0).toUpperCase(),
);

function setTheme(nextTheme: string) {
  theme.value = nextTheme;
  isThemeMenuOpen.value = false;
}

const activeThemeOption = computed(() => {
  return (
    themeOptions.find((option) => option.value === theme.value) ||
    themeOptions[0]
  );
});

function toggleThemeMenu() {
  isThemeMenuOpen.value = !isThemeMenuOpen.value;
  if (isThemeMenuOpen.value) {
    isUserMenuOpen.value = false;
  }
}

function toggleUserMenu() {
  isUserMenuOpen.value = !isUserMenuOpen.value;
  if (isUserMenuOpen.value) {
    isThemeMenuOpen.value = false;
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
}

function goToProfile() {
  isUserMenuOpen.value = false;
  router.push("/profile");
}

function handleLogout() {
  isUserMenuOpen.value = false;
  authStore.logoutSync();
  router.push("/login");
}
</script>

<style scoped>
/* Scoped styles overrides if needed */
</style>
