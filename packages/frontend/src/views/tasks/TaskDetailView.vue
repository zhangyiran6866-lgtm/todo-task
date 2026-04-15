<template>
  <div class="min-h-screen bg-[#050a0f] flex flex-col items-center">
    <!-- Navbar -->
    <header
      class="w-full h-16 border-b border-white/5 flex items-center px-6 bg-glass sticky top-0 z-20"
    >
      <button
        class="flex items-center gap-2 text-white/50 hover:text-neon transition-colors"
        @click="router.back()"
      >
        <ArrowLeft class="w-5 h-5" />
        <span class="text-sm font-medium">返回基站</span>
      </button>
    </header>

    <main class="w-full max-w-3xl px-6 py-12 flex-1 flex flex-col">
      <div
        v-if="isLoading"
        class="flex justify-center mt-20"
      >
        <div
          class="w-8 h-8 border-2 border-neon border-t-transparent rounded-full animate-spin"
        />
      </div>

      <div
        v-else-if="task"
        class="space-y-10 animate-fade-in flex-1 flex flex-col"
      >
        <!-- Header -->
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3 relative">
              <select
                v-model="task.status"
                class="absolute opacity-0 inset-0 cursor-pointer"
                @change="save('status', task.status)"
              >
                <option value="todo">
                  待处理
                </option>
                <option value="in_progress">
                  计算中
                </option>
                <option value="done">
                  已封存
                </option>
              </select>
              <div
                class="px-3 py-1 rounded-full border text-xs font-medium uppercase tracking-wider transition-colors pointer-events-none"
                :class="statusTagClass"
              >
                {{ statusText }}
              </div>
            </div>

            <div class="flex items-center gap-3 relative">
              <select
                v-model="task.priority"
                class="absolute opacity-0 inset-0 cursor-pointer w-full"
                @change="save('priority', task.priority)"
              >
                <option value="routine">
                  日常任务
                </option>
                <option value="critical">
                  重要紧急
                </option>
                <option value="important">
                  重要不紧急
                </option>
                <option value="urgent">
                  紧急不重要
                </option>
                <option value="low">
                  不重要也不紧急
                </option>
              </select>
              <div
                class="flex items-center gap-1.5 px-3 py-1 rounded border text-sm pointer-events-none transition-colors"
                :class="priorityTagClass"
              >
                <AlertCircle
                  v-if="
                    task.priority === 'critical' ||
                      task.priority === 'important'
                  "
                  class="w-4 h-4"
                />
                <ArrowUpCircle
                  v-else-if="task.priority === 'urgent'"
                  class="w-4 h-4"
                />
                <span class="capitalize">{{ priorityText }} 标签</span>
              </div>
            </div>
          </div>

          <!-- Title -->
          <input
            v-model="task.title"
            class="w-full bg-transparent text-4xl font-light text-white outline-none border-b border-transparent focus:border-white/20 pb-2 transition-colors placeholder-white/20"
            placeholder="输入进程代号..."
            @change="save('title', task.title)"
          >
        </div>

        <!-- Meta -->
        <div class="flex flex-wrap gap-6 text-sm text-white/50">
          <div class="relative flex items-center gap-2 group cursor-pointer">
            <Calendar class="w-4 h-4 group-hover:text-neon transition-colors" />
            <span class="group-hover:text-white transition-colors">
              截止日期: {{ formattedDate || "未设置" }}
            </span>
            <input
              v-model="editedDueDate"
              type="datetime-local"
              class="absolute opacity-0 inset-0 cursor-pointer"
              @change="saveDueDate"
            >
          </div>

          <div class="flex items-center gap-2">
            <Clock class="w-4 h-4" />
            <span>构造于:
              {{ new Date(task.created_at).toLocaleDateString() }}</span>
          </div>
        </div>

        <!-- Body -->
        <div class="flex-1">
          <textarea
            v-model="task.description"
            class="w-full h-full min-h-[200px] bg-transparent text-white/80 outline-none border border-transparent focus:border-white/10 rounded-xl p-4 transition-colors placeholder-white/20 resize-none leading-relaxed"
            placeholder="输入全息事件描述..."
            @blur="save('description', task.description)"
          />
        </div>

        <!-- Danger Zone -->
        <div class="pt-8 mt-auto border-t border-white/5 flex justify-end">
          <button
            class="flex items-center gap-2 px-6 py-2.5 rounded-lg border border-rose-500/30 text-rose-400 hover:bg-rose-500 hover:text-white transition-all duration-300"
            @click="confirmDelete"
          >
            <Trash2 class="w-4 h-4" />
            <span>销毁节点</span>
          </button>
        </div>
      </div>

      <div
        v-else
        class="text-white/40 text-center mt-20"
      >
        404: 数据节点丢失或访问权限被抑制。
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { taskApi, type Task, type UpdateTaskReq } from "@/api/task";
import { useTaskStore } from "@/stores/use-task-store";
import {
  ArrowLeft,
  Calendar,
  Clock,
  AlertCircle,
  ArrowUpCircle,
  Trash2,
} from "lucide-vue-next";

