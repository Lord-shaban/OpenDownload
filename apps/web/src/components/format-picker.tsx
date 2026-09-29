"use client";
import Image from "next/image";
import {
  Check,
  Film,
  Headphones,
  ImageIcon,
  Subtitles,
  ArrowDownToLine,
  Clock3,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { type Analysis, type Option } from "@/lib/api";
import { bytes, duration } from "@/lib/media";
export const kinds = [
  { id: "video", label: "Video", icon: Film },
  { id: "audio", label: "Audio", icon: Headphones },
  { id: "image", label: "Images", icon: ImageIcon },
  { id: "subtitle", label: "Subtitles", icon: Subtitles },
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
  const available = kinds.filter((kind) =>
    analysis.options.some((option) => option.kind === kind.id),
  );
  const current = selected?.kind || available[0]?.id;
  return (
    <section
      aria-label="Available media"
      className="panel-enter overflow-hidden rounded-2xl border bg-card"
    >
      <div className="flex gap-5 border-b p-5 sm:p-6">
        <div className="relative flex size-20 shrink-0 items-center justify-center overflow-hidden rounded-xl bg-secondary sm:size-24">
          {fixture || !analysis.hasThumbnail ? (
            <Film size={32} className="text-primary" aria-hidden="true" />
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
            {fixture ? "Test fixture" : analysis.platform}
          </div>
          <h2 className="text-lg leading-snug font-medium tracking-tight sm:text-xl">
            {analysis.title}
          </h2>
          <p className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
            {analysis.creator}
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
          className="mb-5 flex flex-wrap gap-1 rounded-lg bg-muted p-1"
          role="group"
          aria-label="Media type"
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
              className={`flex min-h-11 flex-1 items-center justify-center gap-2 rounded-md px-3 text-sm font-medium ${current === kind.id ? "bg-card text-foreground shadow-xs" : "text-muted-foreground hover:text-foreground"}`}
            >
              <kind.icon size={16} aria-hidden="true" />
              {kind.label}
            </button>
          ))}
        </div>
        <fieldset>
          <legend className="mb-3 flex w-full items-center justify-between text-xs font-medium text-muted-foreground">
            <span>
              {current === "video"
                ? "AVAILABLE QUALITY"
                : current === "audio"
                  ? "AUDIO FORMAT"
                  : current === "subtitle"
                    ? "CAPTION LANGUAGE"
                    : "IMAGE FORMAT"}
            </span>
            <span className="font-normal">From the source</span>
          </legend>
          <div className="grid gap-2 sm:grid-cols-2">
            {analysis.options
              .filter((option) => option.kind === current)
              .map((option, index) => {
                const checked = selected?.id === option.id;
                const tooLarge = (option.bytes || 0) > maxBytes;
                return (
                  <label
                    key={option.id}
                    className={`relative flex min-h-[78px] cursor-pointer items-center gap-3 rounded-xl border p-4 transition-colors has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-offset-2 has-[:focus-visible]:outline-ring ${checked ? "border-primary bg-secondary" : "hover:border-primary/40"} ${tooLarge ? "opacity-50" : ""}`}
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
                        {option.label}
                        {index === 0 && current === "video" ? (
                          <span className="rounded bg-primary/10 px-1.5 py-0.5 text-[9px] font-medium tracking-wide text-primary">
                            BEST AVAILABLE
                          </span>
                        ) : null}
                      </span>
                      <span className="mt-1 block text-[11px] text-muted-foreground">
                        {option.detail}
                      </span>
                      <span className="mt-1 block font-mono text-[10px] text-muted-foreground">
                        {tooLarge
                          ? "Exceeds instance limit"
                          : bytes(option.bytes)}
                      </span>
                    </span>
                    <span
                      className={`flex size-5 shrink-0 items-center justify-center rounded-full border ${checked ? "border-primary bg-primary text-primary-foreground" : "border-border"}`}
                    >
                      {checked ? <Check size={12} aria-hidden="true" /> : null}
                    </span>
                  </label>
                );
              })}
          </div>
        </fieldset>
      </div>
      <div className="flex flex-wrap items-center justify-between gap-4 border-t bg-muted/30 px-5 py-4 sm:px-6">
        <p className="max-w-[260px] text-xs leading-relaxed text-muted-foreground">
          {fixture
            ? "A test file verifies this flow. No real media is extracted."
            : "Saved temporarily on your instance. Original watermarks are preserved."}
        </p>
        <Button
          className="min-h-11 gap-2 px-5"
          onClick={onQueue}
          disabled={busy || !selected || (selected.bytes || 0) > maxBytes}
        >
          <ArrowDownToLine size={16} aria-hidden="true" />
          {busy ? "Adding to queue…" : "Queue download"}
        </Button>
      </div>
    </section>
  );
}
