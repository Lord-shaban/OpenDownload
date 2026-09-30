"use client";
import { useEffect, useRef, useState } from "react";
import { ArrowRightIcon as ArrowRight } from "@phosphor-icons/react/dist/csr/ArrowRight";
import { WarningCircleIcon as CircleAlert } from "@phosphor-icons/react/dist/csr/WarningCircle";
import { ClipboardTextIcon as Clipboard } from "@phosphor-icons/react/dist/csr/ClipboardText";
import { QuestionIcon as CircleHelp } from "@phosphor-icons/react/dist/csr/Question";
import { LinkSimpleIcon as Link2 } from "@phosphor-icons/react/dist/csr/LinkSimple";
import { CircleNotchIcon as LoaderCircle } from "@phosphor-icons/react/dist/csr/CircleNotch";
import { MoonStarsIcon as Moon } from "@phosphor-icons/react/dist/csr/MoonStars";
import { PlusIcon as Plus } from "@phosphor-icons/react/dist/csr/Plus";
import { SunIcon as Sun } from "@phosphor-icons/react/dist/csr/Sun";
import { XIcon as X } from "@phosphor-icons/react/dist/csr/X";
import { useLocale } from "@/components/locale-provider";
import { FlowProgress } from "@/components/flow-progress";
import { Brand } from "@/components/brand";
import Link from "next/link";
import { FormatPicker } from "@/components/format-picker";
import { HelpDialog } from "@/components/help-dialog";
import { JobList } from "@/components/job-list";
import { SupportedSites } from "@/components/supported-sites";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useInstance } from "@/hooks/use-instance";
import {
  type Analysis,
  type Job,
  type Option,
  analysisSchema,
  emptySchema,
  jobSchema,
  request,
} from "@/lib/api";
import { type MessageKey } from "@/lib/i18n";
import { detectSource, displayText } from "@/lib/media";
import { errorMessage } from "@/lib/localized-media";

