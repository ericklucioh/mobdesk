# Mobdesk — MVP 1

## Objetivo

Transformar um Termux praticamente virgem em um ambiente Ubuntu de desenvolvimento acessível pelo próprio celular e por SSH.

O usuário instala o Mobdesk, executa `setup` uma vez e depois usa `start` para iniciar o ambiente.

## Trajetória do usuário

```text
Termux instalado
    ↓
instalar Go e Mobdesk
    ↓
mobdesk setup
    ↓
mobdesk start
    ↓
Ubuntu iniciado + SSH disponível
```

Bootstrap inicial:

```bash
pkg update
pkg upgrade -y
pkg install -y golang
go install github.com/ericklucioh/mobdesk/cmd/mobdesk@latest
mobdesk setup
mobdesk start
```

O usuário executa somente o bootstrap e os comandos Mobdesk. Os comandos abaixo são a referência operacional que o Mobdesk deve executar internamente.

## `mobdesk setup`

O setup é responsável por preparar todo o ambiente necessário para que `mobdesk start` funcione.

### Termux/host

Comandos de preparação:

```bash
pkg update
pkg upgrade -y
pkg install -y proot-distro openssh tmux curl wget
pkg install -y termux-services
```

Pacotes e serviços a instalar ou verificar:

- `proot-distro`;
- `openssh`;
- `tmux`;
- `curl`;
- `git`, se necessário para operações do host;
- `termux-services`, quando a persistência de serviços for habilitada.

Comandos de SSH:

```bash
whoami
passwd
sshd
ss -tln | grep 8022
ip addr
```

O servidor SSH do Termux deve escutar na porta `8022`. O Mobdesk deve detectar o usuário real e os endereços disponíveis, em vez de assumir um valor fixo.

Comandos de persistência:

```bash
termux-wake-lock
sv-enable sshd
sv up sshd
sv status sshd
tmux new-session -d -s mobdesk
tmux has-session -t mobdesk
tmux attach -t mobdesk
```

Também deve verificar:

- arquitetura `aarch64`;
- espaço disponível;
- versão do Termux;
- diretórios privados do Termux;
- existência do Ubuntu;
- estado de instalações anteriores.

### Ubuntu

Baixar e instalar uma imagem Ubuntu ARM64 persistente via PRoot-Distro.

Comandos de instalação e controle:

```bash
proot-distro list
proot-distro install ubuntu
proot-distro login ubuntu
proot-distro login ubuntu -- /bin/uname -a
proot-distro login ubuntu -- bash -lc 'echo ubuntu-ok'
proot-distro remove ubuntu
```

O `remove` é destrutivo e não deve ser executado pelo fluxo normal do Mobdesk.

Dentro do Ubuntu, instalar ou verificar:

- certificados e ferramentas básicas: `ca-certificates`, `curl`, `wget`, `unzip`, `zip`;
- compilação C/C++: `build-essential`, `pkg-config`;
- versionamento: `git`;
- editor: `neovim`;
- sessões: `tmux`;
- Go: `golang`;
- Python: `python3`, `python3-pip`, `python3-venv`;
- JavaScript: `nodejs`, `npm`;
- Java: OpenJDK LTS disponível no Ubuntu;
- navegação e diagnóstico: `ripgrep`, `fd-find`, `fzf`, `btop` ou `htop`.

Comandos de instalação dentro do Ubuntu:

```bash
apt update
apt upgrade -y
apt install -y \
  ca-certificates curl wget unzip zip \
  build-essential pkg-config \
  git neovim tmux \
  golang \
  python3 python3-pip python3-venv \
  nodejs npm \
  ripgrep fd-find fzf btop
```

Java deve usar o pacote OpenJDK LTS disponível na versão Ubuntu instalada:

```bash
apt search openjdk
apt install -y openjdk-21-jdk
java --version
```

O Mobdesk deve tratar nomes de pacotes variáveis como configuração, pois a disponibilidade pode mudar conforme a versão do Ubuntu.

Comandos de verificação do Ubuntu:

```bash
command -v git
command -v nvim
command -v go
command -v python3
command -v node
command -v npm
command -v java
git --version
go version
python3 --version
node --version
npm --version
```

O Mobdesk deve instalar apenas o conjunto essencial no MVP e permitir que ferramentas adicionais sejam adicionadas depois.

