# Application name
APP_NAME=platform

# Docker/Podman image name
IMG ?= dbaas-api:latest
CONTAINER_TOOL ?= podman
IMAGE_NAME=$(APP_NAME):dev


.PHONY: help build run dev test clean


help:
	@echo "Available commands:"
	@echo "  make build   - Build production container image"
	@echo "  make dev     - Run development container with hot reload"
	@echo "  make run     - Run application locally"
	@echo "  make test    - Run Go tests"
	@echo "  make clean   - Remove temporary files"


# Build production image using final stage
build:
	${CONTAINER_TOOL} build \
		-t ${IMG} \
		--target prod \
		-f Containerfile .


# Push production image
push:
	${CONTAINER_TOOL} push ${IMG}


# Run development container
# Mount source code so Air can detect changes
dev:
	${CONTAINER_TOOL} build \
		--platform linux/arm64 \
		-t $(IMAGE_NAME) \
		--target dev \
		-f Containerfile .

	exec ${CONTAINER_TOOL} run \
		--rm \
		-p 8080:8080 \
		-v $$(pwd):/app:Z \
		-v $$HOME/.kube:/root/.kube:Z \
		-v go-cache:/go/pkg/mod \
		$(IMAGE_NAME)


# Run Go application directly without container
run:
	go run ./cmd/server


# Execute unit tests
# v: verbose
# -count=1 - disable test result caching
#  go test -v pkg_path -run test_name
test:
	go test -count=1 -v ./...


# Remove generated files
clean:
	rm -rf tmp
