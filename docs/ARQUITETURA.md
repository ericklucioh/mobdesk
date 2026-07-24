# Arquitetura base do Mobdesk

Este documento registra somente a fundação técnica do projeto. Ele descreve
onde o software roda, como as camadas se relacionam, quais são as fronteiras
de execução e quais limitações vêm do Android e do PRoot.

Regras de negócio, catálogo de ferramentas, telas, roadmap e decisões de
produto devem ser documentados separadamente.

## Topologia de execução

```text
Android / HyperOS
└── Termux
    ├── binário Mobdesk (Go)
    ├── ferramentas e processos do host
    └── PRoot-Distro
        └── Ubuntu ARM64 persistente
            └── processos do ambiente Linux do usuário
```

### Android e HyperOS

O Android fornece o kernel, a rede, o armazenamento, a bateria e as políticas
de suspensão. O HyperOS pode encerrar ou restringir o Termux mesmo quando um
processo ainda estiver em execução.

### Termux

O Termux é o host do Mobdesk. Ele fornece:

- o ambiente de execução do binário Go;
- o `PATH` e os pacotes nativos do host;
- acesso aos comandos e recursos disponíveis no Android;
- o processo SSH gerenciado pelo Mobdesk;
- a execução e a entrada no Ubuntu via PRoot-Distro;
- o armazenamento privado usado pelo estado do aplicativo.

### Ubuntu via PRoot

O Ubuntu é um userland Linux persistente executado a partir do Termux. Ele
fornece filesystem, bibliotecas e ferramentas de distribuição Linux, mas não
possui um kernel separado.

PRoot não é uma VM, um container com isolamento real ou um sistema operacional
independente. Os processos continuam sujeitos ao kernel e às políticas do
Android.

## Camadas do código

```text
cmd/mobdesk
└── internal/cobra       entrada e comandos da CLI
    ├── internal/status  coleta e modelo do estado
    ├── internal/install instalação de perfis de ferramentas
    ├── internal/update  consulta e aplicação de atualizações
    ├── internal/logs    leitura de registros de operações
    └── internal/version metadados de versão

internal/tui             apresentação interativa Bubble Tea
└── backend de comunicação
    ├── backend real: chama o executável e interpreta JSON
    └── backend mock: simula respostas para testes manuais
```

### Entrada e orquestração

`cmd/mobdesk` inicia a aplicação. `internal/cobra` registra os comandos,
interpreta argumentos e coordena os serviços internos.

Os serviços não devem depender da renderização da TUI. A saída humana e a
saída JSON são responsabilidades da camada de apresentação da CLI.

### TUI

`internal/tui` é uma camada de apresentação. Ela usa Bubble Tea para o ciclo
de eventos, Bubbles para componentes interativos e Lip Gloss para estilos.

A TUI não deve duplicar instalação, coleta, atualização ou regras de segurança.
No backend real, ela se comunica com o contrato JSON da própria CLI. O backend
mock implementa a mesma interface apenas para cenários de teste visual.

### Serviços internos

- `internal/status`: coleta uma fotografia do ambiente e produz o modelo de
  estado compartilhado pelo CLI e pela TUI;
- `internal/install`: resolve perfis, executa instalações idempotentes e grava
  seus registros;
- `internal/update`: consulta e aplica atualizações do Mobdesk;
- `internal/logs`: lê registros persistidos sem criar uma tela própria;
- `internal/version`: fornece metadados de versão do binário.

Novas camadas só devem ser criadas quando houver comportamento real que exija
separação.

## Fronteiras de execução

Existem dois ambientes de processo:

```text
Mobdesk no Termux
    ├── comandos do host
    │   └── pkg, sshd e ferramentas Termux
    └── comandos do Ubuntu
        └── proot-distro login ubuntu -- ...
```

Cada comando deve deixar claro em qual ambiente será executado. A aplicação
não deve tratar um processo do Ubuntu como se fosse um processo nativo do
Termux.

Processos simples usam `os/exec` com contexto e cancelamento. Shells,
editores e outros programas interativos exigem PTY, encaminhamento de entrada
e saída e restauração segura do terminal.

Entrada do usuário não deve ser concatenada em comandos sem validação.

## Estado e armazenamento

O estado persistente do Mobdesk fica em diretórios privados do usuário no
Termux. Esses diretórios podem conter configuração, estado de etapas,
registros de instalação e arquivos relacionados ao SSH.

Regras estruturais:

- permissões devem restringir arquivos privados;
- segredos não entram no código, no Git ou em logs;
- estado do host e estado do Ubuntu devem permanecer distinguíveis;
- dependências compiladas para Termux e Ubuntu não devem ser misturadas;
- projetos e dados do usuário devem sobreviver à repetição das operações;
- armazenamento externo do Android não deve ser tratado como filesystem Unix
  completo sem validação.

## Contratos entre camadas

Os comandos da CLI são a fronteira pública de execução. Quando a TUI chama o
backend real, a resposta esperada é uma mensagem final, normalmente derivada
de JSON no stdout. Mensagens auxiliares e erros não devem corromper JSON.

O fluxo normal é:

```text
evento da TUI
  -> backend
  -> comando Cobra ou mock
  -> serviço interno
  -> resultado final
  -> mensagem Bubble Tea
  -> novo estado renderizado
```

O fluxo real não depende de streaming de progresso ou de polling contínuo.
Qualquer estado exibido durante uma operação deve ser tratado como espera até
que exista uma resposta final confiável.

## Limitações da plataforma

O projeto não deve assumir:

- `systemd` funcionando dentro do PRoot;
- namespaces, cgroups, seccomp ou módulos de kernel disponíveis;
- Docker real ou isolamento de container;
- acesso privilegiado a dispositivos Android;
- aceleração gráfica garantida;
- execução contínua após suspensão ou encerramento pelo HyperOS;
- desempenho de produção pesada, carga ou processamento prolongado.

Essas limitações pertencem à plataforma e não devem ser escondidas por uma
abstração que prometa capacidades inexistentes.

## Segurança estrutural

- o SSH deve controlar apenas a própria instância do Mobdesk;
- portas e processos externos não devem ser encerrados sem comprovação de
  propriedade;
- acesso remoto deve preferir rede local ou túnel seguro;
- arquivos de configuração, estado e credenciais devem ser privados;
- operações destrutivas devem ter confirmação explícita;
- comandos longos devem aceitar contexto e cancelamento;
- falhas parciais não devem apagar dados nem deixar processos órfãos.

## Verificação da base

As mudanças estruturais devem ser verificadas com:

```bash
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
go build ./cmd/mobdesk
```

Testes locais validam a lógica e os contratos. A validação definitiva da
integração com Termux, Android, PRoot e HyperOS precisa ocorrer em um aparelho
real.
