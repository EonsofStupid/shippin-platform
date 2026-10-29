import { createFileRoute } from "@tanstack/react-router";
import { RequireOperator } from "../domains/warden/identity";
import { ProxmoxFleetView } from "../domains/warden/proxmox";
import { AdminShell } from "../domains/admin/views/AdminShell";

export const Route = createFileRoute("/admin/proxmox")({
  ssr: false,
  component: () => (
    <RequireOperator>
      <AdminShell>
        <ProxmoxFleetView />
      </AdminShell>
    </RequireOperator>
  ),
});
