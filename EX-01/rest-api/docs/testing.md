# Testing & Static Analysis

← [Back to README](../README.md)

This page explains how to test this project. It covers the test suite, coverage reports, and static analysis tools.

## Testing

```bash
make test
# or
go test ./... -v
```

**Docker must be running**. `TestMain`, in [controllers_test/main_test.go](../controllers_test/main_test.go), starts one real MySQL 8 container, using Testcontainers. Every test in the package shares this container. `TestMain` removes it after the suite finishes.

### Coverage

```bash
make test-coverage
```

Runs the suite with `-coverprofile`, then opens an HTML report in your browser. The report shows exactly which lines a test exercises, and which it does not. Run this locally, any time you want to check.

This uses `-coverpkg=./controllers/...`, rather than plain `go test -cover ./...`. The tests live in the separate `controllers_test` package, not in `controllers` itself.

Test files, by what they cover:

| File | Covers |
|---|---|
| [helpers_test.go](../controllers_test/helpers_test.go) | Shared router setup and HTTP request helper used by every other test file |
| [StudentController_test.go](../controllers_test/StudentController_test.go) | Malformed or invalid JSON bodies for Create and Update (unclosed JSON, wrong types, non-object payloads) |
| [student_endpoints_test.go](../controllers_test/student_endpoints_test.go) | Per-endpoint behavior: health and readiness checks, listing, not-found (404), duplicate email (409), missing required fields, rejection of unknown fields |
| [student_lifecycle_test.go](../controllers_test/student_lifecycle_test.go) | Full create → read → update → read → delete → read flow against the real database |

If you do not know a test's exact name, list every runnable test first (no Docker required):

```bash
make test-list
```

Then run a single test by name:

```bash
make test-one TEST=TestStudentLifecycle_PersistsAcrossRealDB
# or directly
go test ./controllers_test/... -run TestStudentLifecycle_PersistsAcrossRealDB -v
```

## Static Analysis

### Staticcheck

[Staticcheck](https://staticcheck.dev/) catches issues `go vet` misses, like unused code, suspicious type conversions, deprecated API usage, and simplifiable expressions.

**Install it once:**

```bash
./scripts/install-prerequisites.sh
# or directly
go install honnef.co/go/tools/cmd/staticcheck@latest
```

**Run it:**

```bash
make staticcheck
# or directly
staticcheck ./...
```

### Dockerfile Linting

[hadolint](https://github.com/hadolint/hadolint) catches Dockerfile issues, like missed layer-consolidation opportunities, unpinned base images, and other common anti-patterns.

**Install it once:**

```bash
./scripts/install-prerequisites.sh   # macOS
# or directly
brew install hadolint
```

See [hadolint's install docs](https://github.com/hadolint/hadolint#install) for other platforms.

**Run it:**

```bash
make hadolint
# or directly
hadolint Dockerfile
```

Staticcheck and hadolint both also run automatically in CI on every push. See [CI/CD](ci-cd.md). Failures surface there even if you skip running them locally.
