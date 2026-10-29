# WardenClyffe plugins (W-rooted, template-shaped)

Turnkey plugin packages. **This repo dir is the source; the distribution root
is the W drive** (`W:/plugins` = `/workspace/warden-storage/plugins`, per the
W rule — same path from the local workstation and the devstation). Adding a
plugin to a tool = browse to W, open the plugin folder, copy its connect
snippet.

## Template (every plugin follows this shape — no exceptions)

```text
plugins/<plugin-slug>/
  plugin.json                  manifest: name, slug, kind, version, mcp_id,
                               entry{command,args}, requires{secrets: REFS ONLY, network}
  connect/claude-desktop.json  paste into Claude Desktop config
  connect/codex.toml           paste into ~/.codex/config.toml
  connect/claude-code.sh       one-liner: claude mcp add ...
  README.md                    what it serves + how to verify
```

Every `plugin.json` MUST validate against the master contract
`schemas/contracts/plugin.v1.schema.json`. `cortex-control` is the master
control-layer plugin; Clyffy-Dean minions **derive** from it and satisfy the
same contract. (Draft 2020-12; CI validates on change.)

Rules:
- `kind`: `control` | `intelligence` | `minion`.
- **Secrets never live in a package** — `requires.secrets` lists Infisical
  refs; values materialize at `/run/shippin-secrets` on the devstation.
- `mcp_id` follows the mesh grammar; status stays `scaffold` until the ADR
  0030 bar (Streamable HTTP, OAuth 2.1, server card) is met.
- One shared binary may serve many plugins via `entry.args` (cortex-mcp does).
- Updates ship as `.pulse` packets (kind `app`/`clyffy`/`minion`).

## Current plugins

| Slug | Kind | Serves |
|---|---|---|
| `cortex-control` | control | decision capture, validator, registry, naming/structure law, strict prompts |
| `cortex-intelligence` | intelligence | direct read-only Qdrant (LXC 106) + self-hosted SurrealDB (LXC 104) |
