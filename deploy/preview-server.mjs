import { startServers } from "./process-supervisor.mjs";

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

console.warn(
  "OpenDownload PREVIEW: fixture mode is forced; no real media is downloaded.",
);
await startServers(env, { ...process.env, HOSTNAME: "0.0.0.0" });
