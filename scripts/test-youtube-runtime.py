"""Regression checks for the image-owned YouTube entrypoint; no network."""
import importlib.util
import asyncio
import pathlib
import sys
from types import SimpleNamespace
import unittest

sys.dont_write_bytecode = True

spec = importlib.util.spec_from_file_location("youtube_runtime", pathlib.Path(__file__).resolve().parents[1] / "deploy/youtube-runtime.py")
runtime = importlib.util.module_from_spec(spec)
spec.loader.exec_module(runtime)

browser_spec = importlib.util.spec_from_file_location("chromium_egress", pathlib.Path(__file__).resolve().parents[1] / "deploy/chromium-egress.py")
chromium = importlib.util.module_from_spec(browser_spec)
browser_spec.loader.exec_module(chromium)

session_spec = importlib.util.spec_from_file_location("youtube_browser_session", pathlib.Path(__file__).resolve().parents[1] / "deploy/youtube-browser-session.py")
session = importlib.util.module_from_spec(session_spec)
session_spec.loader.exec_module(session)


class SessionTests(unittest.IsolatedAsyncioTestCase):
    async def test_preserves_only_bounded_chromium_startup_failure(self):
        for message, expected in (
            ("Chromium startup failed: " + "x" * 5000, 4096),
            ("other exception content", None),
        ):
            stages = []

            async def start(**_):
                raise RuntimeError(message)

            with self.assertRaises(RuntimeError):
                await session.launch_browser(None, start, None, SimpleNamespace(debug=stages.append))
            self.assertIn("Guest browser failed at browser-start: RuntimeError", stages)
            if expected is None:
                self.assertNotIn(message, stages)
            else:
                self.assertEqual(stages[-1], message[:expected])

    async def test_deadline_covers_cookie_clear_after_driver_start(self):
        stages = []
        stopped = []
        cookie_started = asyncio.Event()

        async def window():
            return 1, None

        async def send(_):
            return None

        async def clear():
            cookie_started.set()
            await asyncio.Event().wait()

        browser = SimpleNamespace(main_tab=SimpleNamespace(get_window=window, send=send),
            cookies=SimpleNamespace(clear=clear), stop=lambda: stopped.append(True))

        async def start(**_):
            return browser

        cdp = SimpleNamespace(browser=SimpleNamespace(set_window_bounds=lambda **_: None,
            Bounds=lambda **_: None, WindowState=SimpleNamespace(MINIMIZED="minimized")))
        with self.assertRaises(asyncio.TimeoutError):
            await session.launch_browser(None, start, cdp, SimpleNamespace(debug=stages.append), timeout=0.1)
        self.assertTrue(cookie_started.is_set())
        self.assertEqual(stopped, [True])
        self.assertEqual(stages[-1], "Guest browser failed at guest-cookie-clear: TimeoutError")

    async def test_request_cancellation_closes_launched_browser(self):
        started = asyncio.Event()
        stopped = []

        async def window():
            started.set()
            await asyncio.Event().wait()

        browser = SimpleNamespace(main_tab=SimpleNamespace(get_window=window), stop=lambda: stopped.append(True))

        async def start(**_):
            return browser

        task = asyncio.create_task(session.launch_browser(None, start, None, SimpleNamespace(debug=lambda _: None)))
        await asyncio.wait_for(started.wait(), 1)
        task.cancel()
        with self.assertRaises(asyncio.CancelledError):
            await task
        self.assertEqual(stopped, [True])


class RuntimeTests(unittest.TestCase):
    def test_browser_profile_uses_only_image_owned_wpc_plugin(self):
        args = ["--no-plugin-dirs", "--proxy", "http://127.0.0.1:8090", "--", "https://youtu.be/abc"]
        actual = runtime.attestation_args(args, True)
        self.assertEqual(actual[-2:], args[-2:])
        self.assertIn("/opt/youtube-wpc-plugins", actual)
        self.assertIn("youtubepot-wpc:browser_path=/usr/local/bin/opendownload-chromium", actual)
        self.assertNotIn("/opt/youtube-plugins", actual)

    def test_browser_cannot_bypass_proxy_or_grant_remote_debugging(self):
        actual = chromium.chromium_args(["--proxy-server=http://evil.example", "--proxy-bypass-list=*", "--remote-allow-origins=*", "--disable-features=IsolateOrigins,site-per-process", "--remote-debugging-port=12345"])
        self.assertNotIn("--proxy-bypass-list=*", actual)
        self.assertNotIn("--remote-allow-origins=*", actual)
        self.assertNotIn("--disable-features=IsolateOrigins,site-per-process", actual)
        self.assertIn("--proxy-server=http://127.0.0.1:8090", actual)
        self.assertIn("--proxy-bypass-list=<-loopback>", actual)
        self.assertIn("--remote-debugging-address=127.0.0.1", actual)
        self.assertIn("--site-per-process", actual)

    def test_only_explicit_youtube_sources_enable_plugin(self):
        for url in ("https://youtu.be/abc", "https://www.youtube.com/watch?v=abc", "https://m.youtube.com/shorts/abc"):
            self.assertTrue(runtime.youtube_request(["--no-plugin-dirs", "--", url]))
        for args in (["--version"], ["https://youtu.be/abc"], ["--", "https://youtube.com.evil.example/watch?v=abc"],
                     ["--", "https://notyoutube.com/abc"], ["--", "https://youtube.com@evil.example/abc"],
                     ["--", "https://account:secret@youtube.com/abc"], ["--", "https://www.tiktok.com/video/abc"]):
            self.assertFalse(runtime.youtube_request(args))

    def test_proxy_and_source_preserved_and_user_plugins_disabled(self):
        args = ["--ignore-config", "--no-plugin-dirs", "--proxy", "http://127.0.0.1:8090", "--", "https://youtu.be/abc"]
        result = runtime.attestation_args(args)
        self.assertEqual(result[:4], args[:4])
        self.assertEqual(result[-2:], args[-2:])
        self.assertEqual(result[result.index("--plugin-dirs") + 1], "/opt/youtube-plugins")
        self.assertNotIn("default", result)
        self.assertEqual(result[result.index("--impersonate") + 1], "chrome")
        self.assertIn("youtube:player_client=mweb;fetch_pot=always", result)

    def test_other_platforms_keep_original_arguments(self):
        args = ["--no-plugin-dirs", "--", "https://www.tiktok.com/video/abc"]
        self.assertEqual(runtime.run(args, lambda actual: actual), args)


if __name__ == "__main__":
    unittest.main()
