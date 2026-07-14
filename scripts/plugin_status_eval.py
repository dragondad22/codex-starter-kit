"""Deterministic policy oracle for the skills-only status tracer evaluations.

This is development and CI evaluation code, not a plugin runtime dependency. The skill
uses the same fail-closed decision table while the standalone engine remains authoritative.
"""

ALLOWED_LIFECYCLES = {
    "managed",
    "managed_degraded",
    "setup_incomplete",
    "unmanaged",
}


def _compatible(capability):
    return (
        capability.get("schema_version") == 1
        and capability.get("engine", {}).get("name") == "starter-kit"
        and capability.get("protocol")
        == {"name": "starter-kit.lifecycle", "version": 1}
        and "status" in capability.get("operations", [])
        and 1 in capability.get("status_schema_versions", [])
    )


def _valid_status(status):
    if not isinstance(status, dict):
        return False
    if status.get("schema_version") != 1:
        return False
    if not isinstance(status.get("repository"), str) or not status["repository"]:
        return False
    if status.get("lifecycle") not in ALLOWED_LIFECYCLES:
        return False
    return all(
        isinstance(status.get(field), list)
        and all(isinstance(item, str) for item in status[field])
        for field in ("problems", "recovery", "evidence")
    )


def evaluate_scenario(scenario):
    """Return the bounded behavior a conforming status skill may expose."""
    if scenario.get("route") is False:
        return {"routed": False, "mode": "not-routed", "invoked_status": False}

    capability = scenario.get("capability")
    if not isinstance(capability, dict):
        return {"routed": True, "mode": "unsupported", "invoked_status": False}

    state = capability.get("state")
    if state in {"missing", "administratively-unavailable"}:
        return {
            "routed": True,
            "mode": "degraded-guidance",
            "invoked_status": False,
        }
    if state != "available":
        return {"routed": True, "mode": "unsupported", "invoked_status": False}
    if not _compatible(capability):
        return {
            "routed": True,
            "mode": "degraded-guidance",
            "invoked_status": False,
        }
    if capability.get("engine", {}).get("provenance") != "verified":
        return {
            "routed": True,
            "mode": "degraded-guidance",
            "invoked_status": False,
        }
    if capability.get("status_authorized") is not True:
        return {"routed": True, "mode": "unsupported", "invoked_status": False}

    status = scenario.get("status")
    if not _valid_status(status):
        return {"routed": True, "mode": "unsupported", "invoked_status": True}

    mode = "full" if capability.get("mutation_authorized") is True else "verification-only"
    return {
        "routed": True,
        "mode": mode,
        "invoked_status": True,
        "repository": status["repository"],
        "lifecycle": status["lifecycle"],
        "problems": list(status["problems"]),
        "recovery": list(status["recovery"]),
        "evidence": list(status["evidence"]),
    }
