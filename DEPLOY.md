# Deploy Artifacts

This repo publishes deployment artifacts through Bazel:

- API and web container images.
- An app Flux OCI bundle from `deploy/app`.
- A database Flux OCI bundle from `deploy/database`.
- A generated manifest that lists bundle name, OCI bundle reference, overlay
  path, and force semantics.

Production uses stamped platform values and `overlays/prod`:

```sh
bazel build //deploy:publish_manifest
bazel run //deploy:publish_all
```

The production publish set pushes images and Flux OCI bundles with both the
current commit tag and `latest`, skips pushes when `latest` already matches, and
writes manifest bundle references with the `release` tag.

Preview uses stamped platform values and `overlays/preview`:

```sh
bazel build //deploy:publish_preview_manifest
bazel run //deploy:publish_preview_all
```

The platform publish harness provides these stamped values at publish time:

- `{STABLE_PRODUCTION_REFERENCE_PREFIX}` for production image and artifact
  references.
- `{STABLE_GIT_COMMIT}` for the production commit tag.
- `{STABLE_PREVIEW_REFERENCE_PREFIX}` for preview image and artifact references.

Preview publishes use the default `latest` tag unless a publish invocation
overrides the tag at runtime.

The preview app overlay injects the `registry-pull` image pull secret for the API
and web workloads. The database bundle is separate from the app bundle so the
platform can reconcile those concerns independently.

The public route is exposed through the platform Gateway:

- `t-testuser.infrapad.ai`
- `app.adiom.io`

The `HTTPRoute` sends `/api` to the Go API service and all other paths to the
frontend service.
