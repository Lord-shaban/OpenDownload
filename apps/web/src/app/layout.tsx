import type { Metadata } from "next";
import { GeistSans } from "geist/font/sans";
import { GeistMono } from "geist/font/mono";
import localFont from "next/font/local";
import { cookies } from "next/headers";
import { LocaleProvider } from "@/components/locale-provider";
import { direction, LOCALE_COOKIE, parseLocale, translate } from "@/lib/i18n";
import "./globals.css";
const arabicFont = localFont({
  src: [
    {
      path: "./fonts/IBMPlexSansArabic-Regular.woff2",
      weight: "400",
      style: "normal",
    },
    {
      path: "./fonts/IBMPlexSansArabic-Medium.woff2",
      weight: "500",
      style: "normal",
    },
    {
      path: "./fonts/IBMPlexSansArabic-SemiBold.woff2",
      weight: "600",
      style: "normal",
    },
  ],
  variable: "--font-arabic",
  display: "swap",
  preload: false,
});
export async function generateMetadata(): Promise<Metadata> {
  const locale = parseLocale((await cookies()).get(LOCALE_COOKIE)?.value);
  return {
    title: `OpenDownload — ${translate(locale, "saveGood")}`,
    icons: { icon: "/icon.svg" },
    description: translate(locale, "intro"),
  };
}
export default async function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  const preferences = await cookies();
  const locale = parseLocale(preferences.get(LOCALE_COOKIE)?.value);
  const dark = preferences.get("od_theme")?.value === "dark";
  return (
    <html
      lang={locale}
      dir={direction(locale)}
      className={`${GeistSans.variable} ${GeistMono.variable} ${arabicFont.variable}${dark ? " dark" : ""}`}
      suppressHydrationWarning
    >
      <body>
        <LocaleProvider initialLocale={locale}>{children}</LocaleProvider>
      </body>
    </html>
  );
}
