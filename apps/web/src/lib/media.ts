import { translate, type Locale } from "./i18n";
export const bidiControls = /[\u061c\u200e\u200f\u202a-\u202e\u2066-\u2069]/u;
export function displayText(value: string) {
  return value.replace(/[\u061c\u200e\u200f\u202a-\u202e\u2066-\u2069]/gu, "");
}
export const sources = [
  "Instagram",
  "TikTok",
  "X / Twitter",
  "Facebook",
  "Reddit",
  "Vimeo",
  "SoundCloud",
];
const platforms: Record<string, string> = {
  "youtube.com": "YouTube",
  "youtu.be": "YouTube",
  "instagram.com": "Instagram",
  "tiktok.com": "TikTok",
  "twitter.com": "X / Twitter",
  "x.com": "X / Twitter",
  "facebook.com": "Facebook",
  "fb.watch": "Facebook",
  "reddit.com": "Reddit",
  "redd.it": "Reddit",
  "vimeo.com": "Vimeo",
  "soundcloud.com": "SoundCloud",
};
export function detectSource(raw: string): { name: string; valid: boolean } {
  try {
    if (
      bidiControls.test(raw) ||
      /%(?:d8%9c|e2%80%(?:8e|8f|aa|ab|ac|ad|ae)|e2%81%(?:a6|a7|a8|a9))/i.test(
        raw,
      )
    )
      return { name: "", valid: false };
    const url = new URL(raw.trim());
    if (
      !["https:", "http:"].includes(url.protocol) ||
      url.username ||
      url.password
    )
      return { name: "Public links only", valid: false };
    const hostname = url.hostname.toLowerCase();
    for (const [domain, name] of Object.entries(platforms))
      if (hostname === domain || hostname.endsWith(`.${domain}`))
        return { name, valid: true };
    return { name: hostname, valid: hostname.includes(".") };
  } catch {
    return { name: "Waiting for a link", valid: false };
  }
}
export function bytes(value?: number, locale: Locale = "en") {
  if (!value || value < 0) return translate(locale, "sizeVaries");
  const unit = value >= 1024 ** 3 ? "GB" : value >= 1024 ** 2 ? "MB" : "KB";
  const amount =
    unit === "GB"
      ? value / 1024 ** 3
      : unit === "MB"
        ? value / 1024 ** 2
        : value / 1024;
  return `${amount.toFixed(amount < 10 ? 1 : 0)} ${unit}`;
}
export function duration(seconds: number) {
  const total = Math.max(0, Math.floor(seconds));
  return `${Math.floor(total / 60)}:${String(total % 60).padStart(2, "0")}`;
}
export function expiryLabel(
  expiresAt: string,
  now: number,
  locale: Locale = "en",
) {
  const mins = Math.ceil((Date.parse(expiresAt) - now) / 60_000);
  return mins <= 0
    ? translate(locale, "expired")
    : mins >= 60
      ? translate(locale, "hoursRemaining", { count: Math.ceil(mins / 60) })
      : translate(locale, "minutesRemaining", { count: mins });
}
