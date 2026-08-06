#!/usr/bin/env python3
"""Smart router: build a spread task.yaml from a scenario or a standalone doc page.

Modes
-----
Scenario mode  (--target matches a name in scenarios.yaml):
    Chains multiple documentation pages into one task.yaml.
    Injects ``set -e`` and ``export SCENARIO_MODE=true`` at the top so
    individual pages can detect they are running inside a chain.

Standalone mode  (--target is a path to a .md file):
    Generates a task.yaml for a single page.
    Injects ``set -e`` and ``export SCENARIO_MODE=false`` so pages can
    explicitly gate scenario-only checks behind ``SCENARIO_MODE=true``.

Both modes prepend ``cd "${SPREAD_PATH:-.}"`` before each section so that
path-sensitive commands always start from a known root.
"""

from __future__ import annotations

import argparse
import re
import shlex
import subprocess
import sys
import tempfile
from pathlib import Path

try:
    import yaml
except ImportError:
    print(
        "Error: PyYAML is required. Install with: pip install pyyaml\n"
        "       or: sudo apt-get install python3-yaml",
        file=sys.stderr,
    )
    sys.exit(1)


def _get_operator_workflows_ref() -> str:
    try:
        wf = Path(__file__).resolve().parents[2] / ".github/workflows/docs-spread-tests.yaml"
        m = re.search(r"OPERATOR_WORKFLOWS_REF:\s*([a-f0-9]+)", wf.read_text(encoding="utf-8"))
        if m:
            return m.group(1)
    except Exception:
        pass
    return "main"


UPSTREAM_SCRIPT_URL = (
    "https://raw.githubusercontent.com/canonical/operator-workflows"
    f"/{_get_operator_workflows_ref()}/spread/create_spread_task_file.py"
)

# Compiled pattern for the SPREAD SUITE marker used in documentation pages.
_SUITE_MARKER_RE = re.compile(r"SPREAD SUITE:\s*([a-z_]+)")

# Matches `trap <expression> EXIT` line (no leading whitespace)
_TRAP_EXIT_RE = re.compile(r"^trap\s+(.+)\s+EXIT\s*$")
SCENARIO_ONLY_SUITE = "scenario_only"
SUPPORTED_SUITES = {"snap_bootstrapped", "snap_clean", SCENARIO_ONLY_SUITE}
KILL_TIMEOUT = "30m"


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def find_repo_root() -> Path:
    """Return the absolute path to the repository root, detected via git."""
    result = subprocess.run(
        ["git", "rev-parse", "--show-toplevel"],
        capture_output=True,
        text=True,
        check=True,
    )
    return Path(result.stdout.strip())


def detect_suite(file_path: Path) -> str | None:
    """Return the SPREAD SUITE value declared in a docs file, or None if absent."""
    content = file_path.read_text(encoding="utf-8")
    for line in content.splitlines():
        match = _SUITE_MARKER_RE.search(line)
        if match:
            return match.group(1)
    return None


def validate_suite(suite: str) -> None:
    """Raise ValueError if *suite* is not supported by the spread workflow."""
    if suite not in SUPPORTED_SUITES:
        raise ValueError(
            f"Unsupported SPREAD SUITE marker '{suite}'. "
            f"Supported values: {', '.join(sorted(SUPPORTED_SUITES))}"
        )


def run_upstream_script(upstream_script: Path, doc_file: Path, output_task: Path) -> None:
    """Invoke the upstream create_spread_task_file.py to extract commands from *doc_file*."""
    subprocess.run(
        [sys.executable, str(upstream_script), str(doc_file), str(output_task)],
        check=True,
    )


def extract_execute_block(task_file: Path) -> list[str]:
    """Return the lines inside ``execute: |`` with the two-space YAML indent removed."""
    lines = iter(task_file.read_text(encoding="utf-8").splitlines())

    for line in lines: # first loop — scans to "execute: |"
        if line == "execute: |":
            break
    else:
        raise ValueError(f"Missing 'execute: |' block in {task_file}")

    block: list[str] = []
    for line in lines:  # second loop — continues from where the first left off
        if line.startswith("  "):  
            block.append(line[2:])
        elif line.strip() == "":
            block.append("")
        else:
            break  

    if not block:
        raise ValueError(f"Empty execute block in {task_file}")

    return block


# ---------------------------------------------------------------------------
# Trap helpers
# ---------------------------------------------------------------------------


