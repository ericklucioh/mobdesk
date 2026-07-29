# Roadmap do Mobdesk

O Mobdesk evolui em quatro estágios. Termux continua sendo o host de controle, enquanto Ubuntu via PRoot permanece como o ambiente principal de desenvolvimento.

## Visão geral

| Estágio | Categoria | Nome | Resultado |
|---|---|---|---|
| 1 | MVP | Bootstrap Ubuntu | Instalar e acessar Ubuntu persistente por shell e SSH |
| 2 | MVP | Workstation TUI | Trabalhar com ferramentas textuais organizadas |
| 3 | MVP | Ambiente persistente | Recuperar sessões, serviços e acesso remoto |
| 4 | Aplicação | Mobdesk Manager | Administrar projetos, sessões e serviços |

## Estágio 1 - Bootstrap Ubuntu

**Status:** implementação inicial concluída; validação completa em Termux real ainda pendente.

### Objetivo

Levar o usuário de um Termux com Go instalado a um Ubuntu persistente acessível pelo celular e por SSH.

### Escopo

- instalar `proot-distro`, OpenSSH e ferramentas de diagnóstico;
- instalar ou verificar Ubuntu ARM64 persistente;
- criar workspace e estado local;
- configurar SSH dedicado do Mobdesk;
- iniciar e parar o servidor com segurança;
- abrir o shell Ubuntu local com `mobdesk shell`;
- consultar o estado do ambiente com `mobdesk status`;
- instalar os perfis iniciais de desenvolvimento com `mobdesk install`;
- detectar endereços locais e manter o wake-lock quando disponível;
- repetir e retomar o setup sem apagar dados.

### Fora deste estágio

- TUI;
- projetos, serviços e sessões persistentes;
- `doctor`;
- Tailscale e encaminhamento de portas.

## Estágio 2 - Workstation TUI

**Status:** implementação inicial em andamento; validação em Termux real ainda
pendente.

### Objetivo

Oferecer uma interface textual organizada para trabalhar no Ubuntu pelo celular ou por SSH.

### Escopo

- TUI para status, setup, start, stop, shell e atualização;
- catálogo e instalação dos perfis iniciais de ferramentas;
- suporte a teclado, mouse e terminais móveis;
- identificação da sessão SSH no Ubuntu e bloqueio das ações exclusivas do
  host Termux.
- perfis opcionais de configuração para apps, começando por Neovim/LazyVim.
- comandos CLI para instalar, desinstalar e aplicar/remover configurações com
  contrato JSON versionado.
- status distingue apps detectados, apps gerenciados e configuração aplicada,
  modificada ou em conflito.
- popup touch-first para detalhes, ações e confirmações destrutivas dos apps.

### Critério

O usuário consegue estudar e desenvolver no ambiente sem depender de uma sequência extensa de comandos internos.

## Estágio 3 - Ambiente persistente e remoto

### Objetivo

Permitir reconectar e continuar o trabalho após troca de rede, desconexão ou tela desligada.

### Escopo

- sessões persistentes com tmux;
- inicialização automática e recuperação;
- `status` e `doctor` completos;
- logs e health checks;
- Tailscale opcional;
- encaminhamento de portas;
- backups;
- orientações de bateria e execução em segundo plano no HyperOS.

### Critério

O usuário consegue iniciar, parar, reconectar e continuar um projeto sem reconstruir o ambiente.

## Estágio 4 - Mobdesk Manager

### Objetivo

Transformar os comandos do Termux, PRoot e Ubuntu em uma central local de gerenciamento.

### Escopo

- projetos e ambientes;
- ferramentas instaladas;
- sessões e serviços;
- portas, túneis e logs;
- diagnóstico e recuperação;
- backups;
- atualização e remoção controladas;
- configuração persistida e observável.

## Princípios de evolução

1. Cada estágio deve preservar o funcionamento do anterior.
2. A CLI e a TUI devem usar os mesmos serviços internos.
3. Não antecipar plugins, Nix ou múltiplos usuários sem necessidade validada.
4. O próximo estágio só começa depois de o fluxo atual ser validado no Termux real.

## Fora do núcleo

- Docker real;
- VM Linux;
- desktop gráfico completo;
- Nix como requisito;
- Neko como requisito;
- múltiplos usuários;
- cargas de produção pesada.
