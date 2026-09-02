# CI/CD

← [Back to README](../README.md)

This page documents the CI/CD pipeline. It covers when the pipeline runs, what each step does, and where it publishes the image.

[.github/workflows/build-test-push.yml](../../../.github/workflows/build-test-push.yml) builds, tests, lints, and publishes a Docker image. It runs in two ways:

- Automatically, on every push to `feature/ci` or `main` that changes a file under `EX-01/rest-api/`.
- Manually, using "Run workflow" in the Actions tab.

**Pipeline steps, in order:**

1. Check out the code, with full history and tags.
2. Determine the image version, from the current git tag.
3. Build the app.
4. Run the test suite. This starts a real MySQL container.
5. Install lint tools, then run static analysis and formatting checks.
6. Log in to the container registry.
7. Build and publish the Docker image.

**Image publishing:** the built image is pushed to [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry) (GHCR), under this repo's owner, as `ghcr.io/<owner>/student-rest-api:<version>`.

**Runner:** this pipeline runs on a [self-hosted runner](https://docs.github.com/en/actions/hosting-your-own-runners/managing-self-hosted-runners/adding-self-hosted-runners), not GitHub-hosted infrastructure.

Note: the test step needs a working Docker daemon to launch the MySQL container. Docker (Desktop) must be running on the runner machine before triggering the pipeline.
