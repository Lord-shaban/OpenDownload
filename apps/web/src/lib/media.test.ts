import { describe, expect, it } from "vitest";
import { bytes, detectSource, duration } from "./media";
import { analysisSchema, jobSchema } from "./api";
describe("public source detection", () => {
  it.each([
    ["https://www.linkedin.com/posts/creator-video-123-abcd", "LinkedIn"],
    ["https://lnkd.in/abc", "LinkedIn"],
    ["https://pinterest.co.uk/pin/123/", "Pinterest"],
    ["https://www.pinterest.com/pin/123/", "Pinterest"],
    ["https://pin.it/abc", "Pinterest"],
    ["https://www.threads.com/@creator/post/ABC", "Threads"],
    ["https://threads.net/@creator/post/ABC", "Threads"],
  ])("recognizes %s", (url, name) => {
    expect(detectSource(url)).toEqual({ name, valid: true });
    expect(
      detectSource(
        url.replace(
          new URL(url).hostname,
          `${new URL(url).hostname}.attacker.net`,
        ),
      ).name,
    ).not.toBe(name);
  });
  it("uses exact hostname boundaries", () => {
    expect(detectSource("https://www.youtube.com/watch?v=1").name).toBe(
      "YouTube",
    );
    expect(detectSource("https://youtube.com.attacker.net/a").name).toBe(
      "youtube.com.attacker.net",
    );
  });
  it("rejects unsafe schemes and credentials", () => {
    expect(detectSource("file:///tmp/video").valid).toBe(false);
    expect(detectSource("https://user:pass@example.com").valid).toBe(false);
  });
});
describe("display helpers", () => {
  it("handles missing sizes and short durations", () => {
    expect(bytes(0)).toBe("Size varies");
    expect(bytes(1048576)).toBe("1.0 MB");
    expect(duration(154)).toBe("2:34");
  });
});
describe("network response validation", () => {
  it("rejects corrupt analysis and unknown job states", () => {
    expect(
      analysisSchema.safeParse({ id: "bad", options: "anything" }).success,
    ).toBe(false);
    expect(jobSchema.safeParse({ state: "invented" }).success).toBe(false);
  });
});
