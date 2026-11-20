#!/usr/bin/env python3
"""Apply strict variant overlay to snap/snapcraft.yaml."""

import os
from pathlib import Path

import yaml

DIR = Path(__file__).resolve().parent
_PROJECT_DIR = os.getenv("SNAPCRAFT_PROJECT_DIR") or ""
PROJECT_DIR = Path(_PROJECT_DIR) if _PROJECT_DIR else DIR.parents[2]
SNAPCRAFT_YAML = PROJECT_DIR / "snap" / "snapcraft.yaml"
OVERLAY_YAML = DIR / "strict-overlay.yaml"


def apply_strict_overlay() -> None:
    """Read snapcraft.yaml, apply strict variant overlay, and write back."""
    with open(SNAPCRAFT_YAML, "r", encoding="utf-8") as f:
        doc = yaml.safe_load(f)

    with open(OVERLAY_YAML, "r", encoding="utf-8") as f:
        overlay = yaml.safe_load(f)

    doc["confinement"] = "strict"
    doc["assumes"] = overlay.get("assumes", ["snapd2.62"])

    if "environment" in overlay:
        doc.setdefault("environment", {}).update(overlay["environment"])

    if "plugs" in overlay:
        doc.setdefault("plugs", {}).update(overlay["plugs"])

    if "layout" in overlay:
        doc.setdefault("layout", {}).update(overlay["layout"])

    if "hooks" in overlay:
        doc.setdefault("hooks", {}).update(overlay["hooks"])

    for part_name, part in doc.get("parts", {}).items():
        if "build-attributes" in part:
            part["build-attributes"] = [
                "no-patchelf" if attr == "enable-patchelf" else attr
                for attr in part["build-attributes"]
            ]
        if part_name in ("etcd", "containerd") and "override-prime" in part:
            del part["override-prime"]

    overlay_apps = overlay.get("apps", {})
    doc_apps = doc.get("apps", {})
    if set(overlay_apps) != set(doc_apps):
        missing = sorted(set(doc_apps) - set(overlay_apps))
        extra = sorted(set(overlay_apps) - set(doc_apps))
        raise ValueError(f"strict overlay app mismatch (missing={missing}, extra={extra})")
    for app_name, app_overlay in overlay_apps.items():
        doc_apps[app_name].update(app_overlay)
    with open(SNAPCRAFT_YAML, "w", encoding="utf-8") as f:
        yaml.safe_dump(doc, f, sort_keys=False)


def main() -> None:
    """Entry point for applying strict variant overlay."""
    apply_strict_overlay()


if __name__ == "__main__":
    main()
