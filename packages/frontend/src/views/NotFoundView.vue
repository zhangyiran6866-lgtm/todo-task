<script setup lang="ts">
import { useI18n } from "vue-i18n";

const { t } = useI18n();
</script>

<template>
  <main class="min-h-screen flex items-center justify-center bg-[var(--bg-dark)]">
    <div class="text-center">
      <h1 class="text-9xl font-bold text-neon">
        404
      </h1>
      <p class="text-[var(--text-secondary)] mt-4">
        {{ t("notFound.title") }}
      </p>
      <router-link
        to="/"
        class="mt-6 inline-block text-neon hover:shadow-neon transition-all"
      >
        ← {{ t("notFound.backHome") }}
      </router-link>
    </div>
  </main>
</template>
