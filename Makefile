VERSION ?= $(shell git describe --exact-match --tags $$(git log -n1 --pretty='%h') 2>/dev/null || echo "$$(git rev-parse --abbrev-ref HEAD)-$$(git rev-parse --short HEAD)")
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
	go fmt `go list ./...`

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
	goreleaser release --rm-dist --auto-snapshot --skip-validate --skip-publish --debug --skip-sign
