import { z } from "zod";

// These schemas validate the existing Shippin API responses. They are not the
// provider-neutral compute contract; that migration needs a separate API review.
export const proxmoxStatusSchema = z.looseObject({
  configured: z.boolean(),
  host: z.string(),
  port: z.number().int(),
  node: z.string(),
  reachable: z.boolean(),
  version: z.record(z.string(), z.unknown()).optional(),
  error: z.string().optional(),
  message: z.string(),
});

export const proxmoxGuestSchema = z.looseObject({
  node: z.string(),
  kind: z.string(),
  vmid: z.number().int(),
  name: z.string(),
  status: z.string(),
  cpus: z.number().optional(),
  maxmem: z.number().optional(),
  maxdisk: z.number().optional(),
  uptime: z.number().optional(),
  template: z.number().optional(),
  stable_label: z.string(),
});

// Go encodes a nil slice as null; accept that only for an empty inventory.
export const proxmoxGuestsSchema = z
  .looseObject({
    count: z.number().int().nonnegative(),
    items: z.array(proxmoxGuestSchema).nullable(),
  })
  .refine((value) => value.count === (value.items?.length ?? 0), {
    message: "Inventory count does not match items",
  })
  .transform((value) => ({ ...value, items: value.items ?? [] }));

export const guestActionSchema = z.looseObject({
  action_request_id: z.string(),
  kind: z.string(),
  node: z.string(),
  guest_kind: z.string(),
  vmid: z.number().int(),
  upid: z.string().optional(),
  status: z.string(),
  exit_status: z.string().optional(),
  error: z.string().optional(),
});
