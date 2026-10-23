#
# Copyright 2026 Canonical, Ltd.
#
"""Deterministic Markdown reports for CI failure metrics."""

from __future__ import annotations

from pathlib import Path
from typing import Any, Dict, List, Optional, Sequence

from metrics.rollup import delta, signature_catalogue, top_signatures

CLASS_OWNER_HINT = {
    "infra.runner": "CI/infra",
    "infra.provisioning": "CI/infra",
    "external.dependency": "external dependency",
    "ci.config": "CI owners",
    "product.bug": "product team",
    "test.bug": "test owners",
    "unknown": "unowned triage",
}


def write_reports(
    reports_dir: Path,
    signatures_dir: Path,
    rollups: Sequence[Dict[str, Any]],
    catalogue: Sequence[Dict[str, Any]],
    open_issue_rows: Optional[Sequence[Dict[str, Any]]] = None,
) -> tuple[Path, Path]:
    """Write committed human-facing report exports."""

    latest_path = reports_dir / "latest.md"
    catalogue_path = signatures_dir / "catalogue.md"
    _write_text(latest_path, render_latest_markdown(rollups, open_issue_rows))
    _write_text(catalogue_path, render_catalogue_markdown(catalogue))
    return latest_path, catalogue_path


def build_catalogue(conn: Any) -> List[Dict[str, Any]]:
    """Return the current signature catalogue payload."""

    return signature_catalogue(conn)


def render_latest_markdown(
    rollups: Sequence[Dict[str, Any]],
    open_issue_rows: Optional[Sequence[Dict[str, Any]]] = None,
) -> str:
    """Render the latest trend report with attribution and integrity up front."""

    lines: List[str] = ["# CI failure report", ""]
    if not rollups:
        lines.append("_No CI metrics have been ingested._")
        return "\n".join(lines) + "\n"

    latest = rollups[-1]
    previous = rollups[-2] if len(rollups) > 1 else None
    lines.append(
        f"Data period: `{latest['period']}`"
        + (
            f" (through `{latest['data_through']}`)"
            if latest.get("data_through")
            else ""
        )
    )
    lines.append("")
    lines.extend(_attribution_and_delta(latest, previous))
    lines.append("")
    lines.extend(_headline_table(rollups))
    lines.append("")
    lines.extend(_class_table(latest))
    lines.append("")
    lines.extend(_integrity_section(latest, previous))
    lines.append("")
    lines.extend(_signature_section(latest))
    lines.append("")
    lines.extend(_open_issues_section(open_issue_rows))
    lines.append("")
    lines.append(
        f"_Derived from durable facts at report time. Rollup v{latest['rollup_version']}._"
    )
    return "\n".join(lines) + "\n"


def render_digest_markdown(
    rollups: Sequence[Dict[str, Any]],
    report_url: Optional[str] = None,
    open_issue_rows: Optional[Sequence[Dict[str, Any]]] = None,
) -> str:
    """Render the short Mattermost digest body: attribution and delta first.

    This is deliberately smaller than ``render_latest_markdown`` — a chat digest
    is read in seconds, not reviewed like the committed report — but it reuses
    the same attribution-and-delta section so the two never drift apart.
    """

    lines: List[str] = ["#### CI failure digest", ""]
    if not rollups:
        lines.append("_No CI metrics have been ingested yet._")
        return "\n".join(lines) + "\n"

    latest = rollups[-1]
    previous = rollups[-2] if len(rollups) > 1 else None
    lines.append(f"Data period: `{latest['period']}`")
    lines.append("")
    lines.extend(_attribution_and_delta(latest, previous))
    lines.append("")
    lines.extend(
        _open_issues_section(open_issue_rows, heading="#### Open CI-filed issues")
    )
    if report_url:
        lines.append("")
        lines.append(f"[Full report and signature catalogue]({report_url})")
    return "\n".join(lines) + "\n"


def render_catalogue_markdown(catalogue: Sequence[Dict[str, Any]]) -> str:
    """Render the deterministic signature catalogue."""

    lines = [
        "# CI failure signature catalogue",
        "",
        "| Signature | Occurrences | Runs | Class | Rule | State | First seen | Last seen | Contexts |",
        "|---|---:|---:|---|---|---|---|---|---|",
    ]
    for row in catalogue:
        contexts = ", ".join(
            f"{ctx['substrate']}:{ctx['occurrences']}"
            for ctx in row.get("contexts", [])
        )
        lines.append(
            "| `{signature_id}` | {occurrences} | {runs_seen} | {klass} | {rule} | "
            "{state} | {first} | {last} | {contexts} |".format(
                signature_id=row["signature_id"],
                occurrences=row["occurrences"],
                runs_seen=row["runs_seen"],
                klass=_class_label(row),
                rule=row.get("rule_id") or "_unowned_",
                state=row.get("state") or "",
                first=row.get("first_seen") or "",
                last=row.get("last_seen") or "",
                contexts=contexts,
            )
        )
    return "\n".join(lines) + "\n"


