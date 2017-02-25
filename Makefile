
.DEFAULT_GOAL := test

PACKAGES=$(go list ./... | grep -v /vendor/)

.PHONY: test
test:
	ginkgo -cover -failOnPending -r -race --randomizeAllSpecs -randomizeSuites --trace $(PACKAGES)

.PHONY: bench
bench:
	go test -run - -bench . -benchmem $(PACKAGES)

.PHONY: lint
lint:
	golint $(PACKAGES)

.PHONY: vet
vet:
	go vet $(PACKAGES)

.PHONY: example
example:
	go build -o ./example/server ./example/server.go
	@echo use ctrl-c to shutdown the example server
	./example/server
