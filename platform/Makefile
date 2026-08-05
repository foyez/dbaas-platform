# Application name
APP_NAME=platform

# Docker/Podman image name
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
	podman build \
		--platform linux/arm64 \
		-t $(APP_NAME):latest \
		--target production \
		-f Containerfile .


# Run development container
# Mount source code so Air can detect changes
dev:
	podman build \
		--platform linux/arm64 \
		-t $(IMAGE_NAME) \
		--target dev \
		-f Containerfile .

	exec podman run \
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
test:
	go test ./...


# Remove generated files
clean:
	rm -rf tmp