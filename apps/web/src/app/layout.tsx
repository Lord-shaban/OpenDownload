import type { Metadata } from "next";
import { GeistSans } from "geist/font/sans";
import { GeistMono } from "geist/font/mono";
import localFont from "next/font/local";
import { cookies } from "next/headers";
import { LocaleProvider } from "@/components/locale-provider";
import { direction, LOCALE_COOKIE, parseLocale, translate } from "@/lib/i18n";
import "./globals.css";
const arabicFont = localFont({
  src: "./fonts/NotoSansArabic.ttf",
  variable: "--font-arabic",
  display: "swap",
  weight: "100 900",
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
  const locale = parseLocale((await cookies()).get(LOCALE_COOKIE)?.value);
  return (
    <html
      lang={locale}
      dir={direction(locale)}
      className={`${GeistSans.variable} ${GeistMono.variable} ${arabicFont.variable}`}
      suppressHydrationWarning
    >
      <body>
        <LocaleProvider initialLocale={locale}>{children}</LocaleProvider>
      </body>
    </html>
  );
}
