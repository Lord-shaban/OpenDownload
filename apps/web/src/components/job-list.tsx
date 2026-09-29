"use client";
import {
  Check,
  CircleAlert,
  Clock3,
  Download,
  Film,
  Headphones,
  ImageIcon,
  LoaderCircle,
  RotateCcw,
  Subtitles,
  Trash2,
  X,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { type Job } from "@/lib/api";
import { bytes, expiryLabel } from "@/lib/media";
const mediaIcons: Record<string, typeof Film> = {
  video: Film,
  audio: Headphones,
  image: ImageIcon,
  subtitle: Subtitles,
};
export function JobList({
  jobs,
  busy,
  onAction,
  fixture,
  now,
}: {
  jobs: Job[];
  busy: string;
  onAction: (id: string, action: "cancel" | "retry" | "delete") => void;
  fixture: boolean;
  now: number;
}) {
  if (jobs.length === 0)
    return (
      <div className="rounded-2xl border border-dashed bg-card/50 px-6 py-14 text-center">
        <span className="mx-auto mb-4 flex size-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
          <Download size={22} aria-hidden="true" />
        </span>
        <h2 className="text-sm font-medium">
          A little space for what you save.
        </h2>
        <p className="mx-auto mt-2 max-w-xs text-xs leading-6 text-muted-foreground">
          Your queued and completed downloads will appear here. Start with a
          public link above.
        </p>
      </div>
    );
  return (
    <ul
      className="divide-y overflow-hidden rounded-2xl border bg-card"
      aria-label="Download jobs"
    >
      {jobs.map((job) => {
        const Icon = mediaIcons[job.kind] || Film;
        const active = job.state === "queued" || job.state === "processing";
        const expired = Date.parse(job.expiresAt) <= now;
        const stateText =
          job.state === "processing"
            ? job.phase === "processing"
              ? "Preparing file"
              : job.phase === "analyzing"
                ? "Checking source"
                : "Downloading"
            : job.state === "complete"
              ? "Ready to save"
              : job.state === "failed"
                ? "Couldn’t finish"
                : job.state === "canceled"
                  ? "Canceled"
                  : "In queue";
        return (
          <li key={job.id} className="p-4 sm:p-5">
            <div className="flex items-start gap-3 sm:gap-4">
              <span
                className={`flex size-11 shrink-0 items-center justify-center rounded-xl ${job.state === "complete" ? "bg-secondary text-primary" : "bg-muted text-muted-foreground"}`}
              >
                <Icon size={20} aria-hidden="true" />
              </span>
              <div className="min-w-0 flex-1">
                <h3 className="truncate text-sm font-medium" title={job.title}>
                  {job.title}
                </h3>
                <div className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-[11px] text-muted-foreground">
                  <span>{fixture ? "Test fixture" : job.platform}</span>
                  <span aria-hidden="true">·</span>
                  <span>
                    {job.label} / {job.extension.toUpperCase()}
                  </span>
                  {job.files[0] ? (
                    <>
                      <span aria-hidden="true">·</span>
                      <span>
                        {bytes(
                          job.files.reduce(
                            (total, file) => total + file.bytes,
                            0,
                          ),
                        )}
                      </span>
                    </>
                  ) : null}
                </div>
                <div
                  className={`mt-2 flex flex-wrap items-center gap-1.5 text-xs ${job.state === "failed" ? "text-destructive" : job.state === "complete" ? "text-primary" : "text-muted-foreground"}`}
                >
                  {job.state === "complete" ? (
                    <Check size={13} aria-hidden="true" />
                  ) : job.state === "failed" ? (
                    <CircleAlert size={13} aria-hidden="true" />
                  ) : job.state === "queued" ? (
                    <Clock3 size={13} aria-hidden="true" />
                  ) : job.state === "processing" ? (
                    <LoaderCircle
                      size={13}
                      className="spinner"
                      aria-hidden="true"
                    />
                  ) : null}
                  <span>{expired ? "Expired" : stateText}</span>
                  {job.state === "complete" && !expired ? (
                    <span className="ml-2 text-[10px] text-muted-foreground">
                      {expiryLabel(job.expiresAt, now)}
                    </span>
                  ) : null}
                  {job.state === "processing" && job.percent >= 0 ? (
                    <span className="ml-auto font-mono">
                      {Math.round(job.percent)}%
                    </span>
                  ) : null}
                </div>
                {job.state === "processing" ? (
                  <div
                    role="progressbar"
                    aria-label={`${stateText}: ${job.title}`}
                    aria-valuemin={0}
                    aria-valuemax={100}
                    aria-valuenow={
                      job.percent >= 0 ? Math.round(job.percent) : undefined
                    }
                    className="mt-3 h-1 overflow-hidden rounded-full bg-muted"
                  >
                    <div
                      className="progress-value h-full origin-left bg-primary"
                      style={{
                        transform: `scaleX(${job.percent >= 0 ? Math.min(1, job.percent / 100) : 0.12})`,
                      }}
                    />
                  </div>
                ) : null}
                {job.error ? (
                  <p className="mt-2 text-xs leading-5 text-destructive">
                    {job.error}
                  </p>
                ) : null}
              </div>
              <div className="flex shrink-0 items-center gap-1">
                {active ? (
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={`Cancel ${job.title}`}
                    disabled={busy === job.id}
                    onClick={() => onAction(job.id, "cancel")}
                  >
                    <X size={16} />
                  </Button>
                ) : job.state === "failed" || job.state === "canceled" ? (
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={`Retry ${job.title}`}
                    disabled={busy === job.id || expired}
                    onClick={() => onAction(job.id, "retry")}
                  >
                    <RotateCcw size={16} />
                  </Button>
                ) : null}
                {!active ? (
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={`Delete ${job.title} and its files`}
                    disabled={busy === job.id}
                    onClick={() => onAction(job.id, "delete")}
                  >
                    <Trash2 size={15} />
                  </Button>
                ) : null}
              </div>
            </div>
            {job.state === "complete" && !expired ? (
              <div className="mt-4 flex flex-wrap gap-2 sm:ml-[60px]">
                {job.files.map((file, index) => (
                  <Button
                    asChild
                    key={file.id}
                    variant="secondary"
                    size="sm"
                    className="min-h-10 gap-2"
                  >
                    <a
                      href={`/api/v1/jobs/${job.id}/files/${file.id}`}
                      download
                    >
                      <Download size={14} aria-hidden="true" />
                      {fixture
                        ? "Save test file"
                        : job.files.length > 1
                          ? `Save file ${index + 1}`
                          : "Save file"}
                    </a>
                  </Button>
                ))}
              </div>
            ) : null}
          </li>
        );
      })}
    </ul>
  );
}
