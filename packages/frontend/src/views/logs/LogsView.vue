<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import {
  ArrowLeft,
  ChevronLeft,
  ChevronDown,
  ChevronRight,
  LoaderCircle,
  RefreshCw,
  ShieldAlert,
  TerminalSquare,
  X,
} from "lucide-vue-next";
import { VueDatePicker } from "@vuepic/vue-datepicker";
import "@vuepic/vue-datepicker/dist/main.css";

import { logApi, type GetLogsReq, type LogItem } from "@/api/log";

const router = useRouter();
const { t } = useI18n();

const channelFilter = ref("");
const levelFilter = ref("");
const moduleFilter = ref("");
const startAtFilter = ref<Date | null>(null);
const endAtFilter = ref<Date | null>(null);
const channelMenuRef = ref<HTMLElement | null>(null);
const levelMenuRef = ref<HTMLElement | null>(null);
const isChannelMenuOpen = ref(false);
const isLevelMenuOpen = ref(false);

const logs = ref<LogItem[]>([]);
const currentPage = ref(1);
const pageSize = ref(20);
const total = ref(0);
const totalPages = ref(0);
const hasNext = ref(false);
const hasPrev = ref(false);

const isLoading = ref(false);
const listErrorMessage = ref("");

const selectedLog = ref<LogItem | null>(null);
const isDetailOpen = ref(false);
const isDetailLoading = ref(false);
const detailErrorMessage = ref("");

const pageLabel = computed(() => {
  if (totalPages.value <= 0) {
    return `1 / 1`;
  }
  return `${currentPage.value} / ${totalPages.value}`;
});

const channelOptions = computed(() => [
  { label: t("logs.channelAll"), value: "" },
  { label: t("logs.channelApp"), value: "app" },
  { label: t("logs.channelError"), value: "error" },
  { label: t("logs.channelAudit"), value: "audit" },
] as const);

const levelOptions = computed(() => [
  { label: t("logs.levelAll"), value: "" },
  { label: t("logs.levelDebug"), value: "debug" },
  { label: t("logs.levelInfo"), value: "info" },
  { label: t("logs.levelWarn"), value: "warn" },
  { label: t("logs.levelError"), value: "error" },
] as const);

const activeChannelLabel = computed(() => {
  return (
    channelOptions.value.find((item) => item.value === channelFilter.value)?.label ||
    channelOptions.value[0].label
  );
});

const activeLevelLabel = computed(() => {
  return (
    levelOptions.value.find((item) => item.value === levelFilter.value)?.label ||
    levelOptions.value[0].label
  );
});

void fetchLogs(1);

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
  unlockBodyScroll();
});

watch(isDetailOpen, (opened) => {
  if (opened) {
    lockBodyScroll();
    return;
  }
  unlockBodyScroll();
});

function toIsoDateTime(value: Date | null): string | undefined {
  if (!value) return undefined;
  if (Number.isNaN(value.getTime())) return undefined;
  return value.toISOString();
}

function buildFilters(page: number): GetLogsReq {
  return {
    channel: (channelFilter.value || undefined) as GetLogsReq["channel"],
    level: (levelFilter.value || undefined) as GetLogsReq["level"],
    module: moduleFilter.value || undefined,
    start_at: toIsoDateTime(startAtFilter.value),
    end_at: toIsoDateTime(endAtFilter.value),
    page,
    page_size: pageSize.value,
  };
}

function formatTimestamp(value: string | undefined): string {
  if (!value) return "--";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "--";
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "short",
    timeStyle: "medium",
  }).format(date);
}

function getLevelClass(level: string | undefined): string {
  switch (level) {
    case "error":
      return "bg-rose-500/15 text-rose-300 border-rose-400/40";
    case "warn":
      return "bg-amber-500/15 text-amber-200 border-amber-300/40";
    case "debug":
      return "bg-violet-500/15 text-violet-200 border-violet-400/40";
    default:
      return "bg-cyan-500/15 text-cyan-200 border-cyan-400/40";
  }
}

function getChannelClass(channel: string | undefined): string {
  switch (channel) {
    case "error":
      return "bg-rose-500/15 text-rose-300 border-rose-400/40";
    case "audit":
      return "bg-emerald-500/15 text-emerald-300 border-emerald-400/40";
    default:
      return "bg-slate-500/15 text-slate-200 border-slate-300/40";
  }
}

