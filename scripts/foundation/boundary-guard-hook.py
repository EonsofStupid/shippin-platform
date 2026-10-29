#!/usr/bin/env python3
"""Claude Code PreToolUse hook: force a boundary review before a NEW boundary is
created. Wired to Bash (catches `mkdir`) and Write (new files in an unrecognized
top-level boundary).

Reads the hook JSON on stdin. Advisory only: it injects a duplicate-candidate
report into context (additionalContext) so the agent reviews before proceeding;
it does not deny the call. Always exits 0.
"""

from __future__ import annotations

import json
import re
import subprocess
import sys
from pathlib import Path

TREE = "scripts/foundation/boundary-guard-tree.py"


def find_repo_root(start: Path) -> Path:
    cur = start.resolve()
    for parent in [cur, *cur.parents]:
        if (parent / "AGENTS.md").exists() and (parent / "docs").is_dir():
            return parent
    return Path.cwd()


def candidate_paths(tool: str, ti: dict) -> list[str]:
    paths: list[str] = []
    if tool == "Bash":
        cmd = str(ti.get("command", ""))
        for m in re.finditer(r"mkdir\b([^\n;&|]*)", cmd):
            for tok in m.group(1).split():
                if tok.startswith("-"):
                    continue
                paths.append(tok.strip("'\""))
    elif tool == "Write":
        fp = str(ti.get("file_path", ""))
        if fp:
            paths.append(str(Path(fp).parent))
    return paths


def check(root: Path, path: str) -> tuple[int, str]:
    rel = path
    try:
        rel = str(Path(path).resolve().relative_to(root))
    except (ValueError, OSError):
        pass
    try:
        proc = subprocess.run(
            [sys.executable, str(root / TREE), "--check", rel, "--root", str(root)],
            capture_output=True, text=True, timeout=20,
        )
        return proc.returncode, proc.stdout.strip()
    except (OSError, subprocess.SubprocessError) as exc:
        return 0, f"(boundary-guard check skipped: {exc})"


def main() -> int:
    try:
        data = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0
    tool = data.get("tool_name", "")
    ti = data.get("tool_input", {}) or {}
    cwd = Path(data.get("cwd", ".")) if data.get("cwd") else Path.cwd()
    root = find_repo_root(cwd)

    reports = []
    for p in candidate_paths(tool, ti):
        code, out = check(root, p)
        if code == 2:
            reports.append(out)
    if not reports:
        return 0

    msg = (
        "[FORCED: WardenClyffe Boundary-Guard] A new folder/boundary is about to "
        "be created, but the duplicate check flagged a likely existing match. "
        "Review before proceeding (reuse the existing boundary if it fits); "
        "creating a genuinely new domain needs operator approval. See "
        "docs/ai/WARDENCLYFFE_BOUNDARY_GUARD_SKILL.md.\n\n" + "\n\n".join(reports)
    )
    print(json.dumps({
        "hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "additionalContext": msg,
        }
    }))
    return 0


if __name__ == "__main__":
    sys.exit(main())
