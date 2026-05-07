##@ OCP Integration tests

.PHONY: build-integration-tests
build-integration-tests: ## Build integration tests suite binary with OTE
	CGO_ENABLED=0 go build -mod=vendor -o bin/netobserv-tests-ext ./cmd/netobserv-tests-ext

.PHONY: image-tests-ext
image-tests-ext: ## Build tests extension container image
	podman build -f Dockerfile.tests-ext -t $(REPO)/netobserv-operator-tests-ext:$(VERSION) .
