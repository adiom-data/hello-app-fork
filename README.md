# Hello App

A small sample app with a Go API and a React frontend.

## Run the API

```sh
go run ./api
```

The API listens on `http://localhost:18081`.

Endpoints:

- `GET /api/hello`
- `GET /healthz`

## Run the frontend

```sh
cd web
npm install
npm run dev
```

The frontend runs on `http://127.0.0.1:5173` and proxies `/api` requests to the Go API.

## Containers and Kubernetes

This app has two production-style containers:

- `Dockerfile.api` builds the Go API image.
- `Dockerfile.web` builds the React static bundle and serves it with nginx.

The Kubernetes manifests live in `deploy/` with overlays for the `dev` and
`prod` namespaces. The route follows the platform pattern from `../mtest`:

- `/api` routes to `hello-app-api`
- `/` routes to `hello-app-web`

See [DEPLOY.md](DEPLOY.md) for the image and Flux OCI deploy bundle commands.
