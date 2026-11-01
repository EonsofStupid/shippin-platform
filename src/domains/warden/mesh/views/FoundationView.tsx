import { useEffect, useState } from "react";
import { api, type ClyffyHome, type FoundationService } from "../../../../lib/api";
import { Card, Badge } from "../../../../lib/design";
import { useAuth } from "../../identity";

const ROLE_LABEL: Record<string, string> = {
  sql: "Postgres", vector: "Qdrant", graph: "SurrealDB", identity: "Authentik",
  dns: "PowerDNS", edge: "Caddy edge", ca: "step-ca", embedder: "Harrier",
  "llm-gateway": "Bifrost", observability: "Observatory", operator: "Capsule",
  network: "OPNsense", devstation: "Devstation", portal: "Portal", warden: "Warden",
};

function reachBadge(state: string) {
  if (state === "reachable" || state === "running") return "running";
  if (state === "planned") return "requested";
  if (state === "down" || state === "legacy") return "error";
  return "stopped";
}

const DESIGNATION_TONE = {
  plugin: "info", connector: "brand", platform: "neutral", core: "success",
} as const;

export function FoundationView() {
  const { token } = useAuth();
  const [home, setHome] = useState<ClyffyHome | null>(null);
  const [services, setServices] = useState<FoundationService[]>([]);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    setHome(null);
    setServices([]);
    setErr(null);
    if (token) {
      Promise.all([api.clyffyHome(token), api.clyffyServices(token)])
        .then(([h, s]) => {
          if (active) {
            setHome(h);
            setServices(s.items);
          }
        })
        .catch((e) => {
          if (active) setErr((e as Error).message);
        });
    }
    return () => {
      active = false;
    };
  }, [token]);

  return (
    <div>
      <h1 className="page-title">Foundation · Clyffy orchestrator</h1>
      <p className="page-sub">
        Warden's captured model of the managed platforms &amp; services. Clyffy runs on the Warden server and reads <em>this</em> — not raw IPs.
      </p>
      {err && <Card title="Error"><span className="muted" style={{ color: "var(--wc-danger)" }}>{err}</span></Card>}

      {home && (
        <div className="grid-cards" style={{ marginBottom: "1rem" }}>
          <Card title="Platforms">
            <div style={{ display: "grid", gap: "0.5rem" }}>
              {home.platforms.map((p) => (
                <div key={p.stable_id} className="spread">
                  <div><strong>{p.kind}</strong> <span className="muted">{p.network_name ? `· net ${p.network_name}` : ""}</span></div>
                  <Badge state={reachBadge(p.state)} />
                </div>
              ))}
            </div>
            <p className="muted" style={{ fontSize: "var(--wc-fs-xs)", marginTop: "0.5rem" }}>
              The Coolify platform is the headless Docker service — now captured, not hand-waved.
            </p>
          </Card>
          <Card title="Intelligence plane (live probe)">
            <div style={{ display: "grid", gap: "0.4rem" }}>
              {home.intelligence.map((p) => (
                <div key={p.stable_id} className="spread">
                  <div className="row" style={{ gap: ".4rem" }}>{ROLE_LABEL[p.role] ?? p.role}
                    {p.credential_gated && <Badge tone="warning" dot={false}>credential-gated</Badge>}
                  </div>
                  <Badge state={p.reachable ? "running" : "error"} />
                </div>
              ))}
            </div>
          </Card>
          <Card title="AI Plugins · ai-only" action={<Badge tone="info" dot={false}>plugin</Badge>}>
            <div style={{ display: "grid", gap: "0.4rem" }}>
              {services.filter((s) => s.designation === "plugin").map((s) => (
                <div key={s.stable_id} className="spread">
                  <div className="row" style={{ gap: ".4rem" }}>{ROLE_LABEL[s.role] ?? s.role}
                    <span className="muted" style={{ fontSize: "var(--wc-fs-xs)" }}>{s.name}</span>
                  </div>
                  <Badge state={reachBadge(s.state)} />
                </div>
              ))}
            </div>
            <p className="muted" style={{ fontSize: "var(--wc-fs-xs)", marginTop: "0.5rem" }}>
              SurrealDB · Qdrant · Harrier · <strong>RRD</strong> (Reason Ready Daemon, reranker). Operator sees status only.
            </p>
          </Card>
          <Card title="Summary">
            <dl className="kv">
              <dt>platforms</dt><dd>{home.summary.platforms}</dd>
              <dt>services</dt><dd>{home.summary.services}</dd>
              <dt>plugins (ai-only)</dt><dd>{home.summary.services_by_designation?.plugin ?? 0}</dd>
              <dt>connectors (shared)</dt><dd>{home.summary.services_by_designation?.connector ?? 0}</dd>
              <dt>credential-gated</dt><dd>{home.summary.credential_gated_services}</dd>
              <dt>config owed</dt><dd>{home.summary.config_items_owed} items</dd>
            </dl>
          </Card>
        </div>
      )}

      <Card title={`Captured services (${services.length})`}>
        <div className="ui-table__wrap">
          <table className="ui-table">
            <thead><tr><th>vmid</th><th>service</th><th>role</th><th>designation</th><th>audiences</th><th>endpoint</th><th>auth</th><th>credential ref</th><th>state</th></tr></thead>
            <tbody>
              {services.map((s) => (
                <tr key={s.stable_id}>
                  <td>{s.vmid ?? "—"}</td>
                  <td>{s.name}</td>
                  <td>{ROLE_LABEL[s.role] ?? s.role}</td>
                  <td><Badge tone={DESIGNATION_TONE[s.designation]} dot={false}>{s.designation}</Badge></td>
                  <td>{s.audiences.join(" + ")}</td>
                  <td>{s.endpoint ?? "—"}</td>
                  <td>{s.auth_required ? "🔒" : "open"}</td>
                  <td>{s.credential_ref ?? "—"}</td>
                  <td><Badge state={reachBadge(s.state)} /></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}
