SHELL := /bin/sh

COMPOSE ?= docker compose
SERVICE ?= termux
TERMUX_ARCH ?= latest

.PHONY: help termux build-image shell test vet build run dev clean-env reset-env clean-image arm64-image

help:
	@printf '%s\n' \
		'make build-image  - constrói a imagem local do Termux' \
		'make termux       - abre um shell interativo no ambiente' \
		'make shell        - abre um shell no container existente' \
		'make test         - executa go test ./...' \
		'make vet          - executa go vet ./...' \
		'make build        - compila o Mobdesk dentro do container' \
		'make run          - executa o binário do Mobdesk' \
		'make dev          - inicia o Air com hot-reload' \
		'make clean-env    - apaga os volumes persistentes do Termux' \
		'make reset-env    - recria o ambiente do Termux do zero' \
		'make arm64-image  - constrói a imagem Termux para linux/arm64' \
		'make clean-image  - remove a imagem local do ambiente'

build-image:
	TERMUX_ARCH=$(TERMUX_ARCH) $(COMPOSE) build

termux:
	TERMUX_ARCH=$(TERMUX_ARCH) $(COMPOSE) run --rm $(SERVICE) bash

shell:
	TERMUX_ARCH=$(TERMUX_ARCH) $(COMPOSE) run --rm $(SERVICE) bash

test:
	TERMUX_ARCH=$(TERMUX_ARCH) $(COMPOSE) run --rm $(SERVICE) bash -lc 'go test ./...'

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
	@printf '%s\n' 'Ambiente Termux recriado. Execute make termux, make test ou make dev.'

arm64-image:
	docker buildx build --platform linux/arm64 --build-arg TERMUX_ARCH=aarch64 -f Dockerfile.termux -t mobdesk-termux:aarch64 --load .

clean-image:
	docker image rm mobdesk-termux:$(TERMUX_ARCH)

fmt:
	go fmt ./...
	
check: fmt vet test build