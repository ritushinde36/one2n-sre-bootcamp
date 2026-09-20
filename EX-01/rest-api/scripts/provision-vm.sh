#!/bin/bash
# Provisions the Debian Vagrant VM (see ../Vagrantfile): installs Docker +
# the `docker compose` CLI plugin, then brings up the proxy stack so the
# API is reachable on the forwarded port (8080).
set -euo pipefail

# Vagrant only streams provisioner output to the terminal running `vagrant up`,
# so it is lost once that scrollback is gone. Send a copy to a timestamped file
# under the synced folder instead: /vagrant is the host's rest-api directory, so
# the log lands on the host, survives `vagrant destroy`, and is already covered
# by logs/ in .gitignore. `exec` with only redirections rebinds this script's
# own stdout/stderr for the rest of the run, and tee keeps the output streaming
# to Vagrant as well. Timestamps are UTC (the VM's clock), not host local time.
LOG_DIR=/vagrant/logs
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/provision-vm-$(date -u +%Y%m%d-%H%M%S).log"
exec > >(tee "$LOG_FILE") 2>&1
echo "==> Provisioning started at $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "==> Logging to $LOG_FILE (host: EX-01/rest-api/logs/)"

# The utm/debian11 box runs bullseye (Debian 11), which is now EOL. Once a
# release goes EOL, deb.debian.org stops reliably serving it - point apt at
# archive.debian.org instead, which permanently freezes its packages in
# place. bullseye-security isn't archived there yet, and its live mirror at
# security.debian.org has been serving a package index out of sync with its
# own pool (causing 404s on install) - drop that source entirely. This VM
# doesn't need security-patched versions, just working ones from bullseye/updates.
sed -i -e 's|deb\.debian\.org|archive.debian.org|g' /etc/apt/sources.list
sed -i '/security\.debian\.org/d' /etc/apt/sources.list

# archive.debian.org's own Release files are also past their Valid-Until
# date (frozen in time), so this check still needs skipping.
apt-get -o Acquire::Check-Valid-Until=false update
apt-get install -y curl git make docker.io

systemctl enable --now docker
usermod -aG docker vagrant

# Debian 11's docker.io package doesn't ship the `docker compose` plugin
# (that's what the Makefile uses, not the standalone docker-compose binary),
# so install it as a CLI plugin per Docker's docs.
COMPOSE_VERSION="v2.29.7"
mkdir -p /usr/local/lib/docker/cli-plugins
curl -SL "https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-linux-$(uname -m)" \
	-o /usr/local/lib/docker/cli-plugins/docker-compose
chmod +x /usr/local/lib/docker/cli-plugins/docker-compose

docker --version
docker compose version

cd /vagrant

echo "==> Deploying the proxy stack (mysql/migrate/api-1/api-2/nginx)..."
make compose-proxy-up
