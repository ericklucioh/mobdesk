# Roadmap do Mobdesk

O Mobdesk evolui em quatro estágios. Termux é a workstation principal; Ubuntu
via PRoot é uma camada de compatibilidade opcional.

## Visão geral

| Estágio | Categoria | Nome | Resultado |
|---|---|---|---|
| 1 | MVP | Bootstrap Termux | Instalar e acessar a workstation Termux por shell e SSH |
| 2 | MVP | Workstation TUI | Trabalhar com ferramentas textuais organizadas |
| 3 | MVP | Ambiente persistente | Recuperar sessões, serviços e acesso remoto |
| 4 | Aplicação | Mobdesk Manager | Administrar projetos, sessões e serviços |

## Estágio 1 - Bootstrap Termux

**Status:** implementação inicial concluída; validação completa em Termux real ainda pendente.

### Objetivo

Levar o usuário de um Termux com Go instalado a uma workstation nativa acessível pelo celular e por SSH.

### Escopo

- instalar OpenSSH e ferramentas de diagnóstico;
- criar workspace e estado local;
- configurar SSH dedicado do Mobdesk;
- iniciar e parar o servidor com segurança;
- abrir o shell Termux local com `mobdesk shell`;
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

Oferecer uma interface textual organizada para trabalhar no Termux pelo celular ou por SSH.

### Escopo

- TUI para status, setup, start, stop, shell e atualização;
- catálogo e instalação dos perfis iniciais de ferramentas;
- suporte a teclado, mouse e terminais móveis;
- identificação de runtimes de compatibilidade sem tratar sua ausência como
  falha da workstation.

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

Transformar os comandos do Termux e os guests opcionais em uma central local de gerenciamento.

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
