"""Apply a small reviewed startup patch to pinned nodriver 0.50.3."""
import importlib.metadata
import pathlib

assert importlib.metadata.version("nodriver") == "0.50.3"
source = pathlib.Path(importlib.metadata.distribution("nodriver").locate_file("nodriver/core/browser.py"))
text = source.read_text(encoding="utf8")
replacements = {
    """        for _ in range(5):
            try:
                self.info = ContraDict(await self._http.get("version"), silent=True)

            except (Exception,):
                if _ == 4:
                    logger.debug("could not start", exc_info=True)
                await asyncio.sleep(0.5)
            else:
                break
""": """        startup_deadline = asyncio.get_running_loop().time() + 15
        for _ in range(30):
            remaining = startup_deadline - asyncio.get_running_loop().time()
            if remaining <= 0:
                break
            try:
                info = await asyncio.wait_for(self._http.get("version"), min(1, remaining))
                self.info = ContraDict(info, silent=True)
            except (Exception,):
                remaining = startup_deadline - asyncio.get_running_loop().time()
                await asyncio.sleep(min(0.5, max(0, remaining)))
            else:
                break
""",
    "urllib.request.urlopen(request, timeout=10)": "urllib.request.urlopen(request, timeout=1 if endpoint == 'version' else 10)",
    "        if not self.info:\n            raise Exception(": """        if not self.info:
            # Preserve a bounded startup failure for the extractor's encrypted
            # operator diagnostics. Never expose it in visitor responses.
            process = self._process
            if process is not None and process.stderr is not None:
                try:
                    startup_error = await asyncio.wait_for(process.stderr.read(4096), 0.2)
                    if startup_error:
                        raise RuntimeError('Chromium startup failed: ' + startup_error.decode('utf8', errors='replace'))
                except asyncio.TimeoutError:
                    pass
            raise Exception(""",
}
for old, new in replacements.items():
    assert text.count(old) == 1, "Pinned browser source changed; review the patch."
    text = text.replace(old, new)
compile(text, str(source), "exec")
source.write_text(text, encoding="utf8")
