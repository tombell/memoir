NAME = memoir
PLATFORMS := windows-amd64 windows-arm64 darwin-amd64 darwin-arm64 linux-amd64 linux-arm64

dev:
	@echo "building bin/${NAME}"
	@go build -o bin/${NAME} ./cmd/${NAME}

prod: $(PLATFORMS)

$(PLATFORMS):
	@echo "building ${NAME}-$@"
	@GOOS=$(word 1,$(subst -, ,$@)) GOARCH=$(word 2,$(subst -, ,$@)) \
		go build -o bin/${NAME}-$@ ./cmd/${NAME}

run:
	@pkill -f bin/${NAME} || true
	@bin/${NAME}

watch:
	@while sleep 1; do \
		trap "exit" INT TERM; \
		rg --files --glob '{*.json,*.go,*.tmpl.*}' | \
		entr -c -d -r make dev run; \
	done

clean:
	@rm -fr bin

sqlc:
	@sqlc generate

.DEFAULT_GOAL := dev
.PHONY: dev prod $(PLATFORMS) run watch clean sqlc
