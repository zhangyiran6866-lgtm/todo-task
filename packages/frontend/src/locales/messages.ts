import en from "./en";
import zh from "./zh";

export const messages = {
  zh,
  en,
} as const;

export type AppLanguage = keyof typeof messages;
