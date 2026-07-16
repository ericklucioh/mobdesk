# Mobdesk: evolução em seis estágios

Este documento organiza a evolução do Mobdesk do menor produto útil até uma aplicação completa e robusta.

A arquitetura principal considera o Termux como host de controle e integração com o Android, enquanto um Ubuntu persistente via PRoot fornece o ambiente de desenvolvimento compatível com Linux tradicional.

Os três primeiros estágios são MVPs. Os três últimos representam versões de aplicação, com maior automação, experiência de uso e confiabilidade.

## Arquitetura-base

```text
Android/HyperOS
└── Termux — host e integração Android
    ├── proot-distro
    ├── OpenSSH
    ├── Tailscale
    ├── wake-lock e inicialização
    └── Mobdesk launcher
        └── Ubuntu persistente — ambiente principal
            ├── apt e glibc
            ├── linguagens e toolchains
            ├── projetos
            ├── serviços
            └── ferramentas Linux
```

## Visão geral

| Estágio | Categoria | Nome | Resultado principal |
|---|---|---|---|
| 1 | MVP | Ubuntu remoto | Usar o POCO como servidor de desenvolvimento Linux por SSH |
| 2 | MVP | Workstation TUI Ubuntu | Trabalhar com editor, arquivos, Git e processos dentro do Ubuntu |
| 3 | MVP | Ambiente Ubuntu persistente | Manter serviços, sessões e acesso remoto de forma confiável |
| 4 | Aplicação | Mobdesk Manager | Centralizar projetos, sessões, serviços e diagnósticos |
| 5 | Aplicação | Mobdesk Desktop | Oferecer uma experiência visual integrada pelo navegador |
| 6 | Aplicação | Mobdesk Platform | Entregar uma workstation reproduzível, extensível e distribuível |

## Estágio 1 — MVP: Ubuntu remoto

### Escopo

O Mobdesk estabelece um Ubuntu persistente via PRoot como ambiente principal de desenvolvimento. O usuário acessa o aparelho por SSH e executa projetos, ferramentas e servidores dentro desse Ubuntu.

O Termux permanece responsável por instalar e iniciar o PRoot, oferecer a rede e fornecer a integração mínima com o Android.

Este estágio valida a hipótese central: o POCO F6 pode funcionar como uma workstation Linux ARM64 remota sem root, VM ou desktop gráfico completo.

### Ferramentas

- Android/HyperOS;
- Termux como host;
- PRoot-Distro;
- Ubuntu ARM64 persistente;
- `apt` e `glibc`;
- OpenSSH;
- Git;
- tmux;
- Neovim;
- Go, Node.js, Python, Rust e Java;
- servidores HTTP locais;
- Tailscale como opção de acesso privado.

### Limites

- configuração predominantemente manual;
- acesso principal por terminal;
- sem gerenciador próprio;
- sem interface visual do Mobdesk;
- sem Docker real ou recursos adicionais de kernel;
- integração com Android ainda limitada ao Termux hospedeiro.

### Resultado

Um Ubuntu persistente, acessível por SSH, capaz de hospedar código, dependências, ferramentas Linux e servidores de desenvolvimento.

## Estágio 2 — MVP: Workstation TUI Ubuntu

### Escopo

O Mobdesk transforma o Ubuntu remoto em uma workstation textual organizada, reunindo as principais tarefas de desenvolvimento em uma área de trabalho única.

O ambiente visual roda dentro do Ubuntu, enquanto o Termux continua funcionando como camada de inicialização e suporte.

### Ferramentas

- tudo do Estágio 1;
- VTM ou Zellij;
- lf, Yazi, broot ou TUIFI Manager;
- lazygit;
- btop ou htop;
- ripgrep;
- fd;
- fzf;
- Neovim/LazyVim;
- configurações personalizadas de editor e terminal;
- tmux como camada de persistência e recuperação.

### Limites

- interface dependente do emulador de terminal;
- maior consumo de armazenamento e memória que o Termux nativo;
- atalhos e layout precisam ser configurados;
- suporte touchscreen parcial;
- sem gerenciamento integrado de projetos ou serviços;
- PRoot continua sem isolamento real de container.

### Resultado

Uma workstation Linux textual com editor, arquivos, Git, processos, logs e terminais organizados em uma experiência única.

## Estágio 3 — MVP: Ambiente Ubuntu persistente

### Escopo

O Mobdesk deixa de ser apenas um Ubuntu acessado manualmente e passa a manter o ambiente operacional após desconexões, mudanças de rede ou períodos com a tela desligada.

Este estágio valida o uso cotidiano e remoto do sistema, incluindo recuperação de sessões e serviços.

### Ferramentas

- tudo do Estágio 2;
- Tailscale;
- Termux:Boot;
- termux-services;
- termux-wake-lock;
- tmux;
- scripts de inicialização do Ubuntu;
- scripts de health check;
- scripts de diagnóstico;
- encaminhamento de portas SSH;
- backups com tar, rsync ou armazenamento externo;
- configuração de bateria e inicialização do HyperOS.

### Limites

- ainda não há uma aplicação central de gerenciamento;
- recuperação de falhas depende de scripts e comandos;
- serviços são organizados por convenções;
- atualizações e rollback ainda são manuais;
- o Ubuntu continua dependente das limitações do kernel Android.

### Resultado

Um ambiente Ubuntu remoto persistente, acessível pela rede privada e capaz de recuperar sessões, serviços e informações básicas de saúde.

