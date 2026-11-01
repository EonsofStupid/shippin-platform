# Shippin Platform

Shippin is being developed as a self-hosted infrastructure and customer panel.
The first infrastructure target is Incus; Proxmox is a separate provider. The
React application is under `src/`, and existing Go services are under `services/`.

**Status:** This repository is not yet a verified server release. The current
API still has authorization and configuration gaps, and primary installation,
node enrollment, real provider operation, upgrades and recovery need end-to-end
verification. See the [release roadmap](docs/shippin-platform-roadmap.md) and
[contribution guide](CONTRIBUTING.md).

The material below is retained as historical project context. Its product names,
performance claims, topology and installation commands have not been verified as
the current Shippin platform and must be reviewed before use.

---

# WardenClyffe - Warden Server Manager And Clyffe Customer Portal

<div align="center">

**A free, modern server and customer operations platform for Proxmox-backed infrastructure**

[![Rust](https://img.shields.io/badge/rust-1.70%2B-orange.svg)](https://www.rust-lang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Sponsor](https://img.shields.io/badge/Sponsor-Heart-ea4aaa.svg)](https://github.com/sponsors/wardenclyffesoftwaresystemsltd)

**[wardenclyffescale.org](https://wardenclyffescale.org)** | **[wardenclyffe.uk.com](https://wardenclyffe.uk.com)** | **[Discord](https://discord.gg/q9qMjHjUQY)** | **[Reddit](https://www.reddit.com/r/WardenClyffe/)**

Copyright 2024-2026 WardenClyffe Software Systems Ltd

</div>

---

WardenClyffe is being built as two coordinated product surfaces:

- **Warden**: the operator/server-control platform for our hosts, Proxmox nodes, VMs, LXC containers, storage, networking, backups, AI operations, and automation.
- **Clyffe**: the customer portal, knowledge base, tickets, CRM, and customer-safe service panel built on Warden's API.

The first target is an internal, turnkey two-server deployment. Proxmox VE is the infrastructure substrate for KVM, LXC, storage, networking, backups, tasks, and consoles. WardenClyffe wraps that substrate with a modern product identity, tenant-safe APIs, customer panels, and AI-assisted operations.

Canonical alignment docs:

- [Module map](docs/WARDENCLYFFE_MODULE_MAP.md): Module 1 is Warden, Module 2 is Clyffe.
- [Naming conventions](docs/WARDENCLYFFE_NAMING_CONVENTIONS.md): boring product, API, database, and MCP names.
- [DDD folder structure](docs/HYPERMODULAR_DDD_FOLDER_STRUCTURE.md): self-describing Warden/Clyffe module scaffold.
- [Host fleet onboarding](docs/HOST_FLEET_AND_ONBOARDING.md): Wisconsin and Virginia host onboarding path.
- [Public IP homebase](docs/PUBLIC_IP_HOMEBASE_FOUNDATION.md): domain, edge, public IP, remote host, and Tailscale-exit foundation.
- [Foundation service matrix](docs/FOUNDATION_SERVICE_MATRIX.md): current VM/LXC services, app choices, and configuration gates.
- [Foundation app research](docs/FOUNDATION_APP_RESEARCH_2026_05.md): May 2026 primary-source review of OPNsense, Authentik, Caddy, Traefik, DNS, Postgres, Qdrant, SurrealDB, and Infisical.
- [Architecture](docs/WARDEN_CLYFFE_ARCHITECTURE.md): current Warden/Clyffe boundary.
- [Proxmox free cheat sheet](docs/PROXMOX_FREE_CHEATSHEET.md): Warden's Proxmox API coverage map.
- [Backend options review](docs/WARDENCLYFFE_BACKEND_OPTIONS_2026_05.md): MariaDB, Postgres, SurrealDB, Qdrant, Dragonfly, and distributed SQL review.
- [Modern patterns](docs/WARDENCLYFFE_MODERN_PATTERNS_2026_05.md): May 2026 patterns to preserve.

| Tool | Description | Status |
|------|-------------|--------|
| **Warden** | Server, VM, LXC, storage, automation, and Proxmox-native operator panel | Pilot |
| **Clyffe** | Customer portal, tickets, CRM, service panel, and knowledge base | Planned |
| **[WardenClyffeScale](#wardenclyffescale--database-replication)** | MariaDB/MySQL replication, clustering, and load balancing | Existing component |
| **[WardenClyffeDisk](#wardenclyffedisk--distributed-filesystem)** | Disk sharing and replication across networks | Existing component |
| **[WardenClyffeNet](#wardenclyffenet--private-networking)** | Secure private networking across the internet | Existing component |
| **[Mediastack](mediastack/README.md)** | Hades-owned private media VM for premium community members and invited friends (invite-only, isolated network) | Scaffold |

---

## WardenClyffeScale â€” Database Replication

**Database replication, clustering, and load balancing â€” the easy way**

WardenClyffeScale is a lightweight, high-availability replication layer for MariaDB/MySQL. It provides **automatic leader election** with deterministic failover, **WAL-based replication** for strong consistency, and a **MySQL-compatible proxy** for transparent routingâ€”all in a single Rust binary.

Works with MySQL, Percona, and Amazon RDS â€¢ **MariaDB recommended**

### Why WardenClyffeScale?

| Feature | Benefit |
|---------|---------|
| **Sub-Millisecond Replication** | Push-based replication faster than MySQL, MariaDB, or Galera |
| **Zero Write Conflicts** | Single-leader model eliminates certification failures |
| **Predictable Failover** | Lowest node ID always winsâ€”you know exactly who becomes leader |
| **Safe Node Rejoin** | Returning nodes sync via WAL before taking leadership |
| **Zero-Config Discovery** | Nodes find each other automatically via UDP broadcast |
| **Transparent Proxy** | Connect via MySQL protocolâ€”no application changes needed |
| **Built-in Load Balancer** | Distribute connections across cluster nodes with automatic failover |
| **Single Binary** | No patched databases, no complex dependencies |

### Fastest Replication in the Industry

| System | Typical Replication Lag |
|--------|------------------------|
| MySQL Async | 100ms - seconds |
| MySQL Semi-Sync | 10-50ms |
| MariaDB Galera | 10-20ms |
| **WardenClyffeScale** | **<1ms** âš¡ |

### Quick Start

> **All cluster nodes MUST have identical data before starting WardenClyffeScale.** WardenClyffeScale replicates new changes only â€” it does NOT sync existing data between nodes.

```bash
# Install WardenClyffeScale on each server
curl -sSL https://raw.githubusercontent.com/wardenclyffesoftwaresystemsltd/WardenClyffeScale/main/setup.sh | bash
```

### Load Balancer

Install the load balancer directly on any server that needs database access. It auto-discovers your cluster â€” no configuration needed.

```bash
curl -sSL https://raw.githubusercontent.com/wardenclyffesoftwaresystemsltd/WardenClyffeScale/main/setup_lb.sh | bash
```

```
Web Server 1 â”€â”€â”€ WardenClyffeScale LB â”€â”€â”€â”
             (auto-discovers)    â”‚
                                 â”‚
Web Server 2 â”€â”€â”€ WardenClyffeScale LB â”€â”€â”€â”¼â”€â”€â–º WardenClyffeScale DB Cluster
             (auto-discovers)    â”‚
                                 â”‚
Web Server 3 â”€â”€â”€ WardenClyffeScale LB â”€â”€â”€â”˜
             (auto-discovers)
```

### Cluster Commands

```bash
wardenclyffectl list servers     # Check cluster status
wardenclyffectl stats            # Live throughput monitoring
wardenclyffectl migrate --from 10.0.10.111:8080   # Migrate data to new node
wardenclyffectl reset --force    # Reset WAL and state (DESTRUCTIVE)
```

---

## WardenClyffeDisk â€” Distributed Filesystem

**Disk sharing and replication across networks â€” the easy way**

WardenClyffeDisk is a FUSE-based distributed filesystem that shares and replicates files across Linux servers. Mount a shared directory on any number of machines and have your data automatically synchronised. Supports leader, follower, and client modes.

### Why WardenClyffeDisk?

| Feature | Benefit |
|---------|---------|
| **FUSE-Based** | Mount as a regular filesystemâ€”works with any application |
| **Automatic Replication** | Files sync to all nodes automatically |
| **Content-Addressed Storage** | Efficient deduplication via SHA256 chunking |
| **Leader-Follower Model** | Strong consistency with automatic failover |
| **Client Mode** | Workstations can access the shared drive without local storage |
| **Multiple Drives** | Run multiple independent filesystems per node |

### Quick Start

```bash
# Interactive installer - prompts for node ID, role, and discovery
curl -sSL https://raw.githubusercontent.com/wardenclyffesoftwaresystemsltd/WardenClyffeScale/main/wardenclyffedisk/setup.sh | bash
```

### Node Roles

| Role | Storage | Description |
|------|---------|-------------|
| **Leader** | âœ… Local | Primary node â€” accepts writes, replicates to followers |
| **Follower** | âœ… Local | Receives replicated data, can become leader on failover |
| **Client** | âŒ None | Mount-only â€” reads/writes forwarded to leader, no local data |
| **Auto** | âœ… Local | Auto-election â€” lowest ID becomes leader |

See [`wardenclyffedisk/README.md`](wardenclyffedisk/README.md) for full documentation.

---

## WardenClyffeNet â€” Private Networking

WardenClyffeNet creates a secure, encrypted private network between your machines over the internet. Machines on WardenClyffeNet can see each other as if they were on the same LAN, but all traffic is encrypted with modern cryptography (X25519 + ChaCha20-Poly1305 â€” the same crypto as WireGuard).

### Why WardenClyffeNet?

| Feature | Benefit |
|---------|---------|
| **WireGuard-Class Crypto** | X25519 key exchange + ChaCha20-Poly1305 AEAD encryption |
| **Mesh Networking** | Every node can reach every other node directly â€” no single point of failure |
| **Invite/Join System** | Connect new peers with a single token â€” no manual key exchange |
| **Relay Forwarding** | Nodes behind NAT can communicate through a relay â€” no port forwarding needed |
| **Gateway Mode** | Route internet traffic through a gateway node with NAT masquerading |
| **LAN Auto-Discovery** | Nodes find each other automatically on the same network |
| **TUN-Based** | Uses kernel TUN interfaces for near-native performance |
| **Hostname/DynDNS** | Use hostnames in endpoints â€” re-resolved every 60s for dynamic IPs |
| **Single Binary** | No dependencies â€” just `wardenclyffenet` and `wardenclyffenetctl` |
| **Systemd Service** | Runs as a background service with automatic startup |

### Quick Start

```bash
# Interactive installer â€” downloads binary, generates keys, creates systemd service
curl -sSL https://raw.githubusercontent.com/wardenclyffesoftwaresystemsltd/WardenClyffeScale/main/wardenclyffenet/setup.sh | sudo bash
```

The installer will:
- Check for `/dev/net/tun` (with Proxmox/LXC guidance if missing)
- Download and install `wardenclyffenet` and `wardenclyffenetctl`
- Generate an X25519 keypair
- Prompt for WardenClyffeNet IP address, port, and gateway mode
- Create a systemd service for automatic startup

### Easy Peer Setup (Invite/Join)

Connect two machines in seconds â€” no manual key exchange:

```bash
# On the first machine (the one with a public IP / port forwarding):
sudo wardenclyffenet invite

# Output:  sudo wardenclyffenet --config /etc/wardenclyffenet/config.toml join eyJwa...

# Copy that command and run it on the second machine:
sudo wardenclyffenet --config /etc/wardenclyffenet/config.toml join eyJwa...

# It will output a reverse token â€” run that on the first machine:
sudo wardenclyffenet --config /etc/wardenclyffenet/config.toml join eyJlc...

# Restart WardenClyffeNet on both:
sudo systemctl restart wardenclyffenet
```

The invite token auto-detects the node's public IP, includes the public key, and assigns WardenClyffeNet IPs automatically.

### NAT Traversal (Relay Forwarding)

WardenClyffeNet supports **relay forwarding** so machines behind NAT firewalls can communicate without port forwarding:

```
Laptop (behind NAT)          Server (public IP)           Home PC (behind NAT)
    10.0.10.1                   10.0.10.2                   10.0.10.3
        â”‚                           â”‚                           â”‚
        â””â”€â”€â”€â”€â”€ encrypted UDP â”€â”€â”€â”€â”€â”€â–ºâ”‚â—„â”€â”€â”€â”€â”€â”€ encrypted UDP â”€â”€â”€â”€â”€â”˜
                                    â”‚
                             Relay forwards
                            packets between
                           Laptop â—„â”€â”€â–º Home PC
```

**How it works:**
1. Both the laptop and home PC connect to the server (which has a public IP)
2. When the laptop sends a packet to the home PC, the server detects it's not the destination
3. The server decrypts, re-encrypts for the home PC, and forwards it
4. This happens automatically â€” no configuration needed
5. Any node that can be reached by both parties can act as a relay

### Peer Discovery Methods

WardenClyffeNet supports three ways to find and connect to peers â€” mix and match as needed:

| Method | Use Case | Config |
|--------|----------|--------|
| **LAN Auto-Discovery** | Machines on the same network | `discovery = true` (default) |
| **Static IP** | VPS, dedicated servers, data centres | `endpoint = "203.0.113.5:9600"` |
| **Hostname / DynDNS** | Home broadband, dynamic IPs | `endpoint = "myhome.dyndns.org:9600"` |

Hostnames are resolved on startup and **re-resolved every 60 seconds**, so DynDNS changes are picked up automatically. Works with any DNS provider â€” DynDNS, No-IP, Cloudflare, DuckDNS, or your own domain.

### Multi-Server Deployment (Static IPs)

Link multiple standalone servers across different locations into a single WardenClyffeNet mesh:

```
Server A (London)              Server B (New York)           Server C (Tokyo)
  Public: 203.0.113.5            Public: 198.51.100.10         Public: 192.0.2.50
  WardenClyffeNet: 10.0.10.1             WardenClyffeNet: 10.0.10.2            WardenClyffeNet: 10.0.10.3
       â”‚                              â”‚                             â”‚
       â””â”€â”€â”€â”€â”€â”€ encrypted UDP â”€â”€â”€â”€â”€â”€â”€â”€â–ºâ”‚â—„â”€â”€â”€â”€â”€ encrypted UDP â”€â”€â”€â”€â”€â”€â”€â”€â”˜
```

1. Install WardenClyffeNet on each server, giving each a unique WardenClyffeNet IP
2. Use `sudo wardenclyffenet invite` on one server to generate invite tokens
3. Run the invite command on each other server to exchange keys
4. Restart WardenClyffeNet â€” PEX automatically propagates the full mesh topology

> ðŸ’¡ **You don't need a full mesh in the config.** Each server only needs to know about at least one other server. PEX shares the rest automatically within 30 seconds.

### Architecture

```
Machine A (10.0.10.1)          Machine B (10.0.10.2)
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”            â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚  wardenclyffenet0 (TUN) â”‚â—„â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â–ºâ”‚  wardenclyffenet0 (TUN) â”‚
â”‚  10.0.10.1/24   â”‚  Encrypted â”‚  10.0.10.2/24   â”‚
â”‚  ChaCha20-Poly  â”‚  UDP/9600  â”‚  ChaCha20-Poly  â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜            â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
         â–²                              â–²
         â”‚       Encrypted UDP          â”‚
         â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
                    â”‚
           â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â–¼â”€â”€â”€â”€â”€â”€â”€â”€â”
           â”‚  Machine C      â”‚
           â”‚  (Gateway)      â”‚
           â”‚  10.0.10.3/24   â”‚
           â”‚  NAT â†’ Internet â”‚
           â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
```

### CLI Reference

```bash
# Daemon
wardenclyffenet                          # Start the daemon (usually via systemd)
wardenclyffenet init --address 10.0.10.1 # Generate config and keypair
wardenclyffenet genkey                   # Generate a new X25519 keypair
wardenclyffenet pubkey                   # Show this node's public key
wardenclyffenet token                    # Show join token for sharing
wardenclyffenet invite                   # Generate invite token for a new peer
wardenclyffenet join <token>             # Join a network using an invite token

# Control utility
wardenclyffenetctl status                # Show node status, IP, uptime
wardenclyffenetctl peers                 # List peers with connection status
wardenclyffenetctl info                  # Combined status and peer list

# Service management
sudo systemctl start wardenclyffenet     # Start service
sudo systemctl status wardenclyffenet    # Check status
sudo journalctl -u wardenclyffenet -f    # View logs
```

### Configuration Example

```toml
[network]
address = "10.0.10.1"
listen_port = 9600
discovery = true        # LAN auto-discovery (default)

# Static IP peer
[[peers]]
public_key = "BASE64_PUBLIC_KEY"
endpoint = "203.0.113.5:9600"
allowed_ip = "10.0.10.2"
name = "london-vps"

# DynDNS hostname peer (re-resolved every 60s)
[[peers]]
public_key = "ANOTHER_PUBLIC_KEY"
endpoint = "myhome.dyndns.org:9600"
allowed_ip = "10.0.10.3"
name = "home-server"
```

### Security

| Layer | Technology |
|-------|------------|
| Key Exchange | **X25519** (Curve25519 Diffie-Hellman) |
| Encryption | **ChaCha20-Poly1305** AEAD (256-bit) |
| Replay Protection | Counter-based nonces with monotonic validation |
| Network Isolation | iptables firewall blocks all external inbound traffic |
| Key Storage | Private keys stored with 0600 permissions |

> âš ï¸ **Proxmox/LXC Users:** The TUN device (`/dev/net/tun`) is blocked by default in LXC containers. See [wardenclyffescale.org/wardenclyffenet.html](https://wardenclyffescale.org/wardenclyffenet.html) for setup instructions.

---

## Architecture (WardenClyffeScale)

| Layer        | Component                                      |
|--------------|-------------------------------------------------|
| Applications | Connect via HTTP API or MySQL Protocol          |
| WardenClyffeScale    | Leader + Followers replicate via WAL            |
| Database     | Each node has local MariaDB (localhost:3306)    |

**Write Flow:** Client â†’ Any Node â†’ Forwarded to Leader â†’ Replicated to All Nodes

**Read Flow:** Client â†’ Any Node â†’ Local Data (or forwarded to Leader if node is behind)

## Cluster Sizing

| Nodes | Fault Tolerance   | Use Case                        |
|-------|-------------------|---------------------------------|
| 1     | None              | Development only                |
| 2     | 1 node failure    | Basic HA (not recommended)      |
| 3     | 2 node failures   | Minimum for production          |
| 5     | 4 node failures   | Recommended for production      |
| 7     | 6 node failures   | High availability               |

**Geo-Distribution:** Nodes can be deployed across different data centres or regions. Connect to your nearest node for low-latency reads â€” if the data isn't up-to-date, the request is automatically forwarded to the leader.

> **Note:** WardenClyffeScale doesn't use quorum â€” only one node needs to survive. While the cluster can run on a single remaining node, it's recommended to maintain at least 2 active nodes for redundancy.

## Documentation

- **Website:** [wardenclyffescale.org](https://wardenclyffescale.org)
- **Full Docs:** [docs/DOCUMENTATION.md](docs/DOCUMENTATION.md)
- **WardenClyffeDisk Docs:** [wardenclyffedisk/README.md](wardenclyffedisk/README.md)
- **WardenClyffeNet Docs:** [wardenclyffescale.org/wardenclyffenet.html](https://wardenclyffescale.org/wardenclyffenet.html)
- **MariaDB/MySQL Editor:** [wardenclyffescale.org/wardenclyffe-mysql.html](https://wardenclyffescale.org/wardenclyffe-mysql.html)

---

## Support

- â¤ï¸ **Sponsor:** [Support development](https://github.com/sponsors/wardenclyffesoftwaresystemsltd)
- ðŸ’¬ **Discord:** [Join our community](https://discord.gg/q9qMjHjUQY)
- ðŸ”¥ **Reddit:** [r/WardenClyffe](https://www.reddit.com/r/WardenClyffe/)
- ðŸŒ **Website:** [wardenclyffe.uk.com](https://wardenclyffe.uk.com)
- â­ **GitHub:** [Star this repo](https://github.com/wardenclyffesoftwaresystemsltd/WardenClyffeScale)
- ðŸ› **Issues:** [Report a bug](https://github.com/wardenclyffesoftwaresystemsltd/WardenClyffeScale/issues)

---

## License

[MIT License](LICENSE) â€” Free to use, modify, and distribute without restriction.

Â© 2024-2026 [WardenClyffe Software Systems Ltd](https://wardenclyffe.uk.com/)

## âš ï¸ Disclaimer

**USE AT YOUR OWN RISK.** This software is provided "as is" without warranty of any kind, express or implied, including but not limited to the warranties of merchantability, fitness for a particular purpose, and noninfringement. In no event shall WardenClyffe Software Systems Ltd be liable for any claim, damages, or other liability arising from the use of this software.

By using WardenClyffe tools, you acknowledge that you are solely responsible for your data and any consequences of using this software.

---

*We hand code and use AI to assist with the development of this software.*
