"""Bound the complete guest-browser launch; never load an account profile."""
import asyncio


async def mint_token(mint, *args, timeout=30, **kwargs):
    """Bound a provider's promise; never log the returned token or binding."""
    try:
        return await asyncio.wait_for(mint(*args, **kwargs), timeout)
    except asyncio.TimeoutError:
        kwargs["logger"].debug("Guest attestation deadline reached")
        raise


async def launch_browser(config, start, cdp, logger, url="https://www.youtube.com?themeRefresh=1", timeout=30):
    browser = None
    stage = "browser-start"

    async def session():
        nonlocal browser, stage
        logger.debug("Guest browser stage: browser-start")
        browser = await start(config=config)
        stage = "window-lookup"
        logger.debug("Guest browser stage: window-lookup")
        window_id, _ = await browser.main_tab.get_window()
        stage = "window-minimize"
        logger.debug("Guest browser stage: window-minimize")
        await browser.main_tab.send(cdp.browser.set_window_bounds(
            window_id=window_id,
            bounds=cdp.browser.Bounds(window_state=cdp.browser.WindowState.MINIMIZED)))
        stage = "guest-cookie-clear"
        logger.debug("Guest browser stage: guest-cookie-clear")
        await browser.cookies.clear()
        stage = "page-navigation"
        logger.debug("Guest browser stage: page-navigation")
        await browser.get(url)
        stage = "ready"
        logger.debug("Guest browser stage: ready")
        return browser

    try:
        return await asyncio.wait_for(session(), timeout)
    except BaseException as error:
        # Emit only a fixed stage and exception class, never cookies, page data,
        # tokens or URLs. The extractor's optional diagnostics remain encrypted.
        logger.debug(f"Guest browser failed at {stage}: {type(error).__name__}")
        if stage == "browser-start" and isinstance(error, RuntimeError):
            startup_error = str(error)
            if startup_error.startswith("Chromium startup failed: "):
                # The pinned nodriver patch supplies at most 4096 bytes from
                # Chromium startup, before any page is opened. Provider debug
                # output is enabled only for encrypted operator diagnostics.
                logger.debug(startup_error[:4096])
        if browser is not None:
            try:
                browser.stop()
            except Exception:
                # Preserve the launch/cancellation cause. The parent owns the
                # process group and the disposable profile's final cleanup.
                logger.debug("Guest browser cleanup failed")
        raise
