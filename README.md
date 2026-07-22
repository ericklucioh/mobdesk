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
- comandos `setup`, `start`, `stop`, `shell`, `status` e `install`;
- instalação idempotente de Go, Python, Node.js, C, C++ e Lua no Ubuntu;
- status humano e JSON para automação e futura TUI;
- execução no celular ou remotamente pelo computador;
- ambiente reproduzível para desenvolvimento e testes.

O MVP-1 é deliberadamente pequeno. Ele ainda não oferece TUI nem gerencia
projetos. Essas capacidades fazem parte dos próximos estágios.

Consulte rapidamente o ambiente com:

```bash
mobdesk status
mobdesk status --json
```

O status é somente leitura e verifica setup, Ubuntu, SSH, rede, espaço livre
do dispositivo, bateria e Wi-Fi quando o Termux:API estiver disponível.

Instale uma linguagem no Ubuntu com:

```bash
mobdesk install go
mobdesk install python
mobdesk install node
mobdesk install c
mobdesk install cpp
mobdesk install lua
```

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
make test
make vet
make build
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

1. Workstation TUI para setup, start, stop e diagnóstico;
2. instalação de ferramentas e organização do trabalho;
3. sessões persistentes, projetos, serviços e acesso remoto confiável;
4. central local de gerenciamento;
5. interface de gerenciamento acessível pelo navegador;
6. plataforma reproduzível e extensível.

Veja o [roadmap em seis estágios](docs/project/ROADMAP.md).

## Documentação

- [MVP-1](docs/project/MVP.md) — escopo e funcionamento atual;
- [Missão](docs/project/MISSAO.md) — problema, público e valor;
- [Roadmap](docs/project/ROADMAP.md) — evolução do produto;
- [Arquitetura](docs/project/ARQUITETURA.md) — camadas e limites técnicos;
- [Decisões](docs/project/DECISOES.md) — decisões do projeto;
- [Ferramentas](docs/project/FERRAMENTAS.md) — catálogo técnico;
- [Como contribuir](.github/CONTRIBUTING.md) — fluxo para colaboradores.

## Licença

Distribuído sob a [licença MIT](LICENSE).
