# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOMOD=$(GOCMD) mod
GOTEST=$(GOCMD) test
GOFLAGS := -v 
LDFLAGS := -s -w

ifneq ($(shell go env GOOS),darwin)
LDFLAGS := -extldflags "-static"
endif

all: build build-hyperx
build:
	$(GOBUILD) $(GOFLAGS) -ldflags '$(LDFLAGS)' -o "dnsx" cmd/dnsx/dnsx.go
build-hyperx:
	$(GOBUILD) $(GOFLAGS) -ldflags '$(LDFLAGS)' -o "hyperx" cmd/hyperx/hyperx.go
test:
	$(GOTEST) $(GOFLAGS) ./...
tidy:
	$(GOMOD) tidy
