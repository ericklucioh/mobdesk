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

## Regras de fronteira e contrato CLI

1. Todo comando Cobra nao interativo e consumido pela TUI ou por automacao deve
   oferecer `--json` e manter o schema versionado. `shell` e `tui` sao excecoes
   por serem fluxos interativos.
2. Em modo JSON, `stdout` deve conter somente JSON valido. Mensagens humanas,
   progresso e diagnosticos nao podem poluir a resposta; progresso deve usar o
   formato de eventos documentado pelo comando.
3. O contrato JSON deve preservar `schema_version`, `command`, `success`,
   `state` e `message`. Campos novos devem ser aditivos e compativeis com o
   schema existente.
4. Cobra deve adaptar flags, argumentos, contexto e saida. Regras de negocio,
   instalacao, status, configuracao e operacao de host pertencem aos servicos
   internos, nao aos handlers dos comandos.
5. Separe sempre os ambientes: Termux e o host Android; Ubuntu via PRoot e o
   userland de desenvolvimento. Nao assuma que um processo no Ubuntu possui
   acesso ao host, root real, systemd ou namespaces completos.
6. Acoes exclusivas do host devem validar o runtime antes de executar e, quando
   chamadas a partir do Ubuntu ou de uma sessao SSH, retornar uma explicacao
   objetiva orientando o usuario a sair para o Termux.
7. Nao use `runtime.GOOS` sozinho para detectar Termux: o binario Linux ARM64
   pode rodar no userland Android. Use os marcadores e caminhos canonicos do
   projeto, como `PREFIX` e `paths.Current()`.
8. Processos simples devem passar por `executil`; comandos dentro do Ubuntu
   devem atravessar a fronteira declarada por `proot-distro`. A TUI nao deve
   chamar `apt`, `pipx`, `npm`, `proot-distro` ou scripts diretamente.
9. Todo processo longo deve receber `cmd.Context()` ou contexto equivalente,
   suportar cancelamento e deixar estado e logs consistentes em falhas parciais.
10. Cada novo comando deve ter testes para argumentos invalidos, modo texto,
    modo JSON quando aplicavel, erro de runtime e cancelamento quando houver
    operacao longa.

## Regras de UX da TUI

1. A TUI e touch-first: toda acao importante deve ter um alvo visivel e
   clicavel por mouse/toque; nao dependa de o usuario descobrir uma tecla.
2. Todo controle clicavel deve ter uma regiao de hit-test coerente com o
   controle renderizado. Nao use a linha inteira como botao quando isso puder
   disparar uma acao demorada, destrutiva ou inesperada.
3. Toda tela nova ou popup deve oferecer `Voltar`, `Fechar` ou `X` visivel e
   clicavel, alem de um equivalente de teclado. O usuario nunca deve ficar
   preso em uma tela.
4. Toda acao disponivel por mouse/toque tambem deve funcionar por teclado; toda
   acao indisponivel deve explicar o motivo no proprio fluxo.
5. Toque em uma linha de app deve abrir detalhes antes de instalar, remover ou
   alterar configuracao. Acoes destrutivas exigem confirmacao dentro da tela.
6. Estados `busy`, erro, conflito, concluido e bloqueado devem ser visiveis e
   impedir cliques duplicados ou acoes incompatíveis.
7. Telas devem funcionar em terminais estreitos: nao depender de largura fixa,
   nao cortar botoes e manter navegacao e confirmacoes acessiveis.
8. Cada fluxo novo de mouse/toque deve ter teste de hit-test, navegacao e
   teclado equivalente. Validar tambem visualmente em terminal estreito e em
   um dispositivo Termux real quando a mudanca envolver interacao.

## Validacao e documentacao

Execute `make check` antes de concluir alteracoes. Docker valida logica e o
userland; a integracao final precisa ser testada no Termux/POCO F6.

Leia `docs/MISSAO.md` antes de alterar arquitetura ou escopo. Atualize:

- `docs/DECISOES.md` para decisoes;
- `docs/ARQUITETURA.md` para fronteiras tecnicas;
- `docs/ROADMAP.md` para escopo e etapas.
