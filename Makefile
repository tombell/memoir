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

dev: $(BINARIES)

$(BINARIES):
	go build ${MODFLAGS} ${LDFLAGS} -o dist/$@ ./cmd/$@

clean:
	rm -fr dist/

test:
	go test ${MODFLAGS} ${TESTFLAGS} ./...

create-migration:
	echo "-- UP\n\n-- DOWN" > 'migrations/$(shell date "+%Y%m%d%H%M%S")_$(NAME).sql'

.PHONY: all            \
        dev            \
        clean          \
        test           \
        create-migration
