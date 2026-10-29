// Slice 0/1 — Proxmox substrate: fleet inventory + backup tab.
import { useCallback, useEffect, useState } from "react";
import { Badge, Button, Card, Tabs } from "../../../../lib/design";
import { proxmoxService } from "../proxmox.svc";
import type { GuestActionResult, ProxmoxGuest, ProxmoxStatus } from "../types";
import { BackupPanel } from "./BackupPanel";

function fmtMem(bytes?: number) {
  if (!bytes) return "—";
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GiB`;
}

export function ProxmoxFleetView() {
  const [status, setStatus] = useState<ProxmoxStatus | null>(null);
  const [guests, setGuests] = useState<ProxmoxGuest[]>([]);
  const [err, setErr] = useState<string | null>(null);
  const [lastAction, setLastAction] = useState<GuestActionResult | null>(null);
  const [busyKey, setBusyKey] = useState<string | null>(null);

  const load = useCallback(async () => {
    setErr(null);
    try {
      const st = await proxmoxService.status();
      setStatus(st);
      if (st.configured && st.reachable) {
        const g = await proxmoxService.guests();
        setGuests(g.items.filter((x) => !x.template));
      } else {
        setGuests([]);
      }
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e));
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const act = async (g: ProxmoxGuest, op: "start" | "stop") => {
    const key = `${g.stable_label}:${op}`;
    setBusyKey(key);
    setErr(null);
    try {
      const res =
        op === "start"
          ? await proxmoxService.start(g.node, g.kind, g.vmid)
          : await proxmoxService.stop(g.node, g.kind, g.vmid);
      setLastAction(res);
      await load();
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e));
    } finally {
      setBusyKey(null);
    }
  };

  return (
    <div>
      <h1 className="page-title">Proxmox fleet</h1>
      <p className="page-sub">
        Proxmox is the substrate (OS for VMs/LXCs). WardenClyffe is the internal name for these
        layers — inventory, task-true start/stop, and off-box backup.
      </p>
      <Tabs
        items={[
          { id: "fleet", label: "Fleet", content: <FleetBody status={status} guests={guests} err={err} lastAction={lastAction} busyKey={busyKey} load={load} act={act} /> },
          { id: "backup", label: "Backup", content: <BackupPanel /> },
        ]}
      />
    </div>
  );
}

function FleetBody({
  status,
  guests,
  err,
  lastAction,
  busyKey,
  load,
  act,
}: {
  status: ProxmoxStatus | null;
  guests: ProxmoxGuest[];
  err: string | null;
  lastAction: GuestActionResult | null;
  busyKey: string | null;
  load: () => Promise<void>;
  act: (g: ProxmoxGuest, op: "start" | "stop") => Promise<void>;
}) {
  return (
    <div>

      {status && (
        <Card
          title="Substrate status"
          action={
            <Badge state={status.reachable ? "running" : status.configured ? "error" : "requested"} />
          }
        >
          <dl className="kv">
            <dt>message</dt>
            <dd>{status.message}</dd>
            <dt>host</dt>
            <dd className="mono">
              {status.host}:{status.port}
            </dd>
            <dt>node</dt>
            <dd className="mono">{status.node}</dd>
            <dt>configured</dt>
            <dd>{status.configured ? "yes" : "no"}</dd>
            <dt>reachable</dt>
            <dd>{status.reachable ? "yes" : "no"}</dd>
            {status.error && (
              <>
                <dt>error</dt>
                <dd className="mono">{status.error}</dd>
              </>
            )}
            {status.version && (
              <>
                <dt>pve version</dt>
                <dd className="mono">{String(status.version.version ?? JSON.stringify(status.version))}</dd>
              </>
            )}
          </dl>
          {!status.configured && (
            <p className="muted" style={{ marginTop: "0.75rem", fontSize: "var(--wc-fs-sm)" }}>
              Create <code>secrets/proxmox.env</code> from <code>secrets/proxmox.env.example</code> with
              a Proxmox API token, restart warden-api, then refresh.
            </p>
          )}
        </Card>
      )}

      {err && (
        <Card title="Error">
          <span className="muted" style={{ color: "var(--wc-danger, #f88)" }}>
            {err}
          </span>
        </Card>
      )}

      {lastAction && (
        <Card title="Last action" action={<Badge state={lastAction.status === "succeeded" ? "running" : "error"} />}>
          <dl className="kv">
            <dt>kind</dt>
            <dd>{lastAction.kind}</dd>
            <dt>guest</dt>
            <dd className="mono">
              {lastAction.node}/{lastAction.guest_kind}/{lastAction.vmid}
            </dd>
            <dt>action_request</dt>
            <dd className="mono">{lastAction.action_request_id}</dd>
            <dt>upid</dt>
            <dd className="mono">{lastAction.upid ?? "—"}</dd>
            <dt>status</dt>
            <dd>{lastAction.status}</dd>
            {lastAction.error && (
              <>
                <dt>error</dt>
                <dd>{lastAction.error}</dd>
              </>
            )}
          </dl>
        </Card>
      )}

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", margin: "1.25rem 0 0.75rem" }}>
        <h2 style={{ margin: 0, fontSize: "var(--wc-fs-lg)" }}>Guests ({guests.length})</h2>
        <Button size="sm" tone="neutral" variant="outline" onPress={() => void load()}>
          Refresh
        </Button>
      </div>

      <div className="grid-cards">
        {guests.map((g) => {
          const running = g.status === "running";
          return (
            <Card
              key={g.stable_label}
              title={g.name || `VMID ${g.vmid}`}
              action={<Badge state={running ? "running" : "stopped"} />}
            >
              <dl className="kv">
                <dt>kind</dt>
                <dd>
                  {g.kind} · {g.vmid}
                </dd>
                <dt>node</dt>
                <dd className="mono">{g.node}</dd>
                <dt>status</dt>
                <dd>{g.status}</dd>
                <dt>resources</dt>
                <dd>
                  {g.cpus ?? "—"} cpu · {fmtMem(g.maxmem)}
                </dd>
                <dt>label</dt>
                <dd className="mono">{g.stable_label}</dd>
              </dl>
              <div className="row" style={{ marginTop: "0.75rem" }}>
                <Button
                  size="sm"
                  tone="brand"
                  variant="solid"
                  isDisabled={running || busyKey !== null}
                  onPress={() => void act(g, "start")}
                >
                  {busyKey === `${g.stable_label}:start` ? "Starting…" : "Start"}
                </Button>
                <Button
                  size="sm"
                  tone="neutral"
                  variant="outline"
                  isDisabled={!running || busyKey !== null}
                  onPress={() => void act(g, "stop")}
                >
                  {busyKey === `${g.stable_label}:stop` ? "Stopping…" : "Stop"}
                </Button>
              </div>
            </Card>
          );
        })}
      </div>

      {status?.reachable && guests.length === 0 && (
        <p className="muted">No non-template guests returned from Proxmox.</p>
      )}
    </div>
  );
}
