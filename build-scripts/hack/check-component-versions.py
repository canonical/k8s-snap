#!/usr/bin/env python3
import sys
import re
import argparse
from packaging.version import InvalidVersion, Version

VERSION_FILE_PATTERN = re.compile(r'^(?:\+\+\+|---) [ab](/build-scripts/components/[^/]+/version)')

def parse_version(line: str):
    try:
        return Version(line.strip())
    except InvalidVersion:
        return None

def main(git_diff_file):
    current_file = None
    old_version = None
    new_version = None
    invalid_changes = []

    with open(git_diff_file) as f:
        for line in f:
            # Track which file we're looking at
            if match := VERSION_FILE_PATTERN.match(line):
                current_file = "." + match.group(1)  # Extract file path from regex match
                old_version = None
                new_version = None
                continue

            if not current_file:
                continue

            if line.startswith('-'):
                old_version = parse_version(line[1:])
            elif line.startswith('+'):
                new_version = parse_version(line[1:])

            if old_version and new_version:
                # Compare major/minor
                if old_version.release[:2] != new_version.release[:2]:
                    invalid_changes.append((current_file, old_version, new_version))
                old_version = None
                new_version = None

    for f, old_v, new_v in invalid_changes:
        print(f"Invalid version change in {f}: {old_v} → {new_v}", file=sys.stderr)
    sys.exit(1 if invalid_changes else 0)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Check component version changes in a git diff file.")
    parser.add_argument("git_diff_file", help="Path to the git diff file")
    args = parser.parse_args()
    main(args.git_diff_file)
