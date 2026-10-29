#!/usr/bin/env python3
"""Boundary guard: snapshot the file tree, diff on change, and check a proposed
path against existing boundaries before a new one is created.

Dependency-free. A "boundary" is a directory that defines a domain: top-level
directories, modules/module-* directories, and any directory that carries its
own AGENTS.md. The manifest is generated and disposable — regenerate, never
hand-edit it.

Exit codes for --check: 0 = looks new and safe, 2 = likely duplicate (review).
"""

from __future__ import annotations

import argparse
import difflib
import json
import re
import sys
from pathlib import Path

MANIFEST = Path("docs/ai/parking-lot/filetree.manifest.json")
SKIP = {
    ".git", ".idea", ".vscode", "node_modules", "target", "dist", "build",
    ".next", ".turbo", "__pycache__", ".venv", "venv", ".warden-code-server",
}
SIMILARITY_THRESHOLD = 0.72


def find_repo_root(start: Path) -> Path:
    cur = start.resolve()
    for parent in [cur, *cur.parents]:
        if (parent / "AGENTS.md").exists() and (parent / "docs").is_dir():
            return parent
    return Path.cwd()


def read_workspace_id(agents: Path) -> str | None:
    try:
        text = agents.read_text(encoding="utf-8", errors="replace")
    except OSError:
        return None
    m = re.search(r"(?m)^\s*workspace_id:\s*(.+?)\s*$", text)
    if not m:
        return None
    v = m.group(1).strip().strip('"')
    return None if v in {"null", "None", ""} else v


def collect_boundaries(root: Path) -> dict:
    boundaries: dict[str, dict] = {}
    for path in sorted(root.rglob("*")):
        if not path.is_dir():
            continue
        rel_parts = path.relative_to(root).parts
        if any(p in SKIP for p in rel_parts):
            continue
        rel = path.relative_to(root).as_posix()
        is_top = len(rel_parts) == 1
        is_module = bool(re.match(r"module-\d+-[a-z0-9-]+", path.name))
        agents = path / "AGENTS.md"
        has_agents = agents.exists()
        if not (is_top or is_module or has_agents):
            continue
        boundaries[rel] = {
            "name": path.name,
            "depth": len(rel_parts),
            "is_module": is_module,
            "has_agents": has_agents,
            "workspace_id": read_workspace_id(agents) if has_agents else None,
        }
    return boundaries


def collect_all_dirs(root: Path) -> dict:
    """Every directory in the tree (SKIP-filtered) keyed by repo-relative path.
    Used by --check so duplicate detection catches non-boundary folders too."""
    dirs: dict[str, dict] = {}
    for path in sorted(root.rglob("*")):
        if not path.is_dir():
            continue
        rel_parts = path.relative_to(root).parts
        if any(p in SKIP for p in rel_parts):
            continue
        dirs[path.relative_to(root).as_posix()] = {"name": path.name}
    return dirs


def normalize(name: str) -> str:
    return re.sub(r"[^a-z0-9]+", "", name.lower())


def best_matches(proposed: str, boundaries: dict, k: int = 5) -> list[tuple[float, str]]:
    target = normalize(Path(proposed).name)
    scored = []
    for rel, info in boundaries.items():
        ratio = difflib.SequenceMatcher(None, target, normalize(info["name"])).ratio()
        # bonus if the normalized proposed name is a substring of an existing one
        if target and (target in normalize(info["name"]) or normalize(info["name"]) in target):
            ratio = max(ratio, 0.85)
        scored.append((round(ratio, 3), rel))
    scored.sort(reverse=True)
    return scored[:k]


def cmd_snapshot(root: Path) -> int:
    boundaries = collect_boundaries(root)
    out = root / MANIFEST
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps({"boundaries": boundaries}, indent=2, sort_keys=True), encoding="utf-8")
    print(f"snapshot: {len(boundaries)} boundaries -> {MANIFEST.as_posix()}")
    return 0


def load_manifest(root: Path) -> dict:
    path = root / MANIFEST
    if not path.exists():
        return {}
    try:
        return json.loads(path.read_text(encoding="utf-8")).get("boundaries", {})
    except (OSError, json.JSONDecodeError):
        return {}


def cmd_diff(root: Path) -> int:
    old = load_manifest(root)
    new = collect_boundaries(root)
    added = sorted(set(new) - set(old))
    removed = sorted(set(old) - set(new))
    if not added and not removed:
        print("diff: no boundary changes since last snapshot")
        return 0
    for rel in added:
        print(f"  + {rel}")
    for rel in removed:
        print(f"  - {rel}")
    print(f"diff: {len(added)} added, {len(removed)} removed "
          f"(run --snapshot to accept)")
    return 0


def cmd_check(root: Path, proposed: str) -> int:
    # Compare against every directory, not just domain boundaries, so a near-name
    # duplicate (e.g. parkinglot vs parking-lot) is caught before creation.
    universe = collect_all_dirs(root)
    if proposed in universe or (root / proposed).exists():
        print(f"check: '{proposed}' already exists — reuse it, do not recreate.")
        return 2
    matches = best_matches(proposed, universe)
    top = matches[0] if matches else (0.0, None)
    print(f"check: proposed '{proposed}'")
    print("  nearest existing boundaries:")
    for score, rel in matches:
        print(f"    {score:>5}  {rel}")
    if top[0] >= SIMILARITY_THRESHOLD:
        print(f"  LIKELY DUPLICATE of '{top[1]}' (score {top[0]}). "
              "Reuse or extend it instead of creating a new boundary.")
        return 2
    print("  looks new and safe. Establishing a new domain still needs operator "
          "approval (see boundary-guard skill).")
    return 0


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    g = ap.add_mutually_exclusive_group(required=True)
    g.add_argument("--snapshot", action="store_true")
    g.add_argument("--diff", action="store_true")
    g.add_argument("--check", metavar="PATH")
    ap.add_argument("--root", default=None)
    args = ap.parse_args()

    root = Path(args.root).resolve() if args.root else find_repo_root(Path.cwd())
    if args.snapshot:
        return cmd_snapshot(root)
    if args.diff:
        return cmd_diff(root)
    return cmd_check(root, args.check)


if __name__ == "__main__":
    sys.exit(main())
