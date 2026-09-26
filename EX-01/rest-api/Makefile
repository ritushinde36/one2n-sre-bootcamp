.PHONY: deps build run test test-coverage test-list test-one vet fmt staticcheck hadolint newman newman-vagrant migrate-up migrate-down migrate-status docker-build docker-migrate docker-run docker-down docker-network docker-mysql-up docker-mysql-down compose-up compose-migrate compose-down compose-proxy-up compose-proxy-migrate compose-proxy-down vagrant-up vagrant-provision vagrant-status vagrant-ssh vagrant-halt vagrant-destroy k8s-cluster-up k8s-deploy k8s-status k8s-port-forward k8s-logs k8s-down k8s-cluster-down


IMAGE_NAME := student-rest-api
# defaults to the current git version, but can be overridden on the command
# line to build/tag a specific version instead, e.g.:
#   make docker-build VERSION=1.2.3
VERSION := $(shell git describe --tags --always --dirty)
NETWORK_NAME := student-api-net
APP_CONTAINER := student-rest-api
MYSQL_CONTAINER := student-mysql
# override on the command line to run extra instances, e.g.:
#   make docker-run APP_CONTAINER=student-rest-api02 HOST_PORT=8882
HOST_PORT := 8888
# override on the command line to run down/status instead of up, e.g.:
#   make docker-migrate MIGRATE_CMD=status
MIGRATE_CMD := up
# host port mysql's 3306 is published on - override if 3306 is already taken
# on your machine, e.g.: make docker-mysql-up MYSQL_PORT=3307
MYSQL_PORT := 3306
# namespace every k8s-* target works in - matches the manifests
K8S_NAMESPACE := student-api

# Pulls in MYSQL_ROOT_PASSWORD/MYSQL_DATABASE from the gitignored .env file,
# so no credential value ever appears in this committed Makefile. Also picks
# up MYSQL_PORT if set there, overriding the default above.
-include .env

deps:
	go mod tidy

build:
	go build -o bin/rest-api .

run: build
	./bin/rest-api

# test requires Docker to be running (spins up a real MySQL container).
test:
	go test ./... -v

# generates a coverage report and opens it in the browser.
# -coverpkg is needed because the tests live in the separate controllers_test
# package, not in controllers itself 
# requires Docker to be running (spins up a real MySQL container).
test-coverage:
	go test ./controllers_test/... -coverpkg=./controllers/... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# lists all runnable test names (no Docker required) - use these with test-one.
