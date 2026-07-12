from __future__ import annotations

import tempfile
import unittest
from pathlib import Path

from scripts.validate_docs import validate_decision_routes


class DecisionRouteValidationTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory()
        self.root = Path(self.temp_dir.name)
        (self.root / "docs/decisions").mkdir(parents=True)
        (self.root / "docs/discovery").mkdir(parents=True)
        (self.root / "docs/discovery/CODEX_STARTER_KIT_REVIEW.md").write_text(
            "\n".join(f'<a id="d{number}"></a>' for number in range(1, 13)),
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
        rows = [(f"D{number}", self.write_record(number)) for number in range(1, 13)]
        self.write_index(rows)

        self.assertEqual(validate_decision_routes(self.root), [])

    def test_rejects_a_missing_route(self) -> None:
        rows = [(f"D{number}", self.write_record(number)) for number in range(1, 12)]
        self.write_index(rows)

        failures = validate_decision_routes(self.root)

        self.assertTrue(any("missing route: D12" in failure for failure in failures))

    def test_rejects_a_duplicate_route(self) -> None:
        rows = [(f"D{number}", self.write_record(number)) for number in range(1, 13)]
        rows.append(("D1", rows[0][1]))
        self.write_index(rows)

        failures = validate_decision_routes(self.root)

        self.assertTrue(any("duplicate route: D1" in failure for failure in failures))

    def test_rejects_a_record_whose_source_does_not_match_its_route(self) -> None:
        rows = []
        for number in range(1, 13):
            source = "D2" if number == 1 else None
            rows.append((f"D{number}", self.write_record(number, source)))
        self.write_index(rows)

        failures = validate_decision_routes(self.root)

        self.assertTrue(any("D1 record does not declare Source decision: D1" in failure for failure in failures))

    def test_rejects_a_record_missing_required_sections(self) -> None:
        rows = [(f"D{number}", self.write_record(number)) for number in range(1, 13)]
        broken = self.root / "docs/decisions" / rows[0][1]
        broken.write_text("# DEC-0001\n\n**Source decision:** D1\n", encoding="utf-8")
        self.write_index(rows)

        failures = validate_decision_routes(self.root)

        self.assertTrue(any("D1 record missing required marker" in failure for failure in failures))

    def test_rejects_a_mismatched_stable_decision_identity(self) -> None:
        rows = []
        for number in range(1, 13):
            record_id = "DEC-9999" if number == 1 else None
            rows.append((f"D{number}", self.write_record(number, record_id=record_id)))
        self.write_index(rows)

        failures = validate_decision_routes(self.root)

        self.assertTrue(any("D1 record heading does not start with DEC-0001" in failure for failure in failures))

    def test_rejects_one_record_target_reused_by_multiple_decisions(self) -> None:
        rows = [(f"D{number}", self.write_record(number)) for number in range(1, 13)]
        rows[1] = ("D2", rows[0][1])
        self.write_index(rows)

        failures = validate_decision_routes(self.root)

        self.assertTrue(any("record target reused" in failure for failure in failures))

    def test_rejects_a_source_link_without_the_exact_d_item_anchor(self) -> None:
        rows = [(f"D{number}", self.write_record(number)) for number in range(1, 13)]
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
        rows = [(f"D{number}", self.write_record(number)) for number in range(1, 13)]
        discovery = self.root / "docs/discovery/CODEX_STARTER_KIT_REVIEW.md"
        discovery.write_text(
            discovery.read_text(encoding="utf-8").replace('<a id="d1"></a>', ""),
            encoding="utf-8",
        )
        self.write_index(rows)

        failures = validate_decision_routes(self.root)

        self.assertTrue(any("discovery source missing anchor: d1" in failure for failure in failures))


if __name__ == "__main__":
    unittest.main()
