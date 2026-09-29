export const sources = [
  "YouTube",
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
export function bytes(value?: number) {
  if (!value || value < 0) return "Size varies";
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
export function expiryLabel(expiresAt: string, now: number) {
  const mins = Math.ceil((Date.parse(expiresAt) - now) / 60_000);
  return mins <= 0
    ? "Expired"
    : mins >= 60
      ? `${Math.ceil(mins / 60)}h remaining`
      : `${mins}m remaining`;
}
