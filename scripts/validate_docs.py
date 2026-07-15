#!/usr/bin/env python3
"""Validate local Markdown references and required public repository files."""

from __future__ import annotations

import re
import sys
import json
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[1]
REQUIRED = (
    "README.md",
    "CHANGELOG.md",
    "product-version.json",
    "LICENSE",
    "AGENTS.md",
    "CONTRIBUTING.md",
    "SECURITY.md",
    "docs/README.md",
    "docs/product/PRD.md",
    "docs/product/PERSONAS.md",
    "docs/product/GLOSSARY.md",
    "docs/product/RELEASES.md",
    "docs/architecture/SUPPORT_MATRIX.md",
    "docs/decisions/INDEX.md",
    "docs/agents/domain.md",
    "docs/agents/issue-tracker.md",
    "docs/agents/triage-labels.md",
    ".github/pull_request_template.md",
    ".github/ISSUE_TEMPLATE/config.yml",
    ".github/ISSUE_TEMPLATE/feature.yml",
    ".github/ISSUE_TEMPLATE/bug.yml",
    ".github/ISSUE_TEMPLATE/task.yml",
    ".github/labels.yml",
    ".github/workflows/documentation.yml",
    "changes/README.md",
    "changes/schema-v1.json",
    "changes/release-admission-schema-v1.json",
)
LINK = re.compile(r"(?<!!)\[[^]]*]\(([^)]+)\)")
DECISION_ROUTE = re.compile(
    r"^\|\s*(D(?:[1-9]|1[0-5]))\s*\|\s*\[(DEC-\d{4})\]\(([^)]+\.md)\)\s*\|",
    re.MULTILINE,
)
DECISION_INDEX_LINK = re.compile(r"\[(DEC-\d{4})\]\(([^)]+\.md)\)")
APPROVED_DECISIONS = tuple(f"D{number}" for number in range(1, 16))
DECISION_MARKERS = (
    "**Status:** Accepted",
    "**Owner:**",
    "**Date:**",
    "## Context",
    "## Decision",
    "## Consequences",
    "## Source",
)
HEX_COLOR = re.compile(r"^[0-9A-Fa-f]{6}$")
PINNED_ACTION = re.compile(r"^\s*uses:\s*[^@\s]+@([0-9a-f]{40})(?:\s+#.*)?$", re.MULTILINE)
GLOSSARY_TERM = re.compile(r"^### (.+)$", re.MULTILINE)
SENSITIVE_BOUNDARY_MARKERS = {
    "docs/decisions/DEC-0001-first-release-scope.md": (
        "content classification",
        "handling authorization",
        "product assurance",
        "`No`, `Yes`, or `Unsure`",
        "issue #21",
    ),
    "docs/architecture/LIFECYCLES.md": (
        "Missing verified route outcome",
        "`needs-review`",
        "`unsupported`",
    ),
    "docs/product/PRD.md": (
        "Acknowledgment never authorizes tool access or transmission",
        "comprehensive highly sensitive or regulated-content",
    ),
}
PROHIBITED_ACTIVE_ASSURANCE_CLAIMS = (
    "genuinely supports regulated projects",
    "regulated projects must be genuinely supported in the first release",
    "regulated projects fully supported",
    "launch must support regulated projects",
    "regulatory policy and evidence supported at launch",
)


def _load_json_document(path: Path, root: Path) -> tuple[Any | None, list[str]]:
    if not path.is_file():
        return None, [f"{path.relative_to(root)}: missing file"]
    try:
        return json.loads(path.read_text(encoding="utf-8")), []
    except (json.JSONDecodeError, UnicodeDecodeError) as error:
        return None, [
            f"{path.relative_to(root)}: not JSON-compatible YAML: {error}"
        ]


