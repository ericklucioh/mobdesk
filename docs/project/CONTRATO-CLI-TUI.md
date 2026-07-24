# Contrato inicial da CLI e proposta de integração com a TUI

Este documento define o contrato inicial da CLI do Mobdesk e como suas ações
devem ser apresentadas futuramente pela TUI.

O contrato deve ser pequeno, previsível e reutilizável. A CLI e a TUI não devem
implementar regras diferentes: ambas devem chamar os mesmos serviços internos.

## Escopo inicial

O contrato inicial terá somente estes comandos:

```text
mobdesk setup
mobdesk start
mobdesk stop
mobdesk status
mobdesk doctor
mobdesk install <linguagem>
```

## Decisões de escopo

### Shell

Não haverá `mobdesk shell` neste contrato inicial.

O acesso ao Ubuntu continuará acontecendo por:

- `mobdesk start`, que pode abrir o shell local;
- uma conexão SSH na porta `8022`;
- a TUI, quando houver uma ação explícita para abrir o ambiente.

Um comando separado para shell adicionaria pouca utilidade ao fluxo atual.

### Projetos

Não haverá grupo `mobdesk project` no contrato inicial.

O usuário poderá trabalhar no workspace pelo shell Ubuntu ou por SSH. Criação,
listagem e execução de projetos não fazem parte da primeira CLI/TUI.

### Sessões

Não haverá grupo `mobdesk session` no contrato inicial.

Persistência e reconexão de sessões podem ser implementadas internamente com
tmux ou outro mecanismo quando houver necessidade validada, sem expor isso
como uma API pública neste primeiro contrato.

Essa decisão remove comandos da primeira interface; não impede que projetos e
sessões sejam recursos futuros do produto.

## Comandos da CLI

### `mobdesk setup`

Prepara o Termux, o Ubuntu persistente e o SSH dedicado.

```bash
mobdesk setup
mobdesk setup --upgrade-system
```

Regras:

- pode ser executado novamente;
- retoma etapas concluídas;
- não apaga dados do usuário;
- registra o estado de cada etapa;
- pede confirmação antes de ações relevantes;
- `--upgrade-system` permanece uma opção explícita.

### `mobdesk start`

Inicia a workstation, ativa recursos necessários e inicia o SSH.

```bash
mobdesk start
```

O comando deve:

- verificar o setup;
- verificar o Ubuntu;
- garantir a configuração SSH;
- iniciar o wake-lock quando disponível;
- iniciar ou reutilizar o SSH do Mobdesk;
- mostrar os endereços de acesso;
- retornar ao chamador depois de iniciar a workstation; o shell Ubuntu é uma ação explícita separada.

O modo JSON é um resultado único por execução:

```bash
mobdesk start --json
```

Flags futuras possíveis:

```bash
mobdesk start --no-shell
mobdesk start --foreground
```

Essas flags só devem ser adicionadas quando houver necessidade real.

### `mobdesk stop`

Para o SSH do Mobdesk e libera recursos associados.

```bash
mobdesk stop
mobdesk stop --json
```

Regras:

- só pode controlar o processo pertencente ao Mobdesk;
- não deve encerrar processos externos;
- deve tolerar PID obsoleto;
- deve liberar o wake-lock quando aplicável;
- deve ser seguro repetir o comando.

### `mobdesk status`

Apresenta uma fotografia do ambiente.

```bash
mobdesk status
mobdesk status --json
```

Deve informar, conforme o estado disponível:

- versão do Mobdesk;
- arquitetura e host Termux;
- espaço do aparelho;
- espaço ocupado pelo Termux;
- espaço ocupado pelo Ubuntu;
- estado do setup;
- estado do Ubuntu;
- estado do SSH;
- endereços de rede;
- wake-lock;
- linguagens instaladas;
- avisos e erros resumidos.

O modo JSON deve produzir somente JSON no stdout e não pode expor segredos.

### `mobdesk doctor`

Executa diagnóstico detalhado e apresenta evidências e sugestões.

```bash
mobdesk doctor
mobdesk doctor --json
mobdesk doctor --deep
mobdesk doctor --strict
mobdesk doctor --fix
```

O modo padrão é somente leitura. `--fix` só poderá aplicar correções seguras,
reversíveis e confirmadas pelo usuário.

