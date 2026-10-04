# CI/CD

← [Back to README](../README.md)

This page documents the CI/CD pipeline. It covers when the pipeline runs, what each job does, where it publishes the image, and how the new image gets deployed.

[.github/workflows/build-test-push.yml](../../../.github/workflows/build-test-push.yml) builds, tests, lints, and publishes a Docker image. Then it sets the new image tag in the `student-api` Helm chart. It runs in two ways:

- Automatically, on every push to `feature/ci` or `main` that changes a file under `EX-01/rest-api/`. A push that changes only files under `EX-01/rest-api/helm/` does not run it, because a chart change needs no new image.
- Manually, using "Run workflow" in the Actions tab.

**Job 1, `build-test-push`, in order:**

1. Check out the code, with full history and tags.
2. Determine the image version, from the current git tag.
3. Build the app.
4. Run the test suite. This starts a real MySQL container.
5. Install lint tools, then run static analysis and formatting checks.
6. Log in to the container registry.
7. Build and publish the Docker image.

**Job 2, `update-image-tag`, runs only when job 1 succeeds:**

1. Check out the latest commit of the same branch.
2. Set `image.tag` in [helm/student-api/values.yaml](../helm/student-api/values.yaml) to the new version. If it already has that value, the job stops here.
3. Commit the change as `github-actions[bot]`, and push it to the same branch.

That commit changes only a file under `helm/`, so it does not run the pipeline again. Argo CD watches the branch, and deploys the new tag. See [Argo CD](argocd.md).

**Deploy key:** job 2 pushes with a deploy key, not with the default token. The private key is in the Actions secret `DEPLOY_KEY`. `main` accepts changes only through reviewed pull requests, and the ruleset on `main` lets deploy keys bypass that rule.

**Image publishing:** the built image is pushed to [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry) (GHCR), under this repo's owner, as `ghcr.io/<owner>/student-rest-api:<version>`.

**Runner:** this pipeline runs on a [self-hosted runner](https://docs.github.com/en/actions/hosting-your-own-runners/managing-self-hosted-runners/adding-self-hosted-runners), not GitHub-hosted infrastructure.

Note: the test step needs a working Docker daemon to launch the MySQL container. Docker (Desktop) must be running on the runner machine before triggering the pipeline.
