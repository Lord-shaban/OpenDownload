import { ApiError } from "./api";
import { messages, translate, type Locale, type MessageKey } from "./i18n";
const codes: Record<string, MessageKey> = {
  origin_denied: "originDenied",
  json_required: "invalidRequest",
  invalid_request: "invalidRequest",
  rate_limited: "rateLimited",
  unsafe_url: "unsafeUrl",
  analysis_busy: "analysisBusy",
  analysis_timeout: "analysisTimeout",
  source_unavailable: "sourceUnavailable",
  platform_verification_required: "platformVerificationRequired",
  upstream_forbidden: "upstreamForbidden",
  upstream_rate_limited: "upstreamRateLimited",
  source_access_denied: "sourceAccessDenied",
  source_connection_failed: "sourceConnectionFailed",
  analysis_expired: "analysisExpired",
  file_too_large: "fileTooLarge",
  invalid_format: "invalidFormat",
  job_expired: "jobExpired",
  not_found: "notFound",
  queue_full: "queueFull",
  invalid_state: "invalidState",
  operation_failed: "requestFailed",
  file_unavailable: "fileUnavailable",
  unexpected_response: "unexpectedResponse",
};
export function errorMessage(error: unknown, locale: Locale) {
  if (typeof error === "string" && Object.hasOwn(messages, error))
    return translate(locale, error as MessageKey);
  if (error instanceof ApiError) {
    return translate(locale, codes[error.code] || "requestFailed");
  }
  return translate(locale, "networkError");
}
const labels: MessageKey[] = [
  "originalVideo",
  "originalAudio",
  "originalImage",
  "mp3Audio",
  "thumbnail",
];
const details: MessageKey[] = [
  "boundedCollection",
  "unchanged",
  "sourceFile",
  "originalSource",
  "qualityUnknown",
  "codecUnknown",
  "audioIncluded",
  "originalNoConversion",
  "converted",
  "convertedSource",
  "upToBitrate",
  "coverImage",
  "originalCaptions",
  "automaticCaptions",
];
export function mediaLabel(label: string, locale: Locale) {
  const gallery = /^(\d+) original images$/.exec(label);
  if (gallery)
    return translate(locale, "originalImages", { count: gallery[1] });
  const key = labels.find((key) => messages[key][0] === label);
  return key ? translate(locale, key) : label;
}
export function mediaDetail(detail: string, locale: Locale) {
  if (detail === messages.requiresAudio[0])
    return translate(locale, "requiresAudio");
  return detail
    .split(" · ")
    .map((part) => {
      const key = details.find((key) => messages[key][0] === part);
      return key ? translate(locale, key) : part;
    })
    .join(" · ");
}
