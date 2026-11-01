import { z } from "zod";
import { guestActionSchema, proxmoxGuestsSchema, proxmoxStatusSchema } from "./proxmox.schema.ts";

async function req<T>(
  path: string,
  token: string,
  schema: z.ZodType<T>,
  init?: RequestInit,
): Promise<T> {
  if (!token.trim()) throw new Error("Operator authentication is required.");
  const res = await fetch(path, {
    headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
    ...init,
  });
  let body: unknown;
  try {
    body = await res.json();
  } catch {
    throw new Error(`Invalid API response (HTTP ${res.status}).`);
  }
  if (!res.ok) {
    const error = z
      .looseObject({ error: z.union([z.string(), z.looseObject({ message: z.string() })]) })
      .safeParse(body);
    const message = error.success ? error.data.error : undefined;
    throw new Error(
      typeof message === "string"
        ? message
        : (message?.message ?? `Request failed (HTTP ${res.status}).`),
    );
  }
  const parsed = schema.safeParse(body);
  if (!parsed.success) throw new Error("Invalid Proxmox API response.");
  return parsed.data;
}

const nodeName = z.string().regex(/^[A-Za-z0-9][A-Za-z0-9._-]*$/);
const guestKind = z.enum(["qemu", "lxc"]);
const guestId = z.number().int().positive();

function actionPath(node: string, kind: string, vmid: number, action: "start" | "stop"): string {
  const values = z.tuple([nodeName, guestKind, guestId]).safeParse([node, kind, vmid]);
  if (!values.success) throw new Error("Invalid Proxmox guest reference.");
  return `/api/warden/proxmox/guests/${encodeURIComponent(values.data[0])}/${values.data[1]}/${values.data[2]}/${action}`;
}

export const proxmoxService = {
  status: (token: string) => req("/api/warden/proxmox/status", token, proxmoxStatusSchema),
  guests: (token: string) => req("/api/warden/proxmox/guests", token, proxmoxGuestsSchema),
  start: (token: string, node: string, kind: string, vmid: number) =>
    req(actionPath(node, kind, vmid, "start"), token, guestActionSchema, { method: "POST" }),
  stop: (token: string, node: string, kind: string, vmid: number) =>
    req(actionPath(node, kind, vmid, "stop"), token, guestActionSchema, { method: "POST" }),
};
