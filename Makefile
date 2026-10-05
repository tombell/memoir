NAME := memoir
BUILD_DIR := bin
PLATFORMS := windows-amd64 windows-arm64 darwin-amd64 darwin-arm64 linux-amd64 linux-arm64
GO := go
SQLC_VERSION := 1.30.0
MIGRATIONS_DIR := internal/database/migrations

dev:
	@echo "building ${BUILD_DIR}/${NAME}..."
	@$(GO) build -mod=vendor -o ${BUILD_DIR}/${NAME} ./cmd/${NAME}

prod: $(PLATFORMS)

$(PLATFORMS):
	@echo "building ${BUILD_DIR}/${NAME}-$@..."
	@GOOS=$(word 1,$(subst -, ,$@)) GOARCH=$(word 2,$(subst -, ,$@)) \
		$(GO) build -mod=vendor -o ${BUILD_DIR}/${NAME}-$@ ./cmd/${NAME}

run:
	@pkill -f ${BUILD_DIR}/${NAME} || true
	@${BUILD_DIR}/${NAME}

watch:
	@while sleep 1; do \
		trap "exit" INT TERM; \
		rg --files -g '{*.json,*.go,*.sql,*.tmpl.*}' -g '!internal/database/*.go' | \
		entr -c -d -r make sqlc dev run; \
	done

clean:
	@echo "cleaning..."
	@rm -fr ${BUILD_DIR}

sqlc:
	@test "$$(sqlc version)" = "v$(SQLC_VERSION)" || \
		{ echo "sqlc $(SQLC_VERSION) is required; run mise install" >&2; exit 1; }
	@sqlc generate

fmt:
	@find . -path ./vendor -prune -o -path ./.git -prune -o \
		-type f -name '*.go' -print0 | xargs -0 gofmt -w

check-format:
	@files="$$(find . -path ./vendor -prune -o -path ./.git -prune -o \
		-type f -name '*.go' -print0 | xargs -0 gofmt -l)" || exit 1; \
		if [ -n "$$files" ]; then \
			printf 'Run make fmt to format:\n%s\n' "$$files"; exit 1; \
		fi

build:
	@$(GO) build -mod=vendor ./...

vet:
	@$(GO) vet -mod=vendor ./...

test:
	@$(GO) test -mod=vendor -count=1 ./...

test-integration:
	@test -n "$$MEMOIR_TEST_DATABASE_URL" || \
		{ echo "Set MEMOIR_TEST_DATABASE_URL to an isolated PostgreSQL test database" >&2; exit 1; }
	@DATABASE_URL="$$MEMOIR_TEST_DATABASE_URL" $(GO) tool migrate apply \
		--db postgresql --dsn "$$MEMOIR_TEST_DATABASE_URL" --migrations $(MIGRATIONS_DIR)
	@DATABASE_URL="$$MEMOIR_TEST_DATABASE_URL" $(GO) test -mod=vendor -tags=integration -count=1 ./...

check: check-format build vet test

.DEFAULT_GOAL := dev
.PHONY: dev prod $(PLATFORMS) run watch clean sqlc fmt check-format build vet test test-integration check
