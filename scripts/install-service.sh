#!/bin/bash
#
# WardenClyffeScale Service Installation Script
#
# This script installs WardenClyffeScale as a systemd service.
#
# Usage:
#   sudo ./install-service.sh [options]
#
# Options:
#   --node-id ID        Node identifier (default: node-1)
#   --config PATH       Path to configuration file
#   --user USER         User to run the service as (default: wardenclyffescale)
#   --bootstrap         Configure as bootstrap/leader node
#   --uninstall         Remove the service and cleanup
#   --help              Show this help message
#

set -e

# Default values
NODE_ID="node-1"
SERVICE_USER="wardenclyffescale"
INSTALL_DIR="/opt/wardenclyffescale"
CONFIG_DIR="/etc/wardenclyffescale"
DATA_DIR="/var/lib/wardenclyffescale"
LOG_DIR="/var/log/wardenclyffescale"
CONFIG_FILE=""
BOOTSTRAP=""
UNINSTALL=false

# Script location
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --node-id)
            NODE_ID="$2"
            shift 2
            ;;
        --config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        --user)
            SERVICE_USER="$2"
            shift 2
            ;;
        --bootstrap)
            BOOTSTRAP="--bootstrap"
            shift
            ;;
        --uninstall)
            UNINSTALL=true
            shift
            ;;
        --help)
            echo "WardenClyffeScale Service Installation Script"
            echo ""
            echo "Usage:"
            echo "  sudo $0 [options]"
            echo ""
            echo "Options:"
            echo "  --node-id ID        Node identifier (default: node-1)"
            echo "  --config PATH       Path to configuration file"
            echo "  --user USER         User to run the service as (default: wardenclyffescale)"
            echo "  --bootstrap         Configure as bootstrap/leader node"
            echo "  --uninstall         Remove the service and cleanup"
            echo "  --help              Show this help message"
            echo ""
            echo "Example:"
            echo "  sudo $0 --node-id node-1 --bootstrap"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Check for root
if [[ $EUID -ne 0 ]]; then
    print_error "This script must be run as root (use sudo)"
    exit 1
fi

# Uninstall mode
if [[ "$UNINSTALL" == true ]]; then
    echo "Uninstalling WardenClyffeScale service..."
    
    # Stop and disable service
    if systemctl is-active --quiet wardenclyffescale; then
        systemctl stop wardenclyffescale
        print_status "Service stopped"
    fi
    
    if systemctl is-enabled --quiet wardenclyffescale 2>/dev/null; then
        systemctl disable wardenclyffescale
        print_status "Service disabled"
    fi
    
    # Remove service file
    if [[ -f /etc/systemd/system/wardenclyffescale.service ]]; then
        rm /etc/systemd/system/wardenclyffescale.service
        systemctl daemon-reload
        print_status "Service file removed"
    fi
    
    # Ask about data removal
    echo ""
    read -p "Remove WardenClyffeScale binary from ${INSTALL_DIR}? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$INSTALL_DIR"
        print_status "Installation directory removed"
    fi
    
    read -p "Remove data directory ${DATA_DIR}? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$DATA_DIR"
        print_status "Data directory removed"
    fi
    
    read -p "Remove configuration ${CONFIG_DIR}? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$CONFIG_DIR"
        print_status "Configuration directory removed"
    fi
    
    print_status "WardenClyffeScale service uninstalled"
    exit 0
fi

echo "=========================================="
echo "  WardenClyffeScale Service Installer"
echo "=========================================="
echo ""
echo "Node ID:     $NODE_ID"
echo "User:        $SERVICE_USER"
echo "Install Dir: $INSTALL_DIR"
echo "Config Dir:  $CONFIG_DIR"
echo "Data Dir:    $DATA_DIR"
if [[ -n "$BOOTSTRAP" ]]; then
    echo "Mode:        Leader (bootstrap)"
else
    echo "Mode:        Follower"
fi
echo ""

# Check for binary
BINARY=""
if [[ -f "${PROJECT_DIR}/target/release/wardenclyffescale" ]]; then
    BINARY="${PROJECT_DIR}/target/release/wardenclyffescale"
elif [[ -f "${PROJECT_DIR}/target/debug/wardenclyffescale" ]]; then
    BINARY="${PROJECT_DIR}/target/debug/wardenclyffescale"
    print_warning "Using debug build. Consider using release build for production."
else
    print_error "WardenClyffeScale binary not found. Please build first:"
    echo "  cd ${PROJECT_DIR} && cargo build --release"
    exit 1
fi

print_status "Found binary: $BINARY"

# Create user if it doesn't exist
if ! id "$SERVICE_USER" &>/dev/null; then
    useradd --system --no-create-home --shell /usr/sbin/nologin "$SERVICE_USER"
    print_status "Created system user: $SERVICE_USER"
