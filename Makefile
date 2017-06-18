
BENCH_FLAGS ?= -cpuprofile=cpu.pprof -memprofile=mem.pprof -benchmem
PKGS ?= $(shell glide novendor)
HTTP_PROXY = ${http_proxy}
HTTPS_PROXY = ${https_proxy}

# Many Go tools take file globs or directories as arguments instead of packages.
PKG_FILES ?= *.go core ext log demo

# The linting tools evolve with each Go version, so run them only on the latest
# stable release.
GO_VERSION := $(shell go version | cut -d " " -f 3)
GO_MINOR_VERSION := $(word 2,$(subst ., ,$(GO_VERSION)))
LINTABLE_MINOR_VERSIONS := 8
ifneq ($(filter $(LINTABLE_MINOR_VERSIONS),$(GO_MINOR_VERSION)),)
SHOULD_LINT := true
endif

all: lint test

dependencies:
	@echo "Installing Glide and locked dependencies..."
	go get -u -f github.com/Masterminds/glide
	glide install
	@echo "Installing test dependencies..."
	go install ./vendor/github.com/axw/gocov/gocov
	go install ./vendor/github.com/mattn/goveralls
	go install ./vendor/github.com/onsi/ginkgo
ifdef SHOULD_LINT
	@echo "Installing golint..."
	go install ./vendor/github.com/golang/lint/golint
else
	@echo "Not installing golint, since we don't expect to lint on" $(GO_VERSION)
endif

lint:
ifdef SHOULD_LINT
	@rm -rf lint.log
	@echo "Checking formatting..."
	@gofmt -d -s $(PKG_FILES) 2>&1 | tee lint.log
	@echo "Installing test dependencies for vet..."
	@go test -i $(PKGS)
	@echo "Checking vet..."
	@$(foreach dir,$(PKG_FILES),go tool vet $(VET_RULES) $(dir) 2>&1 | tee -a lint.log;)
	@echo "Checking lint..."
	@$(foreach dir,$(PKGS),golint $(dir) 2>&1 | tee -a lint.log;)
	@echo "Checking for unresolved FIXMEs..."
	@git grep -i fixme | grep -v -e vendor -e Makefile | tee -a lint.log
	@[ ! -s lint.log ]
else
	@echo "Skipping linters on" $(GO_VERSION)
endif

test:
	@echo $(PKGS)
	go test -race $(PKGS)

demo-build:
	docker build \
		--build-arg http_proxy=$(HTTP_PROXY) \
		--build-arg https_proxy=$(HTTPS_PROXY) \
		-t ctrace-demo-go .

demo-down:
	docker-compose down

demo-up: demo-down
	docker-compose up

coveralls:
	goveralls -service=travis-ci -ignore=./demos/hellow.go

BENCH ?= .
bench:
	@$(foreach pkg,$(PKGS),go test -bench=$(BENCH) -run="^$$" $(BENCH_FLAGS) $(pkg);)
