<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="modelValue"
        class="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 transition-opacity"
        @click="requestClose"
      />
    </Transition>

    <Transition name="slide">
      <div
        v-if="modelValue"
        class="fixed right-0 top-0 bottom-0 w-full max-w-2xl bg-[#0a1118] border-l border-white/10 z-50 flex flex-col shadow-2xl"
      >
        <div class="px-6 py-5 border-b border-white/10 flex items-center justify-between">
          <h2 class="text-xl font-medium text-white tracking-wide">
            {{ t('tasks.createTask') }}
          </h2>
          <button
            class="p-2 text-white/50 hover:text-neon transition-colors duration-200"
            @click="requestClose"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="flex-1 overflow-y-auto px-6 py-6">
          <form
            class="space-y-8"
            @submit.prevent="submit"
          >
            <div class="space-y-2 relative">
              <div class="flex justify-between items-center">
                <label class="block text-sm font-medium text-white/70">
                  {{ t('tasks.taskName') }} <span class="text-neon ml-1">*</span>
                </label>
                <span class="text-xs text-white/40">{{ form.title.length }}/20</span>
              </div>
              <input
                v-model="form.title"
                required
                type="text"
                maxlength="20"
                :placeholder="t('tasks.taskName')"
                class="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 text-white placeholder-white/30 focus:outline-none focus:border-neon focus:shadow-[0_0_10px_var(--neon-glow)] transition-all duration-300"
              >
            </div>

            <div class="space-y-2">
              <label class="block text-sm font-medium text-white/70">{{ t('tasks.taskDescription') }}</label>
              <textarea
                v-model="form.description"
                rows="4"
                :placeholder="t('tasks.taskDescription')"
                class="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 text-white placeholder-white/30 focus:outline-none focus:border-neon focus:shadow-[0_0_10px_var(--neon-glow)] transition-all duration-300 resize-none"
              />
            </div>

            <div class="space-y-3">
              <label class="block text-sm font-medium text-white/70">{{ t('tasks.priorityFilter') }}</label>
              <div class="grid grid-cols-2 md:grid-cols-3 gap-2">
                <button
                  v-for="p in priorityOptions"
                  :key="p.value"
                  type="button"
                  class="px-4 py-2 rounded-lg text-sm border transition-all duration-300 pointer whitespace-nowrap"
                  :class="[
                    form.priority === p.value
                      ? `${p.activeClass} shadow-[0_0_15px_var(--neon-glow)]`
                      : 'border-white/10 text-white/40 hover:border-white/30 bg-[#111a24]'
                  ]"
                  @click="form.priority = p.value"
                >
                  {{ t(p.key) }}
                </button>
              </div>
            </div>

            <div class="space-y-3">
              <div class="flex items-center justify-between gap-3">
                <label class="block text-sm font-medium text-white/70">
                  {{ t('tasks.startDate') }}
                </label>
                <button
                  type="button"
                  class="text-xs text-neon hover:text-neon/80 transition-colors"
                  @click="setStartNow"
                >
                  {{ t('tasks.setNow') }}
                </button>
              </div>
              <div class="vue-datepicker-wrapper">
                <VueDatePicker
                  v-model="startDate"
                  dark
                  :placeholder="t('tasks.selectStartDate')"
                  format="yyyy/MM/dd HH:mm"
                  :enable-time-picker="true"
                  :clearable="true"
                  teleport="body"
                >
                  <template #action-row="{ selectDate, closePicker }">
                    <div class="flex justify-end gap-4 px-2 py-1">
                      <button
                        type="button"
                        class="text-white/60 text-sm hover:text-white transition-colors"
                        @click="closePicker"
                      >
                        {{ t('common.cancel') }}
                      </button>
                      <button
                        type="button"
                        class="text-neon text-sm font-medium hover:text-neon/80 transition-colors"
                        @click="selectDate"
                      >
                        {{ t('common.confirm') }}
                      </button>
                    </div>
                  </template>
                </VueDatePicker>
              </div>
            </div>

            <div class="space-y-3">
              <label class="block text-sm font-medium text-white/70">
                {{ t('tasks.dueDate') }} (DDL) <span class="text-neon ml-1">*</span>
              </label>
              <div class="vue-datepicker-wrapper">
                <VueDatePicker
                  v-model="dueDate"
                  dark
                  :preset-dates="presetDates"
                  :placeholder="t('tasks.selectDueDate')"
                  format="yyyy/MM/dd HH:mm"
                  :enable-time-picker="true"
                  :clearable="false"
                  teleport="body"
                >
                  <template #action-row="{ selectDate, closePicker }">
                    <div class="flex justify-end gap-4 px-2 py-1">
                      <button
                        type="button"
                        class="text-white/60 text-sm hover:text-white transition-colors"
                        @click="closePicker"
                      >
                        {{ t('common.cancel') }}
                      </button>
                      <button
                        type="button"
                        class="text-neon text-sm font-medium hover:text-neon/80 transition-colors"
                        @click="selectDate"
                      >
                        {{ t('common.confirm') }}
                      </button>
                    </div>
                  </template>
                </VueDatePicker>
              </div>
            </div>
          </form>
        </div>

        <div class="px-6 py-5 border-t border-white/10 bg-[#0a1118] flex justify-end gap-3 mt-auto">
          <button
            type="button"
            class="px-5 py-2.5 rounded-lg text-white/70 hover:bg-white/5 transition-colors duration-200"
            @click="requestClose"
          >
            {{ t('tasks.cancelCreate') }}
          </button>
          <button
            type="button"
            class="px-6 py-2.5 rounded-lg bg-neon text-[#050a0f] font-medium tracking-wide hover:shadow-[0_0_15px_var(--neon-glow)] transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="!canSubmit || isSubmitting"
            @click="submit"
          >
            {{ isSubmitting ? t('tasks.creating') : t('tasks.create') }}
          </button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { X } from 'lucide-vue-next'
