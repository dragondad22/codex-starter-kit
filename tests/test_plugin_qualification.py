import json
import unittest
from pathlib import Path

from scripts.plugin_qualification import evaluate_mode_scenario


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


if __name__ == "__main__":
    unittest.main()
