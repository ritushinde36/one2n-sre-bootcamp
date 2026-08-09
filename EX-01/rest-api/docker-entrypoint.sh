#!/bin/sh
set -e

# Non-secret settings delivered via Compose configs (see docker-compose.yml)
# land here as a plain file, not env vars - load them into the environment
# for the processes below.
if [ -f /run/configs/app.env ]; then
  set -a
  . /run/configs/app.env
  set +a
fi

migrate up
exec rest-api
