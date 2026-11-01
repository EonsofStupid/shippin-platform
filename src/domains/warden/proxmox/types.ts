import type { z } from "zod";
import type {
  guestActionSchema,
  proxmoxGuestSchema,
  proxmoxStatusSchema,
} from "./proxmox.schema.ts";

export type ProxmoxStatus = z.infer<typeof proxmoxStatusSchema>;
export type ProxmoxGuest = z.infer<typeof proxmoxGuestSchema>;
export type GuestActionResult = z.infer<typeof guestActionSchema>;
