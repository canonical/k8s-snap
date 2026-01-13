#
# Copyright 2026 Canonical, Ltd.
#
import argparse
import sys

from cmds.charm import add_charm_cmds
from cmds.mattermost import add_mattermost_cmds


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(prog="k8s-ci", description="k8s CI toolkit")
    subparsers = parser.add_subparsers(dest="subcommand", required=True)

    # register subcommands
    add_charm_cmds(subparsers)
    add_mattermost_cmds(subparsers)

    args = parser.parse_args(argv)

    # subcommands set `func` on the parser defaults
    if hasattr(args, "func") and callable(args.func):
        return args.func(args)

    parser.print_help()
    return 1


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
