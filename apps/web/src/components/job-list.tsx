"use client";
import { CheckIcon as Check } from "@phosphor-icons/react/dist/csr/Check";
import { WarningCircleIcon as CircleAlert } from "@phosphor-icons/react/dist/csr/WarningCircle";
import { ClockIcon as Clock3 } from "@phosphor-icons/react/dist/csr/Clock";
import { DownloadSimpleIcon as Download } from "@phosphor-icons/react/dist/csr/DownloadSimple";
import { FilmStripIcon as Film } from "@phosphor-icons/react/dist/csr/FilmStrip";
import { HeadphonesIcon as Headphones } from "@phosphor-icons/react/dist/csr/Headphones";
import { ImagesIcon as ImageIcon } from "@phosphor-icons/react/dist/csr/Images";
import { CircleNotchIcon as LoaderCircle } from "@phosphor-icons/react/dist/csr/CircleNotch";
import { ArrowCounterClockwiseIcon as RotateCcw } from "@phosphor-icons/react/dist/csr/ArrowCounterClockwise";
import { SubtitlesIcon as Subtitles } from "@phosphor-icons/react/dist/csr/Subtitles";
import { TrashIcon as Trash2 } from "@phosphor-icons/react/dist/csr/Trash";
import { XIcon as X } from "@phosphor-icons/react/dist/csr/X";
import { useLocale } from "@/components/locale-provider";
import { Button } from "@/components/ui/button";
import { type Job } from "@/lib/api";
import { isolate } from "@/lib/i18n";
import { displayText, bytes, expiryLabel } from "@/lib/media";
import { mediaLabel } from "@/lib/localized-media";
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
  const { locale, t } = useLocale();
  const formatBytes = (value?: number) => bytes(value, locale);
  if (jobs.length === 0)
    return (
      <div className="glass-panel px-6 py-10 text-center">
        <span className="mx-auto mb-4 flex size-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
          <Download size={22} aria-hidden="true" />
        </span>
        <h2 className="text-sm font-medium">{t("emptyTitle")}</h2>
        <p className="mx-auto mt-2 max-w-xs text-xs leading-6 text-muted-foreground">
          {t("emptyExplanation")}
        </p>
      </div>
    );
  return (
    <ul className="glass-panel overflow-hidden" aria-label={t("downloadJobs")}>
      {jobs.map((job) => {
        const Icon = mediaIcons[job.kind] || Film;
        const active = job.state === "queued" || job.state === "processing";
        const expired = Date.parse(job.expiresAt) <= now;
        const stateText =
          job.state === "processing"
            ? job.phase === "processing"
              ? t("preparing")
              : job.phase === "analyzing"
                ? t("checkingSource")
                : t("downloading")
            : job.state === "complete"
              ? t("readySave")
              : job.state === "failed"
                ? t("failed")
                : job.state === "canceled"
                  ? t("canceled")
                  : t("inQueue");
        return (
          <li key={job.id} className="job-card p-4 sm:p-5">
            <div className="flex items-start gap-3 sm:gap-4">
              <span
                className={`flex size-11 shrink-0 items-center justify-center rounded-xl ${job.state === "complete" ? "bg-secondary text-primary" : "bg-muted text-muted-foreground"}`}
              >
                <Icon weight="duotone" size={20} aria-hidden="true" />
              </span>
              <div className="min-w-0 flex-1">
                <h3
                  className="truncate text-sm font-medium"
                  title={displayText(job.title)}
                >
                  <bdi>{displayText(job.title)}</bdi>
                </h3>
                <div className="mt-1.5 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground">
                  <span>
                    {fixture ? (
                      t("testFixture")
                    ) : (
                      <bdi>{displayText(job.platform)}</bdi>
                    )}
                  </span>
                  <span aria-hidden="true">·</span>
                  <span>
                    <bdi>{mediaLabel(job.label, locale)}</bdi> /{" "}
                    <bdi dir="ltr">{job.extension.toUpperCase()}</bdi>
                  </span>
                  {job.files[0] ? (
                    <>
                      <span aria-hidden="true">·</span>
                      <span>
                        {formatBytes(
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
                  className={`mt-2 flex flex-wrap items-center gap-1.5 text-xs ${job.state === "failed" ? "text-destructive" : job.state === "complete" ? "text-success" : "text-muted-foreground"}`}
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
                  <span>{expired ? t("expired") : stateText}</span>
                  {job.state === "complete" && !expired ? (
                    <span className="ms-2 text-xs text-muted-foreground">
                      {expiryLabel(job.expiresAt, now, locale)}
                    </span>
                  ) : null}
                  {job.state === "processing" && job.percent >= 0 ? (
                    <span className="ms-auto font-mono">
                      {Math.round(job.percent)}%
                    </span>
                  ) : null}
                </div>
                {job.state === "processing" ? (
                  <div
                    role="progressbar"
                    aria-label={`${stateText}: ${isolate(displayText(job.title))}`}
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
                    {t("jobFailure")}
                  </p>
                ) : null}
              </div>
              <div className="flex shrink-0 items-center gap-1">
                {active ? (
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={t("cancelJob", {
                      title: isolate(displayText(job.title)),
                    })}
                    disabled={busy === job.id}
                    onClick={() => onAction(job.id, "cancel")}
                  >
                    <X size={16} />
                  </Button>
                ) : job.state === "failed" || job.state === "canceled" ? (
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={t("retryJob", {
                      title: isolate(displayText(job.title)),
                    })}
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
                    aria-label={t("deleteJob", {
                      title: isolate(displayText(job.title)),
                    })}
                    disabled={busy === job.id}
                    onClick={() => onAction(job.id, "delete")}
                  >
                    <Trash2 size={15} />
                  </Button>
                ) : null}
              </div>
            </div>
            {job.state === "complete" && !expired ? (
              <div className="mt-4 flex flex-wrap gap-2 sm:ms-[60px]">
                {job.files.map((file, index) => (
                  <Button
                    asChild
                    key={file.id}
                    variant="secondary"
                    size="sm"
                    className="min-h-11 gap-2"
                  >
                    <a
                      href={`/api/v1/jobs/${job.id}/files/${file.id}`}
                      download
                    >
                      <Download size={14} aria-hidden="true" />
                      {fixture
                        ? t("saveTestFile")
                        : job.files.length > 1
                          ? t("saveNumberedFile", { number: index + 1 })
                          : t("saveFile")}
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