test-list:
	@grep -rhoE '^func (Test[A-Za-z0-9_]+)' controllers_test/*.go | sed -E 's/^func //' | grep -v '^TestMain$$'

# runs a single test by name, e.g.: make test-one TEST=TestStudentLifecycle_PersistsAcrossRealDB
test-one:
	go test ./controllers_test/... -run $(TEST) -v

vet:
	go vet ./...

# fails (non-zero exit) and prints filenames if anything isn't gofmt-formatted.
fmt:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "not gofmt-formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi


# requires staticcheck - run ./scripts/install-prerequisites.sh (macOS), or
# see it for the manual install command.
staticcheck:
	staticcheck ./...

# lints the Dockerfile - requires hadolint: ./scripts/install-prerequisites.sh
# (macOS), or brew install hadolint / https://github.com/hadolint/hadolint#install
hadolint:
	hadolint Dockerfile

# runs the Postman collection against a running server (`make run` first).
# requires newman - run ./scripts/install-prerequisites.sh (macOS), or see it
# for the manual install command.
newman:
	newman run postman/student-api.postman_collection.json

# runs the Postman collection against the Vagrant VM (`make vagrant-up` first).
# the environment file points base_url at nginx on port 8080, instead of the
# collection's own 8888. see docs/vagrant.md.
newman-vagrant:
	newman run postman/student-api.postman_collection.json -e postman/vagrant.postman_environment.json

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-status:
	go run ./cmd/migrate status

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

# Requires .env (DSN pointing at the student-mysql container, and
# LOG_FILE for where the app writes its logs on the student-logs volume) and
# the network/mysql container to already be up - see docker-network/docker-mysql-up.
docker-run: docker-network
	docker run -d --name $(APP_CONTAINER) -p $(HOST_PORT):8888 --network $(NETWORK_NAME) --env-file .env -v student-logs:/var/log/app $(IMAGE_NAME):$(VERSION)

# Runs migrations to completion in a one-off container (--rm), separately
# from the app container - like a Kubernetes init container, but for plain
# Docker. Run this yourself, once, before scaling up any number of docker-run
# instances - it's deliberately NOT a docker-run dependency, so that scaling
# to N app containers doesn't mean N containers each redundantly trying to
# migrate. --rm removes the container on exit, but not the
# student-migrate-logs volume, so MIGRATE_LOG_FILE's output survives.
docker-migrate: docker-network
	docker run --rm --network $(NETWORK_NAME) --env-file .env --entrypoint migrate -v student-migrate-logs:/var/log/app $(IMAGE_NAME):$(VERSION) $(MIGRATE_CMD)


# sends SIGTERM so the app's own graceful shutdown gets a chance to drain
# in-flight requests, then removes the stopped container. Logs aren't lost
# in the process - the app writes them to the student-logs volume (see
# docker-run), which outlives the container.
docker-down:
	docker stop $(APP_CONTAINER)
	docker rm $(APP_CONTAINER)

docker-network:
	docker network inspect $(NETWORK_NAME) >/dev/null 2>&1 || docker network create $(NETWORK_NAME)

docker-mysql-up: docker-network
	docker run -d --name $(MYSQL_CONTAINER) \
		--network $(NETWORK_NAME) \
		-v student-mysql-data:/var/lib/mysql \
		-p $(MYSQL_PORT):3306 \
		-e MYSQL_ROOT_PASSWORD=$(MYSQL_ROOT_PASSWORD) \
		-e MYSQL_DATABASE=$(MYSQL_DATABASE) \
		mysql:8.0

docker-mysql-down:
	docker rm -f $(MYSQL_CONTAINER)

# builds the app image and starts both containers via Compose - no manual
# network/mysql/build steps needed, unlike the docker-* targets above.
# Requires .env to already exist - copy it from .env.example first.
# IMAGE_NAME/VERSION are exported so the image gets tagged the same way
# docker-build tags it (git version), instead of Compose's default :latest.
compose-up:
	@if [ -n "$$(docker compose ps -q --status running mysql 2>/dev/null)" ]; then \
		echo "==> mysql is already running"; \
	else \
		echo "==> mysql is not running yet - compose will start it"; \
	fi
	IMAGE_NAME=$(IMAGE_NAME) VERSION=$(VERSION) docker compose --env-file .env up -d --build mysql migrate
	@echo "==> waiting for migrate to finish..."
	@migrate_exit="$$(docker wait student-migrate)"; \
	echo "==> migration status (from the migrate service's own logs):"; \
	docker compose logs migrate --no-color | tail -n 3; \
	if [ "$$migrate_exit" != "0" ]; then \
		echo "==> migrate failed (exit $$migrate_exit) - not starting rest-api" >&2; \
		exit 1; \
	fi
	IMAGE_NAME=$(IMAGE_NAME) VERSION=$(VERSION) docker compose --env-file .env up -d --no-deps rest-api

# runs a one-off migrate command via Compose (default: up), e.g.:
#   make compose-migrate MIGRATE_CMD=status
# `run` overrides the migrate service's own `command: ["up"]` for just this
# invocation, in its own container - separate from the one `compose-up`
# manages, so this is safe to run any time without disturbing that one.
compose-migrate:
	IMAGE_NAME=$(IMAGE_NAME) VERSION=$(VERSION) docker compose --env-file .env run --rm migrate $(MIGRATE_CMD)

# stops and removes both containers (the student-mysql-data volume persists).
compose-down:
	docker compose --env-file .env down

# same as compose-up, but via docker-compose.proxy.yml - builds the app image
# and starts mysql/migrate/api-1/api-2/nginx, load-balancing across the two
# api instances behind nginx on :8080.
compose-proxy-up:
	@if [ -n "$$(docker compose -f docker-compose.proxy.yml ps -q --status running mysql 2>/dev/null)" ]; then \
		echo "==> mysql is already running"; \
	else \
		echo "==> mysql is not running yet - compose will start it"; \
	fi
	IMAGE_NAME=$(IMAGE_NAME) VERSION=$(VERSION) docker compose -f docker-compose.proxy.yml --env-file .env up -d --build
	@echo "==> migration status (from the migrate service's own logs):"
	@docker compose -f docker-compose.proxy.yml logs migrate --no-color | tail -n 3

# runs a one-off migrate command via the proxy compose file (default: up),
# e.g.: make compose-proxy-migrate MIGRATE_CMD=status
compose-proxy-migrate:
	IMAGE_NAME=$(IMAGE_NAME) VERSION=$(VERSION) docker compose -f docker-compose.proxy.yml --env-file .env run --rm migrate $(MIGRATE_CMD)

# stops and removes all proxy-stack containers (the student-mysql-data volume persists).
compose-proxy-down:
	docker compose -f docker-compose.proxy.yml --env-file .env down

# creates the VM and provisions it via scripts/provision-vm.sh, which installs
# Docker and brings up the proxy stack inside the VM - so the API ends up on
# localhost:8080 through the Vagrantfile's forwarded port. Only provisions on
# first create; use vagrant-provision afterwards. Requires vagrant, and on
# Apple Silicon also UTM and the vagrant_utm plugin - see
# ./scripts/install-prerequisites.sh. See docs/vagrant.md.
vagrant-up:
	vagrant up

# re-runs scripts/provision-vm.sh against the existing VM without recreating
# it. Safe to run repeatedly: the apt source edits are idempotent, mkdir -p
# tolerates existing directories, and compose up converges rather than
# duplicating containers.
vagrant-provision:
	vagrant provision

# reports whether the VM is running, poweroff, or not created yet.
vagrant-status:
	vagrant status

# opens a shell in the VM. This directory is mounted there at /vagrant.
vagrant-ssh:
	vagrant ssh

# powers the VM off but keeps its disk, so vagrant-up restarts it with the
# containers and mysql data intact - and without re-provisioning.
vagrant-halt:
	vagrant halt

# deletes the VM and its disk, including the docker volumes inside it (mysql
# data, nginx logs). Prompts for confirmation. Provisioning logs survive -
# they're written to logs/ on the host through the synced folder.
vagrant-destroy:
	vagrant destroy

# creates the 4-node minikube cluster and labels each worker by role. The
# labels are what the manifests' nodeSelectors match on - a pod whose label
# has no matching node stays Pending. See docs/minikube.md.
k8s-cluster-up:
	minikube start --nodes 4 --driver=docker
	kubectl wait --for=condition=Ready nodes --all --timeout=120s
	kubectl label node minikube-m02 type=application --overwrite
	kubectl label node minikube-m03 type=database --overwrite
	kubectl label node minikube-m04 type=dependent_services --overwrite
	kubectl get nodes -L type

# applies the manifests. database first, because the api pod's readiness
# probe checks the database. Requires Vault and the External Secrets Operator
# to be set up already - see docs/secrets-management.md - or the ExternalSecret
# resources have no controller and the pods never get their credentials.
k8s-deploy:
	kubectl apply -f manifests/database.yml
	kubectl apply -f manifests/application.yml

# what's running, and whether the secrets actually synced.
k8s-status:
	kubectl get pods -n $(K8S_NAMESPACE) -o wide
	@echo "==> external secrets:"
	@kubectl get externalsecret -n $(K8S_NAMESPACE) 2>/dev/null || echo "none"

# opens localhost:8888 onto the service, for curl or Postman. This blocks -
# the tunnel lives only as long as the command runs, so leave it in its own
# terminal. On macOS with the docker driver the NodePort is not reachable
# from the host directly, so a tunnel is the only way in.
k8s-port-forward:
	kubectl port-forward -n $(K8S_NAMESPACE) svc/student-api 8888:8888

# tails the api logs. Use MIGRATE_LOGS=1 for the migration init container's
# output instead, e.g.: make k8s-logs MIGRATE_LOGS=1
k8s-logs:
	@if [ -n "$(MIGRATE_LOGS)" ]; then \
		kubectl logs -n $(K8S_NAMESPACE) deploy/rest-api -c migrate; \
	else \
		kubectl logs -n $(K8S_NAMESPACE) deploy/rest-api -f; \
	fi

# removes the app and the database, keeping the cluster. This deletes the
# namespace, and the mysql volume claim with it, so the data is lost.
k8s-down:
	kubectl delete -f manifests/application.yml -f manifests/database.yml

# deletes the whole cluster, Vault and its secrets included.
k8s-cluster-down:
	minikube delete
