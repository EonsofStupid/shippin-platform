export type ProxmoxStatus = {
  configured: boolean;
  host: string;
  port: number;
  node: string;
  reachable: boolean;
  version?: Record<string, unknown>;
  error?: string;
  message: string;
};

export type ProxmoxGuest = {
  node: string;
  kind: "qemu" | "lxc" | string;
  vmid: number;
  name: string;
  status: string;
  cpus?: number;
  maxmem?: number;
  maxdisk?: number;
  uptime?: number;
  template?: number;
  stable_label: string;
};

export type GuestActionResult = {
  action_request_id: string;
  kind: string;
  node: string;
  guest_kind: string;
  vmid: number;
  upid?: string;
  status: string;
  exit_status?: string;
  error?: string;
};

export type S3Status = {
  configured: boolean;
  provider?: string;
  endpoint?: string;
  region?: string;
  bucket?: string;
  prefix?: string;
  path_style?: boolean;
  access_key_suffix?: string;
  restic_password_set?: boolean;
};

export type PveStorage = {
  storage: string;
  type: string;
  content: string;
  active?: number;
  enabled?: number;
};

export type BackupJob = {
  id: string;
  kind: string;
  status: string;
  error?: string;
  created_at?: string;
  updated_at?: string;
};

export type BackupStatus = {
  s3: S3Status;
  restic_available: boolean;
  restic_version?: string;
  workspace_root: string;
  excludes: string[];
  pve_backup_storages?: PveStorage[];
  recent_jobs: BackupJob[];
  message: string;
};

export type S3SaveInput = {
  provider: string;
  endpoint: string;
  region: string;
  bucket: string;
  access_key: string;
  secret_key: string;
  restic_password: string;
  prefix: string;
  path_style?: boolean;
};

export type BackupRunInput = {
  kind: "workspace" | "guest-image";
  node?: string;
  guest_kind?: string;
  vmid?: number;
  storage?: string;
  mode?: string;
};
