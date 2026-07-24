# Plano de execução do `mobdesk doctor`

O `mobdesk doctor` é o diagnóstico profundo do Mobdesk. Ele reutiliza os
coletores do `status`, executa verificações adicionais, explica problemas e
indica os próximos passos para corrigi-los.

## Diferença entre `status` e `doctor`

```text
status  -> fotografia rápida do ambiente
doctor  -> diagnóstico, evidências e sugestões de correção
```

O `status` responde “como está?”. O `doctor` responde “o que está errado, por
que pode estar errado e o que posso fazer?”.

## Objetivos

- diagnosticar falhas do setup;
- verificar Termux, Ubuntu via PRoot, SSH, rede e armazenamento;
- detectar linguagens ausentes, incompletas ou quebradas;
- diferenciar erro de código, ambiente e runtime;
- apresentar evidências compreensíveis;
- sugerir correções sem alterar o sistema por padrão;
- continuar funcionando quando uma verificação individual falhar;
- oferecer saída humana e JSON para automação.

## Fora do escopo inicial

- corrigir tudo automaticamente;
- remover Ubuntu, projetos ou caches sem confirmação;
- monitorar continuamente todos os processos;
- substituir logs detalhados;
- garantir compatibilidade com todas as bibliotecas de cada linguagem;
- configurar Android Studio, emulador ou SDK Android completo;
- diagnosticar projetos e serviços antes de esses recursos existirem.

## Interface

### Diagnóstico padrão

```bash
mobdesk doctor
```

O modo padrão é somente leitura. Pode executar testes, mas não deve instalar,
remover, matar processos ou alterar configurações.

### Saída JSON

```bash
mobdesk doctor --json
```

O stdout deve conter apenas JSON válido. Mensagens auxiliares devem ir para
stderr.

### Diagnóstico profundo

```bash
mobdesk doctor --deep
```

O modo profundo pode medir todo o rootfs, testar login no Ubuntu, executar
testes das toolchains e validar detalhadamente o SSH. Pode ser mais lento.

### Correções seguras

```bash
mobdesk doctor --fix
```

Só aplica ações reversíveis previamente classificadas como seguras, sempre com
confirmação do usuário.

### Modo estrito

```bash
mobdesk doctor --strict
```

Útil para scripts. Avisos relevantes ou verificações desconhecidas podem causar
falha no código de saída.

As flags podem ser combinadas:

```bash
mobdesk doctor --deep --json --strict
```

## Saída humana esperada

```text
Mobdesk doctor

Host
  [OK] arquitetura ARM64 compatível
  [OK] Termux acessível
  [WARN] termux-wake-lock não está disponível

Armazenamento
  [OK] 165 GB livres
  [WARN] Ubuntu ocupa 5.8 GB

Ubuntu
  [OK] rootfs encontrado
  [OK] login via PRoot funcionando
  [OK] workspace disponível

SSH
  [OK] configuração válida
  [ERROR] porta 8022 ocupada por outro processo

Linguagens
  [OK] Go instalado e hello world passou
  [OK] Python instalado e funcionando
  [INFO] Java não instalado
  [WARN] Kotlin incompleto: kotlinc ausente

Resumo
  OK: 8 | INFO: 1 | WARN: 2 | ERROR: 1

Sugestão principal:
  libere a porta 8022 ou configure outra porta antes de executar mobdesk start.
```

## Modelo de uma verificação

Cada verificação deve produzir um resultado estruturado:

```go
type CheckStatus string

const (
	CheckOK      CheckStatus = "ok"
	CheckInfo    CheckStatus = "info"
	CheckWarning CheckStatus = "warning"
	CheckError   CheckStatus = "error"
	CheckUnknown CheckStatus = "unknown"
)

type CheckResult struct {
	ID          string      `json:"id"`
	Category    string      `json:"category"`
	Status      CheckStatus `json:"status"`
	Summary     string      `json:"summary"`
	Evidence    []string    `json:"evidence,omitempty"`
	Suggestions []string    `json:"suggestions,omitempty"`
	Fixable     bool        `json:"fixable"`
	FixApplied  bool        `json:"fix_applied,omitempty"`
	ErrorCode   string      `json:"error_code,omitempty"`
}
```

Uma falha não deve interromper o relatório inteiro. Se a rede falhar, o
diagnóstico ainda deve verificar armazenamento, Ubuntu e toolchains.

## Categorias de diagnóstico

### Host e Termux

Verificar:

- arquitetura ARM64;
- versão do Android e do Termux, quando disponíveis;
- origem/compatibilidade do Termux, quando detectável;
- `$HOME` e `$PREFIX`;
- permissões de leitura e escrita;
- comandos essenciais;
- disponibilidade e estado do `termux-wake-lock`;
- processos básicos do Mobdesk.

Checks iniciais:

```text
host.architecture
host.termux
host.home
host.prefix
host.permissions
host.commands
host.wakelock
```

