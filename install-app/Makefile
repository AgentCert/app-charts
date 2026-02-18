# Makefile for install-app Docker image
# Build and push Helm chart installer

# Configuration - override these with environment variables or command line
IMAGE_REGISTRY ?= docker.io
IMAGE_NAME ?= agentcert-install-app
IMAGE_TAG ?= latest

# Full image reference
IMAGE = $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

# Build context - parent directory (app-charts root)
BUILD_CONTEXT = ..

.PHONY: help
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*##"; } /^[a-zA-Z_-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the Docker image
	@echo "Building Docker image: $(IMAGE)"
	docker build -t $(IMAGE) -f Dockerfile $(BUILD_CONTEXT)

.PHONY: build-no-cache
build-no-cache: ## Build the Docker image without cache
	@echo "Building Docker image (no cache): $(IMAGE)"
	docker build --no-cache -t $(IMAGE) -f Dockerfile $(BUILD_CONTEXT)

.PHONY: push
push: ## Push the Docker image to registry
	@echo "Pushing Docker image: $(IMAGE)"
	docker push $(IMAGE)

.PHONY: build-push
build-push: build push ## Build and push the Docker image

.PHONY: run
run: ## Run the container with help
	docker run --rm $(IMAGE)

.PHONY: run-shell
run-shell: ## Run container with interactive shell
	docker run --rm -it --entrypoint /bin/bash $(IMAGE)

.PHONY: list-charts
list-charts: ## List charts available in the image
	docker run --rm --entrypoint /bin/sh $(IMAGE) -c "ls -la /charts"

.PHONY: install
install: ## Install chart (use FOLDER=<name> NAMESPACE=<ns> KUBECONFIG=<path>)
ifndef FOLDER
	$(error FOLDER is required. Usage: make install FOLDER=sock-shop NAMESPACE=sock-shop)
endif
	docker run --rm \
		$(if $(KUBECONFIG),-v $(KUBECONFIG):/kubeconfig -e KUBECONFIG=/kubeconfig) \
		$(if $(KUBE_CONTEXT),--network host) \
		$(IMAGE) \
		-folder $(FOLDER) \
		$(if $(NAMESPACE),-namespace $(NAMESPACE)) \
		$(if $(RELEASE),-release $(RELEASE)) \
		$(if $(VALUES),-values $(VALUES)) \
		$(if $(DRY_RUN),-dry-run)

.PHONY: install-local
install-local: ## Install using local kubeconfig (auto-detects ~/.kube/config)
ifndef FOLDER
	$(error FOLDER is required. Usage: make install-local FOLDER=sock-shop NAMESPACE=sock-shop)
endif
	docker run --rm \
		-v $(HOME)/.kube/config:/kubeconfig:ro \
		--network host \
		$(IMAGE) \
		-folder $(FOLDER) \
		-kubeconfig /kubeconfig \
		$(if $(NAMESPACE),-namespace $(NAMESPACE)) \
		$(if $(RELEASE),-release $(RELEASE)) \
		$(if $(DRY_RUN),-dry-run)

.PHONY: tag
tag: ## Tag image with additional tag (use NEW_TAG=<tag>)
ifndef NEW_TAG
	$(error NEW_TAG is required. Usage: make tag NEW_TAG=v1.0.0)
endif
	docker tag $(IMAGE) $(IMAGE_REGISTRY)/$(IMAGE_REPO)/$(IMAGE_NAME):$(NEW_TAG)

.PHONY: clean
clean: ## Remove local Docker image
	docker rmi $(IMAGE) || true

.PHONY: test
test: ## Run tests
	@echo "Running Go tests..."
	go test -v ./...

.PHONY: lint
lint: ## Run linter
	@echo "Running Go linter..."
	golangci-lint run ./...

.PHONY: version
version: ## Show current version configuration
	@echo "Image: $(IMAGE)"
	@echo "Registry: $(IMAGE_REGISTRY)"
	@echo "Repository: $(IMAGE_REPO)"
	@echo "Name: $(IMAGE_NAME)"
	@echo "Tag: $(IMAGE_TAG)"
