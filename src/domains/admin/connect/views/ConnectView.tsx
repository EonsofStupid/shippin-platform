// Admin · Connect & Launch — the devstation click-to-auth surface. See per-tool
// auth state and supply the one live credential the devstation needs. Dumb
// view: status + activations go through connectService (Go drives, writes
// secrets only to root-only paths — never echoed back).
import { useEffect, useState, type FormEvent } from "react";
import { Badge, Button, Card } from "../../../../lib/design";
import { connectService } from "../connect.svc";
import type { ToolStatus } from "../types";

export function ConnectView() {
  const [tools, setTools] = useState<ToolStatus[]>([]);
  const [err, setErr] = useState<string | null>(null);
  const [msg, setMsg] = useState<string | null>(null);
  const [infisicalSecret, setInfisicalSecret] = useState("");
  const [githubToken, setGithubToken] = useState("");
  const [busy, setBusy] = useState(false);

  const load = () =>
    connectService.status().then((r) => setTools(r.items)).catch((e) => setErr((e as Error).message));

  useEffect(() => {
    void load();
  }, []);

  const activate = async (e: FormEvent, kind: "infisical" | "github") => {
    e.preventDefault();
    setBusy(true);
    setErr(null);
    setMsg(null);
    try {
      const res =
        kind === "infisical"
          ? await connectService.activateInfisical(infisicalSecret)
          : await connectService.activateGitHub(githubToken);
      if (res.ok) {
        setMsg(res.output ?? "connected");
        if (kind === "infisical") setInfisicalSecret("");
        else setGithubToken("");
      } else {
        setErr(res.error ?? res.output ?? "activation failed");
      }
      await load();
    } catch (e) {
      setErr((e as Error).message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <div>
      <h1 className="page-title">Connect &amp; Launch</h1>
      <p className="page-sub">Authenticate the devstation. Supply one live credential — the broker materializes the rest.</p>
      {err && <Card title="Error"><span className="muted">{err}</span></Card>}
      {msg && <Card title="OK"><pre className="mono">{msg}</pre></Card>}

      <div className="grid-cards">
        {tools.map((t) => (
          <Card key={t.tool} title={t.tool} action={<Badge state={t.connected ? "running" : "requested"} />}>
            <dl className="kv">
              <dt>state</dt><dd>{t.connected ? "connected" : "not connected"}</dd>
              <dt>detail</dt><dd className="mono">{t.detail}</dd>
            </dl>
          </Card>
        ))}
      </div>

      <Card title="Connect Infisical (machine identity / PAT)">
        <p className="muted">Paste a LIVE Universal-Auth client secret. Written root-only to /etc/warden; the broker authenticates and materializes /run/shippin-secrets.</p>
        <form className="edge-add" onSubmit={(e) => activate(e, "infisical")}>
          <input aria-label="infisical client secret" type="password" placeholder="client secret" value={infisicalSecret} onChange={(e) => setInfisicalSecret(e.target.value)} autoComplete="off" />
          <Button type="submit" tone="brand" variant="solid" isDisabled={busy || !infisicalSecret}>
            {busy ? "Connecting…" : "Connect Infisical"}
          </Button>
        </form>
      </Card>

      <Card title="Connect GitHub (PAT)">
        <p className="muted">Paste a GitHub PAT. Stored in the git credential store (0600) and verified against origin.</p>
        <form className="edge-add" onSubmit={(e) => activate(e, "github")}>
          <input aria-label="github token" type="password" placeholder="github_pat_… or ghp_…" value={githubToken} onChange={(e) => setGithubToken(e.target.value)} autoComplete="off" />
          <Button type="submit" tone="brand" variant="solid" isDisabled={busy || !githubToken}>
            {busy ? "Connecting…" : "Connect GitHub"}
          </Button>
        </form>
      </Card>
    </div>
  );
}
