"use client";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  type Job,
  type Status,
  jobListSchema,
  request,
  statusSchema,
} from "@/lib/api";
export function useInstance() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [status, setStatus] = useState<Status | null>(null);
  const [now, setNow] = useState(0);
  const [connection, setConnection] = useState("Connecting to your instance…");
  const controller = useRef<AbortController | null>(null);
  const active = useRef(false);
  const reload = useCallback(async () => {
    if (active.current && !controller.current?.signal.aborted) return;
    active.current = true;
    const abort = new AbortController();
    controller.current = abort;
    const results = await Promise.allSettled([
      request("/jobs", jobListSchema, { signal: abort.signal }),
      request("/status", statusSchema, { signal: abort.signal }),
    ]);
    if (!abort.signal.aborted) {
      setNow(Date.now());
      if (results[0].status === "fulfilled") setJobs(results[0].value.jobs);
      if (results[1].status === "fulfilled") setStatus(results[1].value);
      setConnection(
        results.every((result) => result.status === "fulfilled")
          ? ""
          : "Connection interrupted. Reconnecting; your jobs stay on the server.",
      );
    }
    if (controller.current === abort) active.current = false;
  }, []);
  const busy = jobs.some(
    (job) => job.state === "queued" || job.state === "processing",
  );
  useEffect(() => {
    let disposed = false;
    let timer: ReturnType<typeof setTimeout>;
    const poll = async () => {
      if (!document.hidden) await reload();
      if (!disposed) timer = setTimeout(poll, busy ? 1500 : 8000);
    };
    void poll();
    const visible = () => {
      if (!document.hidden) void reload();
    };
    document.addEventListener("visibilitychange", visible);
    return () => {
      disposed = true;
      clearTimeout(timer);
      controller.current?.abort();
      document.removeEventListener("visibilitychange", visible);
    };
  }, [busy, reload]);
  return { jobs, status, connection, reload, now };
}
