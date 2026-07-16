# Ferramentas e práticas do Mobdesk

Este é um catálogo resumido das ferramentas pesquisadas e do papel de cada uma. Disponibilidade e compatibilidade devem ser verificadas no Ubuntu instalado pelo Mobdesk.

## Base do host

| Ferramenta | Papel |
|---|---|
| Termux | Host Android e execução do Mobdesk |
| `pkg` | Pacotes do host |
| Go | Binário e central de controle |
| PRoot-Distro | Instalação e execução do Ubuntu |
| OpenSSH | Acesso remoto |
| Tailscale | Rede privada fora da LAN |
| Termux:API | Integração com bateria, notificações e informações do Android |
| Termux:Boot | Inicialização após boot |
| termux-services | Serviços persistentes no host |
| Termux:Widget | Atalhos para `mobdesk start` e scripts |

## Ambiente de desenvolvimento

| Ferramenta | Papel |
|---|---|
| Ubuntu ARM64 | Runtime principal |
| `apt` | Pacotes Linux |
| Git | Código e versionamento |
| Go | Backend, CLI e projetos |
| Node.js/TypeScript | Frontend e ferramentas web |
| Python | Scripts, APIs e estudo |
| Rust | CLIs e ferramentas de sistema |
| Java/Kotlin | JVM, aulas e backends |
| Clang/Make/CMake | Compilação nativa |
| Neovim/LazyVim | Editor principal |
| Mason ou gerenciadores nativos | LSPs e formatadores, conforme compatibilidade |

## Experiência TUI

| Ferramenta | Papel | Observação |
|---|---|---|
| tmux | Persistência e recuperação | Deve continuar presente mesmo com VTM/Zellij |
| VTM | Desktop textual com janelas | Principal opção visual a experimentar |
| Zellij | Multiplexador com layouts | Alternativa mais amigável ao tmux |
| lf | Explorador leve | Ponto de partida confiável |
| Yazi | Explorador moderno | Usar quando disponível e estável |
| broot | Navegação e árvore | Complemento de arquivos |
| TUIFI Manager | Explorador visual | Interessante para estética de desktop |
| lazygit | Git visual | Staging, branches, diffs e commits |
| btop/htop | Processos e recursos | Diagnóstico diário |
| ripgrep | Busca de conteúdo | Base para navegação rápida |
| fd | Busca de arquivos | Complemento ao ripgrep |
| fzf | Seleção e filtros | Menus e automações |

Outras ideias pesquisadas — Twin, Desktop-TUI, TermOS, WibWob-DOS, Superfile, ranger e openmux — permanecem como referências ou experimentos. Não são dependências do MVP.

## TUI em Go

Para a central Mobdesk:

- Bubble Tea: ciclo de eventos e modelo de aplicação;
- Bubbles: listas, inputs, tabelas, spinners e componentes;
- Lip Gloss: estilos, cores e layout;
- PTY: shells e processos interativos;
- `os/exec`: execução e supervisão de comandos;
- SQLite: estado local, projetos e ferramentas instaladas;
- YAML/TOML/JSON: configuração exportável.

## Serviços e projetos

O Mobdesk deve tratar ferramentas e projetos como entidades diferentes:

```text
ferramenta:
  Git, Python, Node, Go, Neovim

projeto:
  diretório, runtime, dependências, portas e comandos

serviço:
  processo, logs, estado, porta e sessão
```

Exemplos de servidores educacionais:

- `npm run dev`;
- Vite;
- API Go;
- FastAPI/Uvicorn;
- servidores Java/Kotlin;
- ferramentas locais acessadas pelo navegador do celular.

## Práticas importantes

- manter tudo relacionado ao projeto no `$HOME` privado;
- separar dependências do Termux e do Ubuntu;
- fixar versões importantes;
- registrar logs de instalação;
- não apagar dados durante atualizações;
- testar uma ferramenta por vez;
- instalar primeiro o conjunto mínimo;
- usar tmux para sessões longas;
- usar `127.0.0.1` e túnel SSH quando não houver necessidade de exposição;
- medir RAM, CPU, temperatura, armazenamento e bateria;
- testar com a tela desligada e após reconexão;
- fazer backups antes de resetar o PRoot.

## Compatibilidade e limitações

Termux é excelente para pacotes nativos Android, mas não garante binários Linux/glibc, wheels `manylinux`, AppImages, Snap, Flatpak, `systemd` ou recursos de kernel.

Ubuntu via PRoot melhora a compatibilidade do userland, mas não oferece Docker real, módulos de kernel, namespaces completos, cgroups, acesso direto a dispositivos Android, aceleração gráfica garantida ou isolamento de VM.

## Testes

### Computador ou Docker ARM64

Validar lógica Go, TUI, comandos Ubuntu, instalação de pacotes e tratamento de erros. Isso não substitui Android real.

### Emulador Android

Validar Termux, instalação, PRoot e integração básica. Desempenho e HyperOS não são representados fielmente.

### POCO F6

Validar experiência real, teclado, armazenamento, temperatura, bateria, rede, tela desligada, processos em segundo plano e estabilidade.
