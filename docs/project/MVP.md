# Mobdesk - MVP atual

O MVP atual transforma um Termux com Go instalado em um ambiente Ubuntu persistente, acessível pelo próprio celular e por SSH.

```text
Termux
  -> mobdesk setup
  -> Ubuntu via PRoot-Distro
  -> mobdesk start
  -> SSH :8022 ou shell local
```

## Pré-requisitos

- Termux instalado por uma fonte confiável;
- arquitetura ARM64 compatível;
- Go e Git instalados no Termux;
- espaço livre suficiente para o Ubuntu base e os projetos.

Bootstrap:

```bash
pkg update
pkg install -y golang git
go install github.com/ericklucioh/mobdesk/cmd/mobdesk@latest
~/go/bin/mobdesk setup
```

O binário usado no primeiro setup instala um launcher em `$PREFIX/bin/mobdesk` quando o caminho ainda está livre.

## `mobdesk setup`

O setup é dividido em etapas persistentes e pode ser executado novamente. Ele:

1. cria os diretórios de estado e logs;
2. atualiza os índices de pacotes do Termux;
3. instala `proot-distro`, `openssh` e `net-tools`;
4. instala ou verifica o Ubuntu persistente;
5. cria `/root/workspace` e diretórios do Mobdesk no Ubuntu;
6. solicita a senha do usuário Termux para o acesso SSH;
7. cria uma configuração SSH própria do Mobdesk;
8. instala o launcher global `mobdesk`.

O setup não executa `pkg upgrade` automaticamente. Para solicitar uma atualização completa do Termux:

```bash
mobdesk setup --upgrade-system
```

O estado do setup fica em:

```text
$HOME/.local/share/mobdesk/
├── logs/
├── config/
├── state/
├── password.done
└── setup.done
```

## `mobdesk start`

O start:

1. verifica o setup, a senha e o Ubuntu;
2. garante a configuração SSH dedicada;
3. ativa o wake-lock quando disponível;
4. inicia o `sshd` do Mobdesk na porta `8022`;
5. valida o PID, o processo, a configuração e o banner SSH;
6. mostra os endereços disponíveis;
7. abre o shell Ubuntu local.

O SSH do Mobdesk usa arquivos separados do SSH global do Termux:

```text
$HOME/.config/mobdesk/ssh/sshd_config
$HOME/.local/share/mobdesk/ssh/sshd.pid
$HOME/.local/share/mobdesk/ssh/sshd.log
```

Uma conexão remota é direcionada diretamente para o Ubuntu via PRoot:

```bash
ssh -p 8022 usuario@IP_DO_CELULAR
```

Para sair apenas do shell Ubuntu, use `exit`. Isso não encerra o servidor SSH.

## `mobdesk stop`

O stop lê somente o PID próprio do Mobdesk, confirma que o processo pertence à sua instância SSH e envia `SIGTERM`. Depois aguarda a porta fechar, remove o estado do PID e libera o wake-lock.

Se a porta estiver ocupada por outro processo, o Mobdesk não tenta encerrá-lo.

## Estado atual

### Entregue

- Ubuntu persistente via PRoot-Distro;
- SSH dedicado do Mobdesk na porta `8022`;
- acesso local e remoto diretamente ao Ubuntu;
- detecção de endereços IPv4 via `ifconfig`;
- autenticação por senha do Termux;
- setup idempotente e retomável por etapas;
- logs e PID próprios para o SSH;
- comandos `setup`, `start`, `stop` e `status`;
- comando `install` para Go, Python, Node.js, C, C++ e Lua no Ubuntu;
- saída humana e JSON versionado do `status`;
- coleta rápida de host, setup, Ubuntu, SSH, rede, armazenamento, bateria e Wi-Fi;
- testes unitários para estado, configuração SSH e instalação de linguagens;
- teste de integração Docker para instalação e Hello World das seis linguagens.

### Ainda não implementado

- `mobdesk shell`;
- `mobdesk doctor`;
- TUI;
- projetos, serviços, sessões persistentes e encaminhamento de portas;
- autenticação SSH por chave como fluxo assistido;
- suporte real validado em todos os modelos Android.

## Limites

- PRoot não é VM nem container real;
- o kernel, a rede, a memória e a bateria continuam sendo os do Android;
- o Termux pode ser suspenso ou encerrado pelo Android/HyperOS;
- o SSH não deve ser exposto diretamente na internet;
- o MVP é destinado a estudo, desenvolvimento e servidores leves.

Remover o Ubuntu e apagar projetos não fazem parte do fluxo normal do MVP.
