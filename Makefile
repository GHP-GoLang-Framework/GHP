GO ?= go
BINARY := ghp
COVERAGE_MIN := 90
COVERAGE_FILE := coverage.out

.PHONY: fmt fmt-fix vet lint build test-unit test-integration test-e2e test-coverage docker-build clean

fmt:
	@unformatted=$$($(GO) fmt -n ./src/... >/dev/null; gofmt -l ./src); \
	if [ -n "$$unformatted" ]; then \
		echo "Arquivos não formatados:"; echo "$$unformatted"; exit 1; \
	fi

fmt-fix:
	gofmt -w ./src

vet:
	$(GO) vet ./src/...

lint: vet
	golangci-lint run ./src/...

build:
	$(GO) build -o bin/$(BINARY) ./src/cmd/ghp

test-unit:
	$(GO) test -short ./src/... -race

test-integration:
	$(GO) test ./src/test/integration/... -race

test-e2e: build
	$(GO) test ./src/test/e2e/...

test-coverage:
	$(GO) test ./src/... -coverprofile=$(COVERAGE_FILE) -covermode=atomic -coverpkg=./src/...
	@$(GO) tool cover -func=$(COVERAGE_FILE) | tail -1
	@COVERAGE=$$($(GO) tool cover -func=$(COVERAGE_FILE) | grep total: | awk '{print substr($$3, 1, length($$3)-1)}'); \
	echo "Cobertura total: $$COVERAGE% (mínimo: $(COVERAGE_MIN)%)"; \
	awk -v cov="$$COVERAGE" -v min="$(COVERAGE_MIN)" 'BEGIN { exit (cov+0 < min) }' || \
		{ echo "Cobertura abaixo do mínimo de $(COVERAGE_MIN)%"; exit 1; }

docker-build:
	docker build -t ghp:local .

clean:
	rm -rf bin/ $(COVERAGE_FILE)
