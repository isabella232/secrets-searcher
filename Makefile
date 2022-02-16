APP=secrets-searcher

GOCMD=go
GOIMPORTS=goimports
ERRCHECK=errcheck
GOGENERATE=$(GOCMD) generate
GOBUILD=GO111MODULE=on $(GOCMD) build
GOCLEAN=GO111MODULE=on $(GOCMD) clean
GOTEST=GO111MODULE=on $(GOCMD) test
GOGET=GO111MODULE=on $(GOCMD) get
GOLIST=GO111MODULE=on $(GOCMD) list
BINARY_NAME=$(APP)

all: test build
generate:
	$(GOGENERATE) -v . ./cmd/... ./pkg/...
errcheck:
	$(ERRCHECK) . ./cmd/... ./pkg/...
lint:
	$(GOIMPORTS) -w $(shell $(GOLIST) -f "{{.Dir}}" ./cmd/... ./pkg/...)
build:
	$(GOGENERATE) -v . ./cmd/... ./pkg/...
	$(GOBUILD) -o $(BINARY_NAME) -v
build-linux:
	$(GOGENERATE) -v . ./cmd/... ./pkg/...
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-linux -v
build-race:
	$(GOGENERATE) -v . ./cmd/... ./pkg/...
	$(GOBUILD) -race -o $(BINARY_NAME) -v
test:
	$(GOGENERATE) -v . ./cmd/... ./pkg/...
	$(GOTEST) -race -v . ./cmd/... ./pkg/...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
