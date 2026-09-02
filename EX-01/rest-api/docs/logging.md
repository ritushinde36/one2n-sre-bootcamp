# Logging

← [Back to README](../README.md)

This page explains how the app logs. It covers the log format, and which parts of the app write logs. Read this to understand how to read logs.

All logs, from both the app and the migration tool, use the same JSON format. Each log line is a structured object.

- The request logging middleware logs one line per HTTP request: method, path, status, latency, and client IP.

These two also send their logs through this same JSON format:

- The database layer, for its own query logs.
- The migration tool, for its own logs.

Controllers log at three levels: `info` for success, `warn` for client errors (bad input, not found), and `error` for unexpected failures (database errors).
