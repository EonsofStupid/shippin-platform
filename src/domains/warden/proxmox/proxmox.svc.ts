import type {
  BackupRunInput,
  BackupStatus,
  GuestActionResult,
  ProxmoxGuest,
  ProxmoxStatus,
  S3SaveInput,
} from "./types";

const TOKEN_KEY = "warden.token";

function authHeaders(): HeadersInit {
  const token = localStorage.getItem(TOKEN_KEY) ?? "";
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (token) headers.authorization = `Bearer ${token}`;
  return headers;
}

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: { ...authHeaders(), ...(init?.headers ?? {}) },
    ...init,
  });
  const text = await res.text();
  const body = text ? JSON.parse(text) : null;
  if (!res.ok) {
    const msg = body?.error?.message ?? body?.error ?? res.statusText;
    throw new Error(typeof msg === "string" ? msg : res.statusText);
  }
  return body as T;
}

export const proxmoxService = {
  status: () => req<ProxmoxStatus>("/api/warden/proxmox/status"),
  guests: () => req<{ count: number; items: ProxmoxGuest[] }>("/api/warden/proxmox/guests"),
  start: (node: string, kind: string, vmid: number) =>
    req<GuestActionResult>(`/api/warden/proxmox/guests/${node}/${kind}/${vmid}/start`, {
      method: "POST",
    }),
  stop: (node: string, kind: string, vmid: number) =>
    req<GuestActionResult>(`/api/warden/proxmox/guests/${node}/${kind}/${vmid}/stop`, {
      method: "POST",
    }),
  backupStatus: () => req<BackupStatus>("/api/warden/proxmox/backup/status"),
  saveS3: (input: S3SaveInput) =>
    req<{ ok: boolean; s3: BackupStatus["s3"] }>("/api/warden/proxmox/backup/s3", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  testS3: () =>
    req<{ ok: boolean; message?: string; error?: string }>("/api/warden/proxmox/backup/s3/test", {
      method: "POST",
      body: JSON.stringify({}),
    }),
  runBackup: (input: BackupRunInput) =>
    req<GuestActionResult>("/api/warden/proxmox/backup/run", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  action: (id: string) => req<Record<string, unknown>>(`/api/warden/proxmox/actions/${id}`),
};