const route = useRoute();
const router = useRouter();
const taskStore = useTaskStore();

const taskId = route.params.id as string;
const task = ref<Task | null>(null);
const isLoading = ref(true);

const editedDueDate = ref("");

onMounted(async () => {
  try {
    // Attempt to load from store first for snappiness
    const found = taskStore.tasks.find((t) => t.id === taskId);
    if (found) {
      task.value = { ...found };
      setEditDueDate();
      isLoading.value = false;
    }

    // Always fetch latest
    const fetched = await taskApi.getTaskById(taskId);
    task.value = fetched;
    setEditDueDate();
  } catch (e) {
    console.error("Failed to load task", e);
  } finally {
    isLoading.value = false;
  }
});

function setEditDueDate() {
  if (task.value?.due_at) {
    // datetime-local valid format yyyy-MM-ddThh:mm
    const date = new Date(task.value.due_at);
    const tzoffset = new Date().getTimezoneOffset() * 60000; //offset in milliseconds
    const localISOTime = new Date(date.getTime() - tzoffset)
      .toISOString()
      .slice(0, 16);
    editedDueDate.value = localISOTime;
  }
}

const statusTagClass = computed(() => {
  if (!task.value) return "";
  switch (task.value.status) {
    case "done":
      return "bg-neon/10 text-neon border-neon";
    case "in_progress":
      return "bg-blue-500/10 text-blue-400 border-blue-500/50";
    default:
      return "bg-white/5 text-white/60 border-white/20";
  }
});

const statusText = computed(() => {
  if (!task.value) return "";
  switch (task.value.status) {
    case "done":
      return "已封存";
    case "in_progress":
      return "计算中";
    default:
      return "待处理";
  }
});

const priorityTagClass = computed(() => {
  if (!task.value) return "";
  switch (task.value.priority) {
    case "critical":
      return "text-rose-400 border-rose-500";
    case "important":
      return "text-purple-400 border-purple-500";
    case "urgent":
      return "text-amber-400 border-amber-500";
    case "low":
      return "text-emerald-400 border-emerald-500";
    case "routine":
      return "text-blue-400 border-blue-400";
    default:
      return "text-white/50 border-white/20";
  }
});

const priorityText = computed(() => {
  if (!task.value) return "";
  switch (task.value.priority) {
    case "critical":
      return "重要紧急";
    case "important":
      return "重要不紧急";
    case "urgent":
      return "紧急不重要";
    case "low":
      return "不重要也不紧急";
    case "routine":
      return "日常任务";
    default:
      return "未分类";
  }
});

const formattedDate = computed(() => {
  if (!task.value?.due_at) return "";
  return new Date(task.value.due_at).toLocaleDateString(undefined, {
    year: "numeric",
    month: "long",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
});

async function saveDueDate() {
  if (!task.value) return;
  let isoDate = "";
  if (editedDueDate.value) {
    isoDate = new Date(editedDueDate.value).toISOString();
  }
  task.value.due_at = isoDate || null;
  await save("due_at", isoDate || undefined);
}

async function save(
  field: keyof UpdateTaskReq,
  val: UpdateTaskReq[keyof UpdateTaskReq],
) {
  if (!task.value) return;
  try {
    await taskStore.updateTask(taskId, { [field]: val });
  } catch (e) {
    console.error("Update failed", e);
  }
}

async function confirmDelete() {
  if (confirm("系统警告：永久销毁该数据节点的过程不可逆！是否确认清理？")) {
    try {
      await taskStore.deleteTask(taskId);
      router.back();
    } catch (e) {
      console.error("Failed to delete", e);
    }
  }
}
</script>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.4s ease-out forwards;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
