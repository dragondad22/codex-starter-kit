from pathlib import Path
import unittest


class Issue75WorkflowContractTests(unittest.TestCase):
    def test_apply_compares_nested_delivery_input_from_planning_artifact(self) -> None:
        root = Path(__file__).resolve().parents[1]
        workflow = (
            root / "docs/evidence/issue-75-sandbox-stage-apply.yml"
        ).read_text(encoding="utf-8")
        normalized_workflow = " ".join(workflow.replace("\\\n", "").split())

        self.assertIn(
            "cmp delivery-input/issue-75-delivery-input.json "
            "planning/delivery-input/issue-75-delivery-input.json",
            normalized_workflow,
        )
        self.assertNotIn(
            "cmp delivery-input/issue-75-delivery-input.json "
            "planning/issue-75-delivery-input.json",
            normalized_workflow,
        )


if __name__ == "__main__":
    unittest.main()
