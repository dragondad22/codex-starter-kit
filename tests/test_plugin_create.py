import json
import hashlib
import unittest
from pathlib import Path

from scripts.plugin_create_eval import evaluate_scenario


ROOT = Path(__file__).resolve().parents[1]
PLUGIN = ROOT / "plugins" / "codex-starter-kit"


class PluginCreateContractTests(unittest.TestCase):
    def test_bundled_baseline_is_versioned_and_digest_bound(self):
        baseline_root = PLUGIN / "baselines" / "professional-v1"
        manifest = json.loads(
            (baseline_root / "baseline.json").read_text(encoding="utf-8")
        )
        content = (baseline_root / manifest["content_path"]).read_bytes()

        self.assertEqual(manifest["schema_version"], 1)
        self.assertEqual(manifest["id"], "baseline:professional-engineering:v1")
        self.assertEqual(manifest["source_decision"], "DEC-0017")
        self.assertEqual(
            manifest["content_sha256"], f"sha256:{hashlib.sha256(content).hexdigest()}"
        )

    def test_create_scenarios_preserve_approval_and_failure_boundaries(self):
        scenarios = json.loads(
            (PLUGIN / "evals" / "create-scenarios.json").read_text(encoding="utf-8")
        )
        results = {scenario["id"]: evaluate_scenario(scenario) for scenario in scenarios}

        self.assertEqual(results["happy-path"]["outcome"], "applied")
        self.assertEqual(results["no-change-replay"]["outcome"], "no_change")
        self.assertEqual(results["declaration-no"]["outcome"], "applied")
        self.assertEqual(results["declaration-yes"]["notice"], "acknowledged")
        self.assertEqual(results["declaration-unsure"]["notice"], "acknowledged")
        self.assertFalse(results["notice-not-acknowledged"]["invoked_plan"])
        self.assertFalse(results["declined-approval"]["invoked_apply"])
        self.assertFalse(results["missing-authority"]["invoked_plan"])
        self.assertEqual(results["malformed-plan"]["mode"], "unsupported")
        self.assertEqual(results["stale-precondition"]["outcome"], "failed")
        self.assertEqual(results["existing-content"]["outcome"], "reconciliation_required")
        self.assertFalse(results["existing-content"]["invoked_apply"])
        self.assertEqual(results["interrupted-setup"]["outcome"], "failed")
        self.assertFalse(results["rollback-failure"]["recoverable"])
        self.assertEqual(results["missing-engine"]["mode"], "degraded-guidance")
        self.assertEqual(results["missing-baseline"]["mode"], "degraded-guidance")
        self.assertFalse(results["unrelated-create-request"]["routed"])


if __name__ == "__main__":
    unittest.main()
