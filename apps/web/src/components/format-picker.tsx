"use client";
import Image from "next/image";
import { CheckIcon as Check } from "@phosphor-icons/react/dist/csr/Check";
import { FilmStripIcon as Film } from "@phosphor-icons/react/dist/csr/FilmStrip";
import { HeadphonesIcon as Headphones } from "@phosphor-icons/react/dist/csr/Headphones";
import { ImagesIcon as ImageIcon } from "@phosphor-icons/react/dist/csr/Images";
import { SubtitlesIcon as Subtitles } from "@phosphor-icons/react/dist/csr/Subtitles";
import { DownloadSimpleIcon as ArrowDownToLine } from "@phosphor-icons/react/dist/csr/DownloadSimple";
import { ClockIcon as Clock3 } from "@phosphor-icons/react/dist/csr/Clock";
import { useLocale } from "@/components/locale-provider";
import { Button } from "@/components/ui/button";
import { type Analysis, type Option } from "@/lib/api";
import { displayText, bytes, duration } from "@/lib/media";
import { mediaLabel, mediaDetail } from "@/lib/localized-media";
export const kinds = [
  { id: "video", label: "video", icon: Film },
  { id: "audio", label: "audio", icon: Headphones },
  { id: "image", label: "images", icon: ImageIcon },
  { id: "subtitle", label: "subtitles", icon: Subtitles },
] as const;
export function FormatPicker({
  analysis,
  selected,
  onSelect,
  busy,
  onQueue,
  fixture,
  maxBytes,
}: {
  analysis: Analysis;
  selected: Option | null;
  onSelect: (option: Option) => void;
  busy: boolean;
  onQueue: () => void;
  fixture: boolean;
  maxBytes: number;
}) {
  const { locale, t } = useLocale();
  const formatBytes = (value?: number) => bytes(value, locale);
  const available = kinds.filter((kind) =>
    analysis.options.some((option) => option.kind === kind.id),
  );
  const current = selected?.kind || available[0]?.id;
  const PreviewIcon = available[0]?.icon || Film;
  return (
    <section
      aria-label={t("availableMedia")}
      className="panel-enter glass-panel format-panel"
    >
      <div className="media-summary">
        <div className="relative flex size-14 shrink-0 items-center justify-center overflow-hidden rounded-xl bg-secondary sm:size-16">
          {fixture || !analysis.hasThumbnail ? (
            <PreviewIcon
              weight="duotone"
              size={24}
              className="text-primary"
              aria-hidden="true"
            />
          ) : (
            <Image
              src={`/api/v1/analyses/${analysis.id}/thumbnail`}
              alt=""
              fill
              unoptimized
              className="object-cover"
              onError={(event) => {
                event.currentTarget.style.display = "none";
              }}
            />
          )}
        </div>
        <div className="min-w-0 flex-1">
          <div className="mb-2 flex items-center gap-2 text-[11px] font-medium uppercase tracking-[.12em] text-primary">
            <span className="size-1.5 rounded-full bg-primary" />
            {fixture ? (
              t("testFixture")
            ) : (
              <bdi>{displayText(analysis.platform)}</bdi>
            )}
          </div>
          <h2 className="text-base leading-snug font-medium tracking-tight">
            <bdi>{displayText(analysis.title)}</bdi>
          </h2>
          <p className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
            <bdi>{displayText(analysis.creator)}</bdi>
            {analysis.duration > 0 ? (
              <span className="inline-flex items-center gap-1.5">
                <Clock3 size={12} aria-hidden="true" />
                {duration(analysis.duration)}
              </span>
            ) : null}
          </p>
        </div>
      </div>
      <div className="p-5 sm:p-6">
        <div
          className="media-tabs mb-5 flex flex-wrap gap-1 p-1"
          role="group"
          aria-label={t("mediaType")}
        >
          {available.map((kind) => (
            <button
              type="button"
              key={kind.id}
              onClick={() => {
                const first = analysis.options.find(
                  (option) => option.kind === kind.id,
                );
                if (first) onSelect(first);
              }}
              aria-pressed={current === kind.id}
              className={`flex min-h-11 flex-1 items-center justify-center gap-2 rounded-[10px] px-3 text-sm font-medium ${current === kind.id ? "bg-secondary text-secondary-foreground" : "text-muted-foreground hover:text-foreground"}`}
            >
              <kind.icon weight="duotone" size={16} aria-hidden="true" />
              {t(kind.label)}
            </button>
          ))}
        </div>
        <fieldset>
          <legend className="mb-3 flex w-full items-center justify-between text-xs font-medium text-muted-foreground">
            <span>
              {current === "video"
                ? t("quality")
                : current === "audio"
                  ? t("audioFormat")
                  : current === "subtitle"
                    ? t("captionLanguage")
                    : t("imageFormat")}
            </span>
          </legend>
          <div className="grid gap-2 sm:grid-cols-2">
            {analysis.options
              .filter((option) => option.kind === current)
              .map((option) => {
                const checked = selected?.id === option.id;
                const tooLarge = (option.bytes || 0) > maxBytes;
                return (
                  <label
                    key={option.id}
                    className={`format-option relative flex min-h-[64px] cursor-pointer items-center gap-3 rounded-xl border p-3 transition-colors has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-offset-2 has-[:focus-visible]:outline-ring ${checked ? "border-primary bg-secondary" : "hover:border-primary/40"} ${tooLarge ? "opacity-50" : ""}`}
                  >
                    <input
                      className="sr-only"
                      type="radio"
                      name="format"
                      value={option.id}
                      checked={checked}
                      onChange={() => onSelect(option)}
                      disabled={tooLarge || busy}
                    />
                    <span className="min-w-0 flex-1">
                      <span className="flex flex-wrap items-center gap-2 text-sm font-semibold">
                        <bdi>{mediaLabel(option.label, locale)}</bdi>
                      </span>
                      <span className="sr-only">
                        <bdi>{mediaDetail(option.detail, locale)}</bdi>
                      </span>
                      <span className="mt-2 block font-mono text-xs text-muted-foreground">
                        <bdi>{option.extension.toUpperCase()}</bdi> ·{" "}
                        {tooLarge
                          ? t("exceedsLimit")
                          : formatBytes(option.bytes)}
                      </span>
                    </span>
                    <span
                      className={`flex size-5 shrink-0 items-center justify-center rounded-full border ${checked ? "border-primary bg-primary text-primary-foreground" : "border-border"}`}
                    >
                      {checked ? (
                        <Check
                          size={12}
                          className="selection-check"
                          aria-hidden="true"
                        />
                      ) : null}
                    </span>
                  </label>
                );
              })}
          </div>
        </fieldset>
      </div>
      <div className="format-footer">
        {fixture ? (
          <p className="max-w-[260px] text-xs leading-relaxed text-muted-foreground">
            {t("testFileExplanation")}
          </p>
        ) : null}
        <Button
          className="min-h-11 gap-2 px-5"
          onClick={onQueue}
          disabled={busy || !selected || (selected.bytes || 0) > maxBytes}
        >
          <ArrowDownToLine size={16} aria-hidden="true" />
          {busy ? t("addingQueue") : t("queueDownload")}
        </Button>
      </div>
    </section>
  );
}
