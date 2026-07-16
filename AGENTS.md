# AGENTS.md

## Projeto

Mobdesk é uma central de inicialização e controle para transformar um celular Android em uma workstation pessoal de desenvolvimento. O usuário mantém código, ferramentas, sessões e serviços no próprio aparelho.

## Arquitetura vigente

```text
Android/HyperOS
└── Termux — host e integração com Android
    ├── Mobdesk em Go
    ├── OpenSSH/Tailscale
    ├── wake-lock e inicialização
    └── PRoot-Distro
        └── Ubuntu ARM64 — ambiente principal de desenvolvimento
```

Termux é o host de controle; Ubuntu persistente via PRoot é o runtime principal. PRoot não é VM nem Docker real. Não assumir `systemd`, namespaces completos, cgroups, módulos de kernel, acesso privilegiado a dispositivos ou aceleração gráfica garantida.

## Estado atual

- módulo: `github.com/ericklucioh/mobdesk`;
- Go declarado: `1.26.5`;
- entrada: `cmd/mobdesk/main.go`;
- TUI ainda está no início de implementação;
- missão: `docs/MISSAO.md`;
- arquitetura: `docs/ARQUITETURA.md`;
- roadmap: `docs/ROADMAP.md`;
- decisões: `docs/DECISOES.md`;
- ferramentas: `docs/FERRAMENTAS.md`.

## Dependências e papéis

- `charm.land/bubbletea/v2`: ciclo de eventos e aplicação TUI;
- `charm.land/bubbles/v2`: listas, inputs, tabelas e spinners;
- `charm.land/lipgloss/v2`: estilos e layout;
- `github.com/aymanbagabas/go-osc52/v2`: clipboard via OSC 52;
- `github.com/spf13/cobra`: comandos da CLI;
- `github.com/spf13/pflag`: flags da CLI;
- `golang.org/x/sync`: concorrência e coordenação;
- `golang.org/x/sys`: integração de baixo nível quando necessária;
- pacotes `charmbracelet/x`, terminfo e terminal: suporte de terminal.

O desenho da aplicação é:

```text
Cobra — CLI e roteamento de comandos
└── Bubble Tea — telas TUI interativas
    └── serviços internos
        ├── runtime Termux
        ├── runtime Ubuntu/PRoot
        ├── instalação
        ├── projetos e serviços
        └── diagnóstico
```

`cobra` e `pflag` aparecem como indiretos enquanto não forem importados pelo código. Quando a CLI for implementada, `go mod tidy` deve atualizar o `go.mod` naturalmente. Não criar imports falsos apenas para mudar essa classificação.

## Direção do MVP

O fluxo desejado é instalar Go no Termux, obter o Mobdesk, executar `mobdesk start`, abrir a TUI, configurar Termux e Ubuntu, escolher ferramentas e abrir a workstation Ubuntu.

Comandos previstos:

- `mobdesk start`;
- `mobdesk setup`;
- `mobdesk shell`;
- `mobdesk install <ferramenta>`;
- `mobdesk status`;
- `mobdesk doctor`.

O MVP deve priorizar TUI clara, instalação idempotente, verificação de arquitetura e espaço, PRoot-Distro, Ubuntu ARM64 persistente, execução de comandos nos dois runtimes, captura de saída, progresso, logs, retomada após falha, PTY e diagnóstico.

Não antecipar APK, desktop gráfico, Neko, Nix, Docker real ou múltiplos usuários sem atualizar a missão e o roadmap.

## Regras de implementação

1. Fazer alterações pequenas e coerentes com o estágio atual.
2. Preferir a biblioteca padrão antes de adicionar dependências.
3. Usar Cobra para comandos e flags; usar Bubble Tea para interação TUI.
4. Usar `os/exec` para processos simples e PTY para shells interativos.
5. Separar execução no host Termux da execução dentro do Ubuntu.
6. Não montar comandos concatenando entrada do usuário sem validação.
7. Tornar operações repetíveis; não reinstalar nem apagar dados sem necessidade.
8. Confirmar ações destrutivas, como resetar ou remover o Ubuntu.
9. Manter estado, logs e configurações em caminhos privados e documentados.
10. Não guardar senhas ou segredos no código, logs ou Git.
11. Usar contexto e cancelamento para processos longos.
12. Não bloquear o loop da TUI; executar instalações e diagnósticos em tarefas controladas.

## Estrutura esperada

```text
cmd/mobdesk/          entrada e comandos Cobra
internal/tui/          telas, modelos e componentes Bubble Tea
internal/runtime/      execução Termux/Ubuntu/PRoot
internal/install/      instalação e atualização
internal/projects/     projetos, comandos e portas
internal/services/     processos, sessões e logs
internal/doctor/       verificações e diagnóstico
internal/config/       estado e configuração
```

Não criar toda essa estrutura antecipadamente; introduzir pacotes quando houver comportamento real para organizar.

## Verificação local

Executar, conforme o código existente: `gofmt -w ./cmd ./internal`, `go test ./...`, `go vet ./...` e `go build ./cmd/mobdesk`.

Enquanto ainda não existirem testes ou pacotes `internal`, executar apenas os comandos aplicáveis e registrar limitações. Docker e emuladores podem validar lógica, TUI e Ubuntu, mas o teste definitivo precisa ocorrer no Termux/POCO F6.

## Documentação

Antes de mudar arquitetura ou escopo, ler `docs/MISSAO.md`. Registrar decisões em `docs/DECISOES.md`, arquitetura em `docs/ARQUITETURA.md` e mudanças de escopo em `docs/ROADMAP.md`.