def _attribution_and_delta(
    latest: Dict[str, Any], previous: Optional[Dict[str, Any]]
) -> List[str]:
    lines = ["## Attribution and delta", ""]
    lines.append(
        "- Observed failed jobs: {failed}/{total} = {rate}{change}".format(
            failed=latest["jobs_failure"],
            total=latest["jobs_total"],
            rate=_format_rate(
                latest.get("m3_job_failure_rate"), latest.get("jobs_total"), "jobs"
            ),
            change=_format_delta(delta(latest, previous, "m3_job_failure_rate")),
        )
    )
    lines.append(
        "- Failure pressure including blocked jobs: {rate}{change}".format(
            rate=_format_rate(
                latest.get("m11_failure_pressure_rate"),
                latest.get("m11_failure_pressure_denominator"),
                "jobs plus blocked",
            ),
            change=_format_delta(delta(latest, previous, "m11_failure_pressure_rate")),
        )
    )

    classes = list(_ranked_classes(latest.get("m3_by_class") or {}))
    if not classes:
        lines.append("- No failed jobs have classified facts yet.")
        return lines

    name, payload = classes[0]
    lines.append(
        "- Top attribution: `{}` owned by {} ({} of inspected failures).".format(
            name,
            CLASS_OWNER_HINT.get(name, "unknown owner"),
            _format_rate(
                payload.get("share"), latest.get("m9_failed_job_facts"), "facts"
            ),
        )
    )
    return lines


def _open_issues_section(
    rows: Optional[Sequence[Dict[str, Any]]], heading: str = "## Open CI-filed issues"
) -> List[str]:
    """Render the open-issue-bridge backlog, oldest first, with age in days.

    Surfaces anything the bridge has filed and not yet closed, so nothing sits
    untriaged indefinitely (delivery plan, Movement 3, "Closing the loop").
    """

    lines = [heading, ""]
    if not rows:
        lines.append("_No open CI-filed issues._")
        return lines
    lines.append("| Issue | Signature | Substrate | Class | Age (days) | Occurrences |")
    lines.append("|---|---|---|---|---:|---:|")
    for row in rows:
        age = row.get("age_days")
        lines.append(
            "| #{number} | `{signature}` | {substrate} | `{klass}` | {age} | {occurrences} |".format(
                number=row["issue_number"],
                signature=row["signature_id"],
                substrate=row.get("substrate") or "*",
                klass=row.get("failure_class") or "unknown",
                age=age if age is not None else "n/a",
                occurrences=row.get("last_reported_occurrences", 0),
            )
        )
    return lines


def _headline_table(rollups: Sequence[Dict[str, Any]]) -> List[str]:
    lines = [
        "## Headline metrics",
        "",
        "| Period | Runs | Failed jobs | Failure rate | Scheduled green | "
        "PR first-pass | Unclassified | M12 fact coverage |",
        "|---|---:|---:|---|---|---|---|---|",
    ]
    for row in rollups:
        lines.append(
            "| {period} | {runs} | {failed} | {failure} | {scheduled} | {pr} | "
            "{unclassified} | {coverage} |".format(
                period=row["period"],
                runs=row["runs"],
                failed=row["jobs_failure"],
                failure=_format_rate(
                    row.get("m3_job_failure_rate"), row.get("jobs_total"), "jobs"
                ),
                scheduled=_format_rate(
                    row.get("m1_scheduled_green_rate"),
                    row.get("m1_scheduled_runs"),
                    "runs",
                ),
                pr=_format_rate(
                    row.get("m2_pr_first_pass_rate"), row.get("m2_pr_runs"), "runs"
                ),
                unclassified=_format_rate(
                    row.get("m9_unclassified_rate"),
                    row.get("m9_failed_job_facts"),
                    "facts",
                ),
                coverage=_format_rate(
                    row.get("m12_failed_job_fact_coverage"),
                    row.get("m12_failed_job_fact_denominator"),
                    "expected failed jobs",
                ),
            )
        )
    return lines


