APP=search-secrets

GOCMD=go
GOGENERATE=$(GOCMD) generate
GOBUILD=GO111MODULE=on $(GOCMD) build
GOCLEAN=GO111MODULE=on $(GOCMD) clean
GOTEST=GO111MODULE=on $(GOCMD) test
GOGET=GO111MODULE=on $(GOCMD) get
BINARY_NAME=$(APP)

all: test build
generate:
	$(GOGENERATE) -v . ./cmd/... ./pkg/...
build:
	$(GOGENERATE) -v . ./cmd/... ./pkg/...
	$(GOBUILD) -o $(BINARY_NAME) -v
build-race:
	$(GOGENERATE) -v . ./cmd/... ./pkg/...
	$(GOBUILD) -race -o $(BINARY_NAME) -v
test:
	$(GOGENERATE) -v . ./cmd/... ./pkg/...
	$(GOTEST) -race -v . ./cmd/... ./pkg/...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
