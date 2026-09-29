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
target sizes were adapted. shadcn is licensed under MIT:

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