def _class_table(latest: Dict[str, Any]) -> List[str]:
    lines = [
        "## Failure attribution",
        "",
        "| Class | Count | Share | Owner |",
        "|---|---:|---|---|",
    ]
    rows = list(_ranked_classes(latest.get("m3_by_class") or {}))
    if not rows:
        lines.append("| _none_ | 0 | n/a | n/a |")
        return lines
    for name, payload in rows:
        lines.append(
            f"| `{name}` | {payload['count']} | "
            f"{_format_rate(payload.get('share'), latest.get('m9_failed_job_facts'), 'facts')} | "
            f"{CLASS_OWNER_HINT.get(name, 'unknown owner')} |"
        )
    return lines


def _integrity_section(
    latest: Dict[str, Any], previous: Optional[Dict[str, Any]]
) -> List[str]:
    lines = [
        "## Integrity metrics",
        "",
        "These are report metrics, not appendix material: they show whether "
        "apparent improvement came from running less.",
        "",
        "- M9 unclassified share: {rate}{change}".format(
            rate=_format_rate(
                latest.get("m9_unclassified_rate"),
                latest.get("m9_failed_job_facts"),
                "failed job facts",
            ),
            change=_format_delta(delta(latest, previous, "m9_unclassified_rate")),
        ),
        "- M11 jobs blocked by failed `Prepare Environment`: {blocked} across {prepare} prepare job(s).".format(
            blocked=latest.get("m11_jobs_not_run_due_to_prepare", 0),
            prepare=latest.get("m11_failed_prepare_jobs", 0),
        ),
        "- M12 failed-job fact coverage: {coverage}.".format(
            coverage=_format_rate(
                latest.get("m12_failed_job_fact_coverage"),
                latest.get("m12_failed_job_fact_denominator"),
                "expected failed jobs",
            )
        ),
        "- M12 log availability among failed job facts: {coverage}.".format(
            coverage=_format_rate(
                latest.get("m12_log_available_rate"),
                latest.get("m9_failed_job_facts"),
                "failed job facts",
            )
        ),
    ]
    tests = latest.get("m10_tests_collected") or []
    lines.extend(["", "### M10 tests collected per config", ""])
    if not tests:
        lines.append("_No collected-test counts were stored for this period._")
        return lines
    lines.extend(["| Config | Min | Median | Max |", "|---|---:|---:|---:|"])
    for row in tests:
        lines.append(
            f"| `{row['config']}` | {row['min']} | {row['median']} | {row['max']} |"
        )
    return lines


def _signature_section(latest: Dict[str, Any]) -> List[str]:
    lines = [
        "## Top signatures",
        "",
        "| Signature | Occurrences | Runs | Class | Rule | Tests | Configs | Example |",
        "|---|---:|---:|---|---|---:|---:|---|",
    ]
    signatures = top_signatures(latest)
    if not signatures:
        lines.append("| _none_ | 0 | 0 | n/a | n/a | 0 | 0 | |")
        return lines
    for sig in signatures:
        url = sig.get("exemplar_job_url")
        example = f"[job]({url})" if url else ""
        lines.append(
            "| `{signature}` | {occurrences} | {runs} | {klass} | {rule} | "
            "{tests} | {configs} | {example} |".format(
                signature=sig["signature_id"],
                occurrences=sig["occurrences"],
                runs=sig["runs_seen"],
                klass=_class_label(sig),
                rule=sig.get("rule_id") or "_unowned_",
                tests=sig.get("tests_affected_count", 0),
                configs=sig.get("configs_affected_count", 0),
                example=example,
            )
        )
    return lines


def _ranked_classes(
    rows: Dict[str, Dict[str, Any]],
) -> List[tuple[str, Dict[str, Any]]]:
    return sorted(rows.items(), key=lambda item: (-int(item[1]["count"]), item[0]))


def _format_rate(value: Optional[float], denominator: Optional[int], noun: str) -> str:
    if value is None or not denominator:
        return "n/a"
    return f"{value:.1f}% (n={denominator} {noun})"


def _format_delta(value: Optional[float]) -> str:
    if value is None:
        return ""
    if value == 0:
        return " (no change)"
    return f" ({value:+.1f} pp)"


def _class_label(row: Dict[str, Any]) -> str:
    subclass = row.get("subclass")
    return row.get("failure_class", "unknown") + (f"/{subclass}" if subclass else "")


def _write_text(path: Path, text: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    if path.exists() and path.read_text(encoding="utf-8") == text:
        return
    path.write_text(text, encoding="utf-8")
