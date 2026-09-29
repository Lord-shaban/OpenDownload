"use client";
import { createContext, useContext, useState, type ReactNode } from "react";
import {
  direction,
  LOCALE_COOKIE,
  translate,
  type Locale,
  type MessageKey,
} from "@/lib/i18n";
const LocaleContext = createContext<{
  locale: Locale;
  setLocale: (locale: Locale) => void;
} | null>(null);
export function LocaleProvider({
  initialLocale,
  children,
}: {
  initialLocale: Locale;
  children: ReactNode;
}) {
  const [locale, updateLocale] = useState(initialLocale);
  function setLocale(next: Locale) {
    updateLocale(next);
    document.documentElement.lang = next;
    document.documentElement.dir = direction(next);
    document.title = `OpenDownload — ${translate(next, "saveGood")}`;
    // Language preference only; no URL, job or credential is stored.
    document.cookie = `${LOCALE_COOKIE}=${next}; Path=/; Max-Age=31536000; SameSite=Lax${location.protocol === "https:" ? "; Secure" : ""}`;
  }
  return (
    <LocaleContext.Provider value={{ locale, setLocale }}>
      {children}
    </LocaleContext.Provider>
  );
}
export function useLocale() {
  const context = useContext(LocaleContext);
  if (!context) throw new Error("LocaleProvider is required.");
  return {
    ...context,
    t: (key: MessageKey, values?: Record<string, string | number>) =>
      translate(context.locale, key, values),
  };
}
