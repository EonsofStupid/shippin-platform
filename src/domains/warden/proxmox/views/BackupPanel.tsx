import { useCallback, useEffect, useState, type FormEvent } from "react";
import { Badge, Button, Card, Field, Select } from "../../../../lib/design";
import { proxmoxService } from "../proxmox.svc";
import type { BackupStatus, GuestActionResult, ProxmoxGuest } from "../types";
import "./backup.css";

const PROVIDERS = [
  { id: "r2", label: "Cloudflare R2" },
  { id: "minio", label: "MinIO (private cloud)" },
  { id: "b2", label: "Backblaze B2" },
  { id: "s3", label: "Amazon S3" },
];

export function BackupPanel() {
  const [status, setStatus] = useState<BackupStatus | null>(null);
  const [guests, setGuests] = useState<ProxmoxGuest[]>([]);
  const [err, setErr] = useState<string | null>(null);
  const [msg, setMsg] = useState<string | null>(null);
  const [busy, setBusy] = useState<string | null>(null);
  const [last, setLast] = useState<GuestActionResult | null>(null);

  const [provider, setProvider] = useState("r2");
  const [endpoint, setEndpoint] = useState("");
  const [region, setRegion] = useState("auto");
  const [bucket, setBucket] = useState("");
  const [accessKey, setAccessKey] = useState("");
  const [secretKey, setSecretKey] = useState("");
  const [resticPassword, setResticPassword] = useState("");
  const [prefix, setPrefix] = useState("shippin-backups");
  const [guestKey, setGuestKey] = useState("");
  const [storage, setStorage] = useState("default");

  const load = useCallback(async () => {
    setErr(null);
    try {
      const st = await proxmoxService.backupStatus();
      setStatus(st);
      if (st.s3.provider) setProvider(st.s3.provider);
      if (st.s3.endpoint) setEndpoint(st.s3.endpoint);
      if (st.s3.region) setRegion(st.s3.region);
      if (st.s3.bucket) setBucket(st.s3.bucket);
      if (st.s3.prefix) setPrefix(st.s3.prefix);
      const g = await proxmoxService.guests().catch(() => ({ items: [] as ProxmoxGuest[] }));
      const live = (g.items ?? []).filter((x) => !x.template);
      setGuests(live);
      setGuestKey((cur) => cur || live[0]?.stable_label || "");
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e));
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const save = async (e: FormEvent) => {
    e.preventDefault();
    setBusy("save");
    setErr(null);
    setMsg(null);
    try {
      await proxmoxService.saveS3({
        provider,
        endpoint,
        region,
        bucket,
        access_key: accessKey,
        secret_key: secretKey,
        restic_password: resticPassword,
        prefix,
      });
      setSecretKey("");
      setResticPassword("");
      setAccessKey("");
      setMsg("S3 connector saved to /run/shippin-secrets (and secrets/s3.env). Values are not shown again.");
      await load();
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(null);
    }
  };

  const test = async () => {
    setBusy("test");
    setErr(null);
    setMsg(null);
    try {
      const res = await proxmoxService.testS3();
      if (res.ok) setMsg(res.message ?? "bucket reachable");
      else setErr(res.error ?? "test failed");
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(null);
    }
  };

  const runWorkspace = async () => {
    setBusy("workspace");
    setErr(null);
    setMsg(null);
    try {
      const res = await proxmoxService.runBackup({ kind: "workspace" });
      setLast(res);
      setMsg(`Workspace backup ${res.status} (${res.action_request_id})`);
      await load();
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(null);
    }
  };

  const runImage = async () => {
    const g = guests.find((x) => x.stable_label === guestKey);
    if (!g) {
      setErr("Select a guest to image.");
      return;
    }
    setBusy("image");
    setErr(null);
    setMsg(null);
    try {
      const res = await proxmoxService.runBackup({
        kind: "guest-image",
        node: g.node,
        guest_kind: g.kind,
        vmid: g.vmid,
        storage: storage && storage !== "default" ? storage : undefined,
        mode: "snapshot",
      });
      setLast(res);
      setMsg(`Guest image backup ${res.status} UPID ${res.upid ?? "—"}`);
      await load();
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(null);
    }
  };

  const s3 = status?.s3;
  const storages = status?.pve_backup_storages ?? [];

  return (
    <div>
      <p className="backup-hint">
        WardenClyffe is the internal name for these substrate layers. This tab stores the S3
        connector in <code>/run/shippin-secrets</code> and runs open-source backups: restic to
        object storage for the workspace, Proxmox vzdump for guest images.
      </p>

      {status && (
        <Card
          title="Backup status"
          action={<Badge state={s3?.configured ? "running" : "requested"} />}
        >
          <dl className="kv">
            <dt>message</dt>
            <dd>{status.message}</dd>
            <dt>provider</dt>
            <dd>{s3?.provider || "—"}</dd>
            <dt>endpoint</dt>
            <dd className="mono">{s3?.endpoint || "—"}</dd>
            <dt>bucket</dt>
            <dd className="mono">
              {s3?.bucket || "—"} / {s3?.prefix || "—"}
            </dd>
            <dt>access key</dt>
            <dd className="mono">{s3?.access_key_suffix ? `…${s3.access_key_suffix}` : "not set"}</dd>
            <dt>restic</dt>
            <dd>
              {status.restic_available ? status.restic_version : "not on PATH"} · password{" "}
              {s3?.restic_password_set ? "set" : "missing"}
            </dd>
            <dt>workspace</dt>
            <dd className="mono">{status.workspace_root}</dd>
            <dt>excludes</dt>
            <dd className="mono">{status.excludes.join(", ")}</dd>
          </dl>
        </Card>
      )}

      {err && (
        <Card title="Error">
          <span className="muted" style={{ color: "var(--wc-danger, #f88)" }}>
            {err}
          </span>
        </Card>
      )}
      {msg && (
        <Card title="OK">
          <span className="mono">{msg}</span>
        </Card>
      )}
      {last && (
        <Card title="Last job" action={<Badge state={last.status === "failed" ? "error" : "running"} />}>
          <dl className="kv">
            <dt>kind</dt>
            <dd>{last.kind}</dd>
            <dt>action</dt>
            <dd className="mono">{last.action_request_id}</dd>
            <dt>status</dt>
            <dd>{last.status}</dd>
            <dt>upid</dt>
            <dd className="mono">{last.upid ?? "—"}</dd>
            {last.error && (
              <>
                <dt>error</dt>
                <dd>{last.error}</dd>
              </>
            )}
          </dl>
        </Card>
      )}

      <Card title="S3 connector">
        <p className="backup-hint">
          Secrets are never echoed back. Leave secret/password blank to keep the stored value.
          R2 endpoint looks like <code>https://&lt;accountid&gt;.r2.cloudflarestorage.com</code>.
        </p>
        <form className="backup-form" onSubmit={(e) => void save(e)}>
          <Select label="Provider" selectedKey={provider} onSelect={setProvider} options={PROVIDERS} />
          <Field label="Region" value={region} onChange={setRegion} placeholder="auto or us-east-1" />
          <Field
            label="Endpoint"
            value={endpoint}
            onChange={setEndpoint}
            placeholder="https://….r2.cloudflarestorage.com"
          />
          <Field label="Bucket" value={bucket} onChange={setBucket} placeholder="shippin-backups" />
          <Field label="Prefix" value={prefix} onChange={setPrefix} placeholder="shippin-backups" />
          <Field
            label="Access key"
            value={accessKey}
            onChange={setAccessKey}
            placeholder={s3?.access_key_suffix ? `saved …${s3.access_key_suffix}` : "key id"}
          />
          <Field
            label="Secret key"
            value={secretKey}
            onChange={setSecretKey}
            type="password"
            placeholder="leave blank to keep"
          />
          <Field
            label="restic password"
            value={resticPassword}
            onChange={setResticPassword}
            type="password"
            placeholder="encrypts the restic repo"
          />
          <div className="backup-actions">
            <Button type="submit" tone="brand" variant="solid" isDisabled={busy !== null}>
              {busy === "save" ? "Saving…" : "Save connector"}
            </Button>
            <Button
              type="button"
              tone="neutral"
              variant="outline"
              isDisabled={busy !== null || !s3?.configured}
              onPress={() => void test()}
            >
              {busy === "test" ? "Testing…" : "Test connection"}
            </Button>
          </div>
        </form>
      </Card>

      <Card title="Workspace backup (restic → S3)">
        <p className="backup-hint">
          Optimized image of this host&apos;s W: tree. Skips cache, scratch/tmp, node_modules, and
          cargo target. Fits the 6 TB bucket without uploading 300 GB of rebuildable objects.
        </p>
        <div className="backup-actions">
          <Button
            tone="brand"
            variant="solid"
            isDisabled={busy !== null || !s3?.configured}
            onPress={() => void runWorkspace()}
          >
            {busy === "workspace" ? "Starting…" : "Run workspace backup"}
          </Button>
        </div>
      </Card>

      <Card title="Guest image backup (vzdump)">
        <p className="backup-hint">
          Hypervisor-level image via Proxmox. Lands on a PVE backup storage (snapshot mode). Pulling
          those <code>.vma</code> files into restic/PBS-on-S3 is the next slice.
        </p>
        {guests.length > 0 ? (
          <div className="backup-form">
            <Select
              label="Guest"
              selectedKey={guestKey}
              onSelect={setGuestKey}
              options={guests.map((g) => ({
                id: g.stable_label,
                label: `${g.name || "unnamed"} · ${g.kind} ${g.vmid} (${g.status})`,
              }))}
            />
            <Select
              label="PVE backup storage"
              selectedKey={storage}
              onSelect={setStorage}
              options={[
                { id: "default", label: "Proxmox default" },
                ...storages.map((s) => ({ id: s.storage, label: `${s.storage} (${s.type})` })),
              ]}
            />
            <div className="backup-actions">
              <Button
                tone="neutral"
                variant="outline"
                isDisabled={busy !== null || !guestKey}
                onPress={() => void runImage()}
              >
                {busy === "image" ? "Starting…" : "Run guest image backup"}
              </Button>
            </div>
          </div>
        ) : (
          <p className="muted">No guests in inventory — configure Proxmox on the Fleet tab first.</p>
        )}
      </Card>

      {status?.recent_jobs && status.recent_jobs.length > 0 && (
        <Card title="Recent jobs">
          <dl className="kv">
            {status.recent_jobs.map((j) => (
              <span key={j.id} style={{ display: "contents" }}>
                <dt className="mono">{j.kind}</dt>
                <dd>
                  {j.status}
                  {j.error ? ` — ${j.error}` : ""} · {j.id}
                </dd>
              </span>
            ))}
          </dl>
        </Card>
      )}
    </div>
  );
}
