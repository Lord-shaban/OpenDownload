import { describe, expect, it } from "vitest";
import { ApiError } from "./api";
import { direction, messages, parseLocale, translate } from "./i18n";
import { errorMessage, mediaDetail, mediaLabel } from "./localized-media";
import { bytes, detectSource, displayText, expiryLabel } from "./media";
describe("locale contracts", () => {
  it("falls back safely and keeps both catalogs' placeholders aligned", () => {
    expect(parseLocale("../../bad")).toBe("en");
    expect(direction(parseLocale("ar"))).toBe("rtl");
    for (const [english, arabic] of Object.values(messages)) {
      expect(arabic.trim().length).toBeGreaterThan(0);
      expect(arabic.match(/\{\w+\}/g)?.sort() || []).toEqual(
        english.match(/\{\w+\}/g)?.sort() || [],
      );
    }
  });
  it("translates known and unknown API errors without exposing raw exceptions", () => {
    expect(
      errorMessage(
        new ApiError("analysis_expired", "backend diagnostic"),
        "en",
      ),
    ).toBe(translate("en", "analysisExpired"));
    expect(
      errorMessage(new ApiError("future_error", "internal secret"), "en"),
    ).toBe(translate("en", "requestFailed"));
    expect(
      errorMessage(new ApiError("analysis_expired", "English message"), "ar"),
    ).toBe(translate("ar", "analysisExpired"));
    expect(
      errorMessage(new ApiError("future_error", "internal secret"), "ar"),
    ).toBe(translate("ar", "requestFailed"));
    expect(errorMessage(new TypeError("raw fetch exception"), "en")).toBe(
      translate("en", "networkError"),
    );
  });
  it("keeps actual qualities and codecs while translating product labels", () => {
    expect(mediaLabel("720p", "ar")).toBe("720p");
    expect(mediaLabel("2 original images", "ar")).toBe("صور أصلية: 2");
    expect(mediaDetail("MP4 · avc1 · audio included", "ar")).toBe(
      "MP4 · avc1 · مع الصوت",
    );
    expect(bytes(0, "ar")).toBe("الحجم متغير");
    expect(
      expiryLabel(
        "2026-09-29T15:01:00Z",
        Date.parse("2026-09-29T15:00:00Z"),
        "ar",
      ),
    ).toBe("الوقت المتبقي: 1 د");
  });
  it("distinguishes server refusals from restricted content in both languages", () => {
    for (const locale of ["ar", "en"] as const) {
      for (const [code, key] of [
        ["platform_verification_required", "platformVerificationRequired"],
        ["upstream_forbidden", "upstreamForbidden"],
        ["upstream_rate_limited", "upstreamRateLimited"],
        ["source_access_denied", "sourceAccessDenied"],
        ["source_connection_failed", "sourceConnectionFailed"],
        ["source_metadata_unavailable", "sourceMetadataUnavailable"],
      ] as const) {
        const message = errorMessage(new ApiError(code, "private diagnostic token=secret"), locale);
        expect(message).toBe(translate(locale, key));
        expect(message).not.toContain("secret");
        expect(message).not.toBe(translate(locale, "sourceUnavailable"));
      }
    }
  });
  it("rejects raw and encoded bidi controls without rejecting Arabic URLs", () => {
    for (const control of ["\u202e", "\u2066", "\u200f", "\u061c"]) {
      expect(
        detectSource(`https://exam${encodeURIComponent(control)}ple.com/photo`)
          .valid,
      ).toBe(false);
      expect(detectSource(`https://example.com/${control}photo`).valid).toBe(
        false,
      );
      expect(
        detectSource(`https://example.com/?q=${encodeURIComponent(control)}`)
          .valid,
      ).toBe(false);
    }
    expect(detectSource("https://example.com/صور?title=عنوان").valid).toBe(
      true,
    );
    expect(displayText("title\u202emp4")).toBe("titlemp4");
    // Non-UTF-8 percent escapes are legal URL bytes, unrelated to bidi controls.
    expect(detectSource("https://example.com/photo%E9.jpg").valid).toBe(true);
  });
});
