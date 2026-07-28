# Mobdesk

Transforme seu celular Android em uma workstation Ubuntu pessoal.

O Mobdesk permite levar para a faculdade, viagens ou qualquer lugar um ambiente Linux próprio, sem depender de computadores compartilhados e sem deixar suas contas pessoais conectadas neles.

No MVP atual, o fluxo é simples:

```text
Termux → Mobdesk → SSH → Ubuntu via PRoot
```

O Termux controla o aparelho. O Ubuntu persistente é o ambiente de trabalho. Ao conectar por SSH, você entra diretamente no Ubuntu.

## O que já funciona

- instalação automatizada a partir de um Termux praticamente virgem;
- Ubuntu persistente via PRoot-Distro;
- servidor SSH próprio do Mobdesk na porta `8022`;
- acesso remoto direto ao Ubuntu;
- detecção do IP local via `ifconfig`;
- autenticação por senha;
- comandos `setup`, `start`, `stop`, `shell`, `status`, `install`, `update` e `tui`;
- instalação idempotente de Go, Python, Node.js, C, C++ e Lua no Ubuntu;
- status humano e JSON para automação e TUI;
- TUI com status, setup, ferramentas, shell e atualização;
- versão local e atualização verificável do binário, com rollback automático;
- execução no celular ou remotamente pelo computador;
- ambiente reproduzível para desenvolvimento e testes.

O MVP continua deliberadamente pequeno. A TUI organiza as operações existentes,
mas projetos, serviços, sessões persistentes e interface web permanecem para os
próximos estágios.

Consulte rapidamente o ambiente com:

```bash
mobdesk status
mobdesk status --json
```

O status é somente leitura e verifica setup, Ubuntu, SSH, rede, espaço livre
do dispositivo, bateria e Wi-Fi quando o Termux:API estiver disponível.

## TUI

Abra a interface textual no Termux com:

```bash
mobdesk tui
```

A TUI oferece status, setup, instalação de ferramentas, shell Ubuntu e
atualização do Mobdesk. Quando ela é aberta por SSH, já está dentro do Ubuntu:
nesse modo, mostra o workspace e permite abrir o shell local, mas não oferece
ações que exigem o host Termux, como controlar SSH, executar setup, instalar
ferramentas ou atualizar o binário.

Instale uma linguagem no Ubuntu com:

```bash
mobdesk install go
mobdesk install python
mobdesk install node
mobdesk install c
mobdesk install cpp
mobdesk install lua
```

Consulte a versão compilada e verifique atualizações:

```bash
mobdesk version --json
mobdesk update --check
mobdesk update
```

Builds locais aparecem como `dev` e podem ser atualizados para a última release
estável. Builds do canal `stage` procuram releases `test-v*`.

O update baixa o binário em arquivo temporário, confere seu SHA-256, testa
`version --json` e só então o ativa. Se a ativação falhar, a versão anterior é
restaurada automaticamente. O backup é administrado pelo Mobdesk; não é
necessário mover ou remover arquivos manualmente. O checksum detecta downloads
corrompidos, mas não autentica uma release comprometida que publique binário e
checksum correspondentes.

## Instalação no Termux

Instale o Termux por uma fonte confiável e abra o aplicativo. Depois:

```bash
pkg update
pkg install -y golang git
go install github.com/ericklucioh/mobdesk/cmd/mobdesk@latest
./go/bin/mobdesk setup
```

Na primeira execução, o binário é chamado diretamente pelo caminho criado pelo Go. Depois do setup, o launcher global permite usar `mobdesk` normalmente.

O setup instala apenas as dependências necessárias. Para atualizar todo o ambiente Termux explicitamente, use `mobdesk setup --upgrade-system`.

O setup pode ser executado novamente depois de uma falha. Ele mantém as etapas
concluídas, recupera a etapa pendente e não apaga Ubuntu, workspace ou projetos.
Somente uma execução de setup pode modificar o ambiente por vez.

Durante o setup, o Mobdesk irá:

- instalar `proot-distro`, `openssh` e `net-tools`;
- baixar o Ubuntu base;
- criar o workspace persistente;
- configurar a senha SSH;
- preparar o acesso SSH direto ao Ubuntu.

## Usando o Mobdesk

Inicie a workstation:

```bash
mobdesk start
```

Para abrir o Ubuntu localmente sem iniciar o servidor SSH:

```bash
mobdesk shell
```

O Mobdesk exibirá um comando parecido com:

```bash
ssh -p 8022 android@192.168.3.228
```

Execute esse comando em outro computador conectado à mesma rede. A sessão SSH abrirá diretamente no Ubuntu.

Para encerrar o servidor SSH:

```bash
mobdesk stop
```

Para sair do Ubuntu sem parar o servidor:

```bash
exit
```

O IP local pode mudar quando o celular troca de rede. O Termux precisa permanecer ativo e o Android não pode suspender o aplicativo durante o uso remoto.

## Desenvolvimento

Clone o projeto e entre no diretório:

```bash
git clone https://github.com/ericklucioh/mobdesk.git
cd mobdesk
```

O ambiente Docker simula o userland do Termux e mantém o workspace e o prefixo em volumes persistentes.

```bash
make build-image  # constrói a imagem local
make dev          # inicia o Air com hot-reload
make termux       # abre um shell Termux com a porta SSH publicada
make shell        # abre outro shell no ambiente
```

Verificações:

```bash
make check
make integration-test  # smoke test do Termux/SSH no Docker
```

O teste de integração cria volumes descartáveis, instala o Ubuntu, testa `setup`, `start`, acesso SSH e `stop`, e os remove ao terminar. Ele não reproduz bateria, permissões, wake-lock ou o kernel do Android.

Para apagar o ambiente persistente e começar do zero:

```bash
make reset-env
```

Esse comando remove os volumes do Termux/Ubuntu. O código local não é apagado. A instalação do Ubuntu ocupa aproximadamente `1,5 GB` nos volumes persistentes.

Consulte [CONTRIBUINDO.md](.github/CONTRIBUTING.md) antes de enviar alterações.

## Arquitetura

```text
Android/HyperOS
└── Termux
    ├── Mobdesk em Go
    ├── OpenSSH :8022
    └── PRoot-Distro
        └── Ubuntu ARM64 persistente
```

O projeto não depende de root, VM ou Docker no celular. PRoot melhora a compatibilidade do userland, mas não fornece um kernel Linux separado nem isolamento real de container.

## Próximos estágios

1. consolidar a TUI e validar o fluxo em um aparelho Termux real;
2. sessões persistentes, projetos, serviços e acesso remoto confiável;
3. central local de gerenciamento de projetos, sessões e serviços.

Veja o [roadmap em quatro estágios](docs/ROADMAP.md).

## Documentação

- [Missão](docs/MISSAO.md) — problema, público e valor;
- [Roadmap](docs/ROADMAP.md) — evolução do produto;
- [Arquitetura](docs/ARQUITETURA.md) — camadas e limites técnicos;
- [Decisões](docs/DECISOES.md) — decisões do projeto;
- [Ferramentas](docs/ideias/FERRAMENTAS.md) — catálogo técnico em evolução;
- [Refatoração prioritária](docs/PLANO-REFATORACAO-PRIORITARIA.md) — melhorias estruturais planejadas;
- [Como contribuir](.github/CONTRIBUTING.md) — fluxo para colaboradores.

## Licença

Distribuído sob a [licença MIT](LICENSE).
