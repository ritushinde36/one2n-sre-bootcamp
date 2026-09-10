#!/bin/bash
# Provisions the Debian Vagrant VM (see ../Vagrantfile): installs Docker +
# the `docker compose` CLI plugin, then brings up the proxy stack so the
# API is reachable on the forwarded port (8080).
set -euo pipefail

apt-get update
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
if [ -f env ] && [ ! -f .env ]; then
	mv env .env
fi

echo "==> Deploying the proxy stack (mysql/migrate/api-1/api-2/nginx)..."
make compose-proxy-up
