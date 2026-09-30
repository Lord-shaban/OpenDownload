import { spawn } from "node:child_process";
import { setTimeout as delay } from "node:timers/promises";

// This image deliberately cannot be switched to real extraction by configuration.
const env = {
  ...process.env,
  OD_FIXTURE_MODE: "true",
  OD_PORT: "8080",
  OD_DATA_DIR: "/data",
  OD_WORKERS: "1",
  OD_QUEUE_LIMIT: "3",
  OD_MAX_BYTES: "134217728",
  OD_JOB_TIMEOUT: "5m",
  OD_RETENTION: "15m",
};
const children = new Set();
let stopping = false;
let exitCode = 0;
let killTimer;

function stop(code = 0) {
  if (stopping) return;
  stopping = true;
  exitCode = code;
  for (const child of children) child.kill("SIGTERM");
  killTimer = setTimeout(() => {
    for (const child of children) child.kill("SIGKILL");
  }, 12_000);
  killTimer.unref();
  finish();
}

function finish() {
  if (stopping && children.size === 0) {
    clearTimeout(killTimer);
    process.exit(exitCode);
  }
}

function start(command, args, environment) {
  const child = spawn(command, args, { env: environment, stdio: "inherit" });
  children.add(child);
  child.once("error", (error) => {
    console.error("Preview process failed to start:", error.message);
    stop(1);
  });
  child.once("close", (code, signal) => {
    children.delete(child);
    if (!stopping) {
      console.error(
        "Preview process stopped unexpectedly:",
        command,
        code,
        signal,
      );
      stop(code || 1);
    }
    finish();
  });
  return child;
}

process.on("SIGTERM", () => stop());
process.on("SIGINT", () => stop());
console.warn(
  "OpenDownload PREVIEW: fixture mode is forced; no real media is downloaded.",
);
start("/usr/local/bin/opendownload-api", [], env);

let ready = false;
for (let attempt = 0; attempt < 120 && !stopping; attempt++) {
  try {
    const response = await fetch("http://127.0.0.1:8080/api/v1/health", {
      signal: AbortSignal.timeout(1000),
    });
    if (response.ok) {
      ready = true;
      break;
    }
  } catch {
    // The API has not opened its listener yet.
  }
  await delay(250);
}
if (!ready) {
  if (!stopping) console.error("Preview API did not become ready.");
  stop(1);
} else if (!stopping) {
  start(process.execPath, ["apps/web/server.js"], {
    ...process.env,
    HOSTNAME: "0.0.0.0",
  });
}
