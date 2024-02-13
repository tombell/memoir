NAME=memoir

PLATFORMS:=darwin linux windows

sqlc:
	@sqlc generate

dev:
	@echo "building dist/${NAME}"
	@go build -o dist/${NAME} ./cmd/${NAME}

prod: $(PLATFORMS)

$(PLATFORMS):
	@echo "building ${NAME}-$@-amd64"
	@GOOS=$@ GOARCH=amd64 go build -o dist/${NAME}-$@-amd64 ./cmd/${NAME}

run:
	@dist/${NAME} run

watch:
	@while sleep 1; do \
		trap "exit" SIGINT; \
		find . \
			-type d \( -name vendor \) -prune -false -o \
			-type f \( -name "*.go" \) \
		| entr -c -d -r make dev run; \
	done

clean:
	@rm -fr dist

.DEFAULT_GOAL := dev
.PHONY: sqlc dev prod $(PLATFORMS) run watch clean
