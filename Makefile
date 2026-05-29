API_IMAGE ?= ghcr.io/adiom-data/hello-app-fork-api
WEB_IMAGE ?= ghcr.io/adiom-data/hello-app-fork-web
TAG ?= latest
PLATFORMS ?= linux/amd64
ENV ?= dev
DEPLOY_ARTIFACT ?= oci://ghcr.io/adiom-data/hello-app-fork-deploy
GHCR_OWNER ?= adiom-data
GHCR_API_PACKAGE ?= hello-app-fork-api
GHCR_WEB_PACKAGE ?= hello-app-fork-web
GHCR_DEPLOY_PACKAGE ?= hello-app-fork-deploy
SOURCE ?= $(shell git config --get remote.origin.url 2>/dev/null || pwd)
REVISION ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo local)
KUSTOMIZE ?= kubectl kustomize
DEPLOY_PATH ?= ./deploy

.PHONY: build test web-build docker-build-api docker-build-web docker-push-api docker-push-web docker-buildx-push render flux-push ghcr-api-public ghcr-web-public ghcr-deploy-public ghcr-public flux-push-public

build:
	go build -o /tmp/hello-app-api ./api
	cd web && npm run build

test:
	go test ./...

web-build:
	cd web && npm run build

docker-build-api:
	docker build -f Dockerfile.api -t $(API_IMAGE):$(TAG) .

docker-build-web:
	docker build -f Dockerfile.web -t $(WEB_IMAGE):$(TAG) .

docker-push-api:
	docker push $(API_IMAGE):$(TAG)

docker-push-web:
	docker push $(WEB_IMAGE):$(TAG)

docker-buildx-push:
	docker buildx build --platform $(PLATFORMS) -f Dockerfile.api -t $(API_IMAGE):$(TAG) --push .
	docker buildx build --platform $(PLATFORMS) -f Dockerfile.web -t $(WEB_IMAGE):$(TAG) --push .

render:
	$(KUSTOMIZE) deploy/overlays/$(ENV)

flux-push:
	flux push artifact $(DEPLOY_ARTIFACT):$(TAG) \
		--path=$(DEPLOY_PATH) \
		--source="$(SOURCE)" \
		--revision="$(REVISION)"

ghcr-api-public:
	gh api \
		--method PATCH \
		-H "Accept: application/vnd.github+json" \
		/orgs/$(GHCR_OWNER)/packages/container/$(GHCR_API_PACKAGE)/visibility \
		-f visibility=public

ghcr-web-public:
	gh api \
		--method PATCH \
		-H "Accept: application/vnd.github+json" \
		/orgs/$(GHCR_OWNER)/packages/container/$(GHCR_WEB_PACKAGE)/visibility \
		-f visibility=public

ghcr-deploy-public:
	gh api \
		--method PATCH \
		-H "Accept: application/vnd.github+json" \
		/orgs/$(GHCR_OWNER)/packages/container/$(GHCR_DEPLOY_PACKAGE)/visibility \
		-f visibility=public

ghcr-public: ghcr-api-public ghcr-web-public ghcr-deploy-public

flux-push-public: flux-push ghcr-deploy-public
