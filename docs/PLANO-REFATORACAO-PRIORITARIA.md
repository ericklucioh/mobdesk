# Refatoracao Prioritaria

Este documento registra as tres primeiras frentes de refatoracao do Mobdesk.
O objetivo e reduzir riscos reais de execucao e facilitar a evolucao do MVP sem
reestruturar o projeto inteiro.

## Principio

As mudancas devem ser incrementais, preservando os comandos existentes e seus
testes. Nao criar todos os pacotes previstos pela arquitetura antecipadamente.
Cada extracao deve existir porque passou a concentrar comportamento reutilizavel
ou regras que hoje estao duplicadas.

## 1. Separar as fronteiras Termux e Ubuntu na TUI

**Status:** concluído na TUI atual. A centralização futura das operações de
runtime continua necessária para reduzir duplicação entre CLI e serviços.

### Problema

O Mobdesk possui dois ambientes de execucao:

- Termux: host de controle, com `proot-distro`, SSH e wake-lock;
- Ubuntu via PRoot: ambiente de desenvolvimento e destino das sessoes SSH.

Uma TUI iniciada em uma sessao SSH roda dentro do Ubuntu. Mesmo assim, a TUI
real chama o binario local para executar acoes como `start`, `stop`, `setup` e
`update`. Essas acoes dependem do host Termux e nao devem ser oferecidas como
se estivessem disponiveis no Ubuntu remoto.

### Risco atual

O usuario pode selecionar uma acao visivel na TUI, receber uma falha de
processo e nao entender que a restricao e do ambiente, nao da configuracao da
workstation.

### Direcao

- Detectar explicitamente se a TUI esta no Termux ou no Ubuntu/PRoot.
- No modo Ubuntu remoto, manter apenas acoes que funcionam naquele ambiente,
  como consultar o workspace e abrir o shell local.
- Desabilitar ou substituir acoes de host por uma explicacao objetiva de que
  elas devem ser executadas no Termux.
- Preservar o contrato JSON da CLI usado pela TUI, mas nao assumir que todo
  comando esta disponivel em ambos os runtimes.

### Concluido quando

- A TUI nao tenta iniciar, parar ou configurar SSH/PRoot a partir do Ubuntu.
- A interface deixa claro quais acoes exigem o host Termux.
- Existem testes para os dois modos de runtime.

## 2. Centralizar paths e estado persistente

**Status:** concluído. `internal/paths` é a fonte canônica do layout atual e
os consumidores migrados preservam os diretórios e arquivos existentes.

### Problema

Paths como `$HOME/.local/share/mobdesk`, `$HOME/.config/mobdesk`, logs,
marcadores de setup, registros de instalacao, configuracao SSH e
`/root/workspace` sao montados diretamente em mais de um pacote.

### Risco atual

Uma alteracao no layout de diretorios pode deixar setup, status, instalacao e
SSH discordando entre si. Tambem fica mais dificil testar fluxos com diretorios
temporarios, pois parte das regras depende diretamente do ambiente.

### Direcao

- Criar uma pequena fonte unica para os paths do Mobdesk e seus arquivos de
  estado.
- Receber `HOME`, `PREFIX` e, quando necessario, paths explicitamente em vez
  de espalhar leituras de variaveis de ambiente.
- Definir caminhos distintos para estado do Termux e dados no Ubuntu apenas
  onde isso for necessario.
- Migrar gradualmente `cobra`, `status`, `install` e `logs` para essa fonte,
  sem mudar o layout persistido nesta etapa.

### Concluido quando

- Cada path persistente tem uma definicao canonica.
- Nao existem montagens repetidas dos diretorios base do Mobdesk nos pacotes
  consumidores.
- Os testes conseguem usar um diretorio temporario sem depender do `HOME` real.

## 3. Tirar a orquestracao de start e setup da camada Cobra

**Status:** concluído. `internal/workstation.Service` orquestra `start`,
`stop` e todas as fases de `setup` com paths e dependências explícitos; Cobra
adapta flags, streams e renderização humana/JSON.

### Problema

`internal/cobra/start.go` e `internal/cobra/setup.go` concentram regras de
negocio, acesso a processos, SSH, PRoot, wake-lock, arquivos, locks, saida de
terminal e formato JSON. Isso torna os comandos extensos e dificulta testar
fluxos sem executar comandos reais do sistema.

### Risco atual

Novos recursos aumentarao o acoplamento: uma alteracao em SSH, setup ou estado
exige editar o comando Cobra e pode afetar tanto a CLI humana quanto a TUI.

### Direcao

- Comecar por `start` e `stop`, que concentram PID, porta, lock, SSH e
  wake-lock.
- Extrair operacoes para um servico testavel, recebendo dependencias como
  executor de comandos, paths e relogio quando necessario.
- Manter Cobra como adaptador fino: ler flags, chamar o servico e renderizar
  saida humana ou JSON.
- Aplicar o mesmo padrao ao `setup` somente depois que `start/stop` estiverem
  estaveis.

### Concluido quando

- A logica de iniciar e parar a workstation pode ser testada sem `sshd`,
  `proot-distro` ou Termux reais.
- Os comandos Cobra ficam responsaveis apenas por flags, entrada e saida.
- O comportamento externo de `mobdesk start`, `stop` e `setup` permanece
  compativel.

## Ordem de execucao

1. Corrigir o teste atual da TUI e alinhar a documentacao com a implementacao.
2. Implementar a separacao Termux/Ubuntu na TUI.
3. Extrair `start/stop`; depois aplicar o mesmo desenho ao `setup`.

Esta ordem reduz primeiro um erro de experiencia e de runtime, depois elimina
duplicacao estrutural e, por ultimo, torna a orquestracao testavel.
