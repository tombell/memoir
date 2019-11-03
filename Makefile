VERSION?=dev
COMMIT=$(shell git rev-parse HEAD | cut -c -8)

LDFLAGS=-ldflags "-X github.com/tombell/memoir.Version=${VERSION} -X github.com/tombell/memoir.Commit=${COMMIT}"
MODFLAGS=-mod=vendor

PLATFORMS:=darwin \
           linux  \

BINARIES:=memoir            \
          memoir-db         \
          memoir-tracklists \
          memoir-upload     \

ARCHIVE_PATH:=/tmp/memoir.tar.gz

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

$(BINARIES):
	@echo building dist/$@
	@go build ${MODFLAGS} ${LDFLAGS} -o dist/$@ ./cmd/$@

test:
	@go test ${MODFLAGS} -cover ./...

clean:
	@rm -fr dist $(ARCHIVE_PATH)

modules:
	@go mod download && go mod tidy && go mod vendor

create-migration:
	@echo "-- UP\n\n-- DOWN" > 'datastore/migrations/$(shell date "+%Y%m%d%H%M%S")_$(NAME).sql'

archive:
	@bsdtar -zcf $(ARCHIVE_PATH) \
		-s ,^dist/memoir-linux-amd64,dist/memoir, \
		dist/memoir-linux-amd64 \
		.env.toml \
		configs/Caddyfile \
		configs/memoir.service

.PHONY: all              \
        dev              \
        prod             \
        $(PLATFORMS)     \
        $(BINARIES)      \
        test             \
        clean            \
        modules          \
        create-migration \
        archive          \
