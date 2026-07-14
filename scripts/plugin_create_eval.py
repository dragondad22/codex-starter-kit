"""Deterministic policy oracle for guided-create skill evaluations.

This is development and CI evaluation code, not a plugin runtime dependency.
"""


def _guidance(mode="degraded-guidance"):
    return {
        "routed": True,
        "mode": mode,
        "invoked_plan": False,
        "invoked_apply": False,
    }


def evaluate_scenario(scenario):
    """Evaluate the create workflow's routing, authority, and effect boundaries."""
    if scenario.get("route") is False:
        return {
            "routed": False,
            "mode": "not-routed",
            "invoked_plan": False,
            "invoked_apply": False,
        }

    capability = scenario.get("capability")
    if capability in {"missing-engine", "missing-baseline", "unverified-engine"}:
        return _guidance()
    if capability != "qualified":
        return _guidance("unsupported")

    inputs = scenario.get("inputs")
    if not isinstance(inputs, dict):
        return _guidance("unsupported")
    declaration = inputs.get("special_data")
    if declaration not in {"No", "Yes", "Unsure"}:
        return _guidance("unsupported")
    if not inputs.get("brief") or inputs.get("brief_approved") is not True:
        return {
            **_guidance("full"),
            "outcome": "stopped_before_plan",
            "reason": "approved brief required",
        }
    if inputs.get("owner_persona_confirmed") is not True:
        return {
            **_guidance("full"),
            "outcome": "stopped_before_plan",
            "reason": "owner persona confirmation required",
        }
    if declaration in {"Yes", "Unsure"} and inputs.get("notice_acknowledged") is not True:
        return {
            **_guidance("full"),
            "outcome": "stopped_before_plan",
            "reason": "special-data notice acknowledgment required",
            "notice": "not-acknowledged",
        }

    notice = "not-required" if declaration == "No" else "acknowledged"
    plan = scenario.get("plan")
    if plan == "reconciliation_required":
        result = scenario.get("result", {})
        return {
            "routed": True,
            "mode": "full",
            "invoked_plan": True,
            "invoked_apply": False,
            "outcome": "reconciliation_required",
            "notice": notice,
            "conflicts": list(result.get("conflicts", [])),
            "recovery": list(result.get("recovery", [])),
            "evidence": list(result.get("evidence", [])),
        }
    if plan != "valid":
        return {
            "routed": True,
            "mode": "unsupported",
            "invoked_plan": True,
            "invoked_apply": False,
            "outcome": "malformed_plan",
            "notice": notice,
        }
    if scenario.get("effect_approved") is not True:
        return {
            "routed": True,
            "mode": "full",
            "invoked_plan": True,
            "invoked_apply": False,
            "outcome": "declined",
            "notice": notice,
        }

    result = scenario.get("result")
    if not isinstance(result, dict):
        return {
            "routed": True,
            "mode": "unsupported",
            "invoked_plan": True,
            "invoked_apply": True,
            "outcome": "malformed_result",
            "notice": notice,
        }
    outcome = result.get("outcome")
    if outcome not in {"applied", "no_change", "failed"}:
        return {
            "routed": True,
            "mode": "unsupported",
            "invoked_plan": True,
            "invoked_apply": True,
            "outcome": "malformed_result",
            "notice": notice,
        }
    evaluated = {
        "routed": True,
        "mode": "full",
        "invoked_plan": True,
        "invoked_apply": True,
        "outcome": outcome,
        "notice": notice,
        "recovery": list(result.get("recovery", [])),
        "evidence": list(result.get("evidence", [])),
    }
    if "recoverable" in result:
        evaluated["recoverable"] = result["recoverable"]
    if "conflicts" in result:
        evaluated["conflicts"] = list(result["conflicts"])
    return evaluated
