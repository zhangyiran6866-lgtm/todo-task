<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/65 px-5 backdrop-blur-sm"
      @click.self="handleCancel"
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
              {{ title }}
            </h3>
            <p class="mt-2 text-sm leading-6 text-white/65">
              {{ description }}
            </p>
          </div>
        </div>

        <div class="mt-6 grid grid-cols-2 gap-3">
          <button
            class="flex h-12 w-full items-center justify-center rounded-lg border border-white/15 px-4 text-sm font-medium text-white/75 transition-colors hover:border-white/35 hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
            type="button"
            :disabled="loading"
            @click="handleCancel"
          >
            {{ cancelText }}
          </button>
          <button
            class="flex h-12 w-full items-center justify-center rounded-lg bg-[var(--neon)] px-4 text-sm font-bold text-black transition-all hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
            type="button"
            :disabled="loading"
            @click="emit('confirm')"
          >
            {{ confirmText }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { TriangleAlert } from "lucide-vue-next";

interface Props {
  modelValue: boolean;
  title: string;
  description: string;
  confirmText: string;
  cancelText: string;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
});

const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
  (e: "confirm"): void;
}>();

function handleCancel() {
  if (props.loading) return;
  emit("update:modelValue", false);
}
</script>
