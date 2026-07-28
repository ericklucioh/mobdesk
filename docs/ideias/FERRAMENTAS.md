# Catálogo de aplicativos do Mobdesk

Catálogo editorial de aplicativos úteis para o usuário no ambiente Ubuntu.
Cada entrada segue o formato `nome - tipo - descrição`.

## Ferramentas essenciais no celular

Ferramentas que precisam executar ou operar diretamente o ambiente Ubuntu, os
arquivos, os projetos, as credenciais e os serviços persistidos no celular.

### Ferramentas base sem TUI

Git - git - Controle de versão para projetos

### Ferramentas TUI

#### Multiplexador

Esse é o mais foda, qp teria q escolher apenas um e descartar os demais, pq misturar fica ruim, só q entra no quesito q o tmux é o 01, mas tem ux pra iniciante ruim, mesma coisa com o zellij, muito maduro mas não tão simples de usar.
Ai tem o VTM e o TUIOS, q são bem mais práticos de usar, mas não sei se são tão maduros assim.
Está sendo difícil escolher, pq precisa ser uma mistura de bom para mim, e bom para usuários leigos.

- Zellij - multiplexador - Multiplexador com abas, painéis e layouts
- tmux - multiplexador - Mantém sessões e painéis persistentes
- tuios
- VTM - multiplexador - Desktop textual com janelas e painéis

#### Desenvolvimento, Git e arquivos

- micro - editor - Editor de texto simples para terminal
- TTT - editor - Editor de texto minimalista para terminal

- gh - git - Cliente oficial do GitHub pela linha de comando
- gh-dash - git - Dashboard TUI para GitHub CLI
- lazygit - git - Interface visual para operações Git
- tig - git - Visualizador de histórico Git no terminal
- git-scope - git - Painel de múltiplos repositórios Git
- gitv - git - Cliente de terminal para issues GitHub
- 
- tree - arquivos - Exibe árvores de diretórios no terminal
- fsel - arquivos - Seletor interativo de arquivos no terminal; a confirmar no ambiente
- Yazi - arquivos - Gerenciador de arquivos com previews
- TUIFI Manager - arquivos - Explorador visual de arquivos no terminal

- ripgrep - busca - Busca rápida de texto em arquivos
- fd - busca - Busca rápida de arquivos e diretórios
- fzf - busca - Seleção fuzzy para arquivos e comandos
- television - busca - Fuzzy finder configurável para terminal

- Zsh - shell - Shell interativo com completions configuráveis
- fzf-tab - shell - Seleção fuzzy para completions do Zsh
- Atuin - shell - Histórico interativo de comandos
- Carapace - shell - Completions para vários shells e comandos

#### Dados, APIs e bancos

- jq - dados - Processa e filtra JSON pela linha de comando
- fx - dados - Visualizador interativo de JSON
- csvlens - dados - Visualizador de arquivos CSV no terminal
- Tabiew - dados - Visualizador de CSV, TSV e Parquet
- OTree - dados - Navegador em árvore para dados estruturados
- jqp - dados - Interface para experimentar filtros jq
- Twig - dados - Explorador TUI de JSON e YAML
- ATAC - api - Cliente de APIs HTTP no terminal

- resterm - api - Cliente para HTTP, GraphQL e WebSockets
- Posting - api - Cliente visual para testar APIs HTTP

- Harlequin - banco - Cliente SQL interativo para terminal
- Rainfrog - banco - Cliente TUI para bancos SQL
- Lazysql - banco - Administração visual de bancos SQL
- pgcli - banco - Cliente PostgreSQL com autocomplete
- mycli - banco - Cliente MySQL com autocomplete
- litecli - banco - Cliente SQLite com autocomplete
- sqlite3 - banco - Shell oficial para SQLite
- IRedis - banco - Cliente interativo para Redis
- termdbms - banco - Visualizador de bancos no terminal

#### Operação, rede e infraestrutura

- btop - sistema - Monitor de processos e recursos do sistema
- ncdu - sistema - Analisador interativo de espaço em disco
- gdu - sistema - Analisador visual de uso de disco
- bottom - sistema - Monitor de sistema para terminal
- glances - sistema - Monitor de sistema com painel e API
- diskonaut - sistema - Navegador visual de espaço em disco
- inxi - sistema - Coletor detalhado de informações do sistema

- lnav - logs - Visualizador e analisador de logs
- logradar - logs - Filtro e destaque de logs no terminal
- lazyjournal - logs - Navegador de logs locais e de serviços

- bandwhich - rede - Monitor de uso de rede por processo
- termshark - rede - Inspeção de tráfego com interface TUI
- trippy - rede - Diagnóstico de rede e traceroute
- mitmproxy - rede - Proxy interativo para inspeção HTTPS
- speedtest-cli - rede - Teste de velocidade de conexão pela linha de comando
- curl - rede - Cliente de transferência e requisições HTTP
- wget - rede - Baixador de arquivos pela linha de comando
- ftp - rede - Cliente tradicional de transferência FTP
- ping - rede - Diagnóstico de conectividade de rede
- ip - rede - Administração de interfaces e rotas de rede
- netstat - rede - Visualização de conexões e estatísticas de rede

- K9s - infraestrutura - Administração visual de clusters Kubernetes
- ktop - infraestrutura - Monitor TUI de recursos Kubernetes

- termscp - armazenamento - Cliente de transferência por SCP e SFTP

#### Assistência, depuração e documentação

- OpenCode - assistência - Agente de programação para terminal
- clin - assistência - Ferramenta de assistência de programação no terminal
- glint - assistência - Ferramenta de assistência de programação no terminal

- leetgo - desenvolvimento - Cliente para consultar e resolver exercícios de programação

- dialog - interface - Criação de diálogos e menus em terminal