async function fetchLogs(page: number) {
  isLoading.value = true;
  listErrorMessage.value = "";

  try {
    const data = await logApi.getLogs(buildFilters(page));
    logs.value = data.items || [];
    currentPage.value = data.pagination?.page || page;
    pageSize.value = data.pagination?.page_size || 20;
    total.value = data.pagination?.total || 0;
    totalPages.value = data.pagination?.total_pages || 0;
    hasNext.value = Boolean(data.pagination?.has_next);
    hasPrev.value = Boolean(data.pagination?.has_prev);
  } catch (error) {
    const fallbackMessage = t("logs.listFailed");
    listErrorMessage.value = error instanceof Error ? error.message : fallbackMessage;
  } finally {
    isLoading.value = false;
  }
}

function handleSearch() {
  void fetchLogs(1);
}

function handleReset() {
  channelFilter.value = "";
  levelFilter.value = "";
  moduleFilter.value = "";
  isChannelMenuOpen.value = false;
  isLevelMenuOpen.value = false;
  startAtFilter.value = null;
  endAtFilter.value = null;
  void fetchLogs(1);
}

function handleBack() {
  router.push("/tasks");
}

function handlePrevPage() {
  if (!hasPrev.value || isLoading.value) return;
  void fetchLogs(currentPage.value - 1);
}

function handleNextPage() {
  if (!hasNext.value || isLoading.value) return;
  void fetchLogs(currentPage.value + 1);
}

async function handleOpenDetail(item: LogItem) {
  isDetailOpen.value = true;
  selectedLog.value = item;
  detailErrorMessage.value = "";
  isDetailLoading.value = true;

  try {
    selectedLog.value = await logApi.getLogsId(item.id, {
      channel: item.channel as GetLogsReq["channel"],
    });
  } catch (error) {
    const fallbackMessage = t("logs.detailFailed");
    detailErrorMessage.value = error instanceof Error ? error.message : fallbackMessage;
  } finally {
    isDetailLoading.value = false;
  }
}

function closeDetail() {
  isDetailOpen.value = false;
}

function lockBodyScroll() {
  document.body.style.overflow = "hidden";
}

function unlockBodyScroll() {
  document.body.style.overflow = "";
}

function toggleChannelMenu() {
  isChannelMenuOpen.value = !isChannelMenuOpen.value;
  if (isChannelMenuOpen.value) {
    isLevelMenuOpen.value = false;
  }
}

function toggleLevelMenu() {
  isLevelMenuOpen.value = !isLevelMenuOpen.value;
  if (isLevelMenuOpen.value) {
    isChannelMenuOpen.value = false;
  }
}

function selectChannel(value: string) {
  channelFilter.value = value;
  isChannelMenuOpen.value = false;
  void fetchLogs(1);
}

function selectLevel(value: string) {
  levelFilter.value = value;
  isLevelMenuOpen.value = false;
  void fetchLogs(1);
}

function handleQuickFilter(channel: string, level: string, moduleName: string) {
  channelFilter.value = channel;
  levelFilter.value = level;
  moduleFilter.value = moduleName;
  void fetchLogs(1);
}

function handleClickOutside(event: MouseEvent) {
  if (!(event.target instanceof Node)) return;
  if (channelMenuRef.value && !channelMenuRef.value.contains(event.target)) {
    isChannelMenuOpen.value = false;
  }
  if (levelMenuRef.value && !levelMenuRef.value.contains(event.target)) {
    isLevelMenuOpen.value = false;
  }
}
</script>

