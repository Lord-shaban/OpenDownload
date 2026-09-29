"use client";
import { useLocale } from "@/components/locale-provider";
export default function ErrorPage({ reset }: { reset: () => void }) {
  const { t } = useLocale();
  return (
    <main className="mx-auto max-w-lg p-12">
      <h1 className="text-2xl font-semibold">{t("interruptedWorkspace")}</h1>
      <p className="my-4 text-muted-foreground">{t("reconnectWorkspace")}</p>
      <button
        onClick={reset}
        className="rounded-lg bg-primary px-6 py-3 text-primary-foreground"
      >
        {t("tryAgain")}
      </button>
    </main>
  );
}
