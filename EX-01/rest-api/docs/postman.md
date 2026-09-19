# Postman Collection

← [Back to README](../README.md)

This page explains the Postman collection included with this project. It covers importing it, running it manually, and running it headlessly with Newman.

A ready-to-import collection lives at [postman/student-api.postman_collection.json](../postman/student-api.postman_collection.json). It covers both happy paths (health check, create, get, update, delete) and the main failure cases (missing required fields, not-found lookups, invalid IDs).

**Import into Postman:**

1. Open Postman → **Import** → select `postman/student-api.postman_collection.json`.
2. The collection uses a `base_url` variable (defaults to `http://localhost:8888`).
3. **Run requests top-to-bottom.** "Create Student" captures the new student's ID into a `student_id` collection variable. Later requests (Get, Update, Delete by ID) reuse it. Running requests out of order, or alone, may make the ID-dependent ones fail.

**Run headlessly with Newman** (useful in CI or without the Postman GUI). Install it once with `./scripts/install-prerequisites.sh` (or directly: `npm install -g newman`), then:

```bash
make newman
# or directly
newman run postman/student-api.postman_collection.json
```

Make sure the server is running (`make run`) before running the collection, whether via the Postman GUI or Newman.

## Running against the Vagrant VM

The [Vagrant VM](vagrant.md) serves the API through nginx on port 8080. The collection sets `base_url` to port 8888, so it does not reach the VM.

Do not edit `base_url`. Use the environment file at [postman/vagrant.postman_environment.json](../postman/vagrant.postman_environment.json) instead. It sets `base_url` to `http://localhost:8080`. An environment variable overrides a collection variable with the same name, so one collection works against both setups.

**In Postman:**

1. Click **Import**. Select `postman/vagrant.postman_environment.json`.
2. Look under **Environments** in the left sidebar. The file appears there as "Vagrant VM".
3. Select **Vagrant VM** and set it as **active**
4. Run the collection.

To confirm the environment is active, hold the pointer over `{{base_url}}` in any request. The tooltip shows `http://localhost:8080`.

**With Newman:**

```bash
make newman-vagrant
# or directly
newman run postman/student-api.postman_collection.json -e postman/vagrant.postman_environment.json
```
