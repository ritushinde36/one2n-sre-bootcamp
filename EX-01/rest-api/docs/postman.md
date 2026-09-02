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