def _extract_trap_body(expression: str) -> str:
    """Strip the outer quotes from a trap expression and return the raw body.

    The returned string is unquoted and ready to be joined with other bodies
    before being re-wrapped in a single pair of single-quotes.

    Handles three forms:
    - Single-quoted:  'cmd1; cmd2'  → cmd1; cmd2
    - Double-quoted:  "cmd1; cmd2"  → cmd1; cmd2  (interior " escaped)
    - Unquoted:        cmd1; cmd2   → cmd1; cmd2  (returned as-is)
    """
    if expression.startswith("'") and expression.endswith("'"):
        inner = expression[1:-1]
        # Normalise any escaped single-quotes that were already present.
        inner = inner.replace("'\\''", "'")
        return inner
    if expression.startswith('"') and expression.endswith('"'):
        inner = expression[1:-1]
        inner = inner.replace('\\"', '"')
        return inner
    return expression


def _partition_traps(
    sections: list[tuple[str, list[str]]],
) -> tuple[list[tuple[str, list[str]]], list[str]]:
    """Separate ``trap '...' EXIT`` lines from the rest of each section's commands.

    Returns a tuple of:
    - cleaned_sections: same structure as *sections* but with trap lines removed.
    - trap_bodies: unquoted trap bodies in LIFO order (last encountered first),
      ready to join with ``"; "`` and wrap in a single ``trap '...' EXIT`` line.
    """
    cleaned_sections: list[tuple[str, list[str]]] = []
    trap_bodies: list[str] = []

    for title, commands in sections:
        cleaned_commands: list[str] = []
        for line in commands:
            match = _TRAP_EXIT_RE.match(line)
            if match:
                trap_bodies.append(_extract_trap_body(match.group(1).strip()))
            else:
                cleaned_commands.append(line)
        cleaned_sections.append((title, cleaned_commands))

    return cleaned_sections, list(reversed(trap_bodies))


# ---------------------------------------------------------------------------
# Task writer
# ---------------------------------------------------------------------------

def write_task(
    output_file: Path,
    summary: str,
    sections: list[tuple[str, list[str]]],
    *,
    scenario_mode: bool,
) -> None:
    output_file.parent.mkdir(parents=True, exist_ok=True)

    # Hoist all trap EXIT lines out of sections and merge them into one combined
    # trap in the preamble so that later traps cannot silently overwrite earlier ones.
    cleaned_sections, trap_bodies = _partition_traps(sections)

    with output_file.open("w", encoding="utf-8") as f:
        # Use yaml.dump for scalar fields to handle quotes and special characters safely.
        f.write(yaml.dump({"summary": summary}, default_flow_style=False))
        f.write("\n")
        f.write(yaml.dump({"kill-timeout": KILL_TIMEOUT}, default_flow_style=False))
        f.write("\nexecute: |\n")

        # Preamble
        f.write("  set -e\n")
        if scenario_mode:
            f.write("  export SCENARIO_MODE=true\n")
        else:
            f.write("  export SCENARIO_MODE=false\n")
        if trap_bodies:
            combined = "; ".join(trap_bodies).replace("'", "'\\''")
            f.write(f"  trap '{combined}' EXIT\n")
        f.write("\n")

        # Sections
        for title, commands in cleaned_sections:
            # Single-quote the echo argument; escape any literal single quotes in title
            # to prevent shell expansion of $(...) or backtick sequences.
            safe_title = title.replace("'", "'\\''")
            f.write(f"  echo '=== Starting Section: {safe_title} ==='\n")
            f.write('  cd "${SPREAD_PATH:-.}"\n')
            for line in commands:
                f.write(f"  {line}\n")
            f.write("\n")


# ---------------------------------------------------------------------------
# Scenario mode
# ---------------------------------------------------------------------------


