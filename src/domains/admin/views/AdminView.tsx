// Admin — the authenticated shell. After login at auth.rrflow.ai the operator
// lands here and the role gate decides Admin (Warden) vs Customer surfaces.
// The customer plane (folded in as `admin/storage` for now) must stay
// role-separated from operator views — do not blend them into one screen.
import { useAuth } from "../../warden/identity";

export function AdminView() {
  const { operator } = useAuth();
  return (
    <section className="admin">
      <h1>Admin</h1>
      <p>Signed in as {operator?.username ?? "—"} ({operator?.role ?? "—"}).</p>
      <nav className="admin-nav">
        <a href="/admin/proxmox">
          <strong>Proxmox fleet</strong> — inventory, start/stop, Backup tab
        </a>
        <a href="/warden">Warden overview</a>
        <a href="/admin/connect">Connect &amp; Launch</a>
        <a href="/admin/control">Control layer</a>
        <a href="/admin/intelligence">Intelligence layer</a>
        <a href="/admin/edge">Public IPs</a>
      </nav>
    </section>
  );
}