<template>
  <main class="flex h-screen flex-col overflow-hidden bg-[#050a0f] text-[var(--text-primary)]">
    <header class="border-b border-white/10 bg-black/20 backdrop-blur-xl">
      <div class="mx-auto flex max-w-6xl items-center justify-between px-4 py-5 md:px-6">
        <div class="flex items-center gap-3">
          <button
            class="inline-flex h-10 w-10 items-center justify-center rounded-md border border-white/10 text-white/65 transition-colors hover:border-neon hover:text-neon"
            type="button"
            :aria-label="t('logs.backToTasks')"
            @click="handleBack"
          >
            <ArrowLeft class="h-5 w-5" />
          </button>
          <div class="flex items-center gap-2">
            <TerminalSquare class="h-5 w-5 text-neon" />
            <h1 class="text-xl font-semibold text-white md:text-2xl">
              {{ t("logs.title") }}
            </h1>
          </div>
        </div>

        <button
          class="inline-flex items-center gap-2 rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm text-white/80 transition-colors hover:border-neon/60 hover:text-neon"
          type="button"
          @click="fetchLogs(currentPage)"
        >
          <RefreshCw class="h-4 w-4" />
          {{ t("logs.refresh") }}
        </button>
      </div>
    </header>

    <section class="mx-auto flex w-full max-w-6xl flex-1 flex-col overflow-hidden px-4 py-6 md:px-6 md:py-8">
      <section class="rounded-2xl border border-[rgba(0,243,255,0.18)] bg-[rgba(5,10,15,0.45)] p-4 backdrop-blur-xl md:p-5">
        <div class="flex flex-col gap-3 md:flex-row md:items-end md:gap-2">
          <label class="space-y-1.5 md:w-[170px]">
            <span class="text-xs text-white/55">{{ t("logs.channel") }}</span>
            <div
              ref="channelMenuRef"
              class="relative"
            >
              <button
                class="filter-trigger"
                type="button"
                @click="toggleChannelMenu"
              >
                <span class="truncate">{{ activeChannelLabel }}</span>
                <ChevronDown
                  class="h-4 w-4 shrink-0 text-white/45 transition-transform"
                  :class="isChannelMenuOpen ? 'rotate-180 text-neon' : ''"
                />
              </button>
              <div
                v-if="isChannelMenuOpen"
                class="menu-panel"
              >
                <button
                  v-for="option in channelOptions"
                  :key="option.value"
                  class="menu-item"
                  :class="channelFilter === option.value ? 'menu-item-active' : ''"
                  type="button"
                  @click="selectChannel(option.value)"
                >
                  {{ option.label }}
                </button>
              </div>
            </div>
          </label>

          <label class="space-y-1.5 md:w-[170px]">
            <span class="text-xs text-white/55">{{ t("logs.level") }}</span>
            <div
              ref="levelMenuRef"
              class="relative"
            >
              <button
                class="filter-trigger"
                type="button"
                @click="toggleLevelMenu"
              >
                <span class="truncate">{{ activeLevelLabel }}</span>
                <ChevronDown
                  class="h-4 w-4 shrink-0 text-white/45 transition-transform"
                  :class="isLevelMenuOpen ? 'rotate-180 text-neon' : ''"
                />
              </button>
              <div
                v-if="isLevelMenuOpen"
                class="menu-panel"
              >
                <button
                  v-for="option in levelOptions"
                  :key="option.value"
                  class="menu-item"
                  :class="levelFilter === option.value ? 'menu-item-active' : ''"
                  type="button"
                  @click="selectLevel(option.value)"
                >
                  {{ option.label }}
                </button>
              </div>
            </div>
          </label>

          <label class="space-y-1.5 md:w-[240px]">
            <span class="text-xs text-white/55">{{ t("logs.startAt") }}</span>
            <VueDatePicker
              v-model="startAtFilter"
              :dark="true"
              :enable-time-picker="true"
              :clearable="true"
              :input-icon="false"
              :placeholder="t('logs.startAt')"
              auto-apply
              text-input
              class="filter-date-picker"
            />
          </label>

          <label class="space-y-1.5 md:w-[240px]">
            <span class="text-xs text-white/55">{{ t("logs.endAt") }}</span>
            <VueDatePicker
              v-model="endAtFilter"
              :dark="true"
              :enable-time-picker="true"
              :clearable="true"
              :input-icon="false"
              :placeholder="t('logs.endAt')"
              auto-apply
              text-input
              class="filter-date-picker"
            />
          </label>

          <div class="flex items-center gap-2 md:ml-auto md:pb-[2px]">
            <button
              class="rounded-lg border border-white/10 px-3 py-2 text-sm text-white/70 transition-colors hover:border-white/30 hover:text-white"
              type="button"
              @click="handleReset"
            >
              {{ t("common.clear") }}
            </button>
            <button
              class="rounded-lg border border-neon/70 bg-neon/15 px-3 py-2 text-sm text-neon transition-colors hover:bg-neon/20"
              type="button"
              @click="handleSearch"
            >
              {{ t("logs.search") }}
            </button>
          </div>
        </div>

      </section>

      <section class="mt-6 flex min-h-0 flex-1 flex-col">
        <div class="min-h-0 flex-1 space-y-3 overflow-y-auto pr-1">
        <div
          v-if="isLoading"
          class="flex min-h-56 items-center justify-center text-white/65"
        >
          <LoaderCircle class="mr-2 h-5 w-5 animate-spin text-neon" />
          {{ t("common.loading") }}
        </div>

        <div
          v-else-if="listErrorMessage"
          class="flex min-h-40 items-center justify-center rounded-2xl border border-rose-400/20 bg-rose-500/10 px-4 text-sm text-rose-200"
        >
          {{ listErrorMessage }}
        </div>

        <div
          v-else-if="logs.length === 0"
          class="flex min-h-52 flex-col items-center justify-center rounded-2xl border border-white/10 bg-white/5 text-white/45"
        >
          <ShieldAlert class="mb-2 h-10 w-10 text-white/35" />
          <p>{{ t("logs.empty") }}</p>
        </div>

        <button
          v-for="item in logs"
          :key="item.id"
          class="w-full rounded-2xl border border-white/10 bg-white/[0.03] p-4 text-left transition-all duration-200 hover:-translate-y-[1px] hover:border-neon/55 hover:shadow-[0_0_20px_rgba(0,243,255,0.12)]"
          type="button"
          @click="handleOpenDetail(item)"
        >
          <div class="flex flex-wrap items-start justify-between gap-2">
            <div class="flex flex-wrap items-center gap-2">
              <button
                class="rounded-full border px-2 py-0.5 text-xs uppercase tracking-wide"
                :class="getChannelClass(item.channel)"
                type="button"
                @click.stop="handleQuickFilter(item.channel, levelFilter, moduleFilter)"
              >
                {{ item.channel || "--" }}
              </button>
              <button
                class="rounded-full border px-2 py-0.5 text-xs uppercase tracking-wide"
                :class="getLevelClass(item.level)"
                type="button"
                @click.stop="handleQuickFilter(channelFilter, item.level, moduleFilter)"
              >
                {{ item.level || "--" }}
              </button>
              <button
                class="rounded-full border border-white/15 bg-white/5 px-2 py-0.5 text-xs text-white/70"
                type="button"
                @click.stop="handleQuickFilter(channelFilter, levelFilter, item.module)"
              >
                {{ item.module || "unknown" }}
              </button>
            </div>
            <span class="text-xs text-white/45">
              {{ formatTimestamp(item.timestamp) }}
            </span>
          </div>

          <p class="mt-3 line-clamp-2 text-sm text-white/90">
            {{ item.message || "--" }}
          </p>

          <div class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-white/55">
            <span>{{ t("logs.requestId") }}: {{ item.request_id || "--" }}</span>
            <span>{{ t("logs.path") }}: {{ item.path || item.route || "--" }}</span>
            <span>{{ t("logs.status") }}: {{ item.status_code ?? "--" }}</span>
            <span>{{ t("logs.latency") }}: {{ item.latency_ms ?? "--" }}ms</span>
          </div>
        </button>
        </div>

        <div
          v-if="logs.length > 0"
          class="mt-3 flex shrink-0 flex-wrap items-center justify-between gap-3 rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm"
        >
          <span class="text-white/65">
            {{ t("logs.totalCount", { total }) }}
          </span>
          <div class="flex items-center gap-2">
            <button
              class="inline-flex h-8 w-8 items-center justify-center rounded-md border border-white/15 text-white/70 transition-colors hover:border-neon hover:text-neon disabled:cursor-not-allowed disabled:opacity-35"
              type="button"
              :disabled="!hasPrev || isLoading"
              @click="handlePrevPage"
            >
              <ChevronLeft class="h-4 w-4" />
            </button>
            <span class="min-w-20 text-center text-white/80">
              {{ pageLabel }}
            </span>
            <button
              class="inline-flex h-8 w-8 items-center justify-center rounded-md border border-white/15 text-white/70 transition-colors hover:border-neon hover:text-neon disabled:cursor-not-allowed disabled:opacity-35"
              type="button"
              :disabled="!hasNext || isLoading"
              @click="handleNextPage"
            >
              <ChevronRight class="h-4 w-4" />
            </button>
          </div>
        </div>
      </section>
    </section>

    <div
      v-if="isDetailOpen"
      class="fixed inset-0 z-50 flex bg-black/55 backdrop-blur-[1px]"
      @click.self="closeDetail"
    >
      <aside
        class="ml-auto h-full w-full max-w-xl border-l border-neon/30 bg-[#071018]/95 p-4 md:p-5"
      >
        <div class="mb-4 flex items-center justify-between">
          <h3 class="text-lg font-semibold text-white">
            {{ t("logs.detailTitle") }}
          </h3>
          <button
            class="inline-flex h-9 w-9 items-center justify-center rounded-md border border-white/15 text-white/70 transition-colors hover:border-neon hover:text-neon"
            type="button"
            @click="closeDetail"
          >
            <X class="h-4 w-4" />
          </button>
        </div>

        <div
          v-if="isDetailLoading"
          class="flex h-32 items-center justify-center text-white/65"
        >
          <LoaderCircle class="mr-2 h-5 w-5 animate-spin text-neon" />
          {{ t("common.loading") }}
        </div>

        <div
          v-else-if="detailErrorMessage"
          class="rounded-xl border border-rose-400/20 bg-rose-500/10 px-4 py-3 text-sm text-rose-200"
        >
          {{ detailErrorMessage }}
        </div>

        <div
          v-else-if="selectedLog"
          class="space-y-4"
        >
          <div class="grid grid-cols-2 gap-3 text-sm">
            <div class="rounded-lg border border-white/10 bg-white/5 p-3">
              <p class="mb-1 text-xs text-white/45">{{ t("logs.channel") }}</p>
              <p class="text-white/90">{{ selectedLog.channel || "--" }}</p>
            </div>
            <div class="rounded-lg border border-white/10 bg-white/5 p-3">
              <p class="mb-1 text-xs text-white/45">{{ t("logs.level") }}</p>
              <p class="text-white/90">{{ selectedLog.level || "--" }}</p>
            </div>
            <div class="rounded-lg border border-white/10 bg-white/5 p-3">
              <p class="mb-1 text-xs text-white/45">{{ t("logs.module") }}</p>
              <p class="text-white/90">{{ selectedLog.module || "--" }}</p>
            </div>
            <div class="rounded-lg border border-white/10 bg-white/5 p-3">
              <p class="mb-1 text-xs text-white/45">{{ t("logs.action") }}</p>
              <p class="text-white/90">{{ selectedLog.action || "--" }}</p>
            </div>
          </div>

          <div class="rounded-lg border border-white/10 bg-white/5 p-3 text-sm">
            <p class="mb-1 text-xs text-white/45">{{ t("logs.timestamp") }}</p>
            <p class="text-white/90">{{ formatTimestamp(selectedLog.timestamp) }}</p>
          </div>

          <div class="rounded-lg border border-white/10 bg-white/5 p-3 text-sm">
            <p class="mb-1 text-xs text-white/45">{{ t("logs.message") }}</p>
            <p class="break-words text-white/90">{{ selectedLog.message || "--" }}</p>
          </div>

          <div class="rounded-lg border border-white/10 bg-white/5 p-3 text-xs text-white/60">
            <p>{{ t("logs.requestId") }}: {{ selectedLog.request_id || "--" }}</p>
            <p class="mt-1">{{ t("logs.userId") }}: {{ selectedLog.user_id || "--" }}</p>
            <p class="mt-1">{{ t("logs.ip") }}: {{ selectedLog.client_ip || "--" }}</p>
            <p class="mt-1">{{ t("logs.method") }}: {{ selectedLog.method || "--" }}</p>
            <p class="mt-1">{{ t("logs.path") }}: {{ selectedLog.path || selectedLog.route || "--" }}</p>
            <p class="mt-1">{{ t("logs.status") }}: {{ selectedLog.status_code ?? "--" }}</p>
            <p class="mt-1">{{ t("logs.latency") }}: {{ selectedLog.latency_ms ?? "--" }}ms</p>
          </div>

          <div class="rounded-lg border border-white/10 bg-[#050a0f] p-3">
            <p class="mb-2 text-xs text-white/45">{{ t("logs.raw") }}</p>
            <pre class="max-h-64 overflow-auto rounded-md border border-white/10 bg-black/45 p-2.5 text-xs text-white/75">{{ JSON.stringify(selectedLog.raw || {}, null, 2) }}</pre>
          </div>
        </div>
      </aside>
    </div>
  </main>
