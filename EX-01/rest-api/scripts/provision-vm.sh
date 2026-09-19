#!/bin/bash
# Provisions the Debian Vagrant VM (see ../Vagrantfile): installs Docker +
# the `docker compose` CLI plugin, then brings up the proxy stack so the
# API is reachable on the forwarded port (8080).
set -euo pipefail

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
