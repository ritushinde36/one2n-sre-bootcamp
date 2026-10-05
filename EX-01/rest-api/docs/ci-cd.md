# CI/CD

← [Back to README](../README.md)

This page explains the CI/CD pipeline in [.github/workflows/build-test-push.yml](../../../.github/workflows/build-test-push.yml). The pipeline builds, tests and publishes the Docker image. Then it sets the new image tag in the `student-api` Helm chart, and Argo CD deploys it.

## When it runs

- **On push:** to `feature/ci`, `main` or `feature/helm-argocd`, when the push changes a file under `EX-01/rest-api/`.
- **Manually:** with **Run workflow** in the Actions tab.

A push that changes only files under `EX-01/rest-api/helm/` does not run the pipeline. A chart change needs no new image.

## Job 1: `build-test-push`

1. Check out the code, with full history and tags.
2. Get the image version from the current git tag.
3. Build the app.
4. Run the tests. They start a real MySQL container.
5. Install the lint tools. Then run the static analysis and formatting checks.
6. Log in to the container registry.
7. Build and push the Docker image.

## Job 2: `update-image-tag`

This job runs only when job 1 succeeds.

1. Check out the latest commit of the same branch.
2. Set `image.tag` in [helm/student-api/values.yaml](../helm/student-api/values.yaml) to the new version. If it already has that value, the job stops.
3. Commit the change as `github-actions[bot]`. Push it to the same branch.

This commit changes only a file under `helm/`, so it does not run the pipeline again. Argo CD watches the branch and deploys the new tag. See [Argo CD](argocd.md).

## Runner, image and deploy key

| Item | Details |
|---|---|
| Runner | A [self-hosted runner](https://docs.github.com/en/actions/hosting-your-own-runners/managing-self-hosted-runners/adding-self-hosted-runners), not a GitHub-hosted one. Docker must be running on it before the pipeline starts, because the tests start a MySQL container. |
| Image | Pushed to [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry) (GHCR) as `ghcr.io/<owner>/student-rest-api:<version>`. |
| Deploy key | Job 2 pushes with a deploy key, not the default token. Its private key is in the Actions secret `DEPLOY_KEY`. `main` accepts changes only through reviewed pull requests. Its ruleset lets deploy keys bypass that rule. |
