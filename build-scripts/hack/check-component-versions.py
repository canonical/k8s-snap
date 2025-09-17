#!/usr/bin/env python3
import sys
import re
from packaging.version import Version

VERSION_FILE_PATTERN = re.compile(r'^(\+\+\+|---) a?\.?/build-scripts/components/[^/]+/version')

def parse_version(line):
    line = line.strip()[1:].strip()  # Remove leading + or -
    try:
        return Version(line)
    except Exception:
        return None

def main(diff_file):
    current_file = None
    old_version = None
    new_version = None
    invalid_changes = []

    with open(diff_file) as f:
        for line in f:
            # Track which file we're looking at
            if VERSION_FILE_PATTERN.match(line):
                current_file = line.strip().split()[-1]  # last token: file path
                old_version = None
                new_version = None
                continue

            if not current_file:
                continue

            # Skip the diff headers
            if line.startswith('---') or line.startswith('+++'):
                continue

            if line.startswith('-'):
                old_version = parse_version(line)
            elif line.startswith('+'):
                new_version = parse_version(line)

            if old_version and new_version:
                # Compare major/minor
                if (old_version.major != new_version.major or
                    old_version.minor != new_version.minor):
                    invalid_changes.append((current_file, old_version, new_version))
                old_version = None
                new_version = None

    if invalid_changes:
        for f, old_v, new_v in invalid_changes:
            print(f"Invalid version change in {f}: {old_v} → {new_v}", file=sys.stderr)
        sys.exit(1)
    else:
        sys.exit(0)

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: check-component-versions.py <diff-file>", file=sys.stderr)
        sys.exit(1)
    main(sys.argv[1])
