# Vagrant VM

← [Back to README](../README.md)

This page explains how to deploy the app on a bare metal inside a Vagrant-managed VM. See [Architecture](#architecture) below for what runs where.

Note: **this page is for Apple Silicon (M-series) Macs.** The [Vagrantfile](../Vagrantfile) here is UTM-specific.

## Architecture

```
                Vagrant Box
    ┌─────────────────────────────────────────────┐
    │                                             │
    │  nginx :8080                                │
    │  (load-balances api-1 and api-2)            │
    │  │                                          │
    │  ├────────► api-1 :8081 ──┐                 │
    │  │                        ├──► mysql :3306  │
    │  └────────► api-2 :8082 ──┘                 │
    │                                             │
    └─────────────────────────────────────────────┘
              ▲
              │
    http request (host :8080 → guest :8080)
```

- **nginx** is the only container reachable from the host, on port 8080 (see the forwarded port in the [Vagrantfile](../Vagrantfile)).
- It load-balances across the two `rest-api` containers, `api-1` and `api-2`.
- Both containers talk to the same shared `mysql` container.
- A one-shot `migrate` container applies schema migrations before either API container starts.

See [docker-compose.proxy.yml](../docker-compose.proxy.yml) and [nginx/default.conf](../nginx/default.conf) for the exact definitions, and [Architecture](architecture.md) for how the API talks to MySQL.

## Running the VM

**1. Install prerequisites.** You need Vagrant, UTM, and the `vagrant_utm` plugin. Run `./scripts/install-prerequisites.sh` to install all three on Apple Silicon. See [Prerequisites](prerequisites.md) for details.

**2. Configure the environment.** Copy the example env file, same as for [Local Setup](setup.md) or [Docker & Docker Compose](docker.md):

```bash
cp .env.example env
```

Note the filename: `env`, not `.env`. UTM's Shared Directory doesn't expose dotfiles to the guest, so the file has to be renamed to be visible inside the VM. `scripts/provision-vm.sh` renames it back to `.env` once it's inside the VM, before running Compose.

For this VM-hosted deployment, keep `DSN`'s host as `student-mysql` (the containers all run inside the VM, on the same Docker network - this matches the Docker workflow, not the local one).

**3. Bring up the VM:**

```bash
vagrant up
```

This downloads the `utm/debian11` box (first run only) and provisions it via [scripts/provision-vm.sh](../scripts/provision-vm.sh), which:

1. Installs Docker and the `docker compose` CLI plugin.
2. Renames `env` back to `.env`.
3. Runs `make compose-proxy-up` to build the app image and start `mysql`, `migrate`, `api-1`, `api-2`, and `nginx`.

The [Vagrantfile](../Vagrantfile) shares the project folder into the VM automatically. You do not need to mount it by hand. During boot, UTM may show permission pop-ups. Approve them, or the VM can get stuck.

**4. Verify.** From your host, once `vagrant up` finishes:

```bash
curl http://localhost:8080/healthcheck
```

Run the [Postman collection](postman.md) against `http://localhost:8080` as `base_url` to confirm every endpoint returns 200 through nginx.

**Other handy commands:**

```bash
vagrant ssh          # connect to the VM, for debugging
vagrant provision    # re-run scripts/provision-vm.sh, without recreating the VM
```

**Cleanup:**

```bash
vagrant halt      # stop the VM, keep it for next time
vagrant destroy   # remove the VM entirely
```

See [Troubleshooting](troubleshooting.md) for common issues.
