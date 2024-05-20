GO ?= go
GOFMT ?= gofmt "-s"
GOFILES := $(shell find . -name "*.go")

all: build

.PHONY: build
build: 
	$(GO) mod tidy
	$(GO) build $(BUILD_FLAGS) -o bin/simple example/simple/simple.go
	$(GO) build $(BUILD_FLAGS) -o bin/timed example/timed/timed.go
	$(GO) build $(BUILD_FLAGS) -o bin/multiclient example/multiclient/multiclient.go

.PHONY: run
run: 
	$(GO) mod tidy
	$(GO) run $(BUILD_FLAGS) -race example/simple/simple.go
	$(GO) run $(BUILD_FLAGS) -race example/timed/timed.go
	$(GO) run $(BUILD_FLAGS) -race example/multiclient/multiclient.go

build_network: BUILD_FLAGS=$(shell echo '-tags network')
.PHONY: build_network
build_network: build


run_network: BUILD_FLAGS=$(shell echo '-tags network')
.PHONY: run_network
run_network: run

.PHONY: test
test:
	$(GO) clean -testcache
	$(GO) test -race -v . ./test/
	#parallel prevents multiple TCP servers on the same port
	$(GO) test -race -v -tags network . ./test/ -parallel 1

.PHONY: fmt
fmt:
	$(GOFMT) -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	@diff=$$($(GOFMT) -d $(GOFILES)); \
  if [ -n "$$diff" ]; then \
    echo "Please run 'make fmt' and commit the result:"; \
    echo "$${diff}"; \
    exit 1; \
  fi;
