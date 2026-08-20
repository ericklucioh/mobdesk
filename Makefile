SHELL := /bin/sh

COMPOSE ?= docker compose
SERVICE ?= termux
TERMUX_ARCH ?= latest
CATALOG_TIMEOUT ?= 30m

.PHONY: help termux build-image shell test integration-test vet build run dev clean-env reset-env clean-image arm64-image fmt i18n-check check

help:
	@printf '%s\n' \
		'make build-image     - build the local Termux image' \
		'make termux          - open an interactive shell in the environment' \
		'make shell           - open a shell in the existing container' \
		'make test            - run go test ./...' \
		'make i18n-check      - validate catalogs, presentation and documentation links' \
		'make check           - format, validate and build the project' \
		'make integration-test - test the Termux/SSH flow in Docker' \
		'make vet             - run go vet ./...' \
		'make build           - build Mobdesk inside the container' \
		'make run             - run the Mobdesk binary' \
		'make dev             - start Air with hot reload' \
		'make clean-env       - remove persistent Termux volumes' \
		'make reset-env       - recreate the Termux environment from scratch' \
		'make arm64-image     - build the Termux image for linux/arm64' \
		'make clean-image     - remove the local environment image'

build-image:
	TERMUX_ARCH=$(TERMUX_ARCH) $(COMPOSE) build

termux:
	TERMUX_ARCH=$(TERMUX_ARCH) $(COMPOSE) run --rm --service-ports $(SERVICE) bash

shell:
	TERMUX_ARCH=$(TERMUX_ARCH) $(COMPOSE) run --rm --service-ports $(SERVICE) bash

test:
	TERMUX_ARCH=$(TERMUX_ARCH) $(COMPOSE) run --rm $(SERVICE) bash -lc 'go test ./...'

integration-test:
	TERMUX_ARCH=$(TERMUX_ARCH) COMPOSE="$(COMPOSE)" ./scripts/test-termux.sh

vet:
	TERMUX_ARCH=$(TERMUX_ARCH) $(COMPOSE) run --rm $(SERVICE) bash -lc 'go vet ./...'

build:
	TERMUX_ARCH=$(TERMUX_ARCH) $(COMPOSE) run --rm $(SERVICE) bash -lc 'mkdir -p bin && go build -o bin/mobdesk ./cmd/mobdesk'

run:
	TERMUX_ARCH=$(TERMUX_ARCH) $(COMPOSE) run --rm $(SERVICE) bash -lc 'go run ./cmd/mobdesk'

dev:
	TERMUX_ARCH=$(TERMUX_ARCH) $(COMPOSE) run --rm $(SERVICE) air -c .air.toml

clean-env:
	$(COMPOSE) down --volumes --remove-orphans

reset-env: clean-env build-image
	@printf '%s\n' 'Termux environment recreated. Run make termux, make test or make dev.'

arm64-image:
	docker buildx build --platform linux/arm64 --build-arg TERMUX_ARCH=aarch64 -f Dockerfile.termux -t mobdesk-termux:aarch64 --load .

clean-image:
	docker image rm mobdesk-termux:$(TERMUX_ARCH)

fmt:
	go fmt ./...
	
i18n-check:
	./scripts/i18n-check.sh

check: fmt i18n-check vet test build
