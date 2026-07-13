from __future__ import annotations

import tempfile
import unittest
import json
from pathlib import Path

from scripts.validate_docs import (
    validate_decision_routes,
    validate_issue_templates,
    validate_glossary_order,
    validate_label_manifest,
    validate_sensitive_data_boundary,
    validate_workflow,
)


class GlossaryOrderValidationTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory()
        self.root = Path(self.temp_dir.name)
        (self.root / "docs/product").mkdir(parents=True)

    def tearDown(self) -> None:
        self.temp_dir.cleanup()

    def write_glossary(self, terms: list[str]) -> None:
        content = "# Glossary\n\n" + "\n\n".join(
            f"### {term}\n\nDefinition." for term in terms
        )
        (self.root / "docs/product/GLOSSARY.md").write_text(
            content, encoding="utf-8"
        )

    def test_accepts_case_insensitive_alphabetical_term_order(self) -> None:
        self.write_glossary(["Adapter", "Breadcrumb", "PRD", "Product assurance"])

        self.assertEqual(validate_glossary_order(self.root), [])

    def test_rejects_out_of_order_terms(self) -> None:
        self.write_glossary(["Readiness", "Question work item", "Status"])

        failures = validate_glossary_order(self.root)

        self.assertTrue(any("expected 'Question work item'" in failure for failure in failures))


class DecisionRouteValidationTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory()
        self.root = Path(self.temp_dir.name)
        (self.root / "docs/decisions").mkdir(parents=True)
        (self.root / "docs/discovery").mkdir(parents=True)
        (self.root / "docs/discovery/CODEX_STARTER_KIT_REVIEW.md").write_text(
            "\n".join(f'<a id="d{number}"></a>' for number in range(1, 16)),
            encoding="utf-8",
        )

    def tearDown(self) -> None:
        self.temp_dir.cleanup()

    def write_record(
        self, number: int, source: str | None = None, record_id: str | None = None
    ) -> str:
        name = f"DEC-{number:04d}-decision-{number}.md"
        source_id = source or f"D{number}"
        stable_id = record_id or f"DEC-{number:04d}"
        (self.root / "docs/decisions" / name).write_text(
            "\n".join(
                (
                    f"# {stable_id} — Decision {number}",
                    "",
                    "**Status:** Accepted",
                    "**Owner:** dragondad22",
                    "**Date:** 2026-07-11",
                    f"**Source decision:** {source_id}",
                    "",
                    "## Context",
                    "Context.",
                    "",
                    "## Decision",
                    "Decision.",
                    "",
                    "## Consequences",
                    "Consequences.",
                    "",
                    "## Source",
                    f"[Discovery decision {source_id}](../discovery/CODEX_STARTER_KIT_REVIEW.md#{source_id.lower()}).",
                )
            ),
            encoding="utf-8",
        )
        return name

    def write_index(self, rows: list[tuple[str, str]]) -> None:
        lines = [
            "# Decision Index",
            "",
            "| Source | Record |",
            "|---|---|",
            *(
                f"| {source} | [{name.split('-', 2)[0]}-{name.split('-', 2)[1]}]({name}) |"
                for source, name in rows
            ),
        ]
        (self.root / "docs/decisions/INDEX.md").write_text(
            "\n".join(lines), encoding="utf-8"
        )

    def test_accepts_exactly_one_route_for_each_approved_decision(self) -> None:
        rows = [(f"D{number}", self.write_record(number)) for number in range(1, 16)]
        self.write_index(rows)

        self.assertEqual(validate_decision_routes(self.root), [])

    def test_rejects_a_missing_route(self) -> None:
        rows = [(f"D{number}", self.write_record(number)) for number in range(1, 15)]
        self.write_index(rows)

        failures = validate_decision_routes(self.root)

        self.assertTrue(any("missing route: D15" in failure for failure in failures))

    def test_rejects_a_duplicate_route(self) -> None:
        rows = [(f"D{number}", self.write_record(number)) for number in range(1, 16)]
        rows.append(("D1", rows[0][1]))
        self.write_index(rows)

        failures = validate_decision_routes(self.root)

        self.assertTrue(any("duplicate route: D1" in failure for failure in failures))

    def test_rejects_a_record_whose_source_does_not_match_its_route(self) -> None:
        rows = []
        for number in range(1, 16):
            source = "D2" if number == 1 else None
            rows.append((f"D{number}", self.write_record(number, source)))
        self.write_index(rows)

        failures = validate_decision_routes(self.root)

        self.assertTrue(any("D1 record does not declare Source decision: D1" in failure for failure in failures))

    def test_rejects_a_record_missing_required_sections(self) -> None:
        rows = [(f"D{number}", self.write_record(number)) for number in range(1, 16)]
        broken = self.root / "docs/decisions" / rows[0][1]
        broken.write_text("# DEC-0001\n\n**Source decision:** D1\n", encoding="utf-8")
        self.write_index(rows)

        failures = validate_decision_routes(self.root)

        self.assertTrue(any("D1 record missing required marker" in failure for failure in failures))

    def test_rejects_a_mismatched_stable_decision_identity(self) -> None:
        rows = []
        for number in range(1, 16):
            record_id = "DEC-9999" if number == 1 else None
            rows.append((f"D{number}", self.write_record(number, record_id=record_id)))
        self.write_index(rows)

        failures = validate_decision_routes(self.root)

        self.assertTrue(any("D1 record heading does not start with DEC-0001" in failure for failure in failures))

    def test_rejects_one_record_target_reused_by_multiple_decisions(self) -> None:
        rows = [(f"D{number}", self.write_record(number)) for number in range(1, 16)]
        rows[1] = ("D2", rows[0][1])
        self.write_index(rows)

        failures = validate_decision_routes(self.root)

        self.assertTrue(any("record target reused" in failure for failure in failures))

    def test_rejects_a_source_link_without_the_exact_d_item_anchor(self) -> None:
        rows = [(f"D{number}", self.write_record(number)) for number in range(1, 16)]
        record = self.root / "docs/decisions" / rows[0][1]
        record.write_text(
            record.read_text(encoding="utf-8").replace(
                "CODEX_STARTER_KIT_REVIEW.md#d1", "CODEX_STARTER_KIT_REVIEW.md"
            ),
            encoding="utf-8",
        )
        self.write_index(rows)

        failures = validate_decision_routes(self.root)

        self.assertTrue(any("D1 record missing exact source breadcrumb" in failure for failure in failures))

    def test_rejects_a_missing_source_anchor(self) -> None:
        rows = [(f"D{number}", self.write_record(number)) for number in range(1, 16)]
        discovery = self.root / "docs/discovery/CODEX_STARTER_KIT_REVIEW.md"
        discovery.write_text(
            discovery.read_text(encoding="utf-8").replace('<a id="d1"></a>', ""),
            encoding="utf-8",
        )
        self.write_index(rows)

        failures = validate_decision_routes(self.root)

        self.assertTrue(any("discovery source missing anchor: d1" in failure for failure in failures))


class FoundationManifestValidationTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory()
        self.root = Path(self.temp_dir.name)
        (self.root / ".github/ISSUE_TEMPLATE").mkdir(parents=True)
        (self.root / ".github/workflows").mkdir(parents=True)

    def tearDown(self) -> None:
        self.temp_dir.cleanup()

    def write_labels(self, labels: list[dict[str, str]]) -> None:
        (self.root / ".github/labels.yml").write_text(
            json.dumps({"labels": labels}), encoding="utf-8"
        )

    def write_form(self, name: str, labels: list[str], ids: list[str]) -> None:
        body = [
            {
                "type": "textarea",
                "id": field_id,
                "attributes": {"label": field_id.title()},
                "validations": {"required": True},
            }
            for field_id in ids
        ]
        (self.root / ".github/ISSUE_TEMPLATE" / name).write_text(
            json.dumps(
                {
                    "name": name,
                    "description": "Description",
                    "title": "",
                    "labels": labels,
                    "body": body,
                }
            ),
            encoding="utf-8",
        )

    def test_accepts_valid_labels_and_issue_forms(self) -> None:
        self.write_labels(
            [
                {"name": "type:task", "color": "0075CA", "description": "Task"},
                {"name": "needs-triage", "color": "D4C5F9", "description": "Triage"},
            ]
        )
        self.write_form("task.yml", ["type:task", "needs-triage"], ["summary"])

        self.assertEqual(validate_label_manifest(self.root), [])
        self.assertEqual(validate_issue_templates(self.root), [])

    def test_rejects_malformed_label_json(self) -> None:
        (self.root / ".github/labels.yml").write_text("labels: [", encoding="utf-8")

        failures = validate_label_manifest(self.root)

        self.assertTrue(any("not JSON-compatible YAML" in failure for failure in failures))

    def test_rejects_duplicate_labels_and_invalid_colors(self) -> None:
        self.write_labels(
            [
                {"name": "type:task", "color": "not-hex", "description": "Task"},
                {"name": "type:task", "color": "0075CA", "description": "Duplicate"},
            ]
        )

        failures = validate_label_manifest(self.root)

        self.assertTrue(any("duplicate label" in failure for failure in failures))
        self.assertTrue(any("invalid color" in failure for failure in failures))

    def test_rejects_issue_form_unknown_labels_and_duplicate_ids(self) -> None:
        self.write_labels(
            [{"name": "type:task", "color": "0075CA", "description": "Task"}]
        )
        self.write_form("task.yml", ["type:task", "missing"], ["summary", "summary"])

        failures = validate_issue_templates(self.root)

        self.assertTrue(any("unknown label: missing" in failure for failure in failures))
        self.assertTrue(any("duplicate body id: summary" in failure for failure in failures))

    def test_accepts_native_three_os_workflow_with_pinned_action(self) -> None:
        (self.root / ".github/workflows/documentation.yml").write_text(
            "\n".join(
                (
                    "matrix:",
                    "  os: [ubuntu-latest, macos-latest, windows-latest]",
                    "uses: actions/checkout@34e114876b0b11c390a56381ad16ebd13914f8d5",
                    "run: python scripts/validate_docs.py",
                    "run: python -m unittest discover -s tests",
                    "run: go test ./...",
                    "run: go run ./cmd/phase1-evidence capture --output phase1-native-evidence.json",
                    "run: go run ./cmd/phase1-evidence compare --directory phase1-native-evidence",
                )
            ),
            encoding="utf-8",
        )

        self.assertEqual(validate_workflow(self.root), [])

    def test_rejects_missing_os_mutable_action_and_shell_dependency(self) -> None:
        (self.root / ".github/workflows/documentation.yml").write_text(
            "\n".join(
                (
                    "os: [ubuntu-latest, macos-latest]",
                    "uses: actions/checkout@v4",
                    "shell: bash",
                    "run: python scripts/validate_docs.py",
                )
            ),
            encoding="utf-8",
        )

        failures = validate_workflow(self.root)

        self.assertTrue(any("missing native runner: windows-latest" in failure for failure in failures))
        self.assertTrue(any("action is not pinned" in failure for failure in failures))
        self.assertTrue(any("explicit shell dependency" in failure for failure in failures))
        self.assertTrue(any("phase1-evidence compare" in failure for failure in failures))


class SensitiveDataBoundaryValidationTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory()
        self.root = Path(self.temp_dir.name)
        for relative, content in {
            "docs/decisions/DEC-0001-first-release-scope.md": (
                "content classification; handling authorization; product assurance; "
                "`No`, `Yes`, or `Unsure`; issue #21"
            ),
            "docs/architecture/LIFECYCLES.md": (
                "Missing verified route outcome; `needs-review`; `unsupported`"
            ),
            "docs/product/PRD.md": (
                "Acknowledgment never authorizes tool access or transmission; "
                "comprehensive highly sensitive or regulated-content"
            ),
        }.items():
            path = self.root / relative
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(content, encoding="utf-8")

    def tearDown(self) -> None:
        self.temp_dir.cleanup()

    def test_accepts_complete_sensitive_data_boundary_routes(self) -> None:
        self.assertEqual(validate_sensitive_data_boundary(self.root), [])

    def test_rejects_missing_sensitive_data_boundary_marker(self) -> None:
        path = self.root / "docs/product/PRD.md"
        path.write_text(
            "comprehensive highly sensitive or regulated-content",
            encoding="utf-8",
        )

        failures = validate_sensitive_data_boundary(self.root)

        self.assertTrue(
            any(
                "missing sensitive-data boundary marker" in failure
                for failure in failures
            )
        )

    def test_rejects_unqualified_first_release_assurance_claim_anywhere(self) -> None:
        readme = self.root / "README.md"
        readme.write_text(
            "The launch must support regulated projects.", encoding="utf-8"
        )

        failures = validate_sensitive_data_boundary(self.root)

        self.assertTrue(
            any("unqualified first-release assurance claim" in failure for failure in failures)
        )

    def test_allows_a_superseded_historical_assurance_claim(self) -> None:
        history = self.root / "docs/discovery/HISTORY.md"
        history.parent.mkdir(parents=True, exist_ok=True)
        history.write_text(
            "Regulated projects fully supported. This statement is superseded by the "
            "amended v1 assurance boundary.",
            encoding="utf-8",
        )

        self.assertEqual(validate_sensitive_data_boundary(self.root), [])


if __name__ == "__main__":
    unittest.main()
