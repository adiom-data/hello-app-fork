# Deploy Artifacts

This repo publishes deployment artifacts through Bazel:

- API and web container images.
- An app Flux OCI bundle from `deploy/app`.
- A database Flux OCI bundle from `deploy/database`.
- A generated manifest that lists bundle name, OCI bundle reference, overlay
  path, and force semantics.

Production uses the real `ghcr.io/adiom-data` registry and `overlays/prod`:

```sh
bazel build //deploy:publish_manifest
bazel run //deploy:publish_all
```

Preview uses stamped platform values and `overlays/preview`:

```sh
bazel build //deploy:publish_preview_manifest
bazel run //deploy:publish_preview_all
```

The platform publish harness provides these stamped values at publish time:

- `{STABLE_PREVIEW_REFERENCE_PREFIX}` for preview image and artifact references.
- `{STABLE_PREVIEW_TAG}` for preview image and artifact tags.

The preview app overlay injects the `registry-pull` image pull secret for the API
and web workloads. The database bundle is separate from the app bundle so the
platform can reconcile those concerns independently.

The public route is exposed through the platform Gateway:

- `t-testuser.infrapad.ai`
- `app.adiom.io`

The `HTTPRoute` sends `/api` to the Go API service and all other paths to the
frontend service.
