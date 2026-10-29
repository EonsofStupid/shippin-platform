#!/usr/bin/env python3
"""Claude Code UserPromptSubmit hook: force WardenClyffe parking-lot mode.

Reads the hook JSON on stdin. If the prompt looks like planning or
decision-making, injects context that forces the agent into Parking-Lot mode
(per docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md) even when the operator did not
type /parkinglot. Always exits 0 and never blocks the prompt.
"""

from __future__ import annotations

import json
import re
import sys

TRIGGERS = re.compile(
    r"\b("
    r"plan(ning|s)?|scop(e|ing)|roadmap|sprint|milestone|poa\&?m|backlog|"
    r"lock(ed)?\s*in|lock\s*it|decid(e|ing|ed)|decision|pick(ing)?|choos(e|ing)|"
    r"settle\s+on|go\s+with|commit\s+to|trade[\s-]?off|"
    r"depend(ency|encies|s|\s+on)|feature|capabilit(y|ies)|"
    r"librar(y|ies)|framework|protocol|schema|boundary|boundaries|architect(ure|ing)"
    r")\b",
    re.IGNORECASE,
)

CONTEXT = (
    "[FORCED: WardenClyffe Parking-Lot mode] This turn involves planning or "
    "decision-making. Enter Parking-Lot mode now (do not wait for /parkinglot): "
    "follow docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md. Announce the active focus "
    "feature (default: Building WardenClyffe). When a dependency, feature, or "
    "capability is locked in, capture it with "
    "scripts/foundation/parking-lot-capture.py so it projects into the SurrealDB "
    "intelligence layer. Before creating any new folder/boundary, use the "
    "boundary-guard skill first."
)


def main() -> int:
    try:
        data = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0
    prompt = str(data.get("prompt", ""))
    if not TRIGGERS.search(prompt):
        return 0
    print(json.dumps({
        "hookSpecificOutput": {
            "hookEventName": "UserPromptSubmit",
            "additionalContext": CONTEXT,
        }
    }))
    return 0


if __name__ == "__main__":
    sys.exit(main())
