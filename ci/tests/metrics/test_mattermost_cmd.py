#
# Copyright 2026 Canonical, Ltd.
#
"""Argument-wiring regression tests for the `mattermost` subcommands.

`cmd_post` reads `args.bot_token`/`args.server`/`args.channel_id`, but the
`post` subparser previously never defined those flags, so every invocation
(including the ci-metrics digest workflow step) raised an `AttributeError`
before ever attempting a post. These tests pin the parser/handler contract.
"""

import argparse
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from cmds.mattermost import add_mattermost_cmds  # noqa: E402


def _build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="k8s-ci")
    subparsers = parser.add_subparsers(dest="subcommand", required=True)
    add_mattermost_cmds(subparsers)
    return parser


class MattermostCmdTest(unittest.TestCase):
    def test_post_subparser_defines_bot_token_server_and_channel_id(self) -> None:
        parser = _build_parser()

        args = parser.parse_args(
            [
                "mattermost",
                "post",
                "--file",
                "-",
                "--webhook",
                "https://example.invalid/webhook",
                "--bot-token",
                "tok",
                "--server",
                "https://example.invalid",
                "--channel-id",
                "chan123",
            ]
        )

        self.assertEqual(args.bot_token, "tok")
        self.assertEqual(args.server, "https://example.invalid")
        self.assertEqual(args.channel_id, "chan123")

    def test_post_subparser_allows_omitting_bot_fields(self) -> None:
        # The fields fall back to environment variables in cmd_post, so the
        # parser must not require them outright.
        parser = _build_parser()

        args = parser.parse_args(
            [
                "mattermost",
                "post",
                "--file",
                "-",
                "--webhook",
                "https://example.invalid/webhook",
            ]
        )

        self.assertIsNone(args.bot_token)
        self.assertIsNone(args.server)
        self.assertIsNone(args.channel_id)


if __name__ == "__main__":
    unittest.main()
