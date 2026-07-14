import json
import unittest
from pathlib import Path

from scripts.plugin_qualification import (
    evaluate_approval_scenario,
    evaluate_mode_scenario,
    evaluate_operational_scenario,
)


ROOT = Path(__file__).resolve().parents[1]
PLUGIN = ROOT / "plugins" / "codex-starter-kit"


class PluginQualificationContractTests(unittest.TestCase):
    def test_capability_model_and_scenarios_cover_every_workflow_mode_pair(self):
        model = json.loads(
            (PLUGIN / "contracts" / "capability-model-v1.json").read_text(
                encoding="utf-8"
            )
        )
        scenarios = json.loads(
            (PLUGIN / "evals" / "qualification-scenarios.json").read_text(
                encoding="utf-8"
            )
        )

        self.assertEqual(model["schema_version"], 1)
        self.assertEqual(model["id"], "codex-starter-kit:capability-model:v1")
        self.assertEqual(model["plugin_version"], "0.3.0")
        self.assertEqual(
            set(model["modes"]),
            {"full", "degraded-guidance", "verification-only", "unsupported"},
        )
        self.assertEqual(model["modes"]["full"]["diagnostics"], [])
        for mode in ("degraded-guidance", "verification-only", "unsupported"):
            self.assertTrue(model["modes"][mode]["diagnostics"])

        results = [evaluate_mode_scenario(model, scenario) for scenario in scenarios]
        pairs = {(result["workflow"], result["mode"]) for result in results}
        self.assertEqual(
            pairs,
            {
                (workflow, mode)
                for workflow in ("create", "status", "verify")
                for mode in model["modes"]
            },
        )
        self.assertFalse(
            next(
                result
                for result in results
                if result["workflow"] == "create"
                and result["mode"] == "verification-only"
            )["engine_invocation"]
        )
        self.assertTrue(
            next(
                result
                for result in results
                if result["workflow"] == "verify"
                and result["mode"] == "verification-only"
            )["engine_invocation"]
        )
        self.assertTrue(
            all(not result["lifecycle_effect"] for result in results if result["mode"] != "full")
        )

    def test_evidence_write_authority_applies_only_to_effecting_workflows(self):
        model = json.loads(
            (PLUGIN / "contracts" / "capability-model-v1.json").read_text(
                encoding="utf-8"
            )
        )
        facts = {
            "plugin_enabled": True,
            "engine_available": True,
            "engine_verified": True,
            "engine_compatible": True,
            "baseline_verified": True,
            "requested_authorized": True,
            "mutation_authorized": False,
            "evidence_write_authorized": False,
            "malformed_or_conflicting": False,
        }

        status = evaluate_mode_scenario(
            model,
            {
                "id": "status-read-only-without-evidence-write",
                "workflow": "status",
                "expected_mode": "verification-only",
                "facts": facts,
            },
        )
        verify = evaluate_mode_scenario(
            model,
            {
                "id": "verify-denied-evidence-write",
                "workflow": "verify",
                "expected_mode": "unsupported",
                "facts": facts,
            },
        )
        self.assertTrue(status["engine_invocation"])
        self.assertFalse(status["evidence_effect"])
        self.assertFalse(verify["engine_invocation"])

    def test_operational_failures_preserve_exact_mode_diagnostics_and_recovery(self):
        model = json.loads(
            (PLUGIN / "contracts" / "capability-model-v1.json").read_text(
                encoding="utf-8"
            )
        )
        scenarios = json.loads(
            (PLUGIN / "evals" / "operational-scenarios.json").read_text(
                encoding="utf-8"
            )
        )

        results = {
            scenario["id"]: evaluate_operational_scenario(model, scenario)
            for scenario in scenarios
        }
        self.assertEqual(
            set(results),
            {
                "first-run-offline",
                "restricted-workspace",
                "plugin-admin-disabled",
                "missing-engine",
                "incompatible-engine",
                "malformed-engine-output",
                "interrupted-operation",
                "cancelled-operation",
                "recoverable-operation",
            },
        )
        self.assertEqual(results["first-run-offline"]["mode"], "degraded-guidance")
        self.assertEqual(results["plugin-admin-disabled"]["mode"], "degraded-guidance")
        self.assertEqual(results["malformed-engine-output"]["mode"], "unsupported")
        self.assertEqual(results["interrupted-operation"]["outcome"], "failed")
        self.assertEqual(results["cancelled-operation"]["outcome"], "cancelled")
        self.assertEqual(results["recoverable-operation"]["outcome"], "recovery_available")
        for result in results.values():
            self.assertTrue(result["diagnostics"])
            self.assertTrue(result["remediation"])
            self.assertIn("fallback", result)
            self.assertFalse(result["full_support"])

    def test_approvals_and_installation_domains_remain_separate(self):
        boundaries = json.loads(
            (PLUGIN / "contracts" / "approval-boundaries-v1.json").read_text(
                encoding="utf-8"
            )
        )
        scenarios = json.loads(
            (PLUGIN / "evals" / "approval-scenarios.json").read_text(
                encoding="utf-8"
            )
        )

        results = {
            scenario["id"]: evaluate_approval_scenario(boundaries, scenario)
            for scenario in scenarios
        }
        approval_ids = set(boundaries["approval_boundaries"])
        self.assertEqual(
            approval_ids,
            {
                "plan",
                "repository-effects",
                "network-access",
                "tool-installation",
                "data-handling",
                "authority-change",
            },
        )
        for approval_id in approval_ids:
            inferred = results[f"conversation-does-not-approve-{approval_id}"]
            explicit = results[f"explicitly-approve-{approval_id}"]
            self.assertFalse(inferred["permitted"])
            self.assertEqual(inferred["missing_approvals"], [approval_id])
            self.assertTrue(explicit["permitted"])
            self.assertEqual(explicit["granted_approvals"], [approval_id])

        domains = boundaries["operation_domains"]
        self.assertEqual(domains["plugin-install"], ["plugin"])
        self.assertEqual(domains["plugin-update"], ["plugin"])
        self.assertEqual(domains["engine-install"], ["engine"])
        self.assertEqual(domains["engine-update"], ["engine"])
        self.assertEqual(domains["repository-upgrade"], ["repository"])
        self.assertEqual(domains["baseline-resolution"], ["baseline-policy"])

    def test_live_routing_results_are_bounded_and_costed(self):
        evidence = json.loads(
            (PLUGIN / "evals" / "live-routing-results.json").read_text(
                encoding="utf-8"
            )
        )
        results = {result["scenario_id"]: result for result in evidence["results"]}

        self.assertEqual(len(results), 8)
        for workflow in ("create", "status", "verify"):
            self.assertEqual(
                results[f"explicit-{workflow}"]["routed_skill"],
                f"starter-kit-{workflow}",
            )
            self.assertEqual(
                results[f"implicit-{workflow}"]["routed_skill"],
                f"starter-kit-{workflow}",
            )
        self.assertEqual(results["unrelated-create"]["routed_skill"], "none")
        self.assertEqual(results["unrelated-verify"]["routed_skill"], "none")
        for result in results.values():
            self.assertEqual(result["loaded_references"], [])
            self.assertFalse(result["engine_invocation_planned"])
            self.assertFalse(result["lifecycle_effect_planned"])

        self.assertEqual(evidence["usage"]["input_tokens"], 112920)
        self.assertEqual(evidence["usage"]["cached_input_tokens"], 13056)
        self.assertEqual(evidence["usage"]["output_tokens"], 515)
        self.assertEqual(evidence["effects"], [])

    def test_compatibility_report_and_quality_receipt_are_complete(self):
        report = json.loads(
            (ROOT / "docs" / "evidence" / "phase2-plugin-compatibility.json").read_text(
                encoding="utf-8"
            )
        )
        receipt = json.loads(
            (ROOT / "docs" / "evidence" / "phase2-plugin-quality-receipt.json").read_text(
                encoding="utf-8"
            )
        )

        self.assertEqual(report["schema_version"], 1)
        self.assertEqual(report["plugin"]["version"], "0.3.0")
        self.assertEqual(report["engine"]["qualification"], "not-available")
        self.assertEqual(report["baseline"]["id"], "baseline:professional-engineering:v1")
        self.assertTrue(report["source"]["revision"])
        self.assertTrue(report["freshness"]["captured_at"])
        self.assertEqual(
            {surface["status"] for surface in report["surfaces"]},
            {"pass", "needs-review", "unsupported"},
        )
        self.assertEqual(
            {native["os"] for native in report["native_environments"]},
            {"Linux", "macOS", "Windows"},
        )
        self.assertTrue(report["limitations"])

        categories = {result["category"]: result for result in receipt["results"]}
        self.assertEqual(
            set(categories),
            {
                "functional",
                "security",
                "interaction",
                "accessibility",
                "testing",
                "documentation",
                "compatibility",
                "evidence",
            },
        )
        for result in categories.values():
            self.assertIn(
                result["state"],
                {"pass", "fail", "not-applicable", "not-configured", "needs-review"},
            )
            self.assertTrue(result["evidence"])
        self.assertEqual(categories["accessibility"]["state"], "needs-review")


if __name__ == "__main__":
    unittest.main()