## Estágio 4 — Aplicação: Mobdesk Manager

### Escopo

O Mobdesk ganha uma aplicação de gerenciamento para controlar o Ubuntu e seus projetos sem exigir conhecimento dos comandos internos do Termux ou do PRoot.

O foco é organizar projetos, ambientes, sessões, serviços, portas, logs e diagnósticos em um único ponto de controle.

### Funcionalidades

- cadastro e seleção de projetos;
- escolha do runtime Ubuntu;
- criação e encerramento de sessões;
- inicialização de servidores e workers;
- visualização de status e logs;
- controle de portas e túneis;
- diagnóstico do Termux, PRoot e Ubuntu;
- gerenciamento de backups;
- indicadores de CPU, memória, armazenamento e temperatura;
- comandos de recuperação;
- perfis de ambiente por projeto;
- atualização e remoção controladas do Ubuntu;
- montagem de diretórios compartilhados entre Termux e Ubuntu.

### Ferramentas

- CLI própria `mobdesk`;
- backend em Go;
- SQLite ou arquivos estruturados para estado local;
- YAML, TOML ou JSON para configuração;
- PRoot-Distro como executor do Ubuntu;
- tmux e termux-services;
- OpenSSH;
- Tailscale;
- Termux:API para informações e ações disponíveis no Android;
- logs estruturados;
- testes automatizados da CLI.

### Resultado

Uma aplicação local que transforma os comandos do Termux, PRoot e Ubuntu em um ambiente administrável, observável e repetível.

## Estágio 5 — Aplicação: Mobdesk Desktop

### Escopo

O Mobdesk passa a oferecer uma interface visual acessível pelo navegador de outro computador, tablet ou celular.

A aplicação controla o Ubuntu e fornece uma camada visual para abrir as ferramentas de trabalho, sem eliminar o acesso SSH para operações avançadas.

### Funcionalidades

- dashboard do dispositivo;
- lista de projetos e ambientes Ubuntu;
- iniciar, parar e reiniciar serviços;
- terminal web;
- visualização de logs;
- editor ou integração com editor remoto;
- navegador de arquivos;
- status de sessões;
- abertura de aplicações HTTP;
- controle de túneis;
- autenticação de usuários;
- permissões por função;
- notificações de falhas e consumo elevado;
- configuração de resolução e preferências de interface.

### Ferramentas

- frontend web em TypeScript;
- React, Vue ou Svelte;
- API HTTP em Go;
- WebSocket para terminal, eventos e logs;
- xterm.js ou equivalente para terminal web;
- SQLite;
- Termux como host de rede e integração;
- Ubuntu PRoot como ambiente de execução;
- Tailscale para acesso privado;
- TLS quando o acesso não estiver protegido apenas por túnel;
- Termux:API para integração com o Android.

### Resultado

Uma aplicação web de workstation móvel capaz de administrar e acessar o Ubuntu do POCO sem depender exclusivamente de um terminal SSH.

## Estágio 6 — Aplicação: Mobdesk Platform

### Escopo

O Mobdesk se torna uma plataforma de workstation móvel reproduzível e extensível, com suporte a diferentes perfis de ambiente, dispositivos, usuários e formas de acesso.

Ubuntu via PRoot continua sendo o runtime principal. Nix-on-Droid entra como opção declarativa para instalação, reprodução e gerenciamento de configurações, não como requisito do núcleo.

### Funcionalidades

- instalação assistida do Termux, PRoot e Ubuntu;
- configuração declarativa;
- perfis de dispositivo e usuário;
- ambientes isolados por projeto;
- catálogo de ferramentas e extensões;
- atualizações versionadas;
- rollback;
- backup e restauração completos;
- telemetria local e observabilidade;
- permissões e auditoria;
- sincronização de configurações;
- suporte a múltiplos dispositivos;
- API para automação e agentes;
- plugins para novos runtimes e aplicações;
- navegador remoto opcional com Neko;
- distribuição de imagens ARM64 próprias;
- CI/CD para builds e releases.

### Ferramentas

- CLI e daemon Mobdesk;
- backend em Go;
- frontend web em TypeScript;
- SQLite local, com possibilidade de banco remoto;
- WebSocket e WebRTC;
- Termux, PRoot-Distro e Ubuntu ARM64;
- Nix-on-Droid e Home Manager como camada declarativa opcional;
- Docker ou GitHub Actions para produzir imagens ARM64;
- Neko para navegador remoto;
- Tailscale ou outra VPN privada;
- OpenTelemetry ou observabilidade equivalente;
- testes unitários, integração, E2E e benchmarks;
- GitHub Actions para releases;
- armazenamento externo para backups e artefatos.

### Resultado

Uma plataforma completa para transformar dispositivos Android ARM64 em workstations Linux remotas, configuráveis, monitoráveis e reproduzíveis.

## Ordem de evolução

```text
1. Termux host + Ubuntu persistente + SSH
        ↓
2. Workstation TUI dentro do Ubuntu
        ↓
3. Persistência + acesso remoto confiável
        ↓
4. CLI e gerenciador Mobdesk
        ↓
5. Desktop web
        ↓
6. Plataforma reproduzível e extensível
```

Os estágios devem preservar o funcionamento do anterior. Cada camada nova precisa resolver uma necessidade concreta: primeiro executar, depois organizar, então manter, administrar, visualizar, reproduzir e expandir.