import { useTaskStore } from '@/stores/use-task-store'
import { useAuthStore } from '@/stores/use-auth-store'
import type { CreateTaskReq, TaskPriority } from '@/api/task'

import { VueDatePicker } from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'

const DRAFT_STORAGE_PREFIX = 'task_create_draft_v1'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'created'): void
}>()

interface TaskDraft {
  title: string
  description: string
  priority: TaskPriority
  start_at: string | null
  due_at: string | null
  is_open: boolean
}

interface CreateTaskForm {
  title: string
  description: string
  priority: TaskPriority
}

const taskStore = useTaskStore()
const authStore = useAuthStore()
const { t } = useI18n()

const isSubmitting = ref(false)
const isRestoring = ref(false)

const startDate = ref<Date | null>(null)
const dueDate = ref<Date | null>(null)

const priorityOptions: Array<{ key: string; value: TaskPriority; activeClass: string }> = [
  { key: 'tasks.priorityCritical', value: 'critical', activeClass: 'border-rose-500 text-rose-400 bg-rose-500/10' },
  { key: 'tasks.priorityImportant', value: 'important', activeClass: 'border-purple-500 text-purple-400 bg-purple-500/10' },
  { key: 'tasks.priorityUrgent', value: 'urgent', activeClass: 'border-amber-500 text-amber-400 bg-amber-500/10' },
  { key: 'tasks.priorityLow', value: 'low', activeClass: 'border-emerald-500 text-emerald-400 bg-emerald-500/10' },
  { key: 'tasks.priorityRoutine', value: 'routine', activeClass: 'border-blue-400 text-blue-400 bg-blue-400/10' },
]

const presetDates = computed(() => [
  { label: t('tasks.oneDay'), value: new Date(new Date().setDate(new Date().getDate() + 1)) },
  { label: t('tasks.threeDays'), value: new Date(new Date().setDate(new Date().getDate() + 3)) },
  { label: t('tasks.oneWeek'), value: new Date(new Date().setDate(new Date().getDate() + 7)) },
])

const initialForm = (): CreateTaskForm => ({
  title: '',
  description: '',
  priority: 'routine',
})

const form = ref<CreateTaskForm>(initialForm())

const draftStorageKey = computed(() => {
  const identity = authStore.user?.id || authStore.user?.email || 'anonymous'
  return `${DRAFT_STORAGE_PREFIX}:${identity}`
})

const hasUnsavedChanges = computed(() => {
  return (
    form.value.title.trim().length > 0 ||
    form.value.description.trim().length > 0 ||
    form.value.priority !== 'routine' ||
    startDate.value !== null ||
    dueDate.value !== null
  )
})

const canSubmit = computed(() => {
  return form.value.title.trim().length > 0 && dueDate.value !== null
})

watch(
  [form, startDate, dueDate, () => props.modelValue, draftStorageKey],
  () => {
    persistDraft()
  },
  { deep: true },
)

watch(
  draftStorageKey,
  () => {
    restoreDraft()
  },
  { immediate: true },
)

function setStartNow() {
  startDate.value = new Date()
}

function resetForm() {
  form.value = initialForm()
  startDate.value = null
  dueDate.value = null
}