else
    print_status "User already exists: $SERVICE_USER"
fi

# Create directories
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$DATA_DIR/$NODE_ID"
mkdir -p "$LOG_DIR"

print_status "Created directories"

# Copy binary
cp "$BINARY" "$INSTALL_DIR/wardenclyffescale"
chmod 755 "$INSTALL_DIR/wardenclyffescale"
print_status "Installed binary to $INSTALL_DIR/wardenclyffescale"

# Create or copy configuration
if [[ -n "$CONFIG_FILE" && -f "$CONFIG_FILE" ]]; then
    cp "$CONFIG_FILE" "$CONFIG_DIR/wardenclyffescale.toml"
    print_status "Copied configuration from $CONFIG_FILE"
elif [[ -f "$CONFIG_DIR/wardenclyffescale.toml" ]]; then
    print_status "Using existing configuration"
else
    # Generate default configuration
    cat > "$CONFIG_DIR/wardenclyffescale.toml" << EOF
# WardenClyffeScale Configuration
# Generated by install-service.sh

[node]
id = "${NODE_ID}"
bind_address = "0.0.0.0:7654"
data_dir = "${DATA_DIR}/${NODE_ID}"

[database]
host = "localhost"
port = 3306
user = "wardenclyffescale"
password = "changeme"
database = "myapp"
pool_size = 10
connect_timeout_secs = 30

[wal]
batch_size = 1000
flush_interval_ms = 100
compression = true
segment_size_mb = 64
retention_hours = 168
fsync = true

[cluster]
peers = []
heartbeat_interval_ms = 500
election_timeout_ms = 2000
max_batch_entries = 1000

[api]
enabled = true
bind_address = "0.0.0.0:8080"
cors_enabled = false

[logging]
level = "info"
format = "pretty"
file = "${LOG_DIR}/wardenclyffescale.log"
EOF
    print_status "Created default configuration"
    print_warning "Edit $CONFIG_DIR/wardenclyffescale.toml to configure database connection"
fi

# Set ownership
chown -R "$SERVICE_USER:$SERVICE_USER" "$DATA_DIR"
chown -R "$SERVICE_USER:$SERVICE_USER" "$LOG_DIR"
chown root:root "$CONFIG_DIR/wardenclyffescale.toml"
chmod 640 "$CONFIG_DIR/wardenclyffescale.toml"

print_status "Set file permissions"

# Create systemd service file
EXEC_START="$INSTALL_DIR/wardenclyffescale --config $CONFIG_DIR/wardenclyffescale.toml --log-level info start"
if [[ -n "$BOOTSTRAP" ]]; then
    EXEC_START="$EXEC_START --bootstrap"
fi

cat > /etc/systemd/system/wardenclyffescale.service << EOF
[Unit]
Description=WardenClyffeScale - Distributed MariaDB Synchronization Manager
Documentation=https://github.com/shippinscale/wardenclyffescale
After=network.target mariadb.service
Wants=mariadb.service

[Service]
Type=simple
User=${SERVICE_USER}
Group=${SERVICE_USER}
ExecStart=${EXEC_START}
ExecReload=/bin/kill -HUP \$MAINPID
Restart=on-failure
RestartSec=5
TimeoutStopSec=30

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
ReadWritePaths=${DATA_DIR} ${LOG_DIR}
ReadOnlyPaths=${CONFIG_DIR}

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=wardenclyffescale

# Resource limits
LimitNOFILE=65535
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF

print_status "Created systemd service file"

# Reload systemd
systemctl daemon-reload
print_status "Reloaded systemd"

# Enable service
systemctl enable wardenclyffescale
print_status "Enabled wardenclyffescale service"

echo ""
echo "=========================================="
echo "  Installation Complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo ""
echo "1. Edit the configuration file:"
echo "   sudo nano $CONFIG_DIR/wardenclyffescale.toml"
echo ""
echo "2. Set up the MariaDB user and database:"
echo "   mysql -u root -p"
echo "   CREATE USER 'wardenclyffescale'@'localhost' IDENTIFIED BY 'your-password';"
echo "   GRANT ALL PRIVILEGES ON myapp.* TO 'wardenclyffescale'@'localhost';"
echo "   FLUSH PRIVILEGES;"
echo ""
echo "3. Start the service:"
echo "   sudo systemctl start wardenclyffescale"
echo ""
echo "4. Check the status:"
echo "   sudo systemctl status wardenclyffescale"
echo "   sudo journalctl -u wardenclyffescale -f"
echo ""
echo "5. Access the API:"
echo "   curl http://localhost:8080/health"
echo ""
