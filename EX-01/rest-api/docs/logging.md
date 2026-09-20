# Logging

← [Back to README](../README.md)

This page explains how the app logs. It covers the log format, and which parts of the app write logs. Read this to understand how to read logs.

All logs, from both the app and the migration tool, use the same JSON format. Each log line is a structured object.

- The request logging middleware logs one line per HTTP request: method, path, status, latency, and client IP.

These two also send their logs through this same JSON format:

- The database layer, for its own query logs.
- The migration tool, for its own logs.

Controllers log at three levels: `info` for success, `warn` for client errors (bad input, not found), and `error` for unexpected failures (database errors).

## nginx logs

The Vagrant VM puts nginx in front of two API containers. nginx writes its own logs, separate from the app's.

- The access log is JSON, with one object for each request. [nginx/default.conf](../nginx/default.conf) defines the fields.
- The error log uses nginx's own plain text format.
- nginx writes both logs to two places: stdout and stderr, and the `student-nginx-logs` volume.

See [Reading nginx logs](vagrant.md#reading-nginx-logs) for the commands that read them.

## Provisioning logs

[scripts/provision-vm.sh](../scripts/provision-vm.sh) writes the output of each VM provisioning run to its own file, `logs/provision-vm-<UTC timestamp>.log`. These files are on the host, not inside the VM. This output is plain text, not JSON.
