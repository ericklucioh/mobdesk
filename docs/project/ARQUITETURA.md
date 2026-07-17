# Arquitetura atual do Mobdesk

Este documento registra a arquitetura escolhida para o Mobdesk. A missão e o valor para o usuário estão em [`MISSAO.md`](MISSAO.md); a evolução do produto está em [`estagios_mobdesk.md`](estagios_mobdesk.md).

## Princípio central

O celular executa o ambiente de desenvolvimento. O usuário mantém seus projetos, ferramentas, sessões e informações no próprio aparelho, sem depender de contas pessoais em computadores compartilhados.

```text
Android/HyperOS
└── Termux — host e integração com Android
    ├── Go/Mobdesk
    ├── OpenSSH
    ├── Tailscale
    ├── wake-lock e inicialização
    └── PRoot-Distro
        └── Ubuntu ARM64 persistente — workstation principal
            ├── glibc e apt
            ├── Git
            ├── Go, Node.js, Python, Rust, Java e Kotlin
            ├── Neovim/LazyVim
            ├── ferramentas TUI
            ├── projetos
            └── servidores de desenvolvimento
```

## Responsabilidades das camadas

### Termux

O Termux é o host de controle. Ele fornece:

- execução do binário Mobdesk;
- instalação de pacotes básicos com `pkg`;
- instalação e controle do PRoot-Distro;
- OpenSSH;
- Tailscale e rede;
- wake-lock e inicialização;
- acesso às APIs do Android por Termux:API;
- atalhos por Termux:Widget;
- armazenamento privado do aplicativo.

### Mobdesk

O Mobdesk é a central de inicialização e controle. Ele deve:

- apresentar uma TUI clara;
- verificar o estado do host e do Ubuntu;
- instalar o que estiver ausente;
- executar comandos no Termux;
- executar comandos dentro do Ubuntu;
- instalar ferramentas no Ubuntu;
- iniciar e parar sessões e projetos;
- acompanhar processos, saída e erros;
- exibir diagnóstico e logs.

### Ubuntu via PRoot

O Ubuntu é o ambiente principal do usuário. Ele oferece:

- userland Linux tradicional;
- `apt`;
- `glibc`;
- caminhos e bibliotecas esperados por ferramentas Linux;
- linguagens e dependências de projeto;
- editor, Git, TUIs e servidores.

PRoot melhora a compatibilidade de userland, mas não fornece um kernel Ubuntu, uma VM ou isolamento real de containers. O kernel, a rede, os dispositivos e as limitações de energia continuam sendo os do Android.

## Execução e armazenamento

Projetos e dados devem ficar no armazenamento privado do Termux ou em um diretório compartilhado controlado pelo Mobdesk. Evitar `~/storage`, `/sdcard` e outros caminhos de armazenamento externo para repositórios e builds, pois eles podem não preservar links simbólicos, permissões e atributos Unix.

O Mobdesk deve distinguir:

```text
dados compartilhados:
  código, documentos, configurações exportáveis

dados específicos do runtime:
  node_modules, .venv, caches, binários e bibliotecas compiladas
```

Dependências compiladas para Termux e Ubuntu não devem ser misturadas automaticamente.

## Comandos entre camadas

O Mobdesk pode executar comandos diretamente no host ou no Ubuntu:

```text
Termux:
  pkg install -y proot-distro

Ubuntu:
  proot-distro login ubuntu -- apt install -y git python3
  proot-distro login ubuntu -- bash -lc "cd /projeto && npm run dev"
```

Comandos interativos, como shell e editor, precisam de suporte a PTY. Comandos de instalação e diagnóstico podem ser executados como processos monitorados, capturando stdout, stderr e código de saída.

## Acesso

### MVP

O acesso principal é pelo próprio Termux ou por uma única conexão SSH. Servidores do usuário podem ser executados no Ubuntu e acessados pelo IP do celular ou por encaminhamento SSH.

### Acesso remoto

Tailscale é a opção preferida para acessar o aparelho fora da rede local. A porta SSH não deve ser exposta diretamente na internet.

### Interface futura

Aplicações específicas poderão ser expostas pelo navegador em estágios posteriores. Isso não faz parte do primeiro MVP.

## Persistência

O Mobdesk deve combinar:

- Ubuntu persistente via PRoot-Distro;
- tmux para sessões;
- Termux:Boot ou termux-services;
- `termux-wake-lock` durante uso como servidor;
- permissões de bateria e execução em segundo plano no HyperOS;
- backups do código, configurações e dados importantes.

Persistência de processos não elimina o risco de o Android encerrar o Termux. O diagnóstico deve diferenciar queda de rede, encerramento do PRoot e encerramento do aplicativo host.

## Limites conhecidos

- Docker real não é garantido dentro do PRoot;
- `systemd`, namespaces, cgroups, seccomp e módulos de kernel não funcionam como em uma instalação Linux real;
- acesso direto a GPU, câmera, USB e outros dispositivos Android é limitado;
- aceleração gráfica não é objetivo do núcleo;
- desktops gráficos completos, VNC e X11 podem consumir recursos demais;
- desempenho e estabilidade dependem do HyperOS, bateria, temperatura e memória;
- o sistema é adequado para estudo, desenvolvimento e servidores leves, não para cargas pesadas de produção.

## Segurança

- manter projetos no armazenamento privado;
- preferir autenticação SSH por chave;
- não expor SSH ou aplicações diretamente na internet;
- usar Tailscale ou túnel SSH;
- manter autenticação nas aplicações web futuras;
- revisar scripts antes de executá-los;
- não armazenar segredos em scripts ou no Git;
- manter backup fora do celular.