def validate_label_manifest(root: Path) -> list[str]:
    """Validate the dependency-free JSON-compatible label manifest."""

    path = root / ".github/labels.yml"
    document, failures = _load_json_document(path, root)
    if failures:
        return failures
    if not isinstance(document, dict) or not isinstance(document.get("labels"), list):
        return [f"{path.relative_to(root)}: expected an object with a labels array"]

    seen: set[str] = set()
    for index, label in enumerate(document["labels"]):
        location = f"{path.relative_to(root)}: labels[{index}]"
        if not isinstance(label, dict):
            failures.append(f"{location}: expected an object")
            continue
        name = label.get("name")
        color = label.get("color")
        description = label.get("description")
        if not isinstance(name, str) or not name:
            failures.append(f"{location}: missing non-empty name")
        elif name in seen:
            failures.append(f"{location}: duplicate label: {name}")
        else:
            seen.add(name)
        if not isinstance(color, str) or not HEX_COLOR.fullmatch(color):
            failures.append(f"{location}: invalid color: {color}")
        if not isinstance(description, str) or not description:
            failures.append(f"{location}: missing non-empty description")
    return failures


def validate_issue_templates(root: Path) -> list[str]:
    """Validate GitHub issue forms encoded as JSON-compatible YAML."""

    directory = root / ".github/ISSUE_TEMPLATE"
    label_path = root / ".github/labels.yml"
    label_document, _ = _load_json_document(label_path, root)
    known_labels = {
        label.get("name")
        for label in label_document.get("labels", [])
        if isinstance(label_document, dict)
        and isinstance(label, dict)
        and isinstance(label.get("name"), str)
    } if isinstance(label_document, dict) else set()

    failures: list[str] = []
    forms = sorted(directory.glob("*.yml")) if directory.is_dir() else []
    if not forms:
        return [f"{directory.relative_to(root)}: no issue templates found"]

    for path in forms:
        document, parse_failures = _load_json_document(path, root)
        failures.extend(parse_failures)
        if parse_failures:
            continue
        relative = path.relative_to(root)
        if path.name == "config.yml":
            if not isinstance(document, dict) or not isinstance(
                document.get("blank_issues_enabled"), bool
            ):
                failures.append(f"{relative}: missing boolean blank_issues_enabled")
            continue
        if not isinstance(document, dict):
            failures.append(f"{relative}: expected an object")
            continue
        for key in ("name", "description"):
            if not isinstance(document.get(key), str) or not document[key]:
                failures.append(f"{relative}: missing non-empty {key}")
        labels = document.get("labels")
        if not isinstance(labels, list):
            failures.append(f"{relative}: labels must be an array")
        else:
            for label in labels:
                if label not in known_labels:
                    failures.append(f"{relative}: unknown label: {label}")
        body = document.get("body")
        if not isinstance(body, list) or not body:
            failures.append(f"{relative}: body must be a non-empty array")
            continue
        seen_ids: set[str] = set()
        for index, field in enumerate(body):
            if not isinstance(field, dict):
                failures.append(f"{relative}: body[{index}] must be an object")
                continue
            field_id = field.get("id")
            if field_id is None:
                continue
            if not isinstance(field_id, str) or not field_id:
                failures.append(f"{relative}: body[{index}] has invalid id")
            elif field_id in seen_ids:
                failures.append(f"{relative}: duplicate body id: {field_id}")
            else:
                seen_ids.add(field_id)
    return failures


def validate_workflow(root: Path) -> list[str]:
    """Validate the native three-OS, pinned-action foundation workflow contract."""

    path = root / ".github/workflows/documentation.yml"
    if not path.is_file():
        return [f"{path.relative_to(root)}: missing workflow"]
    text = path.read_text(encoding="utf-8")
    failures: list[str] = []
    for runner in ("ubuntu-latest", "macos-latest", "windows-latest"):
        if runner not in text:
            failures.append(f"{path.relative_to(root)}: missing native runner: {runner}")
    uses_lines = [line for line in text.splitlines() if line.strip().startswith("uses:")]
    for line in uses_lines:
        if not PINNED_ACTION.match(line):
            failures.append(f"{path.relative_to(root)}: action is not pinned: {line.strip()}")
    if "shell:" in text or re.search(r"\brun:\s*(?:bash|sh|pwsh|powershell)\b", text):
        failures.append(f"{path.relative_to(root)}: explicit shell dependency")
    for command in (
        "python scripts/validate_docs.py",
        "python -m unittest discover -s tests",
        "go test ./...",
        "go run ./cmd/starter-kit changes check --repository .",
        "go run ./cmd/phase1-evidence capture --output phase1-native-evidence.json",
        "go run ./cmd/phase1-evidence compare --directory phase1-native-evidence",
    ):
        if command not in text:
            failures.append(f"{path.relative_to(root)}: missing command: {command}")
    return failures


