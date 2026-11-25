#
# Copyright 2026 Canonical, Ltd.
#
import argparse
import io
import json
import sys
import unittest
from contextlib import redirect_stdout
from email.message import Message
from email.utils import formatdate
from pathlib import Path
from unittest.mock import MagicMock, patch
from urllib.error import HTTPError, URLError

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from cmds import charm, store  # noqa: E402


def _channel_entry(name: str, track: str, risk: str, arch: str = "amd64") -> dict:
    return {
        "channel": {
            "architecture": arch,
            "name": name,
            "released-at": "2026-01-15T10:00:00Z",
            "risk": risk,
            "track": track,
        }
    }


def _json_response(payload: dict) -> MagicMock:
    body = io.BytesIO(json.dumps(payload).encode("utf-8"))
    resp = MagicMock()
    resp.__enter__.return_value = body
    resp.__exit__.return_value = False
    return resp


class StoreQueryTests(unittest.TestCase):
    def test_query_info_sends_user_agent_and_series_headers(self):
        payload = {
            "channel-map": [
                _channel_entry("1.35-classic/stable", "1.35-classic", "stable")
            ]
        }
        with patch.object(store, "urlopen", return_value=_json_response(payload)) as m:
            result = store.query_info("snap", "k8s")

        self.assertEqual(result, payload)
        req = m.call_args[0][0]
        headers = {k.lower(): v for k, v in req.header_items()}
        self.assertEqual(headers.get("snap-device-series"), "16")
        self.assertEqual(headers.get("user-agent"), "Mozilla/5.0")

    def test_query_info_honors_retry_after_and_closes_429_response(self):
        hdrs = Message()
        hdrs["Retry-After"] = "12"
        err_fp = MagicMock()
        rate_limited = HTTPError(
            "https://api.snapcraft.io/v2/snaps/info/k8s?fields=channel-map",
            429,
            "Too Many Requests",
            hdrs,
            err_fp,
        )
        payload = {"channel-map": []}
        with patch.object(
            store, "urlopen", side_effect=[rate_limited, _json_response(payload)]
        ), patch.object(store.time, "sleep") as sleep_mock:
            result = store.query_info("snap", "k8s")

        self.assertEqual(result, payload)
        err_fp.close.assert_called_once()
        sleep_mock.assert_called_once()
        self.assertGreaterEqual(sleep_mock.call_args[0][0], 12.0)

    def test_query_info_caps_large_retry_after_and_honors_http_date(self):
        self.assertEqual(store._retry_delay(1, "3600"), store._MAX_RETRY_DELAY)
        with patch.object(store.time, "time", return_value=1_700_000_000.0):
            future_date = formatdate(1_700_000_018.0, usegmt=True)
            delay = store._retry_delay(1, future_date)
            past_date = formatdate(1_699_999_990.0, usegmt=True)
            past_delay = store._retry_delay(1, past_date)
        self.assertEqual(delay, 18.0)
        self.assertGreaterEqual(past_delay, 2.0)
        self.assertLess(past_delay, 4.0)

    def test_query_info_does_not_retry_and_closes_non_transient_http_error(self):
        err_fp = MagicMock()
        not_found = HTTPError(
            "https://api.snapcraft.io/v2/snaps/info/missing?fields=channel-map",
            404,
            "Not Found",
            Message(),
            err_fp,
        )
        with patch.object(store, "urlopen", side_effect=not_found), patch.object(
            store.time, "sleep"
        ) as sleep_mock, self.assertRaises(HTTPError) as ctx:
            store.query_info("snap", "missing")

        self.assertEqual(ctx.exception.code, 404)
        err_fp.close.assert_called_once()
        sleep_mock.assert_not_called()

    def test_query_info_closes_exhausted_retryable_http_errors(self):
        fps = [MagicMock() for _ in range(store._ATTEMPTS)]
        errors = [
            HTTPError(
                "https://api.snapcraft.io/v2/snaps/info/k8s?fields=channel-map",
                429,
                "Too Many Requests",
                Message(),
                fp,
            )
            for fp in fps
        ]
        with patch.object(store, "urlopen", side_effect=errors), patch.object(
            store.time, "sleep"
        ), self.assertRaises(HTTPError):
            store.query_info("snap", "k8s")

        for fp in fps:
            fp.close.assert_called_once()

    def test_charm_channel_available_returns_nonzero_when_store_errors(self):
        args = argparse.Namespace(
            channel="1.37/stable",
            charm_auth=None,
            charm_name="k8s",
        )
        buf = io.StringIO()
        with patch.object(
            charm, "query_info", side_effect=URLError("connection reset")
        ), redirect_stdout(buf):
            rc = charm.cmd_channel_available(args)

        self.assertEqual(rc, 1)


if __name__ == "__main__":
    unittest.main()
