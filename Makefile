ifeq ($(GITHUB_BRANCH_NAME),)
	BRANCH := $(shell git rev-parse --abbrev-ref HEAD)-
else
	BRANCH := $(GITHUB_BRANCH_NAME)-
endif
ifeq ($(GITHUB_SHA),)
	COMMIT := $(shell git describe --no-match --dirty --always --abbrev=8)
else
	COMMIT := $(shell echo $(GITHUB_SHA) | cut -c1-8)
endif
VERSION ?= $(if $(RELEASE_TAG),$(RELEASE_TAG),$(shell git describe --tags || echo '$(subst /,-,$(BRANCH))$(COMMIT)'))

CONTAINER_IMAGE := ghcr.io/kakkoyun/subshells:$(VERSION)

LDFLAGS="-X main.version=$(VERSION)"

.PHONY: build
build: bin/subshells bin/infiniteloop

bin/subshells: deps cmd/subshells/main.go
	mkdir -p bin
	go build -a -ldflags=$(LDFLAGS) -o $@ cmd/subshells/main.go

bin/infiniteloop: deps cmd/infiniteloop/main.go
	mkdir -p bin
	go build -a -ldflags=$(LDFLAGS) -o $@ cmd/infiniteloop/main.go

.PHONY: clean
clean:
	rm -rf bin

.PHONY: deps
deps: go.mod
	go mod tidy

.PHONY: format
format:
	go fmt `go list ./...`

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	 go test -v `go list ./...`

.PHONY: container
container: build
	docker build -t $(CONTAINER_IMAGE) --build-arg VERSION=$(VERSION) .

.PHONY: push-container
push-container: container
	docker push $(CONTAINER_IMAGE)

.PHONY: release-dry-run
release-dry-run:
	goreleaser release --clean --auto-snapshot --skip-validate --skip-publish --debug --skip-sign

.PHONY: dev/up
dev/up:
	source ./local-dev.sh && up

.PHONY: dev/down
dev/down:
	source ./local-dev.sh && down
