# Localization

The workspace supports English (`en`, default) and Arabic (`ar`). The header
selector updates labels and direction without clearing the current URL, selected
format or jobs. A language-only `od_locale` cookie persists the preference for one
year. It stores no media URLs, jobs or credentials.

The server validates that cookie and renders `html[lang]`, `html[dir]` and localized
metadata. Reading the cookie makes the root route dynamic. The client changes the
document title when switching; the metadata description updates on reload.

## Maintaining messages

`apps/web/src/lib/i18n.ts` is the typed English/Arabic catalog. Add both messages
with matching named placeholders. Keep user-facing labels in the catalog and use
`useLocale().t` in components. API errors map stable codes to Arabic messages;
unknown codes use a safe recovery message in both languages. Stored job failures
currently expose only a message; the UI uses a short retry instruction in both
languages. Specific localized job failure reasons would require stable persisted
error codes.

Use logical margins, padding, borders and positioning for mirrored layout. URL
inputs and technical fields stay LTR. Wrap unknown-direction titles with `bdi`;
strip untrusted formatting controls from displayed titles. Accessible string-only
labels use Unicode isolates. Submitted URLs containing bidi controls are rejected
by both the browser and Go URL policy; ordinary Arabic paths/queries are accepted.
See [W3C inline bidi guidance](https://www.w3.org/International/articles/inline-bidi-markup/).

IBM Plex Sans Arabic 400/500/600 is bundled locally as WOFF2; the browser never
requests it from a font CDN. Its full OFL accompanies standalone assets. See
[third-party notices](../THIRD_PARTY_NOTICES.md).

## Theme preference

The header theme button saves `od_theme=light` or `dark` for one year, with
`SameSite=Lax`, path `/` and `Secure` on HTTPS. The server renders the corresponding
root class and initial button state, preventing a preference flash on reload.
The cookie contains only the theme. Glass has a solid fallback when backdrop
blur is unavailable or `prefers-reduced-transparency: reduce` is supported and set.

## Verification

Run the unit and production browser gates described in [testing](TESTING.md).
`e2e/locale.spec.ts` verifies persistence, Arabic keyboard download/delete flows,
errors, RTL reflow, dialog focus and local fonts on desktop and mobile. Include
Arabic in visual review whenever changing layout or mixed-direction fields.
`e2e/experience.spec.ts` also checks theme persistence/server output, step feedback
and focus through Enter → analysis → download → new link.
