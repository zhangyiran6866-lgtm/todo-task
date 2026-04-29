<template>
  <div
    class="task-card bg-glass rounded-xl relative overflow-hidden transition-all duration-300"
    :class="[
      viewMode === 'list' ? 'px-4 py-3' : 'p-4',
      {
        'opacity-60': task.status === 'done',
        'cursor-not-allowed': isOverdue,
        'task-card--disabled': isOverdue,
      },
    ]"
    @contextmenu.prevent="emit('contextmenu', $event, task)"
  >
    <div
      class="flex items-start"
      :class="viewMode === 'list' ? 'gap-2.5' : 'gap-3'"
    >
      <div
        v-if="isOverdue"
        class="flex-shrink-0 rounded-full flex items-center justify-center text-rose-500"
        :class="viewMode === 'list' ? 'mt-0.5 w-4 h-4' : 'mt-1 w-5 h-5'"
      >
        <AlertTriangle
          class="w-4 h-4"
          stroke-width="2.5"
        />
      </div>

      <button
        v-else
        class="task-check-trigger flex-shrink-0 rounded-full border flex items-center justify-center transition-colors duration-200 focus:outline-none relative overflow-visible"
        :class="[
          viewMode === 'list' ? 'mt-0.5 w-4 h-4' : 'mt-1 w-5 h-5',
          isCompleting
            ? 'border-transparent bg-transparent'
            : task.status === 'done'
              ? 'bg-emerald-500/20 border-emerald-400 text-emerald-300'
              : 'border-white/20 text-white/70 hover:border-neon',
        ]"
        @click.stop="toggleStatus"
      >
        <Check
          v-if="task.status === 'done' || isCompleting"
          class="task-check-icon w-3 h-3"
          :class="isCompleting ? 'task-check-icon--pop' : 'text-emerald-300'"
          stroke-width="3"
        />
        <span
          v-if="isCompleting"
          class="task-burst"
          aria-hidden="true"
        >
          <i
            v-for="idx in 8"
            :key="idx"
            class="task-shard"
            :style="{
              '--shard-angle': `${shardAngles[idx - 1]}deg`,
              '--shard-distance': `${shardDistances[idx - 1]}px`,
              '--shard-delay': `${shardDelays[idx - 1]}ms`,
              '--shard-scale': `${shardScales[idx - 1]}`,
              '--shard-hue': `${shardHues[idx - 1]}`,
            }"
          />
        </span>
      </button>

      <!-- Content -->
      <div class="flex-1 min-w-0">
        <h3
          class="font-medium text-white truncate transition-all duration-200"
          :class="[
            viewMode === 'list' ? 'text-base' : 'text-lg',
            { 'line-through text-white/50': task.status === 'done' },
          ]"
        >
          {{ task.title }}
        </h3>
        <p
          v-if="task.description"
          class="text-white/60"
          :class="viewMode === 'list' ? 'mt-0.5 text-xs line-clamp-1' : 'mt-1 text-sm line-clamp-2'"
        >
          {{ task.description }}
        </p>

        <div
          v-if="task.due_at"
          class="flex items-center gap-1.5 text-xs"
          :class="[viewMode === 'list' ? 'mt-2' : 'mt-3', dueDateColor]"
        >
          <Calendar class="w-3.5 h-3.5" />
          <span>{{ t('tasks.dueDateShort') }}{{ formattedDate }}</span>
        </div>

        <div
          class="flex items-center justify-between gap-3 text-xs"
          :class="viewMode === 'list' ? 'mt-2.5' : 'mt-4'"
        >
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
import { useI18n } from 'vue-i18n'
import { Calendar, AlertCircle, AlertTriangle, ArrowUpCircle, Check } from 'lucide-vue-next'
import type { Task } from '@/api/task'

const props = withDefaults(defineProps<{
  task: Task
  viewMode?: 'card' | 'list'
  completingTaskId?: string | null
}>(), {
  viewMode: 'card',
  completingTaskId: null
})

const emit = defineEmits<{
  (e: 'update-status', id: string, status: 'todo' | 'in_progress' | 'done'): void
  (e: 'request-complete', id: string): void
  (e: 'contextmenu', event: MouseEvent, task: Task): void
}>()

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

const isCompleting = computed(() => props.completingTaskId === props.task.id)
const shardAngles = [12, 58, 101, 147, 189, 234, 281, 328]
const shardDistances = [18, 24, 20, 27, 22, 29, 19, 25]
const shardDelays = [0, 10, 4, 14, 6, 18, 8, 12]
const shardScales = [0.9, 1.15, 0.8, 1.25, 1, 1.2, 0.85, 1.1]
const shardHues = [16, 48, 90, 172, 204, 258, 312, 346]

function toggleStatus() {
  if (isOverdue.value) return
  if (props.task.status === 'done') {
    emit('update-status', props.task.id, 'todo')
    return
  }
  if (isCompleting.value) return
  emit('request-complete', props.task.id)
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

.task-card--disabled:hover {
  transform: none;
  border-color: var(--border-dim);
  box-shadow: none;
}

.task-burst {
  position: absolute;
  inset: -18px;
  pointer-events: none;
}

.task-shard {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 4px;
  height: 9px;
  border-radius: 9999px;
  transform-origin: center;
  opacity: 0;
  background: hsl(var(--shard-hue), 98%, 64%);
  box-shadow: 0 0 6px currentColor;
  color: hsl(var(--shard-hue), 98%, 64%);
  animation: shard-burst 300ms ease-out forwards;
  animation-delay: var(--shard-delay);
}

@keyframes shard-burst {
  0% {
    opacity: 1;
    transform: translate(-50%, -50%) rotate(var(--shard-angle)) translateY(0) scale(calc(0.8 * var(--shard-scale)));
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) rotate(var(--shard-angle)) translateY(calc(var(--shard-distance) * -1)) scale(calc(1.06 * var(--shard-scale)));
  }
}

.task-check-icon--pop {
  color: #34d399;
  opacity: 0;
  transform: scale(0.35);
  animation: check-pop 220ms ease-out 110ms forwards;
}

@keyframes check-pop {
  0% {
    opacity: 0;
    transform: scale(0.35);
  }
  70% {
    opacity: 1;
    transform: scale(1.08);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
