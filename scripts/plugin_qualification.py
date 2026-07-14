"""Versioned Phase 2 capability-model evaluator used by tests and qualification reports."""


def _select_mode(workflow, facts):
    required_booleans = (
        "plugin_enabled",
        "engine_available",
        "engine_verified",
        "engine_compatible",
        "baseline_verified",
        "requested_authorized",
        "mutation_authorized",
        "evidence_write_authorized",
    )
    if not isinstance(facts, dict) or any(
        field not in facts or not isinstance(facts[field], bool)
        for field in required_booleans
    ):
        return "unsupported"
    if facts.get("malformed_or_conflicting") is True:
        return "unsupported"
    if not facts["requested_authorized"]:
        return "unsupported"

    compatible = all(
        facts[field]
        for field in (
            "plugin_enabled",
            "engine_available",
            "engine_verified",
            "engine_compatible",
            "baseline_verified",
        )
    )
    if compatible:
        if facts["mutation_authorized"]:
            return "full"
        if workflow in {"status", "verify"} and facts["evidence_write_authorized"]:
            return "verification-only"
        if workflow == "create":
            return "verification-only"
        return "unsupported"

    known = all(
        isinstance(facts[field], bool)
        for field in (
            "plugin_enabled",
            "engine_available",
            "engine_verified",
            "engine_compatible",
            "baseline_verified",
        )
    )
    return "degraded-guidance" if known else "unsupported"


def evaluate_mode_scenario(model, scenario):
    """Evaluate one workflow against capability facts and model permissions."""
    workflow = scenario.get("workflow")
    if workflow not in model.get("workflows", {}):
        raise ValueError(f"unknown workflow: {workflow}")
    mode = _select_mode(workflow, scenario.get("facts"))
    expected_mode = scenario.get("expected_mode")
    if mode != expected_mode:
        raise ValueError(
            f"{scenario.get('id', '<unknown>')}: selected {mode}, expected {expected_mode}"
        )
    mode_contract = model["modes"][mode]
    permissions = mode_contract["workflows"].get(workflow, {})
    return {
        "id": scenario.get("id"),
        "workflow": workflow,
        "mode": mode,
        "engine_invocation": permissions.get("engine_invocation", False),
        "lifecycle_effect": permissions.get("lifecycle_effect", False),
        "evidence_effect": permissions.get("evidence_effect", False),
        "remediation": list(mode_contract["remediation"]),
        "fallback": mode_contract["fallback"],
    }
