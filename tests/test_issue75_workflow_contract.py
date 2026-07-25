from pathlib import Path
import unittest


class Issue75WorkflowContractTests(unittest.TestCase):
    def test_stage_workflows_route_candidate_and_orphan_recovery_to_seeder(self) -> None:
        root = Path(__file__).resolve().parents[1]
        for name in [
            "issue-75-sandbox-stage-plan.yml",
            "issue-75-sandbox-stage-apply.yml",
        ]:
            workflow = (root / "docs/evidence" / name).read_text(encoding="utf-8")
            self.assertIn("file-candidate", workflow)
            self.assertIn("cleanup-orphan-branch", workflow)
            self.assertIn(
                "file-candidate|file-stale|cleanup-orphan-branch|cleanup-delivery",
                workflow,
            )
            self.assertIn("delivery_state_run_id", workflow)
            self.assertIn("issue-75-delivery-state-$run_id", workflow)
            self.assertIn("issue-75-transition.json", workflow)

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

    def test_delivery_input_derives_final_workflow_with_a_bound_placeholder_head(self) -> None:
        root = Path(__file__).resolve().parents[1]
        workflow = (
            root / "docs/evidence/issue-75-delivery-input.yml"
        ).read_text(encoding="utf-8")
        self.assertIn('--branch-head-sha "$SOURCE_REVISION"', workflow)

    def test_issue_identity_handoff_retains_the_postcondition_root(self) -> None:
        root = Path(__file__).resolve().parents[1]
        workflow = (
            root / "docs/evidence/issue-75-sandbox-stage-apply.yml"
        ).read_text(encoding="utf-8")

        self.assertIn(". as $postcondition |", workflow)
        self.assertIn(
            "source_revision: $postcondition.plan.source_revision",
            workflow,
        )

    def test_delivery_episode_reset_uses_the_executable_historical_state_gate(
        self,
    ) -> None:
        root = Path(__file__).resolve().parents[1]
        workflow = (
            root / "docs/evidence/issue-75-contract.yml"
        ).read_text(encoding="utf-8")

        self.assertIn(
            '--name "issue-75-delivery-state-$latest_run_id" --dir latest-state',
            workflow,
        )
        self.assertIn("--historical-state-directory latest-state", workflow)
        self.assertIn('--historical-state-run-id "$latest_run_id"', workflow)
        self.assertNotIn("ref_response=", workflow)
        self.assertLess(
            workflow.index("Build the exact reviewed transition runner"),
            workflow.index("Require paired prior-state inputs"),
        )


if __name__ == "__main__":
    unittest.main()
