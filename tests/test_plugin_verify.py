import json
import unittest
from pathlib import Path

from scripts.plugin_verify_eval import evaluate_scenario


ROOT = Path(__file__).resolve().parents[1]
PLUGIN = ROOT / "plugins" / "codex-starter-kit"


class PluginVerifyContractTests(unittest.TestCase):
    def test_verify_scenarios_preserve_states_approval_and_evidence(self):
        scenarios = json.loads(
            (PLUGIN / "evals" / "verify-scenarios.json").read_text(encoding="utf-8")
        )
        results = {scenario["id"]: evaluate_scenario(scenario) for scenario in scenarios}

        for state in (
            "pass",
            "fail",
            "not-applicable",
            "not-configured",
            "needs-review",
            "accepted-exception",
        ):
            self.assertEqual(results[f"state-{state}"]["control_state"], state)

        self.assertEqual(
            results["state-accepted-exception"]["underlying_state"], "fail"
        )
        self.assertEqual(results["mixed-aggregate"]["overall_state"], "fail")
        self.assertFalse(results["declined-approval"]["invoked_verify"])
        self.assertEqual(results["stale-plan"]["outcome"], "failed")
        self.assertEqual(results["malformed-output"]["mode"], "unsupported")
        self.assertEqual(results["evaluator-failure"]["overall_state"], "fail")
        self.assertNotIn("TOPSECRET", json.dumps(results["redacted-diagnostics"]))
        self.assertEqual(results["plugin-unavailable"]["mode"], "degraded-guidance")
        self.assertFalse(results["unrelated-verify-request"]["routed"])


if __name__ == "__main__":
    unittest.main()
