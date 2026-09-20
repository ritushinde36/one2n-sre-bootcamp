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

See [Architecture](architecture.md) for how the API talks to MySQL.

## Files

These files define the VM and the containers inside it. The table lists them in the order that `vagrant up` uses them.

| File | What it does |
| --- | --- |
| [Vagrantfile](../Vagrantfile) | Defines the VM and forwards host port 8080 to it. |
| [scripts/provision-vm.sh](../scripts/provision-vm.sh) | Installs Docker inside the VM and starts the container stack. |
| [Makefile](../Makefile) | Holds the `vagrant-*` targets and the `compose-proxy-up` target that provisioning calls. |
| [docker-compose.proxy.yml](../docker-compose.proxy.yml) | Defines the five containers: `mysql`, `migrate`, `api-1`, `api-2`, and `nginx`. |
| [Dockerfile](../Dockerfile) | Builds the application image that the migrate and API containers use. |
| [nginx/default.conf](../nginx/default.conf) | Configures nginx to listen on port 8080 and load-balance the two API containers. |
| [postman/vagrant.postman_environment.json](../postman/vagrant.postman_environment.json) | Points the Postman collection at port 8080, instead of the collection's own port 8888. |
| `logs/provision-vm-*.log` | Holds the output of each provisioning run, on the host. |

## Running the VM

**1. Install prerequisites.** Run the script that installs every dependency:

```bash
./scripts/install-prerequisites.sh
```

See [Prerequisites](prerequisites.md) for details.

**2. Configure the environment.** Copy the example env file:

```bash
cp .env.example .env
```

Keep `DSN`'s host as `student-mysql`.

**3. Start the VM:**

```bash
make vagrant-up
```

The [Vagrantfile](../Vagrantfile) shares the project folder into the VM automatically.

**4. Verify.** From your host, once `make vagrant-up` finishes:

```bash
curl http://localhost:8080/healthcheck
```

To test every endpoint through nginx, see [Running against the Vagrant VM](postman.md#running-against-the-vagrant-vm) in the Postman docs.

**Other handy commands:**

```bash
make vagrant-status      # show whether the VM is running
make vagrant-ssh         # connect to the VM, for debugging
make vagrant-provision   # re-run scripts/provision-vm.sh, without recreating the VM
```

**Cleanup:**

```bash
make vagrant-halt      # stop the VM, keep it for next time
make vagrant-destroy   # remove the VM entirely
```

## Reading nginx logs

nginx writes its access log and its error log to two places:

- **stdout and stderr**, which `docker logs` reads. These logs disappear with the container.
- **The `student-nginx-logs` volume**, mounted at `/var/log/nginx-persist`. These logs stay after the container is removed.

The access log is JSON, in the `upstreamlog` format that [nginx/default.conf](../nginx/default.conf) defines.

Both places are inside the VM, so connect to it first:

```bash
make vagrant-ssh
cd /vagrant
```

**Read the live logs** of the running container:

```bash
docker compose -f docker-compose.proxy.yml logs -f nginx
```

**Read the volume** with a disposable container. This works even after `make compose-proxy-down` removes the nginx container:

```bash
docker run --rm -v student-api-proxy_student-nginx-logs:/logs alpine cat /logs/access.log
docker run --rm -v student-api-proxy_student-nginx-logs:/logs alpine cat /logs/error.log
```

**Read the files in the running container** instead:

```bash
docker exec student-nginx cat /var/log/nginx-persist/access.log
```

Compose adds its project name, `student-api-proxy_`, in front of every volume name. The other containers in the stack use the same pattern:

| Container | Volume | Path inside the container |
| --- | --- | --- |
| `nginx` | `student-api-proxy_student-nginx-logs` | `/var/log/nginx-persist` |
| `api-1` | `student-api-proxy_student-api-1-logs` | `/var/log/app` |
| `api-2` | `student-api-proxy_student-api-2-logs` | `/var/log/app` |
| `migrate` | `student-api-proxy_student-migrate-logs` | `/var/log/app` |

Note: these volumes are inside the VM. `make vagrant-halt` keeps them, but `make vagrant-destroy` deletes them with the VM disk. Copy any log you need to `/vagrant` first, because that folder is on the host.

See [Troubleshooting](troubleshooting.md) for common issues.