def build_scenario(
    scenario: dict,
    upstream_script: Path,
    repo_root: Path,
    output_dir: Path,
) -> tuple[Path, str]:
    """Generate a combined task.yaml for a named scenario.

    Returns (output_file, suite_name).
    """
    scenario_name = scenario["name"]
    suite = scenario["suite"]
    validate_suite(suite)
    if suite == SCENARIO_ONLY_SUITE:
        raise ValueError(
            f"Scenario '{scenario_name}' cannot use suite '{SCENARIO_ONLY_SUITE}'."
        )
    pages = scenario["pages"]
    docs_root = (repo_root / "docs" / "canonicalk8s").resolve()
    sections: list[tuple[str, list[str]]] = []

    with tempfile.TemporaryDirectory() as tmpdir:
        for page in pages:
            doc_file = (docs_root / page).resolve()
            if not doc_file.is_relative_to(docs_root):
                raise ValueError(
                    f"Refusing to process {doc_file}: path is outside the {docs_root}"
                )
            if not doc_file.is_file():
                raise FileNotFoundError(f"Page not found: {doc_file}")
            tmp_task = Path(tmpdir) / page.replace("/", "-")
            run_upstream_script(upstream_script, doc_file, tmp_task)
            commands = extract_execute_block(tmp_task)
            # Use the full relative page path as the section title to avoid
            # ambiguity when multiple pages share the same filename.
            sections.append((page, commands))

    output_file = output_dir / suite / f"scenario-{scenario_name}" / "task.yaml"
    write_task(
        output_file,
        summary=f"Scenario: {scenario_name}",
        sections=sections,
        scenario_mode=True,
    )
    return output_file, suite


# ---------------------------------------------------------------------------
# Standalone mode
# ---------------------------------------------------------------------------


def build_standalone(
    doc_path: Path,
    upstream_script: Path,
    repo_root: Path,
    output_dir: Path,
) -> tuple[Path | None, str]:
    """Generate a task.yaml for a single documentation page.

    Returns (output_file, suite_name), or (None, "") if the page should be skipped.
    """
    if not doc_path.is_absolute():
        # Accept paths relative to repo root or relative to docs/canonicalk8s/
        candidate = repo_root / doc_path
        if not candidate.is_file():
            candidate = repo_root / "docs" / "canonicalk8s" / doc_path
        doc_path = candidate

    # Resolve symlinks and normalise ../ components before enforcing the repo boundary.
    doc_path = doc_path.resolve()
    if not doc_path.is_relative_to(repo_root.resolve()):
        raise ValueError(
            f"Refusing to process {doc_path}: path is outside the repository root"
        )

    if not doc_path.is_file():
        raise FileNotFoundError(f"File not found: {doc_path}")

    suite = detect_suite(doc_path)
    if suite is None:
        print(f"Skipping {doc_path.name} (no SPREAD SUITE marker — not a testable page)")
        return None, ""
    validate_suite(suite)
    if suite == SCENARIO_ONLY_SUITE:
        print(f"Skipping {doc_path.name} (scenario_only — will only run as part of a scenario chain)")
        return None, ""

    docs_root = repo_root / "docs" / "canonicalk8s"
    try:
        rel = doc_path.relative_to(docs_root.resolve())
    except ValueError:
        raise ValueError(
            f"{doc_path} is not under {docs_root}. "
            "Pass a path relative to docs/canonicalk8s/ or to the repo root."
        )
    task_name = str(rel.with_suffix("")).replace("/", "-")

    with tempfile.TemporaryDirectory() as tmpdir:
        tmp_task = Path(tmpdir) / "task.yaml"
        run_upstream_script(upstream_script, doc_path, tmp_task)
        commands = extract_execute_block(tmp_task)

    sections = [(doc_path.name, commands)]
    output_file = output_dir / suite / task_name / "task.yaml"
    write_task(
        output_file,
        summary=str(rel),
        sections=sections,
        scenario_mode=False,
    )
    return output_file, suite


# ---------------------------------------------------------------------------
# Detect mode (CI use)
# ---------------------------------------------------------------------------


def run_detect(
    changed_files_input: str,
    scenarios_file: Path,
    upstream_script: Path,
    repo_root: Path,
    output_dir: Path,
) -> None:
    """Generate chained task.yaml for every scenario triggered by changed files.

    Called by CI after the standalone task generation loop.  Exits 0 whether
    or not any scenarios were triggered.
    """
    if not scenarios_file.is_file():
        print(f"No scenarios file found at {scenarios_file}, skipping scenario detection.")
        return

    # Build the set of changed files with full relative path from repo root.
    # shlex.split() handles quoted filenames that contain spaces.
    changed_set = {
        f"docs/canonicalk8s/{p}"
        for p in shlex.split(changed_files_input)
        if p
    }
    if not changed_set:
        print("No changed files provided; skipping scenario detection.")
        return

    data = yaml.safe_load(scenarios_file.read_text(encoding="utf-8")) or {}
    triggered = False
    for scenario in data.get("scenarios", []):
        pages = {f"docs/canonicalk8s/{p}" for p in scenario["pages"]}
        if not pages.intersection(changed_set):
            continue
        name = scenario["name"]
        print(f"Scenario '{name}' triggered — generating chained task.yaml")
        output_file, _ = build_scenario(
            scenario=scenario,
            upstream_script=upstream_script,
            repo_root=repo_root,
            output_dir=output_dir,
        )
        print(f"  Generated: {output_file}")
        triggered = True

    if not triggered:
        print("No scenarios triggered by changed files.")


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Build a spread task.yaml from a scenario (chained pages) "
            "or a standalone documentation page."
        ),
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
examples:
  # scenario mode
  build_spread_task.py --target cis

  # standalone mode
  build_spread_task.py --target snap/howto/install/fips.md

  # CI detect mode (generates all scenarios triggered by changed files)
  build_spread_task.py --detect-scenarios "snap/howto/install/fips.md snap/howto/install/disa-stig.md"

  # override upstream script location (e.g. for local use)
  build_spread_task.py --target cis --upstream-script /path/to/create_spread_task_file.py
