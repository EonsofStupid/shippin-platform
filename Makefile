# WardenClyffe pre-release — methodical bootstrap
# Read docs/STATE.md first.

ROOT := $(abspath .)
export WARDEN_REPO_ROOT := $(ROOT)

.PHONY: help setup stack stop verify-slice0 sync-secrets build-api build-console green status

help:
	@echo "WardenClyffe pre-release targets"
	@echo "  make setup          tools, deps, secrets template"
	@echo "  make sync-secrets   pull from Infisical → secrets/proxmox.env"
	@echo "  make stack          warden-api + clyffe-api + console (foreground)"
	@echo "  make stop           stop stack pid files / ports"
	@echo "  make verify-slice0  health + proxmox status"
	@echo "  make build-api      compile Go APIs"
	@echo "  make build-console  npm run build"
	@echo "  make green          build-api + build-console"
	@echo "  make status         quick ports + health"
	@echo ""
	@echo "Slice 0 UI: /admin/proxmox after login (operator / warden-dev)"

setup:
	bash scripts/dev/setup-prerelease.sh

sync-secrets:
	bash scripts/dev/sync-secrets-from-infisical.sh

stack:
	bash scripts/dev/run-stack.sh

stop:
	bash scripts/dev/stop-stack.sh

verify-slice0:
	bash scripts/dev/verify-slice0.sh

build-api:
	cd services/shippin-api && go build -o /tmp/shippin-api ./cmd/shippin-api
	cd services/clyffe-api && go build -o /tmp/clyffe-api ./cmd/clyffe-api
	@echo "binaries: /tmp/shippin-api /tmp/clyffe-api"

build-console:
	npm run build

green: build-api build-console
	@echo "green: API + console build OK"

status:
	@ss -ltn 2>/dev/null | grep -E ':(5173|8081|8082|5432)\s' || true
	@curl -sS -m 2 http://127.0.0.1:8081/healthz 2>/dev/null || echo "warden-api: down"
	@echo
	@curl -sS -m 2 http://127.0.0.1:8081/api/warden/proxmox/status 2>/dev/null || echo "proxmox status: n/a"
	@echo