def validate_sensitive_data_boundary(root: Path) -> list[str]:
    """Validate durable routes for the amended v1 sensitive-data boundary."""

    failures: list[str] = []
    for relative, markers in SENSITIVE_BOUNDARY_MARKERS.items():
        path = root / relative
        if not path.is_file():
            failures.append(f"{relative}: missing sensitive-data boundary document")
            continue
        text = path.read_text(encoding="utf-8")
        for marker in markers:
            if marker not in text:
                failures.append(
                    f"{relative}: missing sensitive-data boundary marker: {marker}"
                )

    for path in root.rglob("*.md"):
        if ".git" in path.parts:
            continue
        text = path.read_text(encoding="utf-8")
        lowered = text.lower()
        for claim in PROHIBITED_ACTIVE_ASSURANCE_CLAIMS:
            start = lowered.find(claim)
            while start >= 0:
                historical_context = lowered[start : start + 1_000]
                if "superseded" not in historical_context:
                    failures.append(
                        f"{path.relative_to(root)}: unqualified first-release assurance "
                        f"claim: {claim}"
                    )
                start = lowered.find(claim, start + len(claim))
    return failures


def validate_glossary_order(root: Path) -> list[str]:
    """Require canonical glossary terms to remain alphabetically ordered."""

    path = root / "docs/product/GLOSSARY.md"
    if not path.is_file():
        return ["docs/product/GLOSSARY.md: missing glossary"]
    terms = GLOSSARY_TERM.findall(path.read_text(encoding="utf-8"))
    expected = sorted(terms, key=str.casefold)
    if terms == expected:
        return []
    for index, (actual, wanted) in enumerate(zip(terms, expected), start=1):
        if actual != wanted:
            return [
                "docs/product/GLOSSARY.md: terms are not alphabetically ordered: "
                f"position {index} is {actual!r}, expected {wanted!r}"
            ]
    return ["docs/product/GLOSSARY.md: terms are not alphabetically ordered"]


def validate_decision_routes(root: Path) -> list[str]:
    """Return failures for discovery routes and all durable decision records."""

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

    indexed_targets: dict[str, list[str]] = {}
    for record_id, target in DECISION_INDEX_LINK.findall(text):
        indexed_targets.setdefault(target, []).append(record_id)

    for record in sorted(index.parent.glob("DEC-[0-9][0-9][0-9][0-9]-*.md")):
        record_ids = indexed_targets.get(record.name, [])
        if not record_ids:
            failures.append(
                f"{record.relative_to(root)}: unindexed decision record"
            )
            continue
        if len(record_ids) > 1:
            failures.append(
                f"docs/decisions/INDEX.md: decision record indexed more than once: {record.name}"
            )
            continue
        expected_id = record.name[:8]
        if record_ids[0] != expected_id:
            failures.append(
                f"docs/decisions/INDEX.md: {record.name} uses {record_ids[0]}, expected {expected_id}"
            )
        record_text = record.read_text(encoding="utf-8")
        if not record_text.startswith(f"# {expected_id} "):
            failures.append(
                f"{record.relative_to(root)}: heading does not start with {expected_id}"
            )
        for marker in DECISION_MARKERS:
            if marker not in record_text:
                failures.append(
                    f"{record.relative_to(root)}: decision record missing required marker: {marker}"
                )

    for target in indexed_targets:
        record = index.parent / target
        if not record.is_file():
            failures.append(
                f"docs/decisions/INDEX.md: indexed decision record missing: {target}"
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
    failures.extend(validate_label_manifest(ROOT))
    failures.extend(validate_issue_templates(ROOT))
    failures.extend(validate_workflow(ROOT))
    failures.extend(validate_sensitive_data_boundary(ROOT))
    failures.extend(validate_glossary_order(ROOT))

    if failures:
        print("Documentation validation failed:", file=sys.stderr)
        for failure in failures:
            print(f"- {failure}", file=sys.stderr)
        return 1
    print("Documentation validation passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
