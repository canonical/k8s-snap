#
# Copyright 2026 Canonical, Ltd.
#
import argparse
import io
import json
import sys
import unittest
from pathlib import Path
from unittest.mock import patch
from urllib.error import HTTPError

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from cmds import mattermost  # noqa: E402


def _entry(
    channel: str = "latest/edge",
    os_name: str = "ubuntu:24.04",
    arch: str = "amd64",
    test: str = "tests/test_smoke.py::test_smoke",
    status: str = "success",
    run_url: str = "https://github.com/canonical/k8s-snap/actions/runs/100/attempts/1",
    artifact: str = "",
) -> dict:
    return {
        "os": os_name,
        "arch": arch,
        "channel": channel,
        "revision": "1000",
        "artifact": artifact,
        "substrate": "lxd",
        "variant": "classic",
        "runner": "runner-1",
        "attempt": "1",
        "sha": "deadbeef",
        "source": "canonical/k8s-snap@main",
        "run_url": run_url,
        "test": test,
        "status": status,
    }


class MattermostDetailsTablesTests(unittest.TestCase):
    def test_channel_table_appears_before_failing_tests_table(self):
        entries = [
            _entry(
                channel="latest/edge",
                os_name="ubuntu:22.04",
                arch="amd64",
                test="tests/test_a.py::test_alpha",
                status="failure",
            ),
            _entry(
                channel="latest/edge",
                os_name="ubuntu:22.04",
                arch="arm64",
                test="tests/test_a.py::test_alpha",
                status="success",
            ),
            _entry(
                channel="1.36-classic/edge",
                os_name="ubuntu:22.04",
                arch="amd64",
                test="tests/test_a.py::test_alpha",
                status="success",
            ),
            _entry(
                channel="1.36-classic/edge",
                os_name="ubuntu:22.04",
                arch="arm64",
                test="tests/test_a.py::test_alpha",
                status="cancelled",
            ),
        ]

        msg = mattermost.build_details_message(entries)

        self.assertEqual(
            msg,
            "\n".join(
                [
                    "### Channel Matrix",
                    "| Channel | 22.04 amd64 | 22.04 arm64 |",
                    "| --- | :---: | :---: |",
                    "| `latest/edge` | :x: 1 | :white_check_mark: |",
                    "| `1.36-classic/edge` | :white_check_mark: | :stopwatch: 1 |",
                    "",
                    "### Top Failing Tests",
                    "| Test | Count | Breakdown | Channels |",
                    "| --- | ---: | --- | --- |",
                    "| `test_alpha` | 2 | :x: 1, :stopwatch: 1 | `latest/edge`, `1.36-classic/edge` |",
                ]
            ),
        )

    def test_failing_tests_sorted_by_failure_count_desc_and_capped_at_five(self):
        counts = [
            ("test_six", 1),
            ("test_two", 5),
            ("test_four", 3),
            ("test_one", 7),
            ("test_five", 2),
            ("test_three", 4),
            ("test_seven", 1),
        ]
        entries = []
        for name, n in counts:
            for i in range(n):
                entries.append(
                    _entry(
                        channel=f"1.3{i % 6}-classic/edge",
                        os_name="ubuntu:24.04",
                        arch="amd64",
                        test=f"tests/suite.py::{name}",
                        status="failure",
                    )
                )

        msg = mattermost.build_details_message(entries)
        test_section = msg.split("### Top Failing Tests\n", 1)[1]
        lines = test_section.splitlines()

        self.assertEqual(lines[0], "| Test | Count | Breakdown | Channels |")
        self.assertEqual(lines[1], "| --- | ---: | --- | --- |")
        rendered_tests = [line.split("|")[1].strip() for line in lines[2:7]]
        self.assertEqual(
            rendered_tests,
            [
                "`test_one`",
                "`test_two`",
                "`test_three`",
                "`test_four`",
                "`test_five`",
            ],
        )
        self.assertNotIn("test_six", test_section)
        self.assertNotIn("test_seven", test_section)
        self.assertEqual(
            lines[7],
            "_...and [2 more failing tests](https://github.com/canonical/k8s-snap/actions/runs/100/attempts/1)._",
        )

    def test_same_test_name_in_different_files_not_conflated(self):
        entries = [
            _entry(test="tests/test_alpha.py::test_smoke", status="failure"),
            _entry(test="tests/test_beta.py::test_smoke", status="failure"),
        ]

        msg = mattermost.build_details_message(entries)

        self.assertIn(
            "| `test_alpha.py::test_smoke` | 1 | :x: 1 | `latest/edge` |",
            msg,
        )
        self.assertIn(
            "| `test_beta.py::test_smoke` | 1 | :x: 1 | `latest/edge` |",
            msg,
        )

    def test_all_green_omits_failing_tests_table(self):
        entries = [
            _entry(channel="latest/edge", os_name="ubuntu:24.04", arch="amd64"),
            _entry(channel="1.37-classic/edge", os_name="ubuntu:24.04", arch="amd64"),
        ]

        msg = mattermost.build_details_message(entries)

        self.assertIn("### Channel Matrix", msg)
        self.assertNotIn("Top Failing Tests", msg)
        self.assertNotIn("├──", msg)
        self.assertNotIn("└──", msg)

    def test_all_channels_collapsed_when_every_channel_affected(self):
        channels = [
            "latest/edge",
            "1.37-classic/edge",
            "1.36-classic/edge",
            "1.35-classic/edge",
            "1.34-classic/edge",
        ]
        entries = [
            _entry(channel=ch, test="tests/test_up.py::test_upgrade", status="failure")
            for ch in channels
        ]

        msg = mattermost.build_details_message(entries)

        self.assertIn(
            "| `test_upgrade` | 5 | :x: 5 | all 5 channels |",
            msg,
        )

    def test_skipped_entries_do_not_mark_cell_or_test_as_failed(self):
        entries = [
            _entry(status="success"),
            _entry(test="tests/test_skip.py::test_skipped", status="skipped"),
        ]

        msg = mattermost.build_details_message(entries)

        self.assertIn("| `latest/edge` | :white_check_mark: |", msg)
        self.assertNotIn("Top Failing Tests", msg)
        self.assertEqual(mattermost._determine_color(entries), "good")

    def test_large_failure_burst_stays_well_under_mattermost_post_limit(self):
        entries = [
            _entry(
                channel=f"1.3{i % 6}-classic/edge",
                os_name=f"ubuntu:{22 + (i % 3) * 2}.04",
                arch="amd64" if i % 2 == 0 else "arm64",
                test=f"tests/test_{i % 60}.py::test_long_failure_case_name_{i % 60}",
                status="failure" if i % 3 else "cancelled",
            )
            for i in range(500)
        ]

        msg = mattermost.build_details_message(entries)

        self.assertLess(len(msg), 4000)

    def test_summary_payload_includes_breakdown_and_run_link(self):
        entries = [
            _entry(status="success"),
            _entry(status="failure"),
            _entry(status="cancelled"),
            _entry(status="skipped"),
        ]

        payload = mattermost._build_summary_payload(
            "chan-1", entries, "Nightly CI", "danger"
        )

        self.assertEqual(payload["channel_id"], "chan-1")
        att = payload["attachments"][0]
        self.assertEqual(att["title"], ":x: Nightly CI")
        self.assertEqual(att["color"], "danger")
        self.assertEqual(
            att["text"],
            "1/4 passed (25%) · :x: 1 failed · :stopwatch: 1 cancelled · Skipped: 1"
            " — [View Run](https://github.com/canonical/k8s-snap/actions/runs/100/attempts/1)",
        )

    def test_post_bot_logs_http_error_response_body(self):
        err = HTTPError(
            url="https://chat.example.com/api/v4/posts",
            code=400,
            msg="Bad Request",
            hdrs=None,
            fp=io.BytesIO(
                json.dumps({"message": "Post message exceeds 16383 chars"}).encode()
            ),
        )
        stderr = io.StringIO()
        with (
            patch("cmds.mattermost.urlopen", side_effect=err),
            patch("sys.stderr", stderr),
        ):
            with self.assertRaises(SystemExit) as ctx:
                mattermost._post_bot("https://chat.example.com", "tok", "chan", "hello")

        self.assertEqual(ctx.exception.code, 2)
        self.assertIn("Post message exceeds 16383 chars", stderr.getvalue())

    def test_cmd_results_message_dry_run_prints_summary_and_tables(self):
        entries = [_entry(status="failure")]
        stdout = io.StringIO()
        args = argparse.Namespace(
            file="-",
            webhook=None,
            bot_token=None,
            server=None,
            channel_id="chan-1",
            title="Nightly",
            dry_run=True,
        )
        with (
            patch("sys.stdin", io.StringIO(json.dumps(entries))),
            patch("sys.stdout", stdout),
        ):
            rc = mattermost.cmd_results_message(args)

        self.assertEqual(rc, 0)
        out = stdout.getvalue()
        self.assertIn("### Channel Matrix", out)
        self.assertIn("### Top Failing Tests", out)


if __name__ == "__main__":
    unittest.main()
