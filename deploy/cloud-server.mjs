import { startServers } from "./process-supervisor.mjs";

// Executed only after the cloud worker's namespace check has succeeded.
const env = {
  ...process.env,
  DISPLAY: ":99",
  OD_FIXTURE_MODE: "false",
  OD_PORT: "8080",
  OD_DATA_DIR: "/data",
  OD_EGRESS_PROXY: "http://127.0.0.1:8090",
  OD_RESOLVER_SOCKET: "/tmp/opendownload-cloud/resolver.sock",
  OD_WORKERS: "1",
  OD_QUEUE_LIMIT: "3",
  OD_MAX_BYTES: "134217728",
  OD_STORAGE_BUDGET: "536870912",
  OD_JOB_TIMEOUT: "5m",
  OD_RETENTION: "15m",
};
console.info(
  "OpenDownload cloud: real extraction enabled inside the network fence.",
);
await startServers(env, {
  ...process.env,
  HOSTNAME: "127.0.0.1",
  PORT: "3001",
}, [{command: "Xvfb", args: [":99", "-screen", "0", "1280x720x24", "-nolisten", "tcp"]}]);
