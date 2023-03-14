VERSION?=dev
COMMIT=$(shell git rev-parse HEAD | cut -c -8)

LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT)"
MODFLAGS=-mod=vendor
TESTFLAGS=-cover

PLATFORMS:=darwin linux

dev:
	@echo building dist/memoir
	@go build ${MODFLAGS} ${LDFLAGS} -o dist/memoir ./cmd/memoir

prod: $(PLATFORMS)

$(PLATFORMS):
	@echo building dist/memoir-$@-amd64
	@GOOS=$@ GOARCH=amd64 go build ${MODFLAGS} ${LDFLAGS} -o dist/memoir-$@-amd64 ./cmd/memoir

run:
	@dist/memoir run

watch:
	@while sleep 1; do \
		trap "exit" SIGINT; \
		find . \
			-type d \( -name vendor \) -prune -false -o \
			-type f \( -name "*.go" \) \
			| entr -c -d -r make dev run; \
	done

test:
	@go test ${MODFLAGS} ${TESTFLAGS} ./...

clean:
	@rm -fr dist

.DEFAULT_GOAL := dev
.PHONY: dev prod $(PLATFORMS) run watch test clean
