<template>
  <div class="min-h-screen bg-[#050a0f] flex flex-col overflow-hidden">
    <!-- Navbar -->
    <header class="h-16 border-b border-white/5 flex items-center justify-between px-6 bg-glass sticky top-0 z-20">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 rounded-lg bg-neon/20 border border-neon flex items-center justify-center">
          <span class="text-neon font-bold text-lg leading-none">T</span>
        </div>
        <h1 class="text-xl font-medium tracking-wide text-white">TodoTask</h1>
      </div>

      <div class="flex items-center gap-4">
        <select 
          v-model="theme"
          class="bg-transparent text-sm text-white/70 focus:outline-none focus:text-neon cursor-pointer"
        >
          <option value="cyan">赛博靛 (Cyan)</option>
          <option value="purple">霓虹紫 (Purple)</option>
          <option value="green">骇客绿 (Green)</option>
          <option value="pink">极客粉 (Pink)</option>
        </select>

        <button 
          @click="logout"
          class="flex items-center gap-2 text-white/60 hover:text-rose-400 transition-colors"
        >
          <LogOut class="w-4 h-4" />
          <span class="text-sm">安全登出</span>
        </button>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden relative">
      <!-- Sidebar / Filters -->
      <aside class="w-64 border-r border-white/5 bg-[#0a1118]/80 p-6 flex flex-col gap-8 flex-shrink-0 z-10">
        <!-- Status Filter -->
        <div class="space-y-3">
          <h3 class="text-xs font-semibold text-white/40 uppercase tracking-wider">任务状态</h3>
          <div class="space-y-1">
            <button 
              v-for="status in statusOptions" 
              :key="status.value"
              class="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-all duration-200"
              :class="taskStore.filterStatus === status.value ? 'bg-neon/10 text-neon' : 'text-white/60 hover:bg-white/5'"
              @click="setFilter(status.value, taskStore.filterPriority)"
            >
              <component :is="status.icon" class="w-4 h-4" />
              <span>{{ status.label }}</span>
            </button>
          </div>
        </div>

        <!-- Priority Filter -->
        <div class="space-y-3">
          <h3 class="text-xs font-semibold text-white/40 uppercase tracking-wider">标签状态</h3>
          <div class="space-y-1">
            <button 
              v-for="p in priorityOptions" 
              :key="p.value"
              class="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-all duration-200"
              :class="taskStore.filterPriority === p.value ? 'bg-neon/10 text-neon' : 'text-white/60 hover:bg-white/5'"
              @click="setFilter(taskStore.filterStatus, p.value)"
            >
              <div class="w-2 h-2 rounded-full" :class="p.color"></div>
              <span>{{ p.label }}</span>
            </button>
          </div>
        </div>
      </aside>

      <!-- Main Content area with scrolling -->
      <main ref="scrollContainer" class="flex-1 p-6 overflow-y-auto w-full relative">
        <div class="max-w-5xl mx-auto pb-24">
          <!-- Optional Header -->
          <div class="mb-8 flex items-center justify-between">
            <h2 class="text-2xl font-light tracking-wide">
              任务列表
              <span class="text-white/30 text-base ml-2">({{ taskStore.tasks.length }})</span>
            </h2>
          </div>

          <!-- Grid layout -->
          <div v-if="taskStore.tasks.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <TaskCard 
              v-for="task in taskStore.tasks" 
              :key="task.id" 
              :task="task" 
              @update-status="handleUpdateStatus"
              @contextmenu="handleContextMenu"
            />
          </div>

          <!-- Empty State -->
          <div v-else-if="!taskStore.isLoading" class="h-64 flex flex-col items-center justify-center text-white/30">
            <ListTodo class="w-16 h-16 mb-4 opacity-50" />
            <p>暂无任务</p>
          </div>

          <!-- Loading indicator -->
          <div v-if="taskStore.isLoading" class="py-8 flex justify-center">
            <div class="w-6 h-6 border-2 border-neon border-t-transparent rounded-full animate-spin"></div>
          </div>
        </div>
      </main>
      
      <!-- Floating Action Button -->
      <button 
        @click="isDrawerOpen = true"
        class="absolute bottom-8 right-8 w-14 h-14 bg-neon rounded-full flex items-center justify-center text-[#050a0f] hover:scale-110 hover:shadow-[0_0_20px_var(--neon-glow)] transition-all duration-300 z-30"
      >
        <Plus class="w-6 h-6" stroke-width="2.5" />
      </button>
    </div>

    <!-- Create Drawer -->
    <CreateTaskDrawer v-model="isDrawerOpen" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useInfiniteScroll } from '@vueuse/core'
import { useTaskStore } from '@/stores/use-task-store'
import { useAuthStore } from '@/stores/use-auth-store'
import TaskCard from '@/components/tasks/TaskCard.vue'
import CreateTaskDrawer from '@/components/tasks/CreateTaskDrawer.vue'
import { 
  LogOut, Plus, ListTodo, Circle, PlayCircle, CheckCircle2, AlertCircle
} from 'lucide-vue-next'
import type { Task } from '@/api/task'

const router = useRouter()
const authStore = useAuthStore()
const taskStore = useTaskStore()

const scrollContainer = ref<HTMLElement | null>(null)
const isDrawerOpen = ref(false)

// Theme hack (PRD asks for 4 themes using data-theme on html)
const theme = ref(localStorage.getItem('theme') || 'cyan')
watch(theme, (newVal) => {
  document.documentElement.setAttribute('data-theme', newVal)
  localStorage.setItem('theme', newVal)
})

const statusOptions = [
  { label: '全部任务', value: '', icon: ListTodo },
  { label: '待处理', value: 'todo', icon: Circle },
  { label: '进行中', value: 'in_progress', icon: PlayCircle },
  { label: '已过期', value: 'expired', icon: AlertCircle },
  { label: '已完成', value: 'done', icon: CheckCircle2 },
]

const priorityOptions = [
  { label: '所有类型', value: '', color: 'bg-white/20' },
  { label: '重要紧急', value: 'critical', color: 'bg-rose-500' },
  { label: '重要不紧急', value: 'important', color: 'bg-purple-500' },
  { label: '紧急不重要', value: 'urgent', color: 'bg-amber-500' },
  { label: '不重要也不紧急', value: 'low', color: 'bg-emerald-500' },
  { label: '日常任务', value: 'routine', color: 'bg-blue-400' },
]

onMounted(async () => {
  document.documentElement.setAttribute('data-theme', theme.value)
  if (!taskStore.tasks.length) {
    await taskStore.fetchTasks()
  }
})

useInfiniteScroll(
  scrollContainer,
  () => {
    // 只有当有 nextCursor 并且未在加载时触发
    if (taskStore.nextCursor && !taskStore.isLoading) {
      taskStore.fetchTasks(true)
    }
  },
  { distance: 100 }
)

function setFilter(status: string, priority: string) {
  taskStore.applyFilters(status, priority)
}

function handleUpdateStatus(id: string, status: string) {
  taskStore.updateTask(id, { status })
}

function handleContextMenu(_event: MouseEvent, task: Task) {
  // To be expanded in Phase 4. For now, native logging.
  console.log('Context menu initiated for', task.id)
}

function logout() {
  authStore.logoutSync()
  router.push('/login')
}
</script>

<style scoped>
/* Scoped styles overrides if needed */
</style>
