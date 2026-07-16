# Roadmap do Mobdesk

O roadmap parte da missão em [`MISSAO.md`](MISSAO.md) e mantém Ubuntu via PRoot como ambiente principal, com Termux como host.

## MVP 1 — Instalação e central TUI

### Objetivo

O usuário instala o Termux, instala Go, obtém o Mobdesk e executa um comando para configurar o ambiente.

```text
pkg install golang
go install github.com/usuario/mobdesk/cmd/mobdesk@latest
mobdesk start
```

### Escopo

- TUI bonita e simples;
- verificar arquitetura, espaço e dependências;
- instalar `proot-distro` no Termux;
- instalar Ubuntu ARM64 persistente;
- criar diretórios e estado do Mobdesk;
- instalar ferramentas selecionadas dentro do Ubuntu;
- mostrar progresso, logs e erros;
- abrir shell Ubuntu;
- repetir a configuração sem destruir o que já existe;
- retomar uma instalação interrompida.

### Ferramentas iniciais

- Go no Termux para executar o Mobdesk;
- Bubble Tea, Bubbles e Lip Gloss;
- PRoot-Distro;
- Ubuntu ARM64;
- Git;
- Neovim ou LazyVim;
- Python;
- Node.js;
- Go;
- tmux.

### Critério de sucesso

Uma pessoa consegue sair de um Termux recém-instalado para um Ubuntu funcional, com ferramentas básicas, sem precisar conhecer os comandos internos.

## MVP 2 — Workstation TUI

### Objetivo

Usar o Ubuntu como uma workstation textual completa pelo celular ou SSH.

### Escopo

- sessão persistente;
- editor;
- explorador de arquivos;
- Git visual;
- processos e logs;
- múltiplos terminais;
- início de servidores de projeto;
- atalhos e layout configuráveis;
- acesso a projetos pelo navegador do celular;
- opção de iniciar a TUI por atalho do Termux.

### Ferramentas

- tmux como camada de recuperação;
- VTM ou Zellij;
- Neovim/LazyVim;
- lf, Yazi, broot ou TUIFI Manager;
- lazygit;
- btop ou htop;
- ripgrep, fd e fzf;
- Tailscale opcional.

### Critério de sucesso

O usuário consegue estudar e desenvolver em C, JavaScript, HTML, React, Java, Go e Python sem sair do ambiente pessoal do celular.

## MVP 3 — Ambiente persistente e remoto

### Objetivo

Usar o Mobdesk durante o dia, trocar de rede ou sala e continuar o trabalho sem reconstruir o ambiente.

### Escopo

- SSH;
- persistência de sessões e projetos;
- inicialização automática;
- Termux:Boot ou termux-services;
- wake-lock;
- diagnóstico;
- backups;
- encaminhamento de portas;
- acesso remoto por Tailscale;
- configuração recomendada do HyperOS;
- comandos de recuperação.

### Critério de sucesso

O usuário consegue iniciar, parar, reconectar e continuar um projeto com seus dados sob controle, sem depender dos computadores da faculdade.

## Aplicação 1 — Mobdesk Manager

### Objetivo

Transformar scripts e comandos em uma central de gerenciamento do Ubuntu e dos projetos.

### Escopo

- projetos;
- ambientes;
- ferramentas instaladas;
- sessões;
- serviços;
- portas;
- logs;
- processos;
- backups;
- diagnóstico;
- atualização e recuperação;
- montagem de diretórios compartilhados;
- comandos `setup`, `start`, `stop`, `shell`, `status`, `doctor` e `install`.

### Ferramentas

- Go;
- SQLite;
- YAML, TOML ou JSON;
- PRoot-Distro;
- tmux;
- Termux:API;
- Termux:Widget;
- testes automatizados.

## Aplicação 2 — Mobdesk Web/Desktop

### Objetivo

Oferecer uma interface visual leve pelo navegador, sem tentar reproduzir um desktop gráfico completo.

### Escopo

- dashboard;
- iniciar e parar projetos;
- terminal web;
- logs;
- arquivos;
- sessões;
- aplicações HTTP;
- túneis;
- autenticação;
- permissões;
- notificações;
- abertura de editores e ferramentas individuais.

### Ferramentas

- API Go;
- frontend TypeScript;
- React, Vue ou Svelte;
- WebSocket;
- xterm.js;
- SQLite;
- Tailscale;
- TLS quando necessário.

## Aplicação 3 — Mobdesk Platform

### Objetivo

Distribuir uma workstation móvel reproduzível, extensível e fácil de manter.

### Escopo

- instalação assistida;
- perfis de usuário e dispositivo;
- ambientes por projeto;
- catálogo de ferramentas;
- configurações declarativas;
- atualizações versionadas;
- rollback;
- backup e restauração;
- plugins;
- API para automação e agentes;
- suporte a múltiplos dispositivos;
- APK complementar;
- navegador remoto opcional;
- imagens ARM64 próprias;
- CI/CD e releases.

### Ferramentas possíveis

- Go;
- frontend TypeScript;
- WebSocket e WebRTC;
- Nix-on-Droid e Home Manager como camada opcional;
- GitHub Actions;
- Neko para navegador remoto;
- Tailscale;
- observabilidade;
- testes unitários, integração, E2E e benchmarks.

## Fora do MVP

- Docker real;
- desktop Linux gráfico completo;
- VM;
- Nix como requisito;
- Neko;
- múltiplos usuários;
- cargas de produção e testes pesados;
- suporte a todas as ferramentas existentes.