</template>

<style lang="less" scoped>
.filter-trigger {
  width: 100%;
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  border-radius: 0.6rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: linear-gradient(
    145deg,
    rgba(255, 255, 255, 0.08),
    rgba(255, 255, 255, 0.03)
  );
  padding: 0.55rem 0.7rem;
  color: rgba(255, 255, 255, 0.9);
  font-size: 0.875rem;
  outline: none;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.filter-trigger:hover {
  border-color: rgba(255, 255, 255, 0.28);
}

.filter-trigger:focus-visible {
  border-color: rgba(0, 243, 255, 0.75);
  box-shadow: 0 0 0 1px rgba(0, 243, 255, 0.2);
}

.menu-panel {
  position: absolute;
  z-index: 40;
  top: calc(100% + 0.35rem);
  left: 0;
  width: 100%;
  border-radius: 0.7rem;
  border: 1px solid rgba(0, 243, 255, 0.22);
  background: rgba(7, 16, 24, 0.96);
  padding: 0.25rem;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.35);
  backdrop-filter: blur(10px);
}

.menu-item {
  width: 100%;
  border-radius: 0.5rem;
  padding: 0.5rem 0.55rem;
  text-align: left;
  font-size: 0.82rem;
  color: rgba(255, 255, 255, 0.76);
  transition: background-color 0.16s ease, color 0.16s ease;
}