O diagnóstico deve cobrir:

- Termux e arquitetura;
- permissões;
- armazenamento;
- etapas do setup;
- Ubuntu via PRoot;
- configuração e processo SSH;
- rede;
- toolchains instaladas;
- segurança básica.

### `mobdesk install <linguagem>`

Instala uma linguagem selecionada pelo usuário.

```bash
mobdesk install go
mobdesk install node
mobdesk install python
mobdesk install c
mobdesk install cpp
mobdesk install java
mobdesk install kotlin
mobdesk install lua
mobdesk install php
mobdesk install ruby
mobdesk install rust
```

O perfil combinado para C e C++ pode ser exposto como:

```bash
mobdesk install native
```

Cada instalação deve:

- verificar arquitetura;
- verificar espaço disponível;
- verificar se já está instalada;
- instalar apenas o perfil escolhido;
- registrar logs e estado;
- validar a versão;
- executar um teste mínimo;
- preservar os projetos do usuário;
- poder ser retomada após falha.

## Catálogo inicial de linguagens

| Perfil | Pacotes mínimos | Verificação mínima |
|---|---|---|
| `go` | `golang` | `go version` e `go run` |
| `node` | `nodejs` ou `nodejs-lts`, `npm` | `node`, `npm` e script JavaScript |
| `python` | `python` | `python --version` e script Python |
| `c` | `clang` | compilar programa C |
| `cpp` | `clang` | compilar programa C++ |
| `native` | `clang` | validar C e C++ |
| `java` | `openjdk-17` | `javac` e `java` |
| `kotlin` | `openjdk-17`, `kotlinc` | compilar e executar Kotlin/JVM |
| `lua` | `lua54` | executar script Lua |
| `php` | `php` | executar script PHP |
| `ruby` | `ruby` | executar script Ruby |
| `rust` | `rust` | `rustc` e `cargo` |

As linguagens não são instaladas automaticamente. O setup inicial instala
somente as dependências necessárias para o próprio Mobdesk, Termux, SSH e
Ubuntu base.

## Flags globais

O conjunto inicial de flags globais deve ser pequeno:

```bash
mobdesk --help
mobdesk --version
mobdesk --json <comando>
mobdesk --quiet <comando>
mobdesk --verbose <comando>
mobdesk --no-color <comando>
```

Nem todo comando precisa aceitar todas as flags. A CLI deve rejeitar
combinações sem significado com uma mensagem clara.

## Regras de saída

### Texto humano

É a saída padrão para uso direto no Termux:

```bash
mobdesk status
```

Deve ser legível no terminal, funcionar sem cores e indicar claramente erros,
avisos e ações pendentes.

### JSON

É a saída para automação, testes e integração futura:

```bash
mobdesk status --json
mobdesk doctor --json
```

Regras:

- stdout contém apenas o resultado estruturado;
- stderr contém mensagens auxiliares;
- tamanhos são representados em bytes;
- datas usam formato ISO 8601/RFC 3339;
- objetos incluem `schema_version` quando forem contratos públicos;
- nenhum segredo aparece no resultado.

## Códigos de saída

```text
0  sucesso ou diagnóstico sem erro crítico
1  erro operacional ou componente essencial indisponível
2  argumento ou uso inválido
3  resultado parcial em modo estrito
4  ação cancelada ou correção não aplicada
```

Uma linguagem ausente não deve fazer `status` ou `doctor` falhar globalmente.
Ela só será erro quando o usuário solicitar explicitamente sua instalação ou
verificação.

## Regras gerais

### Idempotência

Os comandos abaixo devem aceitar repetição segura:

```text
setup
start
stop
status
doctor
install
```

Repetir um comando não deve reinstalar, apagar ou duplicar recursos sem
necessidade.

### Confirmações

Exigir confirmação para:

- remover Ubuntu;
- apagar projetos ou caches;
- revogar chaves;
- trocar senha;
- matar processos;
- alterar a porta SSH;
- remover uma linguagem.

### Cancelamento

Comandos longos devem respeitar `Ctrl+C`, cancelar processos filhos, liberar
locks, preservar o estado anterior e registrar a etapa interrompida.

### Segurança

