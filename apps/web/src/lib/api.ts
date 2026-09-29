import { z } from "zod";
const option = z.object({
  id: z.string(),
  kind: z.enum(["video", "audio", "image", "subtitle"]),
  label: z.string(),
  extension: z.string(),
  detail: z.string(),
  bytes: z.number().optional(),
});
export const analysisSchema = z.object({
  id: z.string(),
  title: z.string(),
  creator: z.string(),
  platform: z.string(),
  duration: z.number(),
  hasThumbnail: z.boolean(),
  options: z.array(option),
  expiresAt: z.string(),
});
export const jobSchema = z.object({
  id: z.string(),
  title: z.string(),
  platform: z.string(),
  kind: z.string(),
  label: z.string(),
  extension: z.string(),
  state: z.enum(["queued", "processing", "complete", "failed", "canceled"]),
  phase: z.string(),
  percent: z.number(),
  error: z.string().optional(),
  createdAt: z.string(),
  expiresAt: z.string(),
  files: z.array(
    z.object({ id: z.string(), name: z.string(), bytes: z.number() }),
  ),
});
export const statusSchema = z.object({
  ready: z.boolean(),
  fixtureMode: z.boolean(),
  version: z.string(),
  limits: z.object({
    workers: z.number(),
    queue: z.number(),
    maxBytes: z.number(),
    retentionSeconds: z.number(),
  }),
  dependencies: z.object({ ytDlp: z.boolean(), ffmpeg: z.boolean() }),
});
export type Analysis = z.infer<typeof analysisSchema>;
export type Option = z.infer<typeof option>;
export type Job = z.infer<typeof jobSchema>;
export type Status = z.infer<typeof statusSchema>;
export class ApiError extends Error {
  constructor(
    public readonly code: string,
    message: string,
  ) {
    super(message);
  }
}
export async function request<T>(
  path: string,
  schema: z.ZodType<T>,
  init?: RequestInit,
): Promise<T> {
  const response = await fetch(`/api/v1${path}`, {
    ...init,
    credentials: "same-origin",
    cache: "no-store",
    headers: { "Content-Type": "application/json", ...init?.headers },
  });
  if (!response.ok) {
    const payload: unknown = await response.json().catch(() => null);
    const error = z
      .object({
        error: z.object({
          code: z.string().max(80),
          message: z.string().max(500),
        }),
      })
      .safeParse(payload);
    throw new ApiError(
      error.success ? error.data.error.code : "operation_failed",
      error.success
        ? error.data.error.message
        : "The server could not complete this request. Try again shortly.",
    );
  }
  if (response.status === 204) return schema.parse(undefined);
  const result = schema.safeParse(await response.json().catch(() => null));
  if (!result.success)
    throw new ApiError(
      "unexpected_response",
      "The server returned an unexpected response. Refresh to reconnect.",
    );
  return result.data;
}
export const jobListSchema = z.object({ jobs: z.array(jobSchema) });
export const emptySchema = z.undefined();
