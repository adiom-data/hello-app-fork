# Deploy Artifact

This repo builds three artifacts:

- A Go API container image.
- A static React frontend container image.
- A Flux-compatible OCI artifact containing dev/prod Kubernetes manifests in `deploy/`.

Set the registry locations first:

```sh
export API_IMAGE=ghcr.io/adiom-data/hello-app-fork-api
export WEB_IMAGE=ghcr.io/adiom-data/hello-app-fork-web
export DEPLOY_ARTIFACT=oci://ghcr.io/adiom-data/hello-app-fork-deploy
export TAG=latest
```

Build and push both images:

```sh
make docker-build-api docker-build-web
make docker-push-api docker-push-web
```

Or build and push multi-platform images:

```sh
make docker-buildx-push
```

Publish the deploy bundle as an OCI artifact for Flux:

```sh
make flux-push
```

## GitHub Actions

Pushes to `main` run `.github/workflows/publish.yml`, which:

- builds and verifies the app;
- renders both Kustomize overlays;
- pushes `hello-app-fork-api` and `hello-app-fork-web` images to GHCR with `latest`
  and `sha-<commit>` tags;
- pushes the `hello-app-fork-deploy` Flux OCI bundle with `latest` and
  `sha-<commit>` tags;
- marks all three GHCR packages public.

The CI-published deploy bundle pins the API and web image tags to the same
`sha-<commit>` tag it just built. That makes the rendered Deployment pod
templates change on each publish, so Kubernetes rolls the API and frontend pods
instead of continuing to run an older image behind the `latest` tag.

The workflow uses `GITHUB_TOKEN` with `packages: write`. If the organization
blocks workflow-managed package visibility changes, run the same public target
locally with a token that has package admin rights:

```sh
make ghcr-public
```

Render an environment locally:

```sh
make render ENV=dev
make render ENV=prod
```

The deploy bundle contains overlays for `dev` and `prod`. Both overlays deploy
to the app namespace `hello-app-fork` inside their target cluster. The current active
environment is prod; the dev overlay is kept for a future dev cluster and is not
expected to be reconciled in the prod cluster.

The platform tenant registry is expected to create the target namespace in each
cluster and reconcile the bundle with tenant-scoped permissions, following the
pattern in `../mtest`.

The bundle also requests a tenant-local CloudNativePG database named
`hello-app-fork-db`.
See `deploy/DATABASE.md` for the generated service and secret names used by the
API.

The public route attaches to the platform Gateway without declaring
`spec.hostnames`, so it can serve any hostname accepted by the parent Gateway
listener for the tenant.

The `HTTPRoute` sends `/api` to the Go API service and all other paths to the
frontend service.
