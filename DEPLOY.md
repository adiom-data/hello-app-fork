# Deploy Artifact

This repo builds three artifacts:

- A Go API container image.
- A static React frontend container image.
- A Flux-compatible OCI artifact containing dev/prod Kubernetes manifests in `deploy/`.

Set the registry locations first:

```sh
export API_IMAGE=ghcr.io/adiom-data/hello-app-api
export WEB_IMAGE=ghcr.io/adiom-data/hello-app-web
export DEPLOY_ARTIFACT=oci://ghcr.io/adiom-data/hello-app-deploy
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
- pushes `hello-app-api` and `hello-app-web` images to GHCR with `latest`
  and `sha-<commit>` tags;
- pushes the `hello-app-deploy` Flux OCI bundle with `latest` and
  `sha-<commit>` tags;
- marks all three GHCR packages public.

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

The deploy bundle contains overlays for the `dev` and `prod` namespaces. It
expects the platform tenant registry to create the namespace and reconcile the
bundle with tenant-scoped permissions, following the pattern in `../mtest`.

The public route is exposed through the platform Gateway:

- `dev.hello-app.infrapad.local`
- `prod.hello-app.infrapad.local`

The `HTTPRoute` sends `/api` to the Go API service and all other paths to the
frontend service.
