# Root Makefile: local development on a kind cluster.
# Component-specific targets live in platform/, operator/ and client/.

SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

OS   := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')

## Pinned tool versions. Tools are downloaded into ./bin, nothing is installed globally.
KIND_VERSION      ?= v0.33.0
KUBECTL_VERSION   ?= v1.35.10
HELM_VERSION      ?= v3.22.1
TERRAFORM_VERSION ?= 1.12.2

LOCALBIN  ?= $(CURDIR)/bin
KIND      ?= $(LOCALBIN)/kind-$(KIND_VERSION)
KUBECTL   ?= $(LOCALBIN)/kubectl-$(KUBECTL_VERSION)
HELM      ?= $(LOCALBIN)/helm-$(HELM_VERSION)
TERRAFORM ?= $(LOCALBIN)/terraform-$(TERRAFORM_VERSION)

export KIND KUBECTL HELM TERRAFORM

.PHONY: all
all: help

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Local development

.PHONY: dev
dev: tools ## Create the kind cluster and deploy the whole platform.
	hack/dev/up.sh

.PHONY: dev-down
dev-down: tools ## Delete the kind cluster and local state.
	hack/dev/down.sh

.PHONY: dev-status
dev-status: tools ## Show what is running in the local cluster.
	hack/dev/status.sh

##@ Tools

.PHONY: tools
tools: $(KIND) $(KUBECTL) $(HELM) $(TERRAFORM) ## Download pinned tools into ./bin.
	@command -v docker >/dev/null || { echo "docker is required (Docker Desktop, OrbStack, Colima, ...)"; exit 1; }
	@docker info >/dev/null 2>&1 || { echo "docker is installed but not running"; exit 1; }

$(LOCALBIN):
	mkdir -p $(LOCALBIN)

$(KIND): | $(LOCALBIN)
	curl -fsSLo $@ https://kind.sigs.k8s.io/dl/$(KIND_VERSION)/kind-$(OS)-$(ARCH)
	chmod +x $@

$(KUBECTL): | $(LOCALBIN)
	curl -fsSLo $@ https://dl.k8s.io/release/$(KUBECTL_VERSION)/bin/$(OS)/$(ARCH)/kubectl
	chmod +x $@

$(HELM): | $(LOCALBIN)
	curl -fsSL https://get.helm.sh/helm-$(HELM_VERSION)-$(OS)-$(ARCH).tar.gz | tar -xz -C $(LOCALBIN) --strip-components=1 $(OS)-$(ARCH)/helm
	mv $(LOCALBIN)/helm $@

$(TERRAFORM): | $(LOCALBIN)
	curl -fsSLo $(LOCALBIN)/terraform.zip https://releases.hashicorp.com/terraform/$(TERRAFORM_VERSION)/terraform_$(TERRAFORM_VERSION)_$(OS)_$(ARCH).zip
	unzip -o -q $(LOCALBIN)/terraform.zip terraform -d $(LOCALBIN)
	mv $(LOCALBIN)/terraform $@
	rm $(LOCALBIN)/terraform.zip