### Configuração

O setup deve:

- criar o estado do Mobdesk;
- criar diretórios de projetos, logs e configurações;
- configurar o usuário do Ubuntu;
- preparar o `sshd` do Termux;
- gerar ou verificar chaves do servidor;
- preparar tmux;
- registrar versões instaladas;
- ser idempotente;
- poder continuar após uma falha;
- mostrar progresso e erros compreensíveis.

Comandos de diretórios e estado:

```bash
mkdir -p "$HOME/.local/share/mobdesk" \
  "$HOME/.local/share/mobdesk/logs" \
  "$HOME/.local/share/mobdesk/config"

proot-distro login ubuntu -- bash -lc \
  'mkdir -p /root/workspace /root/.config/mobdesk /root/.local/share/mobdesk'
```

Ao finalizar:

```text
Setup concluído.
Ubuntu instalado.
Ferramentas básicas configuradas.
SSH preparado.

Execute: mobdesk start
```

## `mobdesk start`

O start não deve reinstalar ferramentas. Ele deve:

1. verificar se o setup foi concluído;
2. iniciar o `sshd` no Termux;
3. detectar usuário e IP do celular;
4. verificar se o Ubuntu está disponível;
5. iniciar uma sessão Ubuntu via PRoot;
6. exibir o comando de acesso remoto;
7. abrir o shell Ubuntu no celular.

Comandos principais encapsulados pelo `start`:

```bash
termux-wake-lock
sshd
proot-distro login ubuntu -- bash -lc 'tmux attach -t mobdesk || tmux new -s mobdesk'
```

Para iniciar um projeto sem abrir um shell interativo:

```bash
proot-distro login ubuntu -- bash -lc \
  'cd /root/workspace/projeto && npm run dev'
```

Mensagem esperada:

```text
Servidor iniciado!
Ubuntu iniciado!

Acesse de outro computador:

ssh -p 8022 usuario@IP_DO_CELULAR
```

O SSH permanece no Termux e o ambiente de trabalho permanece no Ubuntu:

```text
Termux
├── sshd :8022
└── Mobdesk
    └── Ubuntu via PRoot
```

## Comandos do MVP

```text
mobdesk setup       prepara Termux, PRoot, Ubuntu e ferramentas
mobdesk start       inicia SSH, Ubuntu e shell de trabalho
mobdesk shell       abre o shell Ubuntu sem reiniciar toda a sessão
mobdesk status      mostra o estado do host e do Ubuntu
mobdesk doctor      diagnostica instalação, rede e permissões
mobdesk install     instala uma ferramenta adicional
```

Comandos que os subcomandos devem encapsular:

```bash
mobdesk shell
# equivale a:
proot-distro login ubuntu

mobdesk install git
# equivale a:
proot-distro login ubuntu -- apt install -y git

mobdesk status
# consulta:
ss -tln
proot-distro list
proot-distro login ubuntu -- bash -lc 'ps aux'

mobdesk doctor
# consulta:
uname -a
df -h
free -h
whoami
ss -tln
proot-distro list
```

## Persistência

Devem permanecer após reiniciar o Mobdesk:

- Ubuntu instalado;
- pacotes do Ubuntu;
- projetos;
- configurações;
- sessões e logs;
- estado do setup.

O Mobdesk não deve apagar dados durante uma execução normal. Reset e remoção do Ubuntu exigem confirmação explícita.

## Fora do MVP

- APK próprio;
- interface web;
- VS Code web;
- Neko e navegador remoto;
- Nix-on-Droid;
- Docker real;
- desktop Linux gráfico;
- múltiplos usuários;
- cargas de produção e testes pesados.

## Critérios de sucesso

- o Mobdesk instala o ambiente em um Termux novo;
- `mobdesk setup` pode ser executado novamente sem quebrar a instalação;
- `mobdesk start` inicia SSH e Ubuntu;
- o usuário consegue entrar no Ubuntu pelo celular;
- outro computador consegue acessar com SSH;
- Git, Neovim, Go, Python, Node.js e Java ficam disponíveis;
- o usuário consegue criar, compilar e executar projetos educacionais;
- falhas exibem uma mensagem útil e podem ser diagnosticadas;
- nenhum projeto ou configuração é perdido durante o fluxo normal.
