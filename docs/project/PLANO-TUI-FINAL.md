# Plano de conclusão da TUI

Este plano fecha a primeira versão da TUI do Mobdesk.

## Objetivo

Entregar `mobdesk tui` como uma central fullscreen para Termux e SSH, com a
mesma linguagem visual do protótipo em `docs/figma-manager`, funcionamento
confortável em aproximadamente 350 px de largura e uso das operações reais do
Mobdesk sem duplicar regras de segurança ou instalação.

## Contratos definidos

### Entrada

O ponto de entrada público é:

```text
mobdesk tui
```

A TUI não será iniciada por `start`. `start` inicia a workstation e retorna.

### Operações JSON

`setup`, `start` e `stop` aceitam `--json` e imprimem exatamente um objeto JSON
no stdout. Mensagens auxiliares ficam no stderr. Todo resultado contém:

```json
{
  "schema_version": 1,
  "command": "start",
  "success": true,
  "state": "running",
  "message": "Workstation iniciada"
}
```

Instalação e atualização preservam seus modelos específicos, mas passam a
expor `schema_version` para que a TUI possa validar respostas.

### Catálogo

A TUI exibe somente o catálogo atualmente implementado:

```text
go, python, node, c, cpp, lua
```

Não serão exibidas linguagens ainda não suportadas pelo serviço de instalação.

### Shell

O shell é uma ação explícita. A TUI é suspensa com o terminal liberado, um PTY
é alocado para o processo Ubuntu e a TUI é restaurada quando o shell termina.

## Fase 1: núcleo de execução

### Resultado

Criar uma camada pequena para executar comandos, capturar stdout/stderr,
cancelar processos e converter respostas JSON em mensagens Bubble Tea.

### Tarefas

- criar tipos de operação e erro estruturado;
- validar `schema_version` e `command`;
- separar stdout JSON de stderr;
- preservar código de saída;
- adicionar cancelamento por contexto;
- transportar mensagem, estado e log para a tela;
- impedir execução concorrente da mesma operação;
- manter `os/exec` fora da renderização.

### Critérios

- nenhuma operação longa roda em `Update`;
- stdout dos comandos JSON continua sendo JSON válido;
- erros operacionais aparecem na TUI sem panic;
- `Ctrl+C` cancela ou confirma saída sem deixar processo órfão.

## Fase 2: runtime e PTY

### Resultado

Shell interativo confiável em Termux, SSH e PRoot.

### Tarefas

- adicionar alocação de PTY;
- copiar stdin/stdout entre terminal e PTY;
- colocar terminal local em modo raw durante o shell;
- restaurar terminal em sucesso, erro ou cancelamento;
- encaminhar resize do terminal;
- fechar descritores e processos filhos;
- manter `mobdesk shell` como comando explícito;
- usar o mesmo runner no shell chamado pela TUI.

### Critérios

- Neovim e ferramentas interativas funcionam;
- `Ctrl+D` encerra o shell;
- a TUI retorna depois do shell;
- largura e altura são atualizadas durante o uso.

## Fase 3: modelo e navegação

### Resultado

TUI dividida em modelo, comandos, componentes e renderização.

### Estrutura

```text
internal/tui/
  model.go
  messages.go
  commands.go
  keys.go
  viewport.go
  styles.go
  components.go
  screens.go
  operations.go
  logs.go
```

### Tarefas

- manter tela atual e histórico de navegação;
- criar foco/seleção para cartões e listas;
- mapear `Enter`, setas, `Esc`, `q` e `Ctrl+C`;
- tornar `Início` e `X` ações reais;
- implementar confirmação modal reutilizável;
- limitar scroll ao conteúdo real;
- adaptar conteúdo à largura mínima;
- evitar largura negativa ou overflow.

## Fase 4: visual e viewport

### Resultado

Reproduzir a linguagem do HTML em terminal estreito.

### Regras

- fundo preto;
- borda roxa de um caractere;
- destaque lilás para ações principais;
- texto principal branco;
- texto secundário azul/lilás discreto;
- cabeçalho fixo no topo;
- rodapé de atalhos;
- cards empilhados em largura pequena;
- listas compactas para ferramentas;
- etapas com círculo sem preenchimento;
- modal central com borda dupla;
- conteúdo rolável sem deslocar o cabeçalho.

### Critérios

- layout legível em 350 px;
- nenhum texto essencial é cortado;
- listas podem ser percorridas com teclado;
- operação longa mostra progresso sem travar a tela.

## Fase 5: telas e operações

### Home

- estado Workstation SSH;
- Iniciar quando parada;
- Parar quando ativa;
- cards de configuração, status, ferramentas e shell;
- versão, arquitetura e logs.

### Status

- host e arquitetura;
- setup e fases;
- Ubuntu e workspace;
- SSH, porta e endereços;
- wake-lock;
- bateria e Wi-Fi;
- armazenamento;
- instalações;
- alertas;
- atualizar status.

### Setup

- lista de fases;
- fase concluída, ativa e pendente;
- execução retomável;
- `--upgrade-system` com confirmação;
- cancelamento;
- logs;
- tratamento explícito de senha interativa.

### Start e stop

- confirmação de stop;
- etapas reais;
- resultado de sucesso/erro;
- retorno à home;
- atualização automática do status.

### Ferramentas

- uma linha por linguagem;
- descrição curta;
- estado instalado;
- seleção individual;
- instalação da selecionada;
- tela de resultado;
- log da instalação.

### Sistema e atualização

- versão atual;
- canal e arquitetura;
- verificar atualização;
- confirmar aplicação;
- acompanhar atualização;
- informar resultado e erro.

### Logs

- setup;
- start/stop;
- SSH;
- instalação;
- atualização;
- leitura limitada para não estourar a tela;
- nenhum segredo exibido.

## Fase 6: testes

- contratos JSON válidos e únicos;
- respostas com erro;
- start sem abrir shell;
- stop idempotente;
- navegação e confirmação;
- scroll em largura pequena;
- seleção de ferramenta;
- cancelamento;
- shell e restauração do terminal;
- renderização dos estados principais.

## Ordem de entrega

1. núcleo de execução e contratos;
2. PTY;
3. modelo/navegação;
4. viewport e estilos;
5. home/status/start/stop;
6. setup e instalação;
7. sistema/update/logs;
8. testes de integração;
9. validação em Termux ARM64.

## Fora deste plano

- novos perfis de linguagem;
- interface web;
- APK;
- desktop gráfico;
- múltiplos usuários.
