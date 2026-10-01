"""Apply a small reviewed startup patch to pinned nodriver 0.50.3."""
import importlib.metadata
import hashlib
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

# Bound WPC's entire launch, not just nodriver's version endpoint polling.
assert importlib.metadata.version("yt-dlp-getpot-wpc") == "1.1.2"
wpc = pathlib.Path(importlib.metadata.distribution("yt-dlp-getpot-wpc").locate_file(
    "yt_dlp_plugins/extractor/getpot_wpc.py"))
text = wpc.read_text(encoding="utf8")
mint_begin = text.index("async def mint_po_token(")
mint_end = text.index("async def launch_browser(config):", mint_begin)
mint_original = text[mint_begin:mint_end]
assert hashlib.sha256(mint_original.encode()).hexdigest() == "b927ca298b3c4949ef35bbe3c285d06991ae914c5ab0814778596d367815829d", "Pinned WPC mint changed; review the patch."
mint_patched = mint_original.replace("async def mint_po_token(", "async def _mint_po_token(", 1)
mint_patched = mint_patched.replace("    webpo_client_path = await get_webpo_client_path(tab, logger)",
    '    logger.debug("Guest attestation stage: client-lookup")\n    webpo_client_path = await get_webpo_client_path(tab, logger)', 1)
mint_patched = mint_patched.replace("        po_token = await tab.evaluate(mint_po_token_code, await_promise=True)",
    '        logger.debug("Guest attestation stage: mint-request")\n        po_token = await tab.evaluate(mint_po_token_code, await_promise=True)', 1)
mint_patched += '''async def mint_po_token(tab, logger, content_binding, mint_cold_start_token=False, mint_error_token=False):
    from opendownload_browser_session import mint_token
    try:
        return await mint_token(_mint_po_token, tab=tab, logger=logger, content_binding=content_binding,
            mint_cold_start_token=mint_cold_start_token, mint_error_token=mint_error_token)
    except asyncio.TimeoutError as error:
        raise PoTokenProviderError('guest attestation timed out') from error


'''
text = text[:mint_begin] + mint_patched + text[mint_end:]
begin = text.index("async def launch_browser(config):")
end = text.index("@register_provider", begin)
original = text[begin:end]
assert hashlib.sha256(original.encode()).hexdigest() == "7f3ffb5743ffcee1df238eca5e34a9f72a3adbc77de771dbd11f8ecdf39b86b8", "Pinned WPC launch changed; review the patch."
text = text[:begin] + '''async def launch_browser(config, logger):
    from opendownload_browser_session import launch_browser as launch_session
    try:
        return await launch_session(config, start, nodriver.cdp, logger)
    except Exception as e:
        raise PoTokenProviderError(f'guest browser launch failed: {type(e).__name__}') from e


''' + text[end:]
old = "launch_browser(browser_config)"
assert text.count(old) == 1, "Pinned WPC call changed; review the patch."
text = text.replace(old, "launch_browser(browser_config, self.logger)")
compile(text, str(wpc), "exec")
wpc.write_text(text, encoding="utf8")
