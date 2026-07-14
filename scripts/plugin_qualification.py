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


def evaluate_operational_scenario(model, scenario):
    """Resolve one named operational condition without converting it into support."""
    condition = scenario.get("condition")
    contract = model.get("operational_conditions", {}).get(condition)
    if not isinstance(contract, dict):
        raise ValueError(f"unknown operational condition: {condition}")
    for field in ("mode", "outcome", "diagnostics", "remediation", "fallback"):
        if field not in contract:
            raise ValueError(f"{condition}: missing {field}")
    if contract["mode"] not in model.get("modes", {}):
        raise ValueError(f"{condition}: unknown mode {contract['mode']}")
    if not isinstance(contract["diagnostics"], list) or not contract["diagnostics"]:
        raise ValueError(f"{condition}: diagnostics must be non-empty")
    if not isinstance(contract["remediation"], list) or not contract["remediation"]:
        raise ValueError(f"{condition}: remediation must be non-empty")

    expected = scenario.get("expected", {})
    for field in ("mode", "outcome"):
        if expected.get(field) != contract[field]:
            raise ValueError(
                f"{scenario.get('id', '<unknown>')}: {field} is {contract[field]}, "
                f"expected {expected.get(field)}"
            )
    return {
        "id": scenario.get("id"),
        "mode": contract["mode"],
        "outcome": contract["outcome"],
        "diagnostics": list(contract["diagnostics"]),
        "remediation": list(contract["remediation"]),
        "fallback": contract["fallback"],
        "full_support": False,
    }


def evaluate_approval_scenario(boundaries, scenario):
    """Permit only effects named in the scenario's explicit approval set."""
    known = set(boundaries.get("approval_boundaries", {}))
    requested = scenario.get("requested_approvals")
    granted = scenario.get("explicit_approvals")
    if not isinstance(requested, list) or not isinstance(granted, list):
        raise ValueError("requested_approvals and explicit_approvals must be lists")
    if any(item not in known for item in requested + granted):
        raise ValueError("approval scenario contains an unknown boundary")
    requested_set = set(requested)
    granted_set = requested_set.intersection(granted)
    missing = requested_set - granted_set
    return {
        "id": scenario.get("id"),
        "permitted": not missing,
        "granted_approvals": sorted(granted_set),
        "missing_approvals": sorted(missing),
        "conversation_approval_ignored": scenario.get("conversation_approval") is True,
    }
