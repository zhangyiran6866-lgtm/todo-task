import { defineStore } from "pinia";
import { ref } from "vue";

type Theme = "cyan" | "purple" | "green" | "pink";
type Language = "zh" | "en";

const THEME_KEY = "theme";
const LANGUAGE_KEY = "language";

function readTheme(): Theme {
  const theme = localStorage.getItem(THEME_KEY);
  if (
    theme === "cyan" ||
    theme === "purple" ||
    theme === "green" ||
    theme === "pink"
  ) {
    return theme;
  }
  return "cyan";
}

function readLanguage(): Language {
  const language = localStorage.getItem(LANGUAGE_KEY);
  if (language === "zh" || language === "en") {
    return language;
  }
  return "zh";
}

export const useThemeStore = defineStore("theme", () => {
  const theme = ref<Theme>(readTheme());
  const language = ref<Language>(readLanguage());

  function applyTheme(nextTheme: Theme) {
    document.documentElement.setAttribute("data-theme", nextTheme);
  }

  function setTheme(nextTheme: Theme) {
    theme.value = nextTheme;
    localStorage.setItem(THEME_KEY, nextTheme);
    applyTheme(nextTheme);
  }

  function setLanguage(nextLanguage: Language) {
    language.value = nextLanguage;
    localStorage.setItem(LANGUAGE_KEY, nextLanguage);
  }

  function initialize() {
    applyTheme(theme.value);
  }

  return {
    theme,
    language,
    setTheme,
    setLanguage,
    initialize,
  };
});
