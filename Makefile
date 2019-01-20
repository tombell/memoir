VERSION?=dev
COMMIT=$(shell git rev-parse HEAD | cut -c -8)

LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Commit=${COMMIT}"
MODFLAGS=-mod=vendor
TESTFLAGS=-cover

all: dev

clean:
	rm -fr dist/

dev: memoir-dev memoir-delete-dev memoir-import-dev memoir-migrate-dev

memoir-dev:
	go build ${MODFLAGS} ${LDFLAGS} -o dist/memoir ./cmd/memoir

memoir-delete-dev:
	go build ${MODFLAGS} ${LDFLAGS} -o dist/memoir-delete ./cmd/memoir-delete

memoir-import-dev:
	go build ${MODFLAGS} ${LDFLAGS} -o dist/memoir-import ./cmd/memoir-import

memoir-migrate-dev:
	go build ${MODFLAGS} ${LDFLAGS} -o dist/memoir-migrate ./cmd/memoir-migrate

test:
	go test ${MODFLAGS} ${TESTFLAGS} ./...

create-migration:
	echo "-- UP\n\n-- DOWN" > 'migrations/$(shell date "+%Y%m%d%H%M%S")_$(NAME).sql'

.PHONY: all clean dev memoir-dev memoir-delete-dev memoir-import-dev memoir-migrate-dev test create-migration