.menu-item:hover {
  background: rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.95);
}

.menu-item-active {
  background: rgba(0, 243, 255, 0.14);
  color: var(--neon);
}

:deep(.filter-date-picker .dp__theme_dark) {
  --dp-background-color: rgba(7, 16, 24, 0.96);
  --dp-text-color: rgba(231, 243, 255, 0.92);
  --dp-hover-color: rgba(0, 243, 255, 0.18);
  --dp-hover-text-color: rgba(255, 255, 255, 0.94);
  --dp-primary-color: rgba(0, 243, 255, 0.65);
  --dp-primary-text-color: #051017;
  --dp-border-color: rgba(255, 255, 255, 0.16);
  --dp-menu-border-color: rgba(0, 243, 255, 0.3);
  --dp-border-radius: 10px;
}

:deep(.filter-date-picker .dp__input_wrap) {
  border-radius: 0.6rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: linear-gradient(
    145deg,
    rgba(255, 255, 255, 0.08),
    rgba(255, 255, 255, 0.03)
  );
}

:deep(.filter-date-picker .dp__input) {
  height: 40px;
  border: none;
  background: transparent;
  color: rgba(255, 255, 255, 0.9);
  font-size: 0.875rem;
  padding-left: 0.75rem !important;
  padding-right: 2.35rem !important;
}

:deep(.filter-date-picker .dp__input::placeholder) {
  color: rgba(255, 255, 255, 0.42);
}

:deep(.filter-date-picker .dp__input_icon) {
  display: none !important;
}

:deep(.filter-date-picker .dp__input_icon_pad) {
  padding-inline-start: 0.75rem !important;
}

:deep(.filter-date-picker .dp__clear_icon) {
  right: 0.6rem;
  color: rgba(255, 255, 255, 0.45);
}

:deep(.filter-date-picker .dp__input_focus) {
  box-shadow: 0 0 0 1px rgba(0, 243, 255, 0.24);
}

:deep(.filter-date-picker .dp__menu) {
  backdrop-filter: blur(10px);
}

:deep(.filter-date-picker .dp__today) {
  border-color: rgba(0, 243, 255, 0.65);
}
</style>
