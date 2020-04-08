APP=search-secrets

GOCMD=go
GOGENERATE=$(GOCMD) generate
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=$(APP)

all: test build
build:
	$(GOGENERATE) -v . ./cmd/... ./pkg/...
	$(GOBUILD) -o $(BINARY_NAME) -v
build-race:
	$(GOGENERATE) -v . ./cmd/... ./pkg/...
	$(GOBUILD) -race -o $(BINARY_NAME) -v
test:
	$(GOTEST) -race -v . ./cmd/... ./pkg/...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
