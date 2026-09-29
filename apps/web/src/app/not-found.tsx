"use client";
import Link from "next/link";
import { useLocale } from "@/components/locale-provider";
export default function NotFound() {
  const { t } = useLocale();
  return (
    <main className="mx-auto max-w-lg p-12">
      <h1 className="text-2xl font-semibold">{t("pageMissing")}</h1>
      <Link className="mt-6 inline-block text-primary underline" href="/">
        {t("returnHome")}
      </Link>
    </main>
  );
}
