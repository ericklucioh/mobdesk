# Mobdesk

Mobdesk transforma um celular Android em uma workstation de desenvolvimento.
Termux e o host de controle; Ubuntu ARM64 via PRoot-Distro e o ambiente de
desenvolvimento persistente. PRoot nao e VM nem Docker: nao assumir `systemd`,
cgroups, namespaces completos, root, modulos de kernel ou aceleracao grafica.

## Arquitetura

```text
Termux (host: Android, SSH, wake-lock, PRoot)
└── Ubuntu via PRoot (workspace e ferramentas de desenvolvimento)
```

- Entrada: `cmd/mobdesk/main.go`; modulo: `github.com/ericklucioh/mobdesk`.
- Cobra implementa CLI; Bubble Tea implementa a TUI.
- Servicos nao dependem da renderizacao da TUI.
- A TUI usa o contrato JSON da CLI para operacoes reais.
- Acoes de host (`setup`, SSH, PRoot, instalacoes e atualizacao) so executam
  no Termux. Em uma sessao SSH no Ubuntu, a TUI deve explicar a restricao.

## Escopo atual

Comandos atuais: `start`, `stop`, `setup`, `shell`, `install`, `status`,
`update`, `version` e `tui`. `doctor`, projetos, servicos, interface web, APK,
desktop grafico, Docker real, Nix e multiplos usuarios continuam fora do MVP.


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

## Regras de implementacao

1. Prefira mudancas pequenas e a biblioteca padrao antes de dependencias novas.
2. Separe explicitamente comandos do Termux e do Ubuntu; use `os/exec` para
   processos simples e PTY para shells interativos.
3. Valide entradas antes de formar comandos. Nao concatene entrada do usuario.
4. Operacoes repetiveis devem preservar dados e estado; confirme acoes
   destrutivas.
5. Use contexto e cancelamento em processos longos e nunca bloqueie a TUI.
6. Mantenha estado, logs e configuracoes em caminhos privados. Nunca registre
   segredos no codigo, Git ou logs.
7. Nao crie pacotes antecipadamente; extraia um pacote apenas para comportamento
   real e reutilizavel.

## Validacao e documentacao

Execute `make check` antes de concluir alteracoes. Docker valida logica e o
userland; a integracao final precisa ser testada no Termux/POCO F6.

Leia `docs/MISSAO.md` antes de alterar arquitetura ou escopo. Atualize:

- `docs/DECISOES.md` para decisoes;
- `docs/ARQUITETURA.md` para fronteiras tecnicas;
- `docs/ROADMAP.md` para escopo e etapas.
