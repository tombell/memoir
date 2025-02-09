NAME := memoir
BUILD_DIR := bin
PLATFORMS := windows-amd64 windows-arm64 darwin-amd64 darwin-arm64 linux-amd64 linux-arm64

dev:
	@echo "building ${BUILD_DIR}/${NAME}..."
	@go build -o ${BUILD_DIR}/${NAME} ./cmd/${NAME}

prod: $(PLATFORMS)

$(PLATFORMS):
	@echo "building ${BUILD_DIR}/${NAME}-$@..."
	@GOOS=$(word 1,$(subst -, ,$@)) GOARCH=$(word 2,$(subst -, ,$@)) \
		go build -o ${BUILD_DIR}/${NAME}-$@ ./cmd/${NAME}

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
	@sqlc generate

.DEFAULT_GOAL := dev
.PHONY: dev prod $(PLATFORMS) run watch clean sqlc
