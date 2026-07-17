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
- servidor SSH do Termux na porta `8022`;
- acesso remoto direto ao Ubuntu;
- detecção do IP local via `ifconfig`;
- autenticação por senha;
- comandos `setup`, `start` e `stop`;
- execução no celular ou remotamente pelo computador;
- ambiente reproduzível para desenvolvimento e testes.

O MVP-1 é deliberadamente pequeno. Ele ainda não instala ferramentas de desenvolvimento no Ubuntu, não oferece TUI e não gerencia projetos. Essas capacidades fazem parte dos próximos estágios.

## Instalação no Termux

Instale o Termux por uma fonte confiável e abra o aplicativo. Depois:

```bash
pkg update
pkg upgrade -y
pkg install -y golang git
go install github.com/ericklucioh/mobdesk/cmd/mobdesk@latest
./go/bin/mobdesk setup
```

Na primeira execução, o binário é chamado diretamente pelo caminho criado pelo Go. Depois do setup, o launcher global permite usar `mobdesk` normalmente.

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
```

Para apagar o ambiente persistente e começar do zero:

```bash
make reset-env
```

Esse comando remove os volumes do Termux/Ubuntu. O código local não é apagado. A instalação do Ubuntu ocupa aproximadamente `1,5 GB` nos volumes persistentes.

Consulte [CONTRIBUINDO.md](docs/CONTRIBUINDO.md) antes de enviar alterações.

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
2. instalação de ferramentas de desenvolvimento sob demanda;
3. sessões persistentes, projetos e serviços;
4. central de gerenciamento acessível pelo navegador.

Veja o [roadmap em seis estágios](docs/project/estagios_mobdesk.md).

## Documentação

- [MVP-1](docs/project/MVP.md) — escopo e funcionamento atual;
- [Missão](docs/project/MISSAO.md) — problema, público e valor;
- [Estágios](docs/project/estagios_mobdesk.md) — evolução do produto;
- [Arquitetura](docs/project/ARQUITETURA.md) — camadas e limites técnicos;
- [Decisões](docs/project/DECISOES.md) — decisões do projeto;
- [Ferramentas](docs/project/FERRAMENTAS.md) — catálogo técnico;
- [Como contribuir](docs/CONTRIBUINDO.md) — fluxo para colaboradores.
