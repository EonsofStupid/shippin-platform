// Typed client for warden-api. Same-origin in dev via the Vite proxy.

export type Workspace = {
  id: string;
  stable_id: string;
  tenant_id: string | null;
  kind: string;
  state: string;
  vmid: number | null;
  hostname: string | null;
  tier: string | null;
  cores: number | null;
  memory_mb: number | null;
  disk_gb: number | null;
  ip_cidr: string | null;
  ssh_alias: string | null;
  spec: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

export type SchemaInfo = { name: string; tables: number };
export type TableInfo = {
  schema: string;
  name: string;
  columns: number;
  rows_estimate: number;
  size_bytes: number;
};
export type ColumnInfo = {
  name: string;
  type: string;
  nullable: boolean;
  default: string | null;
  is_primary_key: boolean;
  position: number;
};
export type RowPage = {
  columns: string[];
  rows: Record<string, unknown>[];
  total: number;
  limit: number;
  offset: number;
};
export type QueryResult = { columns: string[]; rows: Record<string, unknown>[]; count: number };
export type ProvisionResult = { action_request_id: string; resource: Workspace };

export type Platform = {
  stable_id: string;
  kind: string;
  api_endpoint: string | null;
  network_name: string | null;
  managed: boolean;
  credential_ref: string | null;
  state: string;
  note: string | null;
};
export type Designation = "connector" | "plugin" | "platform" | "core";
export type FoundationService = {
  stable_id: string;
  name: string;
  role: string;
  designation: Designation;
  audiences: string[];
  platform: string | null;
  vmid: number | null;
  endpoint: string | null;
  auth_required: boolean;
  credential_ref: string | null;
  keep_decision: string | null;
  customer_ready: boolean;
  owed: string[];
  state: string;
};
export type IntelProbe = {
  stable_id: string;
  role: string;
  endpoint: string | null;
  reachable: boolean;
  http_status?: number;
  credential_gated: boolean;
  credential_ref: string | null;
  detail: string;
};
export type ClyffyHome = {
  orchestrator: string;
  persona: string;
  summary: {
    platforms: number;
    services: number;
    services_by_role: Record<string, number>;
    services_by_designation?: Record<string, number>;
    config_items_owed: number;
    credential_gated_services: number;
  };
  platforms: Platform[];
  intelligence: IntelProbe[];
  note: string;
};

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  const text = await res.text();
  const body = text ? JSON.parse(text) : null;
  if (!res.ok) {
    const msg = body?.error?.message ?? res.statusText;
    throw new Error(msg);
  }
  return body as T;
}

function operatorRequest<T>(token: string | null, path: string, init?: RequestInit): Promise<T> {
  if (!token?.trim()) return Promise.reject(new Error("Operator authentication is required."));
  return req<T>(path, {
    ...init,
    headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
  });
}

export const api = {
  health: () => req<{ status: string; service: string }>("/healthz"),
  workspaces: (token: string | null) =>
    operatorRequest<{ items: Workspace[]; count: number }>(token, "/api/warden/workspaces"),
  provision: (token: string | null, input: { tier: string; repo?: string; region?: string }) =>
    operatorRequest<ProvisionResult>(token, "/api/warden/provision", {
      method: "POST",
      body: JSON.stringify(input),
    }),

  schemas: (token: string | null) =>
    operatorRequest<{ items: SchemaInfo[] }>(token, "/api/warden/data/schemas"),
  tables: (token: string | null, schema: string) =>
    operatorRequest<{ items: TableInfo[] }>(
      token,
      `/api/warden/data/schemas/${encodeURIComponent(schema)}/tables`,
    ),
  columns: (token: string | null, schema: string, table: string) =>
    operatorRequest<{ schema: string; table: string; columns: ColumnInfo[] }>(
      token,
      `/api/warden/data/tables/${encodeURIComponent(schema)}/${encodeURIComponent(table)}`,
    ),
  rows: (token: string | null, schema: string, table: string, limit = 50, offset = 0) =>
    operatorRequest<RowPage>(
      token,
      `/api/warden/data/tables/${encodeURIComponent(schema)}/${encodeURIComponent(table)}/rows?limit=${limit}&offset=${offset}`,
    ),
  query: (token: string | null, sql: string) =>
    operatorRequest<QueryResult>(token, "/api/warden/data/query", {
      method: "POST",
      body: JSON.stringify({ sql }),
    }),

  clyffyHome: (token: string | null) => operatorRequest<ClyffyHome>(token, "/api/clyffy/home"),
  clyffyServices: (token: string | null) =>
    operatorRequest<{ items: FoundationService[]; count: number }>(token, "/api/clyffy/services"),
  clyffyPlugins: (token: string | null) =>
    operatorRequest<{
      designation: string;
      audience: string;
      items: FoundationService[];
      count: number;
    }>(token, "/api/clyffy/plugins"),
  clyffyConnectors: (token: string | null) =>
    operatorRequest<{
      designation: string;
      audience: string;
      items: FoundationService[];
      count: number;
    }>(token, "/api/clyffy/connectors"),
  clyffyIntelligence: (token: string | null) =>
    operatorRequest<{ plane: string; probes: IntelProbe[] }>(token, "/api/clyffy/intelligence"),
};

export function cell(v: unknown): string {
  if (v === null || v === undefined) return "∅";
  if (typeof v === "object") return JSON.stringify(v);
  return String(v);
}
