CONTAINER_TOOL ?= podman
IMG ?= frontend:latest

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
	$(CONTAINER_TOOL) build -t ${IMG} \
		--target runtime \
		-f Containerfile .

docker-push:
	$(CONTAINER_TOOL) push ${IMG}

docker-run:
	$(CONTAINER_TOOL) run \
		--rm -p 3000:3000 \
		${IMG}


.PHONY: install dev build test-unit test-e2e lint format docker-build docker-push