### Armazenamento

Verificar:

- total, usado, livre e percentual do filesystem;
- tamanho do `$HOME`;
- tamanho do `$PREFIX`;
- tamanho do Ubuntu;
- tamanho do Mobdesk, logs e caches;
- tamanho do workspace e projetos;
- espaço suficiente para instalar uma linguagem.

Checks iniciais:

```text
storage.device
storage.termux
storage.ubuntu
storage.mobdesk
storage.projects
storage.threshold
```

Ubuntu, workspace e projetos são detalhamento interno do Termux e não podem
ser somados novamente ao total.

### Setup

Verificar:

- `setup.done`;
- marcadores de etapas;
- etapas ausentes ou inconsistentes;
- launcher global;
- senha configurada;
- configuração SSH;
- Ubuntu instalado;
- workspace criado;
- possibilidade de retomar o setup.

Etapas atuais:

```text
directories
packages-updated
system-upgraded
packages-installed
ubuntu-installed
workspace-created
password-configured
ssh-configured
launcher-installed
```

### Ubuntu via PRoot

Verificar:

- instalação e rootfs;
- arquitetura;
- login `proot-distro`;
- execução de comando simples;
- workspace;
- diretórios de configuração;
- permissões;
- espaço disponível;
- último acesso bem-sucedido.

Checks iniciais:

```text
ubuntu.installed
ubuntu.architecture
ubuntu.login
ubuntu.command
ubuntu.workspace
ubuntu.permissions
```

### SSH

Verificar:

- configuração dedicada;
- `sshd -t`;
- host keys;
- PID;
- processo pertencente ao Mobdesk;
- porta `8022`;
- banner SSH;
- endereço de bind;
- senha ou chave configurada;
- conflito com outro processo;
- logs recentes sem exibir conteúdo sensível.

Checks iniciais:

```text
ssh.config
ssh.host_keys
ssh.pid
ssh.process
ssh.port
ssh.banner
ssh.authentication
```

### Rede

Verificar:

- interfaces;
- interface preferida;
- IPv4 local;
- loopback;
- conectividade local;
- conectividade externa, quando solicitada;
- Tailscale opcional;
- porta SSH acessível.

Checks iniciais:

```text
network.interface
network.ip
network.local
network.internet
network.tailscale
```

### Toolchains

Para cada linguagem selecionada, verificar pacote, executável, caminho,
versão, arquitetura, `PATH` e teste mínimo de execução.

```text
go       -> golang
node     -> nodejs ou nodejs-lts, npm
python   -> python
c        -> clang
cpp      -> clang
java     -> openjdk-17
kotlin   -> openjdk-17, kotlinc
lua      -> lua54
php      -> php
ruby     -> ruby
rust     -> rust
```

Checks iniciais:

```text
toolchain.go
toolchain.node
toolchain.python
toolchain.c
toolchain.cpp
toolchain.java
toolchain.kotlin
toolchain.lua
toolchain.php
toolchain.ruby
toolchain.rust
```

Uma linguagem ausente não é erro geral. Ela só deve gerar erro quando o
usuário pedir explicitamente o diagnóstico daquele perfil.

### Segurança

Verificar:

- SSH restrito à rede esperada, quando detectável;
- permissões dos diretórios sensíveis;
- permissões da configuração SSH;
- host keys presentes;
- método de autenticação;
- chaves autorizadas;
- dispositivos pareados;
- tentativas falhas recentes;
- configuração de aprovação futura.

O diagnóstico nunca deve mostrar senha, chave privada, token ou código de
autorização.

## Severidades

```text
info:
  informação que não impede o funcionamento

warning:
  funciona, mas há limitação ou risco

error:
  componente esperado não funciona

critical:
  risco grave ou ambiente não deve continuar operando
```

Exemplos:

- Java ausente: `info`;
- wake-lock indisponível: `warning`;
- Ubuntu instalado, mas inacessível: `error`;
- SSH exposto sem autenticação adequada: `critical`.

## Correções

### Correções seguras

O modo `--fix` poderá, após confirmação:

- recriar diretório privado ausente;
- recriar arquivo de estado não destrutivo;
- regenerar a configuração SSH dedicada;
- recriar o launcher que aponta para o executável correto;
- remover PID obsoleto depois de confirmar que o processo não existe;
- corrigir permissões conhecidas de arquivos do Mobdesk.

Exemplo:

```text
[FIX] Recriar a configuração SSH dedicada?
      Arquivo: $HOME/.config/mobdesk/ssh/sshd_config

      [S] Sim  [N] Não
```

### Correções que exigem confirmação forte

Nunca executar silenciosamente:

- remover Ubuntu;
- apagar projetos ou caches grandes;
- revogar todas as chaves;
- trocar senha;
- matar processos do usuário;
- alterar a porta SSH;
- reinstalar todas as linguagens.

Essas ações devem ser comandos separados ou exigir confirmação explícita com o
nome do alvo.

