#!/usr/bin/env python3
"""Capture a locked-in parking-lot decision as a v2 clyffy_touchpoint.

Dependency-free, mirroring scripts/foundation/validate-touchpoints.py. This does
NOT write to Qdrant or SurrealDB. It writes/updates the Markdown touchpoint that
the sync worker then projects, per docs/ai/TOUCHPOINT_SYNC_PATTERN.md.

Domain is identified from --boundary: the nearest AGENTS.md (its workspace_id)
or the enclosing modules/module-* directory, falling back to a default
workspace. A boundary with no recognizable domain is marked domain_status: new
so the boundary-guard skill can gate creation.
"""

from __future__ import annotations

import argparse
import datetime as _dt
import re
import sys
from pathlib import Path

DEFAULT_WORKSPACE = "wardenclyffe.intelligence"
DECISIONS_DIR = Path("docs/ai/parking-lot/decisions")
SOURCE_SKILL = "docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md"
MESH_REGISTRY = "wardenclyffe/registry/context-mesh.yaml"
VALID_TYPES = ("dependency", "feature", "capability")

START = "<!-- decisions:start -->"
END = "<!-- decisions:end -->"


def slug(text: str) -> str:
    return re.sub(r"[^a-z0-9]+", "-", text.lower()).strip("-") or "x"


def find_repo_root(start: Path) -> Path:
    cur = start.resolve()
    for parent in [cur, *cur.parents]:
        if (parent / "AGENTS.md").exists() and (parent / "docs").is_dir():
            return parent
    return Path.cwd()


def read_frontmatter_scalar(path: Path, name: str) -> str | None:
    try:
        text = path.read_text(encoding="utf-8", errors="replace")
    except OSError:
        return None
    if not text.startswith("---"):
        return None
    fm = text.split("---", 2)[1] if text.count("---") >= 2 else ""
    m = re.search(rf"(?m)^\s*{re.escape(name)}:\s*(.+?)\s*$", fm)
    if not m:
        return None
    value = m.group(1).strip().strip('"')
    return None if value in {"null", "None", ""} else value


def identify_domain(root: Path, boundary: str) -> tuple[str, str | None, str]:
    """Return (workspace_id, module, domain_status) for a boundary path."""
    rel = (root / boundary).resolve()
    module = None
    # A boundary is "existing" when its path is already in the tree; "new" means
    # we are proposing a path that does not exist yet (boundary-guard gates it).
    status = "existing" if rel.exists() else "new"
    # nearest enclosing module directory
    for parent in [rel, *rel.parents]:
        if root not in parent.parents and parent != root:
            continue
        m = re.match(r"module-\d+-[a-z0-9-]+", parent.name)
        if m:
            module = parent.name
            break
    # nearest AGENTS.md (with a workspace_id) walking up from the boundary,
    # bounded by repo root, to inherit the owning workspace.
    probe = rel if rel.is_dir() else rel.parent
    workspace = None
    while True:
        ws = read_frontmatter_scalar(probe / "AGENTS.md", "workspace_id")
        if ws:
            workspace = ws
            break
        if probe == root or root not in probe.parents:
            break
        probe = probe.parent
    return workspace or DEFAULT_WORKSPACE, module, status


def build_frontmatter(meta: dict) -> str:
    module = meta["module"] or "null"
    return (
        "---\n"
        "clyffy_touchpoint:\n"
        "  version: 2\n"
        f"  workspace_id: {meta['workspace_id']}\n"
        f"  project_key: {meta['project_key']}\n"
        "  persona: clyffy-operator\n"
        "  kind: runtime\n"
        f"  owner: {meta['owner']}\n"
        f"  module: {module}\n"
        f"  focus_feature: \"{meta['focus']}\"\n"
        f"  boundary: {meta['boundary']}\n"
        f"  domain_status: {meta['domain_status']}\n"
        f"  source_skill: {SOURCE_SKILL}\n"
        f"  mesh_registry: {MESH_REGISTRY}\n"
        "  sync:\n"
        "    qdrant: true\n"
        "    surreal: true\n"
        "---\n"
    )


def upsert_decision(body: str, key: str, line: str) -> str:
    block = f"<!-- decision:{key} -->\n{line}\n<!-- /decision:{key} -->"
    pat = re.compile(
        rf"<!-- decision:{re.escape(key)} -->.*?<!-- /decision:{re.escape(key)} -->",
        re.DOTALL,
    )
    if pat.search(body):
        return pat.sub(block, body)
    return body.rstrip() + "\n" + block + "\n"


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--type", required=True, choices=VALID_TYPES)
    ap.add_argument("--name", required=True)
    ap.add_argument("--decision", required=True)
    ap.add_argument("--boundary", required=True, help="repo-relative path the decision concerns")
    ap.add_argument("--focus", default="Building WardenClyffe")
    ap.add_argument("--root", default=None)
    args = ap.parse_args()

    root = Path(args.root).resolve() if args.root else find_repo_root(Path.cwd())
    workspace, module, status = identify_domain(root, args.boundary)
    project_key = slug(args.boundary.replace("/", "-")) if args.boundary not in (".", "") else slug(workspace)

    out_dir = root / DECISIONS_DIR
    out_dir.mkdir(parents=True, exist_ok=True)
    out_path = out_dir / f"{project_key}.md"
    owner = f"{DECISIONS_DIR.as_posix()}/{project_key}.md"

    stamp = _dt.date.today().isoformat()
    key = slug(f"{args.type}-{args.name}")
    line = f"- **[{args.type}] {args.name}** (locked {stamp}): {args.decision}"

    meta = {
        "workspace_id": workspace,
        "project_key": project_key,
        "owner": owner,
        "module": module,
        "focus": args.focus,
        "boundary": args.boundary,
        "domain_status": status,
    }

    if out_path.exists():
        text = out_path.read_text(encoding="utf-8")
        fm_end = text.index("---", 3) + 3
        body = text[fm_end:]
        if START in body and END in body:
            pre, mid = body.split(START, 1)
            inner, post = mid.split(END, 1)
            inner = upsert_decision(inner.strip("\n"), key, line)
            body = f"{pre}{START}\n{inner}\n{END}{post}"
        else:
            body = body.rstrip() + f"\n\n{START}\n" + upsert_decision("", key, line) + f"{END}\n"
    else:
        header = (
            f"\n# Parking-Lot Decisions — {project_key}\n\n"
            f"Focus: **{args.focus}**. Boundary: `{args.boundary}`. "
            "Domain status is carried in the frontmatter (`domain_status`).\n\n"
            "Generated by `scripts/foundation/parking-lot-capture.py`. Projected to "
            "Qdrant + SurrealDB by the sync worker (see "
            "`docs/ai/TOUCHPOINT_SYNC_PATTERN.md`). Keep entries terse.\n\n"
        )
        body = header + f"{START}\n" + upsert_decision("", key, line) + f"{END}\n"

    out_path.write_text(build_frontmatter(meta) + body, encoding="utf-8")

    print(f"captured {args.type} '{args.name}' -> {owner}")
    print(f"  workspace_id={workspace} project_key={project_key} domain={status}")
    if status == "new":
        print("  NOTE: new domain — run boundary-guard before creating its folder.")
    print("  next: python scripts/foundation/validate-touchpoints.py --json")
    return 0


if __name__ == "__main__":
    sys.exit(main())