function clearDraft() {
  localStorage.removeItem(draftStorageKey.value)
}

function persistDraft() {
  if (isRestoring.value) return

  if (!props.modelValue && !hasUnsavedChanges.value) {
    clearDraft()
    return
  }

  const draft: TaskDraft = {
    title: form.value.title,
    description: form.value.description,
    priority: form.value.priority,
    start_at: startDate.value ? startDate.value.toISOString() : null,
    due_at: dueDate.value ? dueDate.value.toISOString() : null,
    is_open: props.modelValue,
  }

  localStorage.setItem(draftStorageKey.value, JSON.stringify(draft))
}

function restoreDraft() {
  const raw = localStorage.getItem(draftStorageKey.value)
  if (!raw) return

  try {
    const parsed = JSON.parse(raw) as Partial<TaskDraft>
    isRestoring.value = true

    form.value = {
      title: parsed.title ?? '',
      description: parsed.description ?? '',
      priority: parsed.priority ?? 'routine',
    }
    startDate.value = parsed.start_at ? new Date(parsed.start_at) : null
    dueDate.value = parsed.due_at ? new Date(parsed.due_at) : null

    const hasDraftContent =
      form.value.title.trim().length > 0 ||
      form.value.description.trim().length > 0 ||
      form.value.priority !== 'routine' ||
      startDate.value !== null ||
      dueDate.value !== null

    if (parsed.is_open && hasDraftContent && !props.modelValue) {
      emit('update:modelValue', true)
    }
  } catch {
    clearDraft()
  } finally {
    isRestoring.value = false
  }
}

function closeDrawer() {
  emit('update:modelValue', false)
}

function requestClose() {
  if (hasUnsavedChanges.value) {
    const confirmClose = window.confirm(t('tasks.discardDraftConfirm'))
    if (!confirmClose) {
      return
    }
  }

  resetForm()
  clearDraft()
  closeDrawer()
}

async function submit() {
  if (!canSubmit.value || isSubmitting.value || !dueDate.value) return

  const startAt = startDate.value
  if (startAt && startAt.getTime() > dueDate.value.getTime()) {
    window.alert(t('tasks.timeRangeInvalid'))
    return
  }

  isSubmitting.value = true
  try {
    const payload: CreateTaskReq = {
      title: form.value.title.trim(),
      description: form.value.description,
      priority: form.value.priority,
      due_at: dueDate.value.toISOString(),
    }

    if (startAt) {
      payload.start_at = startAt.toISOString()
    }

    await taskStore.createTask(payload)
    emit('created')
    resetForm()
    clearDraft()
    closeDrawer()
  } catch (e) {
    console.error('Create task failed', e)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.vue-datepicker-wrapper {
  --dp-background-color: rgba(255, 255, 255, 0.05);
  --dp-text-color: rgba(255, 255, 255, 0.8);
  --dp-hover-color: rgba(255, 255, 255, 0.1);
  --dp-hover-text-color: #fff;
  --dp-hover-icon-color: #fff;
  --dp-primary-color: var(--neon);
  --dp-primary-text-color: #050a0f;
  --dp-secondary-color: #a9a9a9;
  --dp-border-color: rgba(255, 255, 255, 0.1);
  --dp-menu-border-color: rgba(255, 255, 255, 0.1);
  --dp-border-color-hover: var(--neon);
  --dp-disabled-color: rgba(255, 255, 255, 0.02);
  --dp-disabled-color-text: rgba(255, 255, 255, 0.2);
  --dp-scroll-bar-background: rgba(255, 255, 255, 0.1);
  --dp-scroll-bar-color: var(--neon);
  --dp-success-color: #00ffaa;
  --dp-icon-color: rgba(255, 255, 255, 0.5);
  --dp-danger-color: #ff3366;
}

:deep(.dp__input) {
  border-radius: 0.5rem;
  padding: 0.75rem 1rem 0.75rem 2.5rem;
  font-family: inherit;
  transition: all 0.3s ease;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.05);
  color: white;
}

:deep(.dp__input:focus) {
  box-shadow: 0 0 10px var(--neon-glow);
  border-color: var(--neon);
}

:deep(.dp__preset_ranges) {
  padding: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

:deep(.dp__preset_ranges span) {
  padding: 0.5rem;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: background 0.2s;
}

:deep(.dp__preset_ranges span:hover) {
  background: var(--dp-hover-color);
  color: var(--neon);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-enter-active,
.slide-leave-active {
  transition: transform 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.slide-enter-from,
.slide-leave-to {
  transform: translateX(100%);
}
</style>
