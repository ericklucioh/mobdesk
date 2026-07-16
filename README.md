# Mobdesk

Workstation pessoal de desenvolvimento para Android, executada no próprio celular.

O Mobdesk foi pensado para estudantes e desenvolvedores que querem ir para a faculdade apenas com o celular e continuar programando em um ambiente próprio, sem deixar contas pessoais abertas em computadores compartilhados.

## Visão

```text
Android/HyperOS
└── Termux
    ├── Mobdesk em Go
    └── Ubuntu ARM64 via PRoot
        ├── ferramentas de desenvolvimento
        ├── editor e TUI
        ├── projetos
        └── servidores locais
```

Termux funciona como host e camada de integração com Android. Ubuntu persistente via PRoot é o ambiente principal, oferecendo `apt`, `glibc` e maior compatibilidade com ferramentas Linux tradicionais.

## Objetivo inicial

```text
instalar Go no Termux
→ instalar/obter Mobdesk
→ executar mobdesk start
→ abrir a TUI
→ configurar Ubuntu
→ escolher ferramentas
→ iniciar a workstation
```

A central será responsável por verificar o ambiente, instalar dependências, executar comandos no Termux e no Ubuntu, mostrar progresso e abrir o ambiente de trabalho.

## MVP

- TUI em Go com Bubble Tea, Bubbles e Lip Gloss;
- CLI com Cobra;
- instalação idempotente do Ubuntu ARM64;
- controle do PRoot-Distro;
- Git, Neovim/LazyVim, Python, Node.js e Go;
- shell Ubuntu com PTY;
- sessões persistentes com tmux;
- diagnóstico e logs;
- SSH e acesso remoto como evolução imediata;
- execução de projetos educacionais em C, JavaScript, HTML, React, Java, Go e Python.

O foco é estudo e desenvolvimento de projetos pequenos e médios. O Mobdesk não pretende substituir uma máquina de produção, executar cargas pesadas ou oferecer um desktop Linux gráfico completo.

## Estado do projeto

O repositório está no início da implementação. A entrada do programa fica em `cmd/mobdesk/main.go`; a CLI e a TUI serão construídas progressivamente.

## Tecnologia

- Go 1.26.5;
- Cobra e pflag para CLI;
- Bubble Tea v2;
- Bubbles v2;
- Lip Gloss v2;
- OSC 52 para clipboard em terminais compatíveis;
- PRoot-Distro;
- Ubuntu ARM64;
- Termux;
- tmux;
- OpenSSH;
- Tailscale como opção de rede privada.

## Documentação

- [Missão](docs/MISSAO.md) — problema, público e valor;
- [Estágios](docs/estagios_mobdesk.md) — evolução em seis níveis;
- [Roadmap](docs/ROADMAP.md) — escopo dos MVPs e aplicações;
- [Arquitetura](docs/ARQUITETURA.md) — camadas, execução, acesso e limites;
- [Decisões](docs/DECISOES.md) — escolhas atuais e alternativas adiadas;
- [Ferramentas](docs/FERRAMENTAS.md) — catálogo técnico e práticas.

## Desenvolvimento

Comandos de verificação, conforme o código existente:

```bash
go test ./...
go vet ./...
go build ./cmd/mobdesk
```

Comandos pelo ambiente Docker/Termux:

```bash
make build-image  # constrói a imagem uma vez
make run          # executa o Mobdesk uma vez
make dev          # mantém o Air ativo e reinicia após alterações Go
make test         # executa os testes no container
make build        # gera bin/mobdesk no volume do projeto
```

`make dev` fica em execução durante a sessão de desenvolvimento. Ele não atualiza a TUI no mesmo processo: o Air recompila e reinicia o Mobdesk quando um arquivo observado é salvo.

O teste definitivo deverá ocorrer no Termux do POCO F6. Docker ARM64 e emuladores podem validar lógica, TUI e partes do Ubuntu, mas não reproduzem completamente Android, PRoot, HyperOS e gerenciamento de energia.

## Direção futura

Depois dos MVPs, o Mobdesk poderá ganhar:

- central de projetos e serviços;
- atalhos pela tela inicial com Termux:Widget;
- terminal e ferramentas acessíveis pelo navegador;
- interface web leve;
- APK complementar;
- ambientes reproduzíveis com Nix como camada opcional;
- navegador remoto com Neko como experimento separado.