""",
    )
    parser.add_argument(
        "--target",
        default=None,
        help=(
            "Scenario name from scenarios.yaml, "
            "or path to a .md file (relative to repo root or docs/canonicalk8s/)."
        ),
    )
    parser.add_argument(
        "--detect-scenarios",
        default=None,
        metavar="CHANGED_FILES",
        help=(
            "(CI use only) Space-delimited docs/canonicalk8s/-relative list of changed files. "
            "Generates a chained task.yaml for every scenario whose pages intersect the changed set."
        ),
    )
    parser.add_argument(
        "--upstream-script",
        required=False,
        default=None,
        type=Path,
        help=(
            "Path to the upstream create_spread_task_file.py "
            "(default: workflow-scripts/spread/create_spread_task_file.py relative to repo root)."
        ),
    )
    parser.add_argument(
        "--repo-root",
        type=Path,
        default=None,
        help="Repo root directory (default: auto-detected via git).",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()

    if not args.detect_scenarios and not args.target:
        print("Error: one of --target or --detect-scenarios is required.", file=sys.stderr)
        sys.exit(1)

    repo_root: Path = args.repo_root or find_repo_root()
    scenarios_file: Path = repo_root / "docs" / "tools" / "scenarios.yaml"
    output_dir: Path = repo_root / "tests" / "spread_generated"
    upstream_script: Path = (
        args.upstream_script
        or repo_root / "workflow-scripts" / "spread" / "create_spread_task_file.py"
    )

    if not upstream_script.is_file():
        print(
            f"Error: upstream script not found at {upstream_script}\n\n"
            f"  Create the dir and fetch the script with:\n"
            f"    mkdir -p {upstream_script.parent} && sudo chown $USER {upstream_script.parent}\n"
            f"    curl -fsSL '{UPSTREAM_SCRIPT_URL}' -o '{upstream_script}'",
            file=sys.stderr,
        )
        sys.exit(1)

    # Pre-create all real suite directories so spread.yaml never encounters a missing
    # path, even when no tasks are generated for that suite in this run.
    for suite in SUPPORTED_SUITES - {SCENARIO_ONLY_SUITE}:
        (output_dir / suite).mkdir(parents=True, exist_ok=True)

    # --- Detect mode (CI) ---
    if args.detect_scenarios is not None:
        run_detect(
            changed_files_input=args.detect_scenarios,
            scenarios_file=scenarios_file,
            upstream_script=upstream_script,
            repo_root=repo_root,
            output_dir=output_dir,
        )
        return

    target = args.target

    # Route: check if target matches a scenario name or is a standalone doc path
    scenarios_data: dict = {}
    if scenarios_file.is_file():
        scenarios_data = yaml.safe_load(scenarios_file.read_text(encoding="utf-8")) or {}
    scenario_entry = next(
        (s for s in scenarios_data.get("scenarios", []) if s["name"] == target),
        None,
    )

    if scenario_entry is not None:
        output_file, suite = build_scenario(
            scenario=scenario_entry,
            upstream_script=upstream_script,
            repo_root=repo_root,
            output_dir=output_dir,
        )
    else:
        output_file, suite = build_standalone(
            doc_path=Path(target),
            upstream_script=upstream_script,
            repo_root=repo_root,
            output_dir=output_dir,
        )
        if output_file is None:
            return

    suite_dir = output_dir.relative_to(repo_root) / suite
    print(f"\nGenerated: {output_file}")
    print(f"\nFor local testing with multipass:")
    print(f"  cd {repo_root}/docs")
    print(f"  spread multipass:ubuntu-24.04-64:{suite_dir}/\n")


if __name__ == "__main__":
    main()