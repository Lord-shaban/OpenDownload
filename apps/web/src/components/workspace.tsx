"use client";
import { useEffect, useRef, useState, type FormEvent } from "react";
import {
  ArrowRight,
  ArrowUpRight,
  Check,
  ChevronRight,
  CircleAlert,
  Clipboard,
  Download,
  Code2,
  Leaf,
  Link2,
  LoaderCircle,
  Moon,
  Plus,
  ShieldCheck,
  Sun,
  X,
} from "lucide-react";
import { Brand } from "@/components/brand";
import Link from "next/link";
import { FormatPicker } from "@/components/format-picker";
import { HelpDialog } from "@/components/help-dialog";
import { JobList } from "@/components/job-list";
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
import { bytes, detectSource, sources } from "@/lib/media";

export function Workspace() {
  const { jobs, status, connection, reload, now } = useInstance();
  const [view, setView] = useState<"workspace" | "downloads">("workspace");
  const [url, setUrl] = useState("");
  const [analysis, setAnalysis] = useState<Analysis | null>(null);
  const [selected, setSelected] = useState<Option | null>(null);
  const [analyzing, setAnalyzing] = useState(false);
  const [queueing, setQueueing] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [actionBusy, setActionBusy] = useState("");
  const [help, setHelp] = useState(false);
  const [dark, setDark] = useState(false);
  const [pendingDelete, setPendingDelete] = useState<Job | null>(null);
  const input = useRef<HTMLInputElement>(null);
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
        input.current?.focus();
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
      updateUrl(await navigator.clipboard.readText());
      input.current?.focus();
    } catch {
      setError(
        "Clipboard access isn’t available. Paste the link into the field with Ctrl+V or ⌘V.",
      );
      input.current?.focus();
    }
  }
  async function analyze(event: FormEvent) {
    event.preventDefault();
    if (analyzing || !url.trim()) return;
    if (!detected.valid) {
      setError(
        "Enter a complete public link starting with https:// or http://.",
      );
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
        body: JSON.stringify({ url: url.trim() }),
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
        setNotice("Analysis complete. Choose an available format.");
        requestAnimationFrame(() => panel.current?.focus());
      }
    } catch (err) {
      if (!controller.signal.aborted)
        setError(
          err instanceof Error
            ? err.message
            : "This link could not be analyzed.",
        );
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
      setNotice("Download added to your queue.");
      await reload();
      setAnalysis(null);
      setSelected(null);
      setUrl("");
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Could not queue this download.",
      );
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
      setNotice(
        operation === "cancel"
          ? "Cancellation requested."
          : "A new retry was added to the queue.",
      );
      await reload();
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "The job could not be updated.",
      );
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
      setNotice("Job and temporary files deleted.");
      await reload();
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "The job could not be deleted.",
      );
    } finally {
      setActionBusy("");
    }
  }
  function switchTheme() {
    const next = !dark;
    setDark(next);
    document.documentElement.classList.toggle("dark", next);
  }
  return (
    <div className="min-h-dvh">
      <a
        href="#main-content"
        className="sr-only fixed top-3 left-3 z-50 rounded-lg bg-card p-3 focus:not-sr-only"
      >
        Skip to workspace
      </a>
      <aside className="fixed inset-y-0 left-0 hidden w-[230px] flex-col border-r bg-card/40 px-5 py-7 lg:flex">
        <Link href="/" className="px-2" aria-label="OpenDownload home">
          <Brand />
        </Link>
        <p className="mt-12 mb-3 px-3 text-[10px] font-medium tracking-[.16em] text-muted-foreground">
          YOUR WORKSPACE
        </p>
        <nav aria-label="Main navigation" className="space-y-1">
          <button
            onClick={() => setView("workspace")}
            aria-current={view === "workspace" ? "page" : undefined}
            className={`flex min-h-11 w-full items-center gap-3 rounded-lg px-3 text-[13px] font-medium ${view === "workspace" ? "bg-secondary text-primary" : "text-muted-foreground hover:bg-muted"}`}
          >
            <Plus size={17} aria-hidden="true" />
            New download
          </button>
          <button
            onClick={() => setView("downloads")}
            aria-current={view === "downloads" ? "page" : undefined}
            className={`flex min-h-11 w-full items-center gap-3 rounded-lg px-3 text-[13px] font-medium ${view === "downloads" ? "bg-secondary text-primary" : "text-muted-foreground hover:bg-muted"}`}
          >
            <Download size={17} aria-hidden="true" />
            Downloads
            {jobs.length > 0 ? (
              <span className="ml-auto rounded bg-muted px-1.5 py-0.5 font-mono text-[10px]">
                {jobs.length}
              </span>
            ) : null}
          </button>
        </nav>
        <div className="mt-auto">
          <div className="mx-2 mb-5 border-b pb-5">
            <Leaf size={20} className="mb-3 text-primary" aria-hidden="true" />
            <p className="text-xs font-medium">Made to stay simple.</p>
            <p className="mt-2 text-[11px] leading-5 text-muted-foreground">
              Open source. Your instance.
              <br />A little more control over your media.
            </p>
          </div>
          <button
            onClick={() => setHelp(true)}
            className="flex min-h-11 w-full items-center justify-between rounded-lg px-3 text-xs text-muted-foreground hover:bg-muted"
          >
            Help & guidelines
            <ArrowUpRight size={14} aria-hidden="true" />
          </button>
          <a
            href="https://github.com/Lord-shaban/OpenDownload"
            target="_blank"
            rel="noreferrer"
            className="flex min-h-11 items-center justify-between rounded-lg px-3 text-xs text-muted-foreground hover:bg-muted"
          >
            View on GitHub
            <Code2 size={14} aria-hidden="true" />
          </a>
          <div className="mt-6 flex items-center gap-2 px-3 font-mono text-[9px] tracking-wide text-muted-foreground">
            <span className="size-1.5 rounded-full bg-primary" />
            SELF-HOSTED / PRE-RELEASE
          </div>
        </div>
      </aside>
      <div className="lg:ml-[230px]">
        <header className="flex h-[76px] items-center justify-between border-b px-5 sm:px-8 xl:px-12">
          <div className="lg:hidden">
            <Brand />
          </div>
          <div className="hidden items-center gap-2 text-xs text-muted-foreground lg:flex">
            Workspace
            <ChevronRight size={13} aria-hidden="true" />
            <span className="text-foreground">
              {view === "workspace" ? "New download" : "Downloads"}
            </span>
          </div>
          <div className="flex items-center gap-2 sm:gap-4">
            <span className="hidden items-center gap-2 text-[11px] text-muted-foreground sm:flex">
              <span
                className={`size-1.5 rounded-full ${connection ? "bg-muted-foreground" : status?.ready ? "bg-primary" : "bg-destructive"}`}
              />
              {fixture
                ? "Test instance"
                : status?.ready
                  ? "Instance ready"
                  : connection
                    ? "Connecting"
                    : "Setup needed"}
            </span>
            <Button
              variant="ghost"
              size="icon"
              onClick={switchTheme}
              aria-label={
                dark ? "Switch to light theme" : "Switch to dark theme"
              }
            >
              {dark ? <Sun size={17} /> : <Moon size={17} />}
            </Button>
          </div>
        </header>
        <nav
          aria-label="Mobile navigation"
          className="flex border-b px-5 lg:hidden"
        >
          <button
            className={`min-h-12 flex-1 text-xs font-medium ${view === "workspace" ? "border-b-2 border-primary text-primary" : "text-muted-foreground"}`}
            aria-current={view === "workspace" ? "page" : undefined}
            onClick={() => setView("workspace")}
          >
            New download
          </button>
          <button
            className={`min-h-12 flex-1 text-xs font-medium ${view === "downloads" ? "border-b-2 border-primary text-primary" : "text-muted-foreground"}`}
            aria-current={view === "downloads" ? "page" : undefined}
            onClick={() => setView("downloads")}
          >
            Downloads{activeJobs ? ` (${activeJobs} active)` : ""}
          </button>
          <button
            className="min-h-12 px-4 text-xs text-muted-foreground"
            onClick={() => setHelp(true)}
          >
            Help
          </button>
        </nav>
        <main
          id="main-content"
          className="mx-auto max-w-[1000px] px-5 pt-9 pb-12 sm:px-8 sm:pt-14 xl:px-12"
        >
          {fixture ? (
            <div className="mb-6 flex items-start gap-2 rounded-lg border border-primary/20 bg-secondary px-4 py-3 text-xs leading-5 text-primary">
              <CircleAlert
                size={15}
                className="mt-0.5 shrink-0"
                aria-hidden="true"
              />
              <p>
                <strong>Fixture mode.</strong> This instance tests the complete
                workflow. It produces a labeled test file, not extracted media.
              </p>
            </div>
          ) : null}
          {status && !status.ready ? (
            <div className="mb-6 rounded-lg border border-destructive/30 bg-card p-4 text-xs leading-6 text-destructive">
              Media tools are missing from this instance. Install yt-dlp and
              FFmpeg, or start the Docker deployment described in the README.
            </div>
          ) : null}
          {connection && status ? (
            <p className="mb-4 text-xs text-muted-foreground" role="status">
              {connection}
            </p>
          ) : null}
          <div className="mb-8 flex items-end justify-between gap-4">
            <div>
              <p className="mb-3 text-[10px] font-medium tracking-[.18em] text-primary">
                {view === "workspace"
                  ? "LESS FRICTION. MORE FREEDOM."
                  : "YOUR TEMPORARY LIBRARY"}
              </p>
              <h1 className="text-[clamp(30px,3.8vw,46px)] leading-[1.14] font-medium tracking-[-.05em]">
                {view === "workspace"
                  ? "Save something good."
                  : "Ready when you are."}
              </h1>
              <p className="mt-4 max-w-lg text-sm leading-6 text-muted-foreground">
                {view === "workspace"
                  ? "A link is all you need. Choose what to keep, and we’ll handle the rest."
                  : "Follow your downloads and save finished files before they expire."}
              </p>
            </div>
            {view === "downloads" ? (
              <Button
                variant="outline"
                onClick={() => setView("workspace")}
                className="shrink-0 gap-2"
              >
                <Plus size={15} aria-hidden="true" />
                <span className="hidden sm:inline">New download</span>
              </Button>
            ) : null}
          </div>
          {view === "workspace" ? (
            <>
              <form
                onSubmit={analyze}
                className="relative rounded-2xl border bg-card p-4 shadow-[0_3px_14px_-8px_#202b2225] sm:p-5"
                aria-busy={analyzing}
              >
                <label
                  htmlFor="media-url"
                  className="mb-3 block text-xs font-medium"
                >
                  Media link
                </label>
                <div className="flex flex-col gap-3 sm:flex-row">
                  <div className="relative min-w-0 flex-1">
                    <Link2
                      size={17}
                      className="pointer-events-none absolute top-4 left-3.5 text-muted-foreground"
                      aria-hidden="true"
                    />
                    <input
                      ref={input}
                      id="media-url"
                      type="text"
                      inputMode="url"
                      autoComplete="off"
                      autoCapitalize="none"
                      spellCheck={false}
                      dir="ltr"
                      value={url}
                      onChange={(event) => updateUrl(event.target.value)}
                      placeholder="Paste a public video, audio, or image link"
                      aria-invalid={!!error}
                      aria-describedby={error ? "workspace-error" : "url-hint"}
                      className="h-12 w-full rounded-lg border bg-background/40 pr-12 pl-11 text-base sm:text-[13px] placeholder:text-muted-foreground focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/15"
                    />
                    {url ? (
                      <button
                        type="button"
                        aria-label="Clear link"
                        className="absolute top-0 right-0 flex size-12 items-center justify-center text-muted-foreground hover:text-foreground"
                        onClick={() => updateUrl("")}
                      >
                        <X size={15} />
                      </button>
                    ) : (
                      <button
                        type="button"
                        aria-label="Paste link from clipboard"
                        title="Paste from clipboard"
                        className="absolute top-0 right-0 flex size-12 items-center justify-center text-muted-foreground hover:text-primary"
                        onClick={() => void paste()}
                      >
                        <Clipboard size={16} />
                      </button>
                    )}
                  </div>
                  <Button
                    type="submit"
                    className="h-12 gap-2 rounded-lg px-6 text-xs"
                    disabled={
                      analyzing || !url.trim() || status?.ready === false
                    }
                  >
                    {analyzing ? (
                      <>
                        <LoaderCircle
                          size={15}
                          className="spinner"
                          aria-hidden="true"
                        />
                        Analyzing…
                      </>
                    ) : (
                      <>
                        Analyze link
                        <ArrowRight size={15} aria-hidden="true" />
                      </>
                    )}
                  </Button>
                </div>
                <div
                  id="url-hint"
                  className="mt-3 flex min-h-5 flex-wrap items-center justify-between gap-2 text-[10px] text-muted-foreground"
                >
                  <span className="flex items-center gap-1.5">
                    {detected.valid ? (
                      <>
                        <Check
                          size={12}
                          className="text-primary"
                          aria-hidden="true"
                        />
                        <span className="font-medium text-primary">
                          {detected.name} detected
                        </span>
                      </>
                    ) : (
                      <>
                        <ShieldCheck size={12} aria-hidden="true" />
                        Only analyzed when you ask. No login required.
                      </>
                    )}
                  </span>
                  {analyzing ? (
                    <button
                      type="button"
                      onClick={() => {
                        abort.current?.abort();
                        setAnalyzing(false);
                        setNotice("Analysis canceled.");
                      }}
                      className="min-h-6 text-primary underline underline-offset-2"
                    >
                      Cancel analysis
                    </button>
                  ) : (
                    <span className="hidden items-center gap-1 sm:flex">
                      Focus link{" "}
                      <kbd className="rounded border px-1.5 py-0.5 font-mono">
                        /
                      </kbd>
                    </span>
                  )}
                </div>
              </form>
              <div className="mt-5 mb-9 flex flex-wrap items-center gap-x-3 gap-y-2 text-[10px] text-muted-foreground">
                <span className="mr-1">POPULAR SOURCES</span>
                {sources.map((source) => (
                  <span key={source} className="font-medium">
                    {source}
                  </span>
                ))}
                <button
                  onClick={() => setHelp(true)}
                  className="min-h-7 text-primary underline underline-offset-3"
                >
                  & other public sources
                </button>
              </div>
            </>
          ) : null}
          {error ? (
            <div
              id="workspace-error"
              role="alert"
              className="mb-6 flex items-start gap-2 rounded-xl border border-destructive/25 bg-card px-4 py-3 text-xs leading-6 text-destructive"
            >
              <CircleAlert
                size={16}
                className="mt-1 shrink-0"
                aria-hidden="true"
              />
              {error}
            </div>
          ) : null}
          <p className="sr-only" role="status" aria-live="polite">
            {notice}
          </p>
          {view === "workspace" && analyzing ? (
            <div className="mb-8 rounded-2xl border bg-card p-6" role="status">
              <div className="mb-5 flex items-center gap-3">
                <span className="flex size-12 items-center justify-center rounded-xl bg-secondary">
                  <LoaderCircle
                    size={22}
                    className="spinner text-primary"
                    aria-hidden="true"
                  />
                </span>
                <div>
                  <p className="text-sm font-medium">Looking at the source.</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    Checking public access and available formats…
                  </p>
                </div>
              </div>
              <div className="h-2 w-3/4 rounded bg-muted" />
              <div className="mt-3 h-2 w-1/2 rounded bg-muted" />
              <p className="mt-5 text-[11px] text-muted-foreground">
                This can take a moment. You can cancel at any time.
              </p>
            </div>
          ) : null}
          {view === "workspace" && analysis ? (
            <div
              ref={panel}
              tabIndex={-1}
              aria-label="Analysis results"
              className="mb-9 rounded-2xl outline-none"
            >
              <FormatPicker
                analysis={analysis}
                selected={selected}
                onSelect={setSelected}
                busy={queueing}
                onQueue={() => void queue()}
                fixture={fixture}
                maxBytes={status?.limits.maxBytes || Infinity}
              />
            </div>
          ) : null}
          {view === "workspace" &&
          !analysis &&
          !analyzing &&
          jobs.length === 0 ? (
            <section
              className="mb-10 grid grid-cols-1 gap-6 border-y py-7 sm:grid-cols-3"
              aria-label="How it works"
            >
              {[
                {
                  title: "One link.",
                  text: "Paste a public media URL.",
                  number: "01",
                },
                {
                  title: "Your choice.",
                  text: "Pick from the available formats.",
                  number: "02",
                },
                {
                  title: "It’s yours to save.",
                  text: "We process. You download.",
                  number: "03",
                },
              ].map((step) => (
                <div key={step.number} className="flex gap-3">
                  <span className="pt-0.5 font-mono text-[10px] text-muted-foreground/80">
                    {step.number}
                  </span>
                  <div>
                    <h2 className="text-xs font-medium">{step.title}</h2>
                    <p className="mt-1.5 text-[11px] leading-5 text-muted-foreground">
                      {step.text}
                    </p>
                  </div>
                </div>
              ))}
            </section>
          ) : null}
          <section aria-labelledby="jobs-heading">
            <div className="mb-4 flex items-center justify-between">
              <h2
                id="jobs-heading"
                className="text-sm font-medium tracking-tight"
              >
                {view === "workspace" ? "Your downloads" : "All downloads"}
                {activeJobs > 0 ? (
                  <span className="ml-2 rounded-full bg-secondary px-2 py-0.5 text-[10px] text-primary">
                    {activeJobs} active
                  </span>
                ) : null}
              </h2>
              {jobs.length === 0 && fixture ? (
                <button
                  onClick={() => {
                    setView("workspace");
                    updateUrl("https://example.com/sample");
                    input.current?.focus();
                  }}
                  className="flex min-h-9 items-center gap-1.5 text-xs text-primary"
                >
                  Try a test link
                  <ArrowUpRight size={12} aria-hidden="true" />
                </button>
              ) : (
                <span className="text-[10px] text-muted-foreground">
                  {status
                    ? `Files kept for ${Math.round(status.limits.retentionSeconds / 3600)}h`
                    : "Temporary storage"}
                </span>
              )}
            </div>
            <JobList
              now={now}
              jobs={jobs}
              busy={actionBusy}
              onAction={(id, operation) => void action(id, operation)}
              fixture={fixture}
            />
          </section>
          <footer className="mt-9 flex flex-wrap items-center justify-between gap-4 border-t pt-5 text-[10px] leading-5 text-muted-foreground">
            <span className="flex items-center gap-1.5">
              <ShieldCheck size={12} aria-hidden="true" />
              No trackers. No ads. Just your media.
            </span>
            <button
              onClick={() => setHelp(true)}
              className="min-h-8 hover:text-primary"
            >
              {status ? `${bytes(status.limits.maxBytes)} per job · ` : ""}
              Public content only
              <ArrowUpRight
                size={11}
                className="ml-1 inline"
                aria-hidden="true"
              />
            </button>
          </footer>
        </main>
      </div>
      <HelpDialog open={help} onOpenChange={setHelp} />
      <Dialog
        open={!!pendingDelete}
        onOpenChange={(open) => {
          if (!open) setPendingDelete(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete this download?</DialogTitle>
            <DialogDescription>
              The job and its temporary files will be removed from this
              instance. Files already saved to your device are unaffected.
            </DialogDescription>
          </DialogHeader>
          <p className="truncate text-sm font-medium">{pendingDelete?.title}</p>
          <div className="mt-4 flex justify-end gap-2">
            <Button variant="outline" onClick={() => setPendingDelete(null)}>
              Keep download
            </Button>
            <Button
              variant="destructive"
              disabled={!!actionBusy}
              onClick={() => void confirmDelete()}
            >
              {actionBusy ? "Deleting…" : "Delete download"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
