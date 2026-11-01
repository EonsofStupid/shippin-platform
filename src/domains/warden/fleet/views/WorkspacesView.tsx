import { useEffect, useState } from "react";
import { api, type Workspace } from "../../../../lib/api";
import { Badge, Card } from "../../../../lib/design";
import { useAuth } from "../../identity";

export function WorkspacesView() {
  const { token } = useAuth();
  const [items, setItems] = useState<Workspace[]>([]);
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) return;
    let active = true;
    setLoading(true);
    api
      .workspaces(token)
      .then((result) => {
        if (active) {
          setItems(result.items);
          setErr(null);
        }
      })
      .catch((error: unknown) => {
        if (active) setErr((error as Error).message);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [token]);

  return (
    <div>
      <h1 className="page-title">Workspaces</h1>
      <p className="page-sub">
        Fleet resources from <code>shippin_infra.resources</code> — capsules &amp; devstations
        Warden provisions on Proxmox.
      </p>
      {err && (
        <Card title="Error">
          <span className="muted">{err}</span>
        </Card>
      )}
      {loading && <p className="muted">Loading…</p>}
      <div className="grid-cards">
        {items.map((w) => (
          <Card key={w.id} title={w.hostname ?? w.stable_id} action={<Badge state={w.state} />}>
            <dl className="kv">
              <dt>stable id</dt>
              <dd className="mono">{w.stable_id}</dd>
              <dt>kind</dt>
              <dd>{w.kind}</dd>
              <dt>tier</dt>
              <dd>{w.tier ?? "—"}</dd>
              <dt>resources</dt>
              <dd>
                {w.cores
                  ? `${w.cores} vCPU · ${(w.memory_mb ?? 0) / 1024} GiB · ${w.disk_gb} GiB`
                  : "—"}
              </dd>
              <dt>address</dt>
              <dd className="mono">{w.ip_cidr ?? "—"}</dd>
              <dt>ssh</dt>
              <dd className="mono">{w.ssh_alias ?? "—"}</dd>
            </dl>
          </Card>
        ))}
      </div>
      {!loading && items.length === 0 && (
        <p className="muted">No workspaces yet. Order one from the Order tab.</p>
      )}
    </div>
  );
}