- nunca imprimir senhas, tokens ou chaves privadas;
- não concatenar entrada do usuário em comandos sem validação;
- não expor SSH diretamente na internet;
- manter arquivos de estado em diretórios privados;
- não gravar segredos nos logs.

## Proposta de navegação da TUI

A TUI deve apresentar as mesmas capacidades como ações visuais:

```text
Mobdesk
├── Início
│   ├── estado geral
│   ├── espaço livre
│   ├── Ubuntu
│   ├── SSH
│   └── alertas
├── Configurar
│   └── setup
├── Iniciar
│   └── start
├── Parar
│   └── stop
├── Linguagens
│   ├── instaladas
│   ├── disponíveis
│   └── instalar selecionadas
└── Configurações
```

### Tela inicial

Deve mostrar somente informações de alto valor:

```text
Mobdesk

Estado:       pronto
Ubuntu:       disponível
SSH:          ativo na porta 8022
IP:           192.168.3.228
Armazenamento: 165 GB livres
Linguagens:   3 instaladas

[S] Status detalhado
[I] Instalar linguagem
[T] Iniciar/parar
[Q] Sair
```

### Tela de instalação

```text
Escolha as linguagens:

[x] Go
[x] Python
[ ] Node.js
[ ] Java
[ ] Kotlin
[ ] C/C++
[ ] Lua
[ ] PHP
[ ] Ruby
[ ] Rust

Espaço estimado: 420 MB

[Enter] Instalar  [Esc] Cancelar
```

## Arquitetura compartilhada

CLI e TUI devem usar os mesmos serviços:

```text
cmd/mobdesk
  └── internal/cobra
        └── serviços internos
              ├── setup
              ├── runtime
              ├── install
              ├── status
              └── doctor

internal/tui
  └── chama os mesmos serviços e renderiza seus resultados
```

Nenhuma regra de instalação, diagnóstico ou segurança deve existir somente na
TUI.

## Ordem de implementação

### Fase 1 — consolidar comandos atuais

- manter `setup`, `start` e `stop`;
- remover a necessidade de `shell` como comando separado;
- documentar os códigos de saída;
- padronizar mensagens e cancelamento.

### Fase 2 — status

- `internal/status` implementado;
- `status` e `status --json` implementados;
- coleta rápida de host, espaço livre, setup, Ubuntu, SSH, rede, bateria e Wi-Fi;
- falhas parciais representadas no JSON sem interromper toda a coleta;
- medições recursivas e toolchains ficam fora do status rápido.

### Fase 3 — doctor

- criar `internal/doctor`;
- registrar checks;
- adicionar evidências e sugestões;
- adicionar `doctor --json`, `--deep` e `--strict`;
- implementar somente correções seguras posteriormente.

### Fase 4 — install

- criar catálogo de perfis;
- implementar instalação idempotente;
- validar versões;
- executar testes mínimos das linguagens;
- registrar estado e logs.

### Fase 5 — TUI inicial

- criar tela inicial;
- integrar status;
- integrar instalação de linguagens;
- integrar setup, start e stop;
- manter o shell local pelo fluxo de start, sem criar comando `shell`.

## Testes do contrato

- `--help` lista apenas os comandos do contrato inicial;
- `--version` retorna a versão sem efeitos colaterais;
- argumentos inválidos retornam código `2`;
- `status --json` produz JSON válido;
- `doctor --json` produz JSON válido mesmo com falha parcial;
- `setup` pode ser retomado;
- `start` e `stop` são idempotentes;
- `install` não reinstala perfil já pronto;
- `Ctrl+C` não deixa locks ou processos órfãos;
- stdout e stderr seguem o contrato;
- a TUI apresenta as mesmas ações e estados da CLI.

## Critério de conclusão

O contrato inicial estará pronto quando:

- os seis comandos definidos tiverem comportamento documentado;
- CLI e TUI compartilharem serviços internos;
- não existir dependência de `shell`, `project` ou `session` para o fluxo básico;
- `status`, `doctor` e `install` tiverem saída previsível;
- comandos longos forem canceláveis;
- ações destrutivas exigirem confirmação;
- o fluxo `setup -> start -> SSH/Ubuntu -> stop` continuar funcionando;
- o contrato estiver validado no Termux ARM64 real.
