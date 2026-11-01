import { useState } from "react";
import { api, type ProvisionResult } from "../../../../lib/api";
import { Badge, Button, Card, Field, Select } from "../../../../lib/design";
import { useAuth } from "../../identity";

const TIERS = [
  { id: "starter", label: "Starter · 2-4 vCPU · 8 GiB" },
  { id: "builder", label: "Builder · 8 vCPU · 16 GiB" },
  { id: "premium-pilot", label: "Premium Pilot · 16 vCPU · 32 GiB" },
  { id: "power", label: "Power · 24-32 vCPU · 64 GiB" },
];

export function OrderDevstationView() {
  const { token } = useAuth();
  const [tier, setTier] = useState("builder");
  const [repo, setRepo] = useState("");
  const [region, setRegion] = useState("us-wi");
  const [result, setResult] = useState<ProvisionResult | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function submit() {
    setBusy(true);
    setErr(null);
    try {
      setResult(await api.provision(token, { tier, repo: repo || undefined, region }));
    } catch (e) {
      setErr((e as Error).message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div>
      <h1 className="page-title">Order a Devstation</h1>
      <p className="page-sub">
        Submit a provision request. Warden records the order, plans the action, and a requested
        workspace appears in the fleet.
      </p>

      <div style={{ maxWidth: "32rem", display: "grid", gap: "1rem" }}>
        <Card title="New workspace">
          <div style={{ display: "grid", gap: "1rem" }}>
            <Select label="Tier" selectedKey={tier} onSelect={setTier} options={TIERS} />
            <Field
              label="Repository (optional)"
              value={repo}
              onChange={setRepo}
              placeholder="github.com/acme/site"
            />
            <Field label="Region" value={region} onChange={setRegion} placeholder="us-wi" />
            <div className="row">
              <Button onClick={submit} isDisabled={busy}>
                {busy ? "Submitting…" : "Order devstation"}
              </Button>
            </div>
          </div>
        </Card>

        {err && (
          <Card title="Error">
            <span className="muted" style={{ color: "var(--wc-danger)" }}>
              {err}
            </span>
          </Card>
        )}

        {result && (
          <Card title="Provision requested" action={<Badge state={result.resource.state} />}>
            <dl className="kv">
              <dt>action request</dt>
              <dd className="mono">{result.action_request_id}</dd>
              <dt>resource id</dt>
              <dd className="mono">{result.resource.id}</dd>
              <dt>stable id</dt>
              <dd className="mono">{result.resource.stable_id}</dd>
              <dt>tier</dt>
              <dd>{result.resource.tier}</dd>
            </dl>
            <p className="muted" style={{ marginTop: "0.75rem", fontSize: "var(--wc-fs-xs)" }}>
              Recorded in <code>shippin_core.action_requests</code> +{" "}
              <code>shippin_infra.resources</code>, audited in <code>shippin_audit.events</code>.
            </p>
          </Card>
        )}
      </div>
    </div>
  );
}