## Recomendações

Cada erro deve conter, quando possível:

- descrição;
- evidência observada;
- impacto;
- próximo comando ou ação sugerida;
- indicação se é corrigível automaticamente;
- aviso sobre alterações ou risco de perda de dados.

Exemplo:

```text
[ERROR] ssh.port

Problema:
  A porta 8022 está ocupada por outro processo.

Evidência:
  PID 2451 está escutando em 0.0.0.0:8022.

Impacto:
  O Mobdesk não consegue iniciar seu servidor SSH.

Próximos passos:
  - execute mobdesk stop se o processo for do Mobdesk;
  - encerre o processo externo se você o reconhecer;
  - configure outra porta, se necessário.
```

## Códigos de saída

```text
0  nenhuma falha crítica
1  erro em componente essencial
2  argumento ou formato inválido
3  coleta parcial em modo estrito
4  correção solicitada, mas não aplicada
```

No modo normal, linguagens ausentes e recursos opcionais não devem causar
falha. No modo `--strict`, avisos e verificações desconhecidas podem causar
falha conforme a política da CLI.

## Arquitetura de implementação

O `doctor` deve reutilizar o pacote de status:

```text
internal/status
  ├── host collector
  ├── storage collector
  ├── ubuntu collector
  ├── ssh collector
  ├── network collector
  └── toolchain collector

internal/doctor
  ├── check registry
  ├── severity rules
  ├── evidence builder
  ├── suggestion builder
  ├── safe fixes
  └── report renderer
```

Cada check deve ser pequeno e testável:

```go
type Check interface {
	ID() string
	Category() string
	Run(context.Context, Snapshot) CheckResult
}
```

O `Snapshot` pode conter dados já coletados pelo status para evitar executar o
mesmo comando várias vezes.

## Ordem de implementação

### Fase 1 — contrato

- criar `internal/doctor`;
- definir `CheckResult` e severidades;
- definir formato JSON;
- definir códigos de saída;
- definir regras de redaction;
- definir política de erro parcial.

### Fase 2 — host, armazenamento e setup

- arquitetura;
- `$HOME` e `$PREFIX`;
- permissões;
- comandos essenciais;
- espaço livre e ocupado;
- estado das etapas;
- launcher;
- wake-lock.

### Fase 3 — Ubuntu e SSH

- Ubuntu instalado;
- login PRoot;
- workspace;
- configuração SSH;
- host keys;
- PID;
- processo;
- porta;
- banner;
- autenticação.

### Fase 4 — rede e toolchains

- interfaces;
- IPv4;
- conectividade;
- Tailscale opcional;
- linguagens instaladas;
- versões;
- testes mínimos.

### Fase 5 — relatório

- comando `mobdesk doctor`;
- saída humana;
- saída JSON;
- resumo de severidades;
- sugestões;
- códigos de saída;
- ocultação de dados sensíveis.

### Fase 6 — correções seguras

- registrar ações corrigíveis;
- implementar confirmação;
- recriar diretórios e estados seguros;
- regenerar configuração SSH;
- corrigir launcher;
- remover PID obsoleto;
- registrar resultado da correção.

### Fase 7 — integração futura

- integrar sessões, projetos e serviços posteriormente.

## Testes

### Unitários

- check OK, aviso, erro e desconhecido;
- falha parcial sem interromper o relatório;
- classificação de severidade;
- geração de sugestões;
- códigos de saída;
- serialização JSON;
- redaction de senhas e tokens;
- confirmação de correção;
- prevenção de correção destrutiva.

### Integração

- setup completo;
- setup incompleto;
- Ubuntu ausente;
- Ubuntu inacessível;
- SSH parado;
- SSH ativo;
- porta ocupada;
- PID obsoleto;
- configuração SSH inválida;
- IP ausente;
- espaço insuficiente;
- linguagem ausente;
- linguagem instalada;
- toolchain incompleta;
- `doctor --json` válido;
- `doctor --strict` com saída esperada.

### Termux real

- Termux ARM64 limpo;
- setup interrompido;
- Ubuntu real via PRoot;
- Termux sem Termux:API;
- Termux suspenso pelo HyperOS;
- troca de Wi-Fi;
- porta ocupada por processo externo;
- pouco espaço disponível;
- toolchains instaladas;
- correções seguras.

## Critério de conclusão

O `mobdesk doctor` estará pronto para o primeiro MVP quando:

- diagnosticar host, armazenamento, setup, Ubuntu, SSH e rede;
- detectar toolchains instaladas, ausentes e incompletas;
- continuar funcionando diante de falhas parciais;
- explicar evidências e próximos passos;
- não revelar segredos;
- não modificar o sistema no modo padrão;
- produzir JSON válido;
- possuir códigos de saída previsíveis;
- contar corretamente o espaço do Termux sem dupla contagem;
- possuir testes unitários e de integração;
- funcionar no Termux ARM64 real.
