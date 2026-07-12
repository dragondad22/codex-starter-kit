#!/usr/bin/env python3
"""Validate local Markdown references and required public repository files."""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
REQUIRED = (
    "README.md",
    "LICENSE",
    "AGENTS.md",
    "CONTRIBUTING.md",
    "SECURITY.md",
    "docs/README.md",
    "docs/product/PRD.md",
    "docs/product/PERSONAS.md",
    "docs/product/GLOSSARY.md",
)
LINK = re.compile(r"(?<!!)\[[^]]*]\(([^)]+)\)")
DECISION_ROUTE = re.compile(
    r"^\|\s*(D(?:[1-9]|1[0-2]))\s*\|\s*\[(DEC-\d{4})\]\(([^)]+\.md)\)\s*\|",
    re.MULTILINE,
)
APPROVED_DECISIONS = tuple(f"D{number}" for number in range(1, 13))
DECISION_MARKERS = (
    "**Status:** Accepted",
    "**Owner:**",
    "**Date:**",
    "## Context",
    "## Decision",
    "## Consequences",
    "## Source",
)


def validate_decision_routes(root: Path) -> list[str]:
    """Return failures for the D1-D12 durable decision routes."""

    index = root / "docs/decisions/INDEX.md"
    if not index.is_file():
        return ["docs/decisions/INDEX.md: missing decision index"]

    text = index.read_text(encoding="utf-8")
    discovery = root / "docs/discovery/CODEX_STARTER_KIT_REVIEW.md"
    discovery_text = discovery.read_text(encoding="utf-8") if discovery.is_file() else ""
    routes: dict[str, list[tuple[str, str]]] = {
        source: [] for source in APPROVED_DECISIONS
    }
    target_sources: dict[str, list[str]] = {}
    for source, record_id, target in DECISION_ROUTE.findall(text):
        routes[source].append((record_id, target))
        target_sources.setdefault(target, []).append(source)

    failures: list[str] = []
    for source in APPROVED_DECISIONS:
        source_anchor = source.lower()
        if f'<a id="{source_anchor}"></a>' not in discovery_text:
            failures.append(
                f"docs/discovery/CODEX_STARTER_KIT_REVIEW.md: discovery source missing anchor: {source_anchor}"
            )
        targets = routes[source]
        if not targets:
            failures.append(f"docs/decisions/INDEX.md: missing route: {source}")
            continue
        if len(targets) > 1:
            failures.append(f"docs/decisions/INDEX.md: duplicate route: {source}")
            continue

        record_id, target = targets[0]
        expected_id = f"DEC-{int(source[1:]):04d}"
        if record_id != expected_id:
            failures.append(
                f"docs/decisions/INDEX.md: {source} route uses {record_id}, expected {expected_id}"
            )

        record = index.parent / target
        if not record.is_file():
            failures.append(f"docs/decisions/INDEX.md: {source} target missing: {target}")
            continue

        record_text = record.read_text(encoding="utf-8")
        if f"**Source decision:** {source}" not in record_text:
            failures.append(
                f"{record.relative_to(root)}: {source} record does not declare "
                f"Source decision: {source}"
            )
        if not record_text.startswith(f"# {expected_id} "):
            failures.append(
                f"{record.relative_to(root)}: {source} record heading does not start with {expected_id}"
            )
        expected_source = (
            f"../discovery/CODEX_STARTER_KIT_REVIEW.md#{source.lower()}"
        )
        if expected_source not in record_text:
            failures.append(
                f"{record.relative_to(root)}: {source} record missing exact source breadcrumb: "
                f"{expected_source}"
            )
        for marker in DECISION_MARKERS:
            if marker not in record_text:
                failures.append(
                    f"{record.relative_to(root)}: {source} record missing required marker: {marker}"
                )
    for target, sources in target_sources.items():
        if len(sources) > 1:
            failures.append(
                f"docs/decisions/INDEX.md: record target reused by {', '.join(sources)}: {target}"
            )
    return failures


def main() -> int:
    failures: list[str] = []
    for relative in REQUIRED:
        if not (ROOT / relative).is_file():
            failures.append(f"missing required file: {relative}")

    for document in ROOT.rglob("*.md"):
        if ".git" in document.parts:
            continue
        text = document.read_text(encoding="utf-8")
        for raw_target in LINK.findall(text):
            target = raw_target.strip().split("#", 1)[0]
            if not target or "://" in target or target.startswith(("mailto:", "#")):
                continue
            resolved = (document.parent / target).resolve()
            try:
                resolved.relative_to(ROOT)
            except ValueError:
                failures.append(f"{document.relative_to(ROOT)}: link escapes repo: {raw_target}")
                continue
            if not resolved.exists():
                failures.append(f"{document.relative_to(ROOT)}: missing link target: {raw_target}")

    failures.extend(validate_decision_routes(ROOT))

    if failures:
        print("Documentation validation failed:", file=sys.stderr)
        for failure in failures:
            print(f"- {failure}", file=sys.stderr)
        return 1
    print("Documentation validation passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
