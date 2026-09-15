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

**Manual UTM step:** the first time, UTM needs the project folder mounted as a Shared Directory with share mode set to `virtFS` - do this in the UTM app once the VM appears, before provisioning can reach `/vagrant`. Also watch for UTM permission pop-ups during boot; missing one can stall the VM.

**4. Verify.** From your host, once `vagrant up` finishes:

```bash
curl http://localhost:8080/healthcheck
```

Run the [Postman collection](postman.md) against `http://localhost:8080` as `base_url` to confirm every endpoint returns 200 through nginx.

**5. SSH into the VM:**

```bash
vagrant ssh
```

**Cleanup:**

```bash
vagrant halt      # stop the VM, keep it for next time
vagrant destroy   # remove the VM entirely
```

`vagrant destroy` removes the VM but not the UTM Shared Directory setting - re-mounting it is still a manual step the next time you `vagrant up` a fresh VM.

## Troubleshooting

| Issue | Solution |
|---|---|
| Port 8080 already in use on the host | Stop whatever else is bound to it; `vagrant up` won't fail loudly, but nginx won't be reachable. |
| Files missing inside the VM / `/vagrant` looks empty | The UTM Shared Directory wasn't mounted, or wasn't set to `virtFS` mode. |
| `.env` not found errors from Compose | Confirm you copied to `env` (no dot) before `vagrant up`, and that provisioning actually ran (check for the "Deploying the proxy stack" line in the `vagrant up` output). |
| Download interrupted / box stuck | A flaky connection during the box download usually means restarting `vagrant up` from scratch rather than resuming. |