export function Workspace({ initialDark = false }: { initialDark?: boolean }) {
  const { locale, setLocale, t } = useLocale();
  const { jobs, status, connection, reload, now } = useInstance();
  const [view, setView] = useState<"workspace" | "downloads">("workspace");
  const [url, setUrl] = useState("");
  const [analysis, setAnalysis] = useState<Analysis | null>(null);
  const [selected, setSelected] = useState<Option | null>(null);
  const [analyzing, setAnalyzing] = useState(false);
  const [queueing, setQueueing] = useState(false);
  const [error, setError] = useState<unknown>(null);
  const [notice, setNotice] = useState<MessageKey | "">("");
  const [actionBusy, setActionBusy] = useState("");
  const [help, setHelp] = useState(false);
  const [dark, setDark] = useState(initialDark);
  const [pendingDelete, setPendingDelete] = useState<Job | null>(null);
  const input = useRef<HTMLInputElement>(null);
  const main = useRef<HTMLElement>(null);
  const panel = useRef<HTMLDivElement>(null);
  const abort = useRef<AbortController | null>(null);
  const detected = detectSource(url);
  const activeJobs = jobs.filter(
    (job) => job.state === "queued" || job.state === "processing",
  ).length;
  const fixture = status?.fixtureMode || false;
  useEffect(() => {
    const shortcut = (event: KeyboardEvent) => {
      const target = event.target as HTMLElement;
      if (
        event.key === "/" &&
        !["INPUT", "TEXTAREA", "SELECT"].includes(target.tagName) &&
        !target.isContentEditable
      ) {
        event.preventDefault();
        setView("workspace");
        requestAnimationFrame(() => input.current?.focus());
      }
    };
    window.addEventListener("keydown", shortcut);
    return () => {
      window.removeEventListener("keydown", shortcut);
      abort.current?.abort();
    };
  }, []);
  function updateUrl(value: string) {
    abort.current?.abort();
    setAnalyzing(false);
    setUrl(value);
    setAnalysis(null);
    setSelected(null);
    setError("");
    setNotice("");
  }
  async function paste() {
    try {
      const pasted = await navigator.clipboard.readText();
      updateUrl(pasted);
      input.current?.focus();
      if (detectSource(pasted).valid && status?.ready) void analyzeUrl(pasted);
    } catch {
      setError("clipboardError");
      input.current?.focus();
    }
  }
  async function analyzeUrl(value: string) {
    if (!value.trim()) return;
    if (!detectSource(value).valid) {
      setError("invalidUrl");
      input.current?.focus();
      return;
    }
    const controller = new AbortController();
    abort.current?.abort();
    abort.current = controller;
    setAnalyzing(true);
    setError("");
    setNotice("");
    setAnalysis(null);
    setSelected(null);
    try {
      const result = await request("/analyze", analysisSchema, {
        method: "POST",
        body: JSON.stringify({ url: value.trim() }),
        signal: controller.signal,
      });
      if (!controller.signal.aborted) {
        setAnalysis(result);
        setSelected(
          result.options.find(
            (option) =>
              (option.bytes || 0) <= (status?.limits.maxBytes || Infinity),
          ) || null,
        );
        setNotice("analysisComplete");
        requestAnimationFrame(() => panel.current?.focus());
      }
    } catch (err) {
      if (!controller.signal.aborted) setError(err);
    } finally {
      if (!controller.signal.aborted) setAnalyzing(false);
    }
  }
  async function queue() {
    if (!analysis || !selected || queueing) return;
    setQueueing(true);
    setError("");
    try {
      await request("/jobs", jobSchema, {
        method: "POST",
        body: JSON.stringify({
          analysisId: analysis.id,
          optionId: selected.id,
        }),
      });
      setNotice("queuedNotice");
      setView("downloads");
      requestAnimationFrame(() => main.current?.focus());
      await reload();
      setAnalysis(null);
      setSelected(null);
      setUrl("");
    } catch (err) {
      setError(err);
    } finally {
      setQueueing(false);
    }
  }
  async function action(id: string, operation: "cancel" | "retry" | "delete") {
    if (operation === "delete") {
      setPendingDelete(jobs.find((job) => job.id === id) || null);
      return;
    }
    setActionBusy(id);
    setError("");
    try {
      await request(`/jobs/${id}/${operation}`, jobSchema, {
        method: "POST",
        body: "{}",
      });
      setNotice(operation === "cancel" ? "cancelNotice" : "retryNotice");
      await reload();
    } catch (err) {
      setError(err);
    } finally {
      setActionBusy("");
    }
  }
  async function confirmDelete() {
    if (!pendingDelete) return;
    setActionBusy(pendingDelete.id);
    setError("");
    try {
      await request(`/jobs/${pendingDelete.id}`, emptySchema, {
        method: "DELETE",
      });
      setPendingDelete(null);
      setNotice("deleteNotice");
      await reload();
      main.current?.focus();
    } catch (err) {
      setError(err);
    } finally {
      setActionBusy("");
    }
  }
  function showWorkspace() {
    setView("workspace");
    requestAnimationFrame(() => input.current?.focus());
  }
  function switchTheme() {
    const next = !dark;
    setDark(next);
    document.documentElement.classList.toggle("dark", next);
    document.cookie = `od_theme=${next ? "dark" : "light"}; Path=/; Max-Age=31536000; SameSite=Lax${location.protocol === "https:" ? "; Secure" : ""}`;
  }
  return (
    <div className="workspace-shell">
      <a href="#main-content" className="skip-link">
        {t("skip")}
      </a>
      <header className="app-header">
        <Link href="/" aria-label={t("home")} className="shrink-0">
          <Brand />
        </Link>
        <div className="header-controls">
          <label className="sr-only" htmlFor="workspace-language">
            {t("language")}
          </label>
          <select
            id="workspace-language"
            value={locale}
            onChange={(event) =>
              setLocale(event.target.value === "ar" ? "ar" : "en")
            }
            className="language-control"
          >
            <option value="en" lang="en">
              English
            </option>
            <option value="ar" lang="ar">
              العربية
            </option>
          </select>
          <Button
            variant="ghost"
            size="icon"
            className="glass-control"
            onClick={switchTheme}
            aria-label={dark ? t("lightTheme") : t("darkTheme")}
          >
            {dark ? (
              <Sun weight="duotone" size={18} aria-hidden="true" />
            ) : (
              <Moon weight="duotone" size={18} aria-hidden="true" />
            )}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="glass-control"
            onClick={() => setHelp(true)}
            aria-label={t("help")}
          >
            <CircleHelp weight="duotone" size={18} aria-hidden="true" />
          </Button>
        </div>
      </header>
      <main
        id="main-content"
        ref={main}
        tabIndex={-1}
        className="workspace-main outline-none"
      >
        <nav aria-label={t("mainNav")} className="workspace-nav">
          <button
            onClick={showWorkspace}
            aria-current={view === "workspace" ? "page" : undefined}
          >
            {t("newDownload")}
          </button>
          <button
            onClick={() => setView("downloads")}
            aria-label={t("downloads")}
            aria-current={view === "downloads" ? "page" : undefined}
          >
            {t("downloads")}
            {jobs.length > 0 ? (
              <span className="nav-count">{jobs.length}</span>
            ) : null}
          </button>
        </nav>
        <div key={view} className="workspace-stage view-enter">
          <h1 className="sr-only">
            {t(view === "workspace" ? "newDownload" : "downloads")}
          </h1>
          {fixture ? (
            <div className="instance-notice" role="status">
              <CircleAlert size={16} className="shrink-0" aria-hidden="true" />
              <p>
                <strong>{t("fixtureMode")}</strong> {t("fixtureExplanation")}
              </p>
            </div>
          ) : null}
          {status && !status.ready ? (
            <div className="instance-notice text-destructive">
              {t("missingTools")}
            </div>
          ) : null}
          {connection ? (
            <p
              className="mb-4 text-center text-sm text-muted-foreground"
              role="status"
            >
              {t(connection)}
            </p>
          ) : null}
          <div className="sr-only">
            <FlowProgress step={view === "downloads" ? 3 : analysis ? 2 : 1} />
          </div>
          {view === "workspace" ? (
            <>
              <form
                onSubmit={(event) => {
                  event.preventDefault();
                  if (!analyzing) void analyzeUrl(url);
                }}
                className="link-composer glass-panel"
                data-filled={!!url}
                aria-busy={analyzing}
              >
                <label htmlFor="media-url" className="sr-only">
                  {t("mediaLink")}
                </label>
                <div className="composer-input" dir="ltr">
                  <Link2
                    weight="duotone"
                    size={20}
                    className="link-icon"
                    aria-hidden="true"
                  />
                  <input
                    ref={input}
                    id="media-url"
                    name="url"
                    type="text"
                    inputMode="url"
                    autoComplete="off"
                    spellCheck={false}
                    dir="ltr"
                    value={url}
                    disabled={queueing}
                    onChange={(event) => updateUrl(event.target.value)}
                    onKeyDown={(event) => {
                      if (event.key === "Enter") {
                        event.preventDefault();
                        if (!analyzing && status?.ready)
                          void analyzeUrl(event.currentTarget.value);
                      }
                    }}
                    onPaste={(event) => {
                      const pasted = event.clipboardData.getData("text").trim();
                      if (detectSource(pasted).valid && status?.ready) {
                        event.preventDefault();
                        updateUrl(pasted);
                        void analyzeUrl(pasted);
                      }
                    }}
                    placeholder={t("placeholder")}
                    aria-invalid={!!error && !!url.trim()}
                    aria-describedby={error ? "workspace-error" : "url-hint"}
                  />
                  {url ? (
                    <button
                      type="button"
                      className="input-action"
                      aria-label={t("clearLink")}
                      disabled={queueing}
                      onClick={() => {
                        updateUrl("");
                        input.current?.focus();
                      }}
                    >
                      <X size={16} aria-hidden="true" />
                    </button>
                  ) : null}
                  <Button
                    type="submit"
                    className="analyze-button"
                    size="icon"
                    disabled={
                      analyzing || !url.trim() || !status?.ready || queueing
                    }
                    aria-label={analyzing ? t("analyzing") : t("analyzeLink")}
                    title={t("analyzeLink")}
                  >
                    {analyzing ? (
                      <LoaderCircle
                        size={19}
                        className="spinner"
                        aria-hidden="true"
                      />
                    ) : (
                      <ArrowRight
                        size={19}
                        className="rtl:rotate-180"
                        aria-hidden="true"
                      />
                    )}
                  </Button>
                </div>
              </form>
              <div className="composer-tools">
                {!url ? (
                  <Button
                    variant="secondary"
                    className="paste-button"
                    onClick={() => void paste()}
                    disabled={!status?.ready}
                    aria-label={t("pasteLink")}
                  >
                    <Clipboard weight="duotone" size={16} aria-hidden="true" />
                    {t("pasteClipboard")}
                  </Button>
                ) : analyzing ? (
                  <button
                    className="quiet-action"
                    onClick={() => {
                      abort.current?.abort();
                      setAnalyzing(false);
                      setNotice("analysisCanceled");
                    }}
                  >
                    {t("cancelAnalysis")}
                  </button>
                ) : (
                  <span id="url-hint" className="source-hint">
                    {detected.valid ? (
                      <bdi>{detected.name}</bdi>
                    ) : (
                      t("publicOnly")
                    )}
                  </span>
                )}
                <span
                  id={!url || analyzing ? "url-hint" : undefined}
                  className="sr-only"
                >
                  {t("publicOnly")}
                </span>
              </div>
              {!analysis && !analyzing ? <SupportedSites /> : null}
            </>
          ) : null}
          {error ? (
            <div id="workspace-error" className="error-panel" role="alert">
              <CircleAlert size={18} className="shrink-0" aria-hidden="true" />
              <p>{errorMessage(error, locale)}</p>
            </div>
          ) : null}
          <p className="sr-only" role="status" aria-live="polite">
            {notice ? t(notice) : ""}
          </p>
          {view === "workspace" && analyzing ? (
            <div
              className="loading-panel glass-panel panel-enter"
              role="status"
            >
              <div className="loading-lines" aria-hidden="true">
                <span />
                <span />
                <span />
              </div>
              <span className="text-sm text-muted-foreground">
                {t("checkingFormats")}
              </span>
            </div>
          ) : null}
          {view === "workspace" && analysis && !analyzing ? (
            <div
              ref={panel}
              tabIndex={-1}
              aria-label={t("results")}
              className="analysis-region outline-none"
            >
              <FormatPicker
                analysis={analysis}
                selected={selected}
                onSelect={setSelected}
                busy={queueing}
                onQueue={() => void queue()}
                fixture={fixture}
                maxBytes={status?.limits.maxBytes || 0}
              />
            </div>
          ) : null}
          {view === "downloads" ? (
            <section
              aria-labelledby="jobs-heading"
              className="downloads-section"
            >
              <div className="downloads-heading">
                <h2 id="jobs-heading" className="sr-only">
                  {t("allDownloads")}
                </h2>
                {activeJobs > 0 ? (
                  <span className="text-xs text-muted-foreground">
                    {t("activeCount", { count: activeJobs })}
                  </span>
                ) : null}
                <Button
                  variant="ghost"
                  className="quiet-action"
                  onClick={showWorkspace}
                >
                  <Plus size={16} aria-hidden="true" />
                  {t("newDownload")}
                </Button>
              </div>
              <JobList
                jobs={jobs}
                onAction={(id, operation) => void action(id, operation)}
                busy={actionBusy}
                fixture={fixture}
                now={now}
              />
            </section>
          ) : null}
          {fixture && view === "workspace" && !analysis && !analyzing ? (
            <button
              className="quiet-action mx-auto mt-4 flex"
              onClick={() => {
                updateUrl("https://example.com/sample");
                input.current?.focus();
              }}
            >
              {t("testLink")}
            </button>
          ) : null}
          <footer className="workspace-footer">
            <button onClick={() => setHelp(true)}>{t("publicOnly")}</button>
            <span aria-hidden="true">·</span>
            <a
              href="https://github.com/Lord-shaban/OpenDownload"
              target="_blank"
              rel="noreferrer"
              aria-label={t("github")}
            >
              GitHub
            </a>
          </footer>
        </div>
      </main>
      <HelpDialog open={help} onOpenChange={setHelp} />
      <Dialog
        open={!!pendingDelete}
        onOpenChange={(open) => {
          if (!open) setPendingDelete(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("deleteQuestion")}</DialogTitle>
            <DialogDescription>{t("deleteExplanation")}</DialogDescription>
          </DialogHeader>
          <p className="truncate text-sm font-medium">
            <bdi>{displayText(pendingDelete?.title || "")}</bdi>
          </p>
          <div className="mt-4 flex flex-wrap justify-end gap-2">
            <Button
              variant="outline"
              className="min-h-11"
              onClick={() => setPendingDelete(null)}
            >
              {t("keepDownload")}
            </Button>
            <Button
              variant="destructive"
              className="min-h-11"
              disabled={!!actionBusy}
              onClick={() => void confirmDelete()}
            >
              {actionBusy ? t("deleting") : t("deleteDownload")}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
