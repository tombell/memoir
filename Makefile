VERSION?=dev
COMMIT=$(shell git rev-parse HEAD | cut -c -8)

LDFLAGS=-ldflags "-X github.com/tombell/memoir.Version=${VERSION} -X github.com/tombell/memoir.Commit=${COMMIT}"
MODFLAGS=-mod=vendor

PLATFORMS:=darwin linux
BINARIES:=memoir memoir-db

all: dev

dev:
	@for target in $(BINARIES); do \
		echo building dist/$$target; \
		go build ${MODFLAGS} ${LDFLAGS} -o dist/$$target ./cmd/$$target || exit 1; \
	done

prod: $(PLATFORMS)

$(PLATFORMS):
	@for target in $(BINARIES); do \
		echo building dist/$$target-$@-amd64; \
		GOOS=$@ GOARCH=amd64 go build ${MODFLAGS} ${LDFLAGS} -o dist/$$target-$@-amd64 ./cmd/$$target || exit 1; \
	done

run:
	@dist/memoir

watch:
	@while sleep 1; do \
		trap "exit" SIGINT; \
		find . \
			-type d \( -name vendor \) -prune -false -o \
			-type f \( -name "*.go" \) \
			| entr -c -d -r make dev run; \
	done

test:
	@go test ${MODFLAGS} -cover ./...

clean:
	@rm -fr dist $(ARCHIVE_PATH)

.PHONY: all          \
        dev          \
        prod         \
        $(PLATFORMS) \
        run          \
        watch        \
        test         \
        clean        \
