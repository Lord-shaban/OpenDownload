// SPDX-License-Identifier: GPL-3.0-only
// Invoke the pinned upstream library without its CLI's persistent token cache.
// This process shares the extractor's process group and isolated network.
import { parseArgs } from "node:util";
import { SessionManager } from "../../build/session_manager.js";
import { VERSION } from "../../build/utils.js";

try {
  const { values } = parseArgs({
    options: {
      version: { type: "boolean" },
      "content-binding": { type: "string", short: "c" },
      proxy: { type: "string", short: "p" },
      "innertube-context": { type: "string" },
      "bypass-cache": { type: "boolean" },
    },
  });
  if (values.version) {
    console.log(VERSION);
  } else {
    // The cloud worker has exactly one guarded bridge. No direct-network or
    // caller-selected proxy fallback, TLS verification overrides, or accounts.
    if (values.proxy !== "http://127.0.0.1:8090") {
      throw new Error("The guarded cloud proxy is required.");
    }
    if (!/^[A-Za-z0-9_-]{1,512}$/.test(values["content-binding"] || "")) {
      throw new Error("Invalid anonymous content binding.");
    }
    const context = values["innertube-context"];
    if (!context || context.length > 65536) {
      throw new Error("A bounded public client context is required.");
    }
    const session = await new SessionManager(false).generatePoToken(
      values["content-binding"], values.proxy, false, undefined, false,
      undefined, JSON.parse(context),
    );
    // The plugin consumes this output; it is never published to the visitor.
    console.log(JSON.stringify(session));
  }
} catch {
  console.error("Anonymous YouTube attestation could not be completed.");
  console.log("{}");
  process.exitCode = 1;
}
