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

test_network: BUILD_FLAGS=$(shell echo '-tags network,randomize_ports')
.PHONY: test_network
test_network: test

.PHONY: test_socket
test_socket: test

.PHONY: test
test:
	$(GO) clean -testcache
	$(GO) test $(BUILD_FLAGS) -race -v -parallel 1 -failfast $(TEST_FLAGS) -run Base .
	$(GO) test $(BUILD_FLAGS) -race -v -parallel 1 -failfast $(TEST_FLAGS) -run Reconnect .
	$(GO) test $(BUILD_FLAGS) -race -v -parallel 1 -failfast $(TEST_FLAGS) -run Unix .
	$(GO) test $(BUILD_FLAGS) -race -v -parallel 1 -failfast $(TEST_FLAGS) -run Handshake .

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
