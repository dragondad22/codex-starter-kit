import json
import unittest
from pathlib import Path

from scripts.plugin_status_eval import evaluate_scenario


ROOT = Path(__file__).resolve().parents[1]
PLUGIN = ROOT / "plugins" / "codex-starter-kit"


class PluginStatusContractTests(unittest.TestCase):
    def test_plugin_exposes_only_the_focused_lifecycle_skills(self):
        manifest = json.loads(
            (PLUGIN / ".codex-plugin" / "plugin.json").read_text(encoding="utf-8")
        )
        self.assertEqual(manifest["skills"], "./skills/")
        self.assertEqual(manifest["interface"]["capabilities"], [])

        skills = sorted((PLUGIN / "skills").glob("*/SKILL.md"))
        self.assertEqual(
            [path.parent.name for path in skills], ["create", "status", "verify"]
        )

        skill = (PLUGIN / "skills" / "status" / "SKILL.md").read_text(encoding="utf-8")
        self.assertIn("name: starter-kit-status", skill)
        self.assertIn("repository lifecycle status", skill.lower())
        self.assertIn("managed", skill.lower())

    def test_supported_status_scenarios_fail_closed(self):
        scenarios = json.loads(
            (PLUGIN / "evals" / "status-scenarios.json").read_text(encoding="utf-8")
        )
        results = {scenario["id"]: evaluate_scenario(scenario) for scenario in scenarios}

        self.assertEqual(results["explicit-managed"]["mode"], "full")
        self.assertEqual(results["implicit-unmanaged"]["lifecycle"], "unmanaged")
        self.assertEqual(
            results["explicit-non-pass"]["lifecycle"], "managed_degraded"
        )
        self.assertEqual(results["malformed-output"]["mode"], "unsupported")
        self.assertEqual(results["missing-engine"]["mode"], "degraded-guidance")
        self.assertEqual(results["incompatible-engine"]["mode"], "degraded-guidance")
        self.assertEqual(results["unverified-engine"]["mode"], "degraded-guidance")
        self.assertEqual(
            results["administratively-unavailable"]["mode"], "degraded-guidance"
        )
        self.assertFalse(results["unrelated-git-status"]["routed"])
        self.assertEqual(
            results["explicit-non-pass"]["problems"],
            ["managed artifact digest mismatch"],
        )


if __name__ == "__main__":
    unittest.main()
