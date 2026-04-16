##@ OCP Integration tests

.PHONY: prereqs-kind
build-integration-tests: ## Build integration tests suite binary with OTE
	CGO_ENABLED=0 go build -mod=mod -o bin/netobserv-tests-ext ./cmd/netobserv-tests-ext
