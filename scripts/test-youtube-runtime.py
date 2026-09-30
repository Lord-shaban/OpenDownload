"""Regression checks for the image-owned YouTube entrypoint; no network."""
import importlib.util
import pathlib
import sys
import unittest

sys.dont_write_bytecode = True

spec = importlib.util.spec_from_file_location("youtube_runtime", pathlib.Path(__file__).resolve().parents[1] / "deploy/youtube-runtime.py")
runtime = importlib.util.module_from_spec(spec)
spec.loader.exec_module(runtime)


class RuntimeTests(unittest.TestCase):
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
