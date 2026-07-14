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


if __name__ == "__main__":
    unittest.main()
