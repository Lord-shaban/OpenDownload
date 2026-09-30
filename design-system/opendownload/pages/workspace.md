# OpenDownload workspace — minimal glass

This page overrides the generated MASTER.md. UI UX Pro Max's Glassmorphism style
lookup informed material fallbacks and contrast checks. The user refined the
brief on 2026-09-29: fewer words, a calmer palette, better icons/type/motion and
an experience as direct as [cobalt.tools](https://cobalt.tools/).

## Identity and composition

The logo is **OpenDownload.**, with one colored period and no pictogram. Keep it
LTR in both locales. A quiet header contains the wordmark, language, theme and
Help. Two small tabs switch between New download and Downloads. The main surface
is a single centered link field with a Paste button. No marketing headline,
visible stepper, platform inventory or repeated instructions.

Analysis results appear only when needed. Display the real media title, available
media types and actual quality/container/size choices. Codec details remain in
radio labels for assistive technology. Keep explanatory detail in Help, not the
main flow. Downloads only appear in the Downloads view.

## Tokens

| Token | Light | Dark |
| --- | --- | --- |
| Canvas | #f4f4f6 | #141418 |
| Solid surface | #fdfdfe | #222228 |
| Text | #26272d | #ededf1 |
| Secondary text | #656570 | #a5a4b0 |
| Primary | #5b57c9 | #b6afff |
| Primary text | #ffffff | #25213e |
| Secondary | #eae9f7 | #34313f |
| Secondary text on tint | #514ca6 | #c8c1ff |
| Border | #d6d6df | #41414c |
| Focus | #6963dc | #b6afff |
| Error | #af3650 | #f4a0af |
| Success | #2e7060 | #95cfb7 |

Glass: white at 68% light / rgb(36,36,43) at 76% dark; blur 16px, a fine edge and
soft shadow. One faint violet halo provides depth. Use glass on the composer,
results and job list; the header and navigation have no glass boxes. Solid cards
are the fallback for unsupported blur and reduced transparency. No animated
filters or decorative continuous motion.

Geist Sans for English and the wordmark, Geist Mono for technical values. IBM
Plex Sans Arabic 400/500/600 is bundled locally as WOFF2, with its OFL. Arabic
body line-height 1.6; no decorative tracking. Input 17px desktop / 16px mobile;
controls 14px, captions 12px, brand 20/23px (18px on narrow phones).

Shell max 1040px; tool max 620px. Desktop gutters 32px, mobile 20px (14px below
360px). Panel radius 22px, actions 14–16px. All interactive targets at least 44px;
primary actions 48px. On mobile the download action spans the panel width.

## Interaction contract

- Clicking Paste or pasting a complete valid URL initiates analysis immediately.
  This is a deliberate paste action; typing never makes background requests.
  Enter/the arrow submits typed links. Clipboard denial gives keyboard recovery
  and focuses the input. Each replacement cancels the previous analysis.
- Loading shows a small skeleton and one short status. Cancellation stays visible.
  Results focus their region, with a useful format already selected. Download
  remains a deliberate action; no fabricated presets or automatic queued files.
- Creating a job opens Downloads and focuses the main region. Actual progress,
  retries, failures, expiry and Save to device remain visible. New download and
  the slash shortcut return focus to the link field.
- English/Arabic and theme preferences persist without storing URLs. Render both
  on the server. URL/file/codec fields remain bidi-safe and URLs stay LTR.
- Messages are short, calm and original: “One moment. Finding your formats…”,
  “Your file is ready”, “ملفك في الطريق…”. Errors include a useful next action.
  Access boundaries and fixture identity always remain explicit where relevant.

## Motion and icons

Phosphor Icons 2.1.10 supplies the core controls, with duotone media/clipboard/theme
icons and plain action arrows/checks. Import individual CSR modules to avoid
loading the entire catalog. shadcn primitives retain their small utility icons.

Panel/view entrance: opacity and 8px translate over 340ms using
cubic-bezier(0.16,1,0.3,1). Selected check: 220ms; tab marker: 240ms; pressed controls
scale to 0.97; progress interpolates over 360ms. Dialog/overlay entrance uses
260/180ms. No animation changes actual progress or delays a request. Reduced
motion disables all transitions/animations. Skeleton bars stay static.

## Verification

Review English/Arabic, both themes, 320–1440px reflow, short landscape dialogs,
keyboard focus and local fonts. Production browser gates cover clipboard/native
paste, denied clipboard recovery, typed Enter, preference persistence and the
complete analysis/download/save/delete flow. Calculate contrast on composed
surfaces. This is a development review, not WCAG certification.

The supported-sites disclosure sits below Paste, collapsed by default. It lists
extractor source hints with monochrome icons, direct public links and a short
availability/photo-gallery limitation. It is keyboard operable as native details.
