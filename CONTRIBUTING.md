# Contributing to Shippin

Use Node.js 24 or newer and the repository's Go workspace. From a clean checkout,
install the web dependencies and opt in to the Git hook:

```sh
npm ci --ignore-scripts
npm run hooks:install
```

Lefthook's npm package normally installs hooks in its postinstall script;
`--ignore-scripts` keeps that step explicit. Review existing hooks before
installing it, particularly when working with multiple Git worktrees.

The hook checks only the reviewed Proxmox client files listed in `biome.json`.
It does not modify or automatically stage files. Hooks can be bypassed, so CI
runs the same check independently. Expand Biome coverage as other contexts are
reviewed and formatted; the current configuration is not a repository-wide audit.

```sh
npm run check:quality
npm run build
npm run typecheck
npm test
```

Run the build before typechecking after a clean checkout: TanStack generates the
route types during the build and those generated files are not committed.

The Proxmox feature's Zod schemas validate the existing API responses and own
the frontend types derived from them. They do not replace a provider-neutral API
contract or validate the Go server's authorization. A nil Go inventory is treated
as empty only when the reported count is zero. Keep those checks aligned with the
server when migrating the API. Never put provider credentials in browser code.

The application is not yet a verified server installation. A passing build or
mocked HTTP test does not establish secure deployment, node enrollment, real
provider operation, upgrade or recovery. Preserve local work and review existing
services and configuration before migrating or deploying them.
