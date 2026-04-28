<template>
  <Teleport to="body">
    <!-- Backdrop -->
    <Transition name="fade">
      <div 
        v-if="modelValue" 
        class="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 transition-opacity"
        @click="close"
      />
    </Transition>

    <!-- Drawer -->
    <Transition name="slide">
      <div 
        v-if="modelValue"
        class="fixed right-0 top-0 bottom-0 w-full max-w-2xl bg-[#0a1118] border-l border-white/10 z-50 flex flex-col shadow-2xl"
      >
        <!-- Header -->
        <div class="px-6 py-5 border-b border-white/10 flex items-center justify-between">
          <h2 class="text-xl font-medium text-white tracking-wide">
            {{ t('tasks.createTask') }}
          </h2>
          <button 
            class="p-2 text-white/50 hover:text-neon transition-colors duration-200"
            @click="close"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Form Body -->
        <div class="flex-1 overflow-y-auto px-6 py-6">
          <form
            class="space-y-8"
            @submit.prevent="submit"
          >
            <!-- Title -->
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

            <!-- Description -->
            <div class="space-y-2">
              <label class="block text-sm font-medium text-white/70">{{ t('tasks.taskDescription') }}</label>
              <textarea 
                v-model="form.description"
                rows="4"
                :placeholder="t('tasks.taskDescription')"
                class="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 text-white placeholder-white/30 focus:outline-none focus:border-neon focus:shadow-[0_0_10px_var(--neon-glow)] transition-all duration-300 resize-none"
              />
            </div>

            <!-- Priority Tags -->
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

            <!-- DDL Date Picker -->
            <div class="space-y-3">
              <label class="block text-sm font-medium text-white/70">
                {{ t('tasks.dueDate') }} (DDL) <span class="text-neon ml-1">*</span>
              </label>
              <div class="vue-datepicker-wrapper">
                <VueDatePicker 
                  v-model="ddlDate" 
                  dark
                  :preset-dates="presetDates"
                  :placeholder="t('tasks.selectDueDate')"
                  format="yyyy/MM/dd HH:mm"
                  :enable-time-picker="true"
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

        <!-- Footer -->
        <div class="px-6 py-5 border-t border-white/10 bg-[#0a1118] flex justify-end gap-3 mt-auto">
          <button 
            type="button"
            class="px-5 py-2.5 rounded-lg text-white/70 hover:bg-white/5 transition-colors duration-200"
            @click="close"
          >
            {{ t('tasks.cancelCreate') }}
          </button>
          <button 
            type="button"
            class="px-6 py-2.5 rounded-lg bg-neon text-[#050a0f] font-medium tracking-wide hover:shadow-[0_0_15px_var(--neon-glow)] transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="!form.title.trim() || !ddlDate || isSubmitting"
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
import type { CreateTaskReq, TaskPriority } from '@/api/task'

// Datepicker
import { VueDatePicker } from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'created'): void
}>()

const taskStore = useTaskStore()
const { t } = useI18n()
const isSubmitting = ref(false)

const ddlDate = ref<Date | null>(null)

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

const initialForm = (): CreateTaskReq => ({
  title: '',
  description: '',
  priority: 'routine',
  due_at: ''
})

const form = ref<CreateTaskReq>(initialForm())

watch(() => props.modelValue, (isOpen) => {
  if (isOpen) {
    form.value = initialForm()
    ddlDate.value = null
  }
})

function close() {
  emit('update:modelValue', false)
}

async function submit() {
  if (!form.value.title.trim() || !ddlDate.value || isSubmitting.value) return
  
  isSubmitting.value = true
  try {
    const payload = { ...form.value }
    if (ddlDate.value) {
      payload.due_at = ddlDate.value.toISOString()
    } else {
      delete payload.due_at
    }

    await taskStore.createTask(payload)
    emit('created')
    close()
  } catch (e) {
    console.error('Create task failed', e)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
/* Inject Neon specific styling onto vue-datepicker */
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
  background: rgba(255,255,255,0.05);
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
</style>

<style scoped>
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
