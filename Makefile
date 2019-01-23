VERSION?=dev
COMMIT=$(shell git rev-parse HEAD | cut -c -8)

LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Commit=${COMMIT}"
MODFLAGS=-mod=vendor
TESTFLAGS=-cover

BINARIES=memoir         \
         memoir-delete  \
         memoir-export  \
         memoir-import  \
         memoir-migrate \
         memoir-upload

all: dev

clean:
	rm -fr dist/

dev: $(BINARIES)

$(BINARIES):
	go build ${MODFLAGS} ${LDFLAGS} -o dist/$@ ./cmd/$@

test:
	go test ${MODFLAGS} ${TESTFLAGS} ./...

create-migration:
	echo "-- UP\n\n-- DOWN" > 'migrations/$(shell date "+%Y%m%d%H%M%S")_$(NAME).sql'

.PHONY: all            \
        clean          \
        dev            \
        test           \
        create-migration
