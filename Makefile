VERSION?=dev
COMMIT=$(shell git rev-parse HEAD | cut -c -8)

LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Commit=${COMMIT}"
MODFLAGS=-mod=vendor
TESTFLAGS=-cover

PLATFORMS:=darwin linux windows

BINARIES:=memoir         \
          memoir-delete  \
          memoir-export  \
          memoir-import  \
          memoir-migrate \
          memoir-upload  \

all: dev

dev:
	@for target in $(BINARIES); do \
		echo building dist/$$target; \
		go build ${MODFLAGS} ${LDFLAGS} -o dist/$$target ./cmd/$$target || exit 1; \
	done

dist: $(PLATFORMS)

$(PLATFORMS):
	@for target in $(BINARIES); do \
		echo building dist/$$target-$@-amd64; \
		GOOS=$@ GOARCH=amd64 go build ${MODFLAGS} ${LDFLAGS} -o dist/$$target-$@-amd64 ./cmd/$$target || exit 1 \
	done

$(BINARIES):
	@echo building dist/$@
	@go build ${MODFLAGS} ${LDFLAGS} -o dist/$@ ./cmd/$@

clean:
	rm -fr dist/

test:
	go test ${MODFLAGS} ${TESTFLAGS} ./...

create-migration:
	echo "-- UP\n\n-- DOWN" > 'migrations/$(shell date "+%Y%m%d%H%M%S")_$(NAME).sql'

.PHONY: all              \
        dev              \
        dist             \
        $(PLATFORMS)     \
        $(BINARIES)      \
        clean            \
        test             \
        create-migration \
