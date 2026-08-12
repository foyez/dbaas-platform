CONTAINER_TOOL ?= podman
IMG ?= registry.onstackit.cloud/dbaas-platform/dbaas-frontend:latest

install:
	npm install

dev:
	npm run dev

build:
	npm run build

test-unit:
	npm run test:unit

test-e2e:
	npm run test:e2e

lint:
	npm run lint

format:
	npm run format

docker-build:
	$(CONTAINER_TOOL) build \
		--platform linux/amd64 \
		-t ${IMG} \
		--build-arg VITE_ZITADEL_AUTHORITY="$(VITE_ZITADEL_AUTHORITY)" \
		--build-arg VITE_ZITADEL_CLIENT_ID="$(VITE_ZITADEL_CLIENT_ID)" \
		--build-arg VITE_ZITADEL_REDIRECT_URI="$(VITE_ZITADEL_REDIRECT_URI)" \
		--build-arg VITE_ZITADEL_POST_LOGOUT_URI="$(VITE_ZITADEL_POST_LOGOUT_URI)" \
		--build-arg VITE_API_URL="$(VITE_API_URL)" \
		--target runtime \
		-f Containerfile .

docker-push:
	$(CONTAINER_TOOL) push ${IMG}

docker-run:
	$(CONTAINER_TOOL) run \
		--rm -p 3000:3000 \
		${IMG}


.PHONY: install dev build test-unit test-e2e lint format docker-build docker-push
