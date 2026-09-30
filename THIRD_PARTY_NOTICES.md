# Third-party software

OpenDownload's original code uses MIT. Dependencies keep their own licenses:
yt-dlp (Unlicense and bundled dependency licenses), FFmpeg (LGPL/GPL depending on
the distribution build), Next.js/React/shadcn (MIT), Go (BSD-style), modernc SQLite
(BSD-style), and SQLite (public domain). Refer to each project's actual license
and the resolved package/image distribution before redistribution.

`apps/web/src/components/ui/button.tsx` and `dialog.tsx` are adapted from the
official [shadcn registry](https://ui.shadcn.com), downloaded on 2026-09-29.
CLI installation failed in the workstation package-manager environment; only
used components were copied through `scripts/sync-shadcn.py`. Imports and control
target sizes, controlled dialog focus restoration, translated labels, logical
placement, bounded dialog height and semantic destructive theme colors were
adapted. shadcn is licensed under MIT:

Copyright (c) 2023 shadcn

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## IBM Plex Sans Arabic

Unmodified Regular, Medium and SemiBold WOFF2 files from the official
[IBM Plex repository](https://github.com/IBM/plex/tree/763c36ef9117782905ae010056dfbe8fd2653a25/packages/plex-sans-arabic),
retrieved on 2026-09-29 at commit `763c36ef9117782905ae010056dfbe8fd2653a25`.
Copyright © 2017 IBM Corp. with Reserved Font Name "Plex".
Licensed under SIL Open Font License 1.1. The full license ships in standalone/
container assets at `apps/web/public/licenses/IBMPlexSansArabic-OFL.txt`.

## Phosphor Icons

Core workspace controls use `@phosphor-icons/react` 2.1.10, from the official
[Phosphor repository](https://github.com/phosphor-icons/react). Copyright (c) 2020
Phosphor Icons. MIT; the full license is included in standalone/container assets
at `apps/web/public/licenses/PhosphorIcons-MIT.txt`. Only individual icon modules
are imported. No Cobalt artwork or source code was copied.