- glow - documentação - Leitor de Markdown no terminal

- less - leitura - Paginação e leitura de texto no terminal

- mermaid-ascii - documentação - Renderizador de diagramas Mermaid em ASCII

Referência: https://github.com/AlexanderGrooff/mermaid-ascii

## Ferramentas não essenciais no celular

Ferramentas de conveniência que o computador cliente normalmente já oferece, ou
que dependem de recursos do host incompatíveis ou pouco úteis no Ubuntu via
PRoot.

### Comunicação e produtividade

- aerc - comunicação - Cliente de email para terminal
- NeoMutt - comunicação - Cliente de email extensível para terminal
- Mutt - comunicação - Cliente tradicional de email no terminal
- Alpine - comunicação - Cliente de email e notícias em terminal
- meli - comunicação - Cliente moderno de email no terminal
- Discordo - comunicação - Cliente de Discord no terminal
- Concord - comunicação - Cliente TUI para Discord
- Endcord - comunicação - Cliente de Discord com recursos amplos
- nchat - comunicação - Cliente de mensagens para terminal
- gomuks - comunicação - Cliente de Matrix no terminal
- Irssi - comunicação - Cliente de IRC para terminal
- Profanity - comunicação - Cliente de XMPP para terminal
- gurk-rs - comunicação - Cliente de Signal para terminal
- tuir - comunicação - Cliente de Reddit no terminal
- Devzat - comunicação - Chat colaborativo acessível por SSH
- calcurse - produtividade - Agenda e calendário no terminal
- taskwarrior-tui - produtividade - Interface visual para tarefas
- hledger-ui - produtividade - Interface de contabilidade no terminal
- taskline - produtividade - Tarefas, quadros e notas no terminal
- Cal - produtividade - Calendário simples no terminal
- keydex - segurança - Gerenciador de credenciais no terminal

### Mídia, web e entretenimento

- cmus - mídia - Player de música local no terminal
- GopherTube - mídia - Cliente de vídeos para terminal
- invidtui - mídia - Cliente TUI para vídeos e canais
- Streamlink - mídia - Encaminha transmissões para um player
- twitch-tui - mídia - Visualizador de chat Twitch no terminal
- ncspot - mídia - Cliente de Spotify no terminal
- spotify-player - mídia - Player de Spotify para terminal
- spotatui - mídia - Cliente de Spotify com letras
- spotify-tui - mídia - Cliente antigo de Spotify no terminal
- ncmpcpp - mídia - Cliente de música para MPD
- kew - mídia - Player de música para terminal
- termusic - mídia - Player TUI de música
- rmpc - mídia - Cliente TUI para MPD
- pyradio - mídia - Player de rádios web
- soundcloud2000 - mídia - Cliente de SoundCloud no terminal
- newsboat - mídia - Leitor de RSS e Atom
- Jellyfin TUI - mídia - Cliente de Jellyfin para terminal
- managarr - mídia - Gerenciador TUI de servidores de mídia
- epy - mídia - Leitor de ebooks no terminal
- tdf - mídia - Visualizador de PDFs no **terminal**
- manga-tui - mídia - Leitor de mangás no terminal
- Elinks - web - Navegador web em modo texto
- ffmpeg - mídia - Ferramenta para conversão e processamento de áudio e vídeo
- ffplay - mídia - Reprodutor de mídia baseado no FFmpeg
- ffprobe - mídia - Analisador de metadados e streams de mídia
- VLC - mídia - Reprodutor multimídia
- cvlc - mídia - Reprodutor VLC controlado pela linha de comando
- qvlc - mídia - Interface gráfica do VLC
- nvlc - áudio - Interface ncurses do VLC; a confirmar no ambiente
- alsamixer - áudio - Mixer de áudio do ALSA no terminal
- webquiz - jogos - Jogo ou quiz acessível pelo terminal; a confirmar no ambiente
- joguinhos - jogos - Jogos simples para terminal; a confirmar quais ferramentas

### Administração não adequada ao PRoot

- WibWob-DOS - desktop-tui - Ambiente experimental inspirado em desktop
- impala - rede - Gerenciador de Wi-Fi no terminal
- bluetuith - rede - Gerenciador de dispositivos Bluetooth
- bluetui - rede - Gerenciador de dispositivos Bluetooth
- nmtui - rede - Gerenciador oficial de conexões de rede
- nmcli - rede - Cliente de linha de comando do NetworkManager
- bluetoothctl - rede - Utilitário de controle de Bluetooth no terminal
- cfdisk - sistema - Editor de partições em modo texto
- cgdisk - sistema - Editor de tabelas GPT em modo texto
- gdisk - sistema - Editor de tabelas de partição GPT; a confirmar no ambiente
- aptitude - sistema - Administração visual de pacotes APT
- aptitude-curses - sistema - Interface curses para administração de pacotes APT
- Caligula - sistema - Criação de imagens e mídias inicializáveis
- neofetch - sistema - Exibição resumida de informações do sistema
- logo-ls - sistema - Comando ls com exibição de logotipos
- distrobox-tui - infraestrutura - Administração visual de ambientes Distrobox
- vctui - infraestrutura - Interface de terminal para vCenter
- ctop - infraestrutura - Monitor de containers no terminal

### Gaveta

- openvpn - rede - Cliente VPN baseado em OpenVPN - tem utilidade, vale colocar na lojinha pra facilitar, mas o usuário q tem q configurar pra ele no contexto dele, não é de minha responsabilidade
- pw-top - áudio - Monitor de fluxos e dispositivos PipeWire - interessante, mas feio, talvez procurar outro mais legal
- moeda - finanças - Ferramenta de linha de comando relacionada a moedas
- leaf - documentação - Ferramenta para leitura ou navegação de Markdown; a confirmar no ambiente
