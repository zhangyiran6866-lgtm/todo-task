<template>
  <div
    class="task-card bg-glass rounded-xl cursor-pointer relative overflow-hidden transition-all duration-300"
    :class="[viewMode === 'list' ? 'p-5' : 'p-4', { 'opacity-60': task.status === 'done' }]"
    @click="goToDetail"
    @contextmenu.prevent="emit('contextmenu', $event, task)"
  >
    <div class="flex items-start gap-3">
      <button
        class="mt-1 flex-shrink-0 w-5 h-5 rounded-full border flex items-center justify-center transition-colors duration-200 focus:outline-none"
        :class="
          task.status === 'done'
            ? 'bg-neon bg-opacity-20 border-neon text-neon'
            : 'border-white/20 hover:border-neon'
        "
        @click.stop="toggleStatus"
      >
        <Check
          v-if="task.status === 'done'"
          class="w-3 h-3"
          stroke-width="3"
        />
      </button>

      <!-- Content -->
      <div class="flex-1 min-w-0">
        <h3
          class="font-medium text-white truncate transition-all duration-200"
          :class="[
            viewMode === 'list' ? 'text-xl' : 'text-lg',
            { 'line-through text-white/50': task.status === 'done' },
          ]"
        >
          {{ task.title }}
        </h3>
        <p
          v-if="task.description"
          class="mt-1 text-sm text-white/60 line-clamp-2"
        >
          {{ task.description }}
        </p>

        <div
          v-if="task.due_at"
          class="mt-3 flex items-center gap-1.5 text-xs"
          :class="dueDateColor"
        >
          <Calendar class="w-3.5 h-3.5" />
          <span>{{ t('tasks.dueDateShort') }}{{ formattedDate }}</span>
        </div>

        <div class="mt-4 flex items-center justify-between gap-3 text-xs">
          <div
            class="px-2 py-1 rounded-md border"
            :class="statusTagClass"
          >
            {{ statusText }}
          </div>

          <div class="flex items-center gap-2">
            <div
              v-if="isOverdue"
              class="px-2 py-1 rounded-md flex items-center gap-1 border text-rose-500 border-rose-500/50 bg-rose-500/10"
            >
              <AlertCircle class="w-3 h-3" />
              <span>{{ t('tasks.statusExpired') }}</span>
            </div>

            <div
              v-if="task.priority"
              class="px-2 py-1 rounded-md flex items-center gap-1 border"
              :class="priorityTagClass"
            >
              <AlertCircle
                v-if="task.priority === 'critical' || task.priority === 'important'"
                class="w-3 h-3"
              />
              <ArrowUpCircle
                v-else-if="task.priority === 'urgent'"
                class="w-3 h-3"
              />
              <span>{{ priorityText }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Calendar, AlertCircle, ArrowUpCircle, Check } from 'lucide-vue-next'
import type { Task } from '@/api/task'

const props = withDefaults(defineProps<{
  task: Task
  viewMode?: 'card' | 'list'
}>(), {
  viewMode: 'card'
})

const emit = defineEmits<{
  (e: 'update-status', id: string, status: 'todo' | 'in_progress' | 'done'): void
  (e: 'contextmenu', event: MouseEvent, task: Task): void
}>()

const router = useRouter()
const { t } = useI18n()

const priorityTagClass = computed(() => {
  switch (props.task.priority) {
    case 'critical': return 'text-rose-400 border-rose-500/30 bg-rose-500/10'
    case 'important': return 'text-purple-400 border-purple-500/30 bg-purple-500/10'
    case 'urgent': return 'text-amber-400 border-amber-500/30 bg-amber-500/10'
    case 'low': return 'text-emerald-400 border-emerald-500/30 bg-emerald-500/10'
    case 'routine': return 'text-blue-400 border-blue-400/30 bg-blue-400/10'
    default: return ''
  }
})

const statusTagClass = computed(() => {
  switch (props.task.status) {
    case 'done': return 'text-neon border-neon/50 bg-neon/10'
    case 'in_progress': return 'text-blue-300 border-blue-400/40 bg-blue-400/10'
    default: return 'text-white/70 border-white/20 bg-white/5'
  }
})

const statusText = computed(() => {
  switch (props.task.status) {
    case 'done': return t('tasks.statusDone')
    case 'in_progress': return t('tasks.statusInProgress')
    default: return t('tasks.statusTodo')
  }
})

const priorityText = computed(() => {
  switch (props.task.priority) {
    case 'critical': return t('tasks.priorityCritical')
    case 'important': return t('tasks.priorityImportant')
    case 'urgent': return t('tasks.priorityUrgent')
    case 'low': return t('tasks.priorityLow')
    case 'routine': return t('tasks.priorityRoutine')
    default: return t('tasks.priorityAll')
  }
})

const isOverdue = computed(() => {
  if (!props.task.due_at || props.task.status === 'done') return false
  return new Date(props.task.due_at).getTime() < Date.now()
})

const dueDateColor = computed(() => {
  if (props.task.status === 'done') return 'text-white/40'
  return isOverdue.value ? 'text-rose-400 font-medium' : 'text-white/50'
})

const formattedDate = computed(() => {
  if (!props.task.due_at) return ''
  return new Date(props.task.due_at).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
})

function toggleStatus() {
  const newStatus = props.task.status === 'done' ? 'todo' : 'done'
  emit('update-status', props.task.id, newStatus)
}

function goToDetail() {
  router.push(`/tasks/${props.task.id}`)
}
</script>

<style scoped>
.task-card {
  border: 1px solid var(--border-dim);
}

.task-card:hover {
  transform: translateY(-2px);
  border-color: var(--neon);
  box-shadow: 0 0 15px var(--neon-glow);
}
</style>
