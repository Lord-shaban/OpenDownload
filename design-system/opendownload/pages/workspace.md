# OpenDownload workspace — final product override

This page overrides the generated MASTER.md recommendations for the app.
UI UX Pro Max's flat design, restrained motion and progress guidance apply. Its
marketing hero/video layout and orange/teal palette do not fit this utility.

## Identity

Quiet, precise, useful. An editorial utility workspace with generous breathing
room, thin rules, warm white canvas, dark olive text and one evergreen action color.
No decorative gradients, glass panels, fake metrics, testimonials or landing-page
sections. Make the URL action the first useful element.

## Tokens

| Token          | Light   | Dark    |
| -------------- | ------- | ------- |
| Canvas         | #f6f7f3 | #121711 |
| Surface        | #ffffff | #1a2119 |
| Text           | #202b22 | #eef2eb |
| Secondary text | #59655b | #a7b3a6 |
| Border         | #dce3d9 | #354132 |
| Primary        | #28613e | #a2d4a5 |
| Primary text   | #ffffff | #142719 |
| Subtle primary | #eaf1e7 | #253826 |
| Error          | #a52b2b | #ffa6a6 |
| Focus          | #337348 | #b8dfad |

Typography: locally bundled Geist Sans (body/display) and Geist Mono (technical
metadata). Body 16/24; secondary 13/20; title clamp(32, 4vw, 48)/1.12, -0.045em.
Tabular numerals for progress. Never fetch fonts from Google at page load.

Spacing: 4, 8, 12, 16, 24, 32, 48, 64. Desktop rail 230, main content max 1000.
Radius: controls 8, cards 14, pills 999. One subtle raised form shadow.

## Layout

Desktop: left utility rail with brand, workspace/downloads/help, and self-host info.
Main header with source status and theme switch. Central URL form and horizontal
source chips. Below: stable media selection panel or helpful first-use steps.
Completed/active jobs use compact rows; empty state is purposeful, not fake data.

Mobile: header replaces rail; 16–20px margins; navigation has text labels; input
and submit stack; quality cards remain large targets; no horizontal overflow.
Test 375, 768, 1024, 1440. Respect safe-area insets.

## States and interactions

Detect locally as URL changes; display a factual source hint. Explicit Analyze
avoids sending pasted data without intent. Errors sit next to the form. A canceled
analysis aborts the request. Metadata reveal uses 180ms fade/6px translation;
format selection updates immediately with clear selected state. Queue status has
unknown-progress handling, separate processing phase and explicit cancellation.
Show expiry beside completed files and surface storage deletion.

Motion: hover colors 140ms, panel enter 200ms, progress interpolation 220ms.
Animate transform/opacity, not layout. No infinite decorative effects. Disable
motion under prefers-reduced-motion. No mandatory motion library.

Accessibility: native forms and labelled controls, 44px targets, visible 2px focus,
screen-reader status announcements only on meaningful changes, keyboard dialogs,
errors with recovery actions, selected states shown with text/icons as well as color.
Contrast/reflow checks required; do not claim WCAG certification.

## Arabic and direction

English and Arabic are selectable in the header. The preference persists in a
language-only cookie and sets the initial document language/direction on reload.
Use local Noto Sans Arabic for Arabic body/headings, normal letter spacing and
1.7 body line height; small Arabic captions have a 12px minimum. Keep the Latin
brand in Geist. The rail moves to the right using logical spacing/borders.

URLs stay LTR. Isolate titles, platform names, filenames and codec/quality values
with `bdi`/explicit direction; use Unicode isolation only where accessibility
labels cannot contain markup. Mirror directional arrows, not media icons.
Dialogs scroll within the viewport. Verify Arabic keyboard selection, focus
restoration, short landscape viewports, light/dark themes and 375–1440px reflow.
