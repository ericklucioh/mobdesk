**POCO F6 COMO  
WORKSTATION LINUX/TUI**

**Termux, SSH, tmux, Neovim e desktops textuais**

Guia consolidado para transformar o celular em um servidor de
desenvolvimento ARM64,  
acessível por SSH e HTTP, com uma camada visual de “sistema operacional”
dentro do terminal.

| **Dispositivo-alvo**        | POCO F6 / Android / HyperOS          |
|-----------------------------|--------------------------------------|
| **Arquitetura recomendada** | Termux nativo + SSH + tmux + Neovim  |
| **Camada visual**           | VTM e ferramentas TUI complementares |

Érick Lúcio • julho de 2026

# 1. Resumo executivo

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th><p><strong>Recomendação principal</strong></p>
<p>Use o Termux nativo como sistema de trabalho. Execute Git, SSH, tmux,
Neovim, Go, Node.js, Python e os servidores HTTP diretamente nele. Use
Debian via PRoot somente para ferramentas que exigem glibc. Use
VM/Podroid somente quando Docker for indispensável.</p></th>
</tr>
</thead>
<tbody>
</tbody>
</table>

O objetivo deste setup não é reproduzir um desktop Linux gráfico
completo. O celular funciona melhor como um servidor ARM64 de
desenvolvimento: o processamento acontece no POCO F6; outro dispositivo
acessa o terminal por SSH e as aplicações web por HTTP ou por túneis
SSH.

A sensação de “sistema operacional visual” pode ser obtida sem X11, VNC
ou noVNC. Em vez de transmitir pixels, utiliza-se um desktop textual
TUI, desenhado dentro do terminal, com janelas, barra, explorador de
arquivos, editor, Git e servidores.

| **Camada**                | **Ferramenta recomendada**    | **Função**                                                              |
|---------------------------|-------------------------------|-------------------------------------------------------------------------|
| Host                      | Termux nativo                 | Executar linguagens, processos e serviços diretamente no Android ARM64. |
| Acesso                    | OpenSSH                       | Entrar remotamente no celular e encaminhar portas HTTP.                 |
| Persistência              | tmux                          | Manter editor e servidores ativos após queda da conexão.                |
| Editor                    | Neovim                        | Editar o projeto pelo terminal, localmente ou via SSH.                  |
| Desktop textual           | VTM                           | Organizar múltiplas aplicações TUI em janelas.                          |
| Arquivos                  | lf / Yazi / TUIFI / Superfile | Navegar por pastas e arquivos visualmente.                              |
| Git                       | lazygit                       | Interface TUI para branches, commits, diffs e staging.                  |
| Compatibilidade GNU/Linux | Debian via PRoot              | Executar ferramentas que exigem glibc.                                  |
| Containers                | Podroid/VM                    | Laboratório opcional para Docker e Podman.                              |

# 2. Arquitetura recomendada

**Arquitetura lógica do ambiente**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>POCO F6 / HyperOS<br />
│<br />
├── Tailscale Android (opcional)<br />
│ └── acesso privado fora da rede local<br />
│<br />
├── Termux nativo<br />
│ ├── OpenSSH Server :8022<br />
│ ├── tmux<br />
│ ├── Neovim<br />
│ ├── Git + lazygit<br />
│ ├── VTM<br />
│ │ ├── janela de arquivos<br />
│ │ ├── janela do editor<br />
│ │ ├── janela do servidor<br />
│ │ └── janela de logs<br />
│ ├── Go<br />
│ ├── Node.js / npm<br />
│ ├── Python<br />
│ └── servidores HTTP<br />
│ ├── Vite :5173<br />
│ ├── API Go :8080<br />
│ └── Python :8000<br />
│<br />
├── Debian via PRoot (sob demanda)<br />
│ └── ferramentas Linux ARM64/glibc incompatíveis com Bionic<br />
│<br />
└── Podroid/VM (laboratório)<br />
└── Docker, Podman e Compose quando realmente necessários</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

A regra de ouro é manter no Termux tudo que funciona nativamente. Cada
camada adicional — PRoot, QEMU, desktop gráfico, VNC — aumenta consumo,
latência e pontos de falha.

# 3. Ferramentas de desktop visual dentro do terminal

Esta seção reúne as ferramentas citadas na pesquisa. Nem todas têm o
mesmo objetivo: algumas são desktops textuais completos, outras são
gerenciadores de janelas, exploradores de arquivos ou componentes que,
combinados, criam a experiência de um sistema operacional.

| **Projeto**        | **Categoria**                           | **Maturidade** | **Termux**                    | **Uso recomendado**        |
|--------------------|-----------------------------------------|----------------|-------------------------------|----------------------------|
| VTM                | Desktop/multiplexador textual           | Alta           | Sim; pacote Termux            | Principal recomendação     |
| Twin               | Servidor de janelas em modo texto       | Média          | Suporte relatado pelo projeto | Experimento avançado       |
| Desktop-TUI        | Desktop com taskbar e janelas           | Baixa/média    | Provável via Rust             | Experimental               |
| TermOS             | Sistema operacional fictício em Textual | Baixa          | Provável via Python           | Demonstração visual        |
| WibWob-DOS         | Desktop TUI controlável por API/IA      | Experimental   | Bun dificulta no Termux       | Inspiração                 |
| TUIFI Manager      | Explorador de arquivos visual           | Média          | Focado em Termux              | Bom complemento            |
| lf                 | Explorador de arquivos                  | Alta           | Sim                           | Leve e confiável           |
| Yazi               | Explorador assíncrono moderno           | Alta           | Depende do pacote/build       | Visual e rápido            |
| Superfile          | Explorador visual em Go                 | Média/alta     | Possível em ARM64             | Interface rica             |
| Midnight Commander | Explorador de dois painéis              | Alta           | Sim                           | Clássico e estável         |
| ranger             | Explorador inspirado no Vim             | Alta           | Sim                           | Familiar para usuários Vim |
| lazygit            | Cliente Git TUI                         | Alta           | Sim                           | Git visual                 |
| tmux               | Multiplexador de terminal               | Muito alta     | Sim                           | Persistência e painéis     |

## 3.1 VTM — Virtual Terminal Multiplexer

É a opção que mais se aproxima de um desktop TUI utilizável no dia a
dia. Ele cria uma área de trabalho textual, permite abrir aplicações CLI
em janelas e funciona bem como camada superior para Neovim, shell,
lazygit, explorador de arquivos, logs e servidores.

- **Ponto forte:** combina múltiplas aplicações dentro de uma interface
  textual organizada.

- **Instalação:** há pacote preparado para o ecossistema Termux.

- **Uso ideal:** entrar por SSH no POCO e iniciar \`vtm\` no terminal
  cliente.

- **Limitação:** é uma categoria de software incomum; atalhos e
  ergonomia exigem adaptação.

**Instalação e inicialização**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>pkg update<br />
pkg install vtm<br />
<br />
vtm</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

**Workspace sugerido**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>VTM<br />
├── janela: lf ou TUIFI Manager<br />
├── janela: Neovim<br />
├── janela: lazygit<br />
├── janela: npm run dev<br />
├── janela: go test ./...<br />
└── janela: tail -f logs/app.log</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

## 3.2 Twin — Textmode WINdow environment

O Twin é conceitualmente parecido com um servidor gráfico, mas em modo
texto. Trabalha com janelas sobrepostas, mouse, clientes e displays
remotos. É tecnicamente interessante e mais próximo de um “X11 textual”,
porém exige mais configuração e tende a ser menos direto que o VTM.

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th><p><strong>Quando testar</strong></p>
<p>Use o Twin como projeto de estudo ou quando quiser explorar um
sistema de janelas textual de verdade. Não o escolha como primeira base
de produtividade.</p></th>
</tr>
</thead>
<tbody>
</tbody>
</table>

## 3.3 Desktop-TUI

Projeto experimental focado exatamente em barra de tarefas, atalhos,
seleção de arquivos e janelas movíveis/redimensionáveis. É uma boa
referência para estudar como construir seu próprio desktop TUI.

**Possível caminho de instalação; pode exigir ajustes de compilação no
Android**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>pkg install rust<br />
cargo install desktop-tui</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

## 3.4 TermOS

Sistema operacional fictício construído com Python e Textual. Simula
menu iniciar, taskbar, relógio, notificações, janelas e pequenos
aplicativos. A proposta visual é forte, mas a maturidade ainda é mais
próxima de uma demonstração do que de uma workstation.

## 3.5 WibWob-DOS

Desktop textual com gerenciador de janelas, editor, explorador, terminal
interno e API HTTP para automação. O diferencial é permitir que humano e
agente de IA controlem o mesmo workspace. O uso de Bun cria dificuldade
no Android/Bionic, portanto o projeto é melhor tratado como inspiração
ou executado em uma camada GNU/Linux compatível.

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th><p><strong>Ideia de projeto</strong></p>
<p>Um caminho interessante para você seria criar uma versão menor desse
conceito em Go: desktop TUI, janelas para comandos, explorador, logs,
Git e uma API local para agentes de IA.</p></th>
</tr>
</thead>
<tbody>
</tbody>
</table>

## 3.6 Exploradores de arquivos TUI

| **Ferramenta**     | **Estilo**              | **Ponto forte**                               | **Escolha sugerida**        |
|--------------------|-------------------------|-----------------------------------------------|-----------------------------|
| lf                 | Minimalista/Vim-like    | Muito leve, rápido e previsível.              | Melhor ponto de partida.    |
| TUIFI Manager      | Desktop/ícones textuais | Focado em uso visual no Termux.               | Para estética de “SO”.      |
| Yazi               | Moderno e assíncrono    | Preview, navegação rápida e experiência rica. | Quando disponível no setup. |
| Superfile          | Interface visual em Go  | Layout moderno, múltiplos painéis.            | Para experimentar.          |
| Midnight Commander | Dois painéis            | Estável, tradicional e completo.              | Para máxima confiabilidade. |
| ranger             | Colunas/Vim-like        | Extensível e conhecido.                       | Alternativa madura.         |

## 3.7 Componentes visuais complementares

| **Ferramenta** | **Função**        | **Por que entra no desktop TUI**                              |
|----------------|-------------------|---------------------------------------------------------------|
| Neovim         | Editor            | Centro do fluxo de desenvolvimento.                           |
| lazygit        | Git               | Staging, commits, branches, logs e diffs em interface visual. |
| btop/htop      | Monitoramento     | CPU, memória e processos do ambiente.                         |
| ripgrep + fzf  | Busca             | Localizar conteúdo e arquivos rapidamente.                    |
| tmux           | Sessões e painéis | Persistência mesmo quando o SSH cai.                          |
| VTM            | Janelas           | Camada visual superior para organizar os demais componentes.  |

# 4. Instalação base do Termux

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th><p><strong>Fonte de instalação</strong></p>
<p>Instale o Termux e seus complementos pela mesma fonte de assinatura —
preferencialmente GitHub oficial ou F-Droid. Evite versões antigas de
lojas que não acompanham o projeto.</p></th>
</tr>
</thead>
<tbody>
</tbody>
</table>

## 4.1 Atualização e pacotes principais

**Pacotes do núcleo de desenvolvimento**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>pkg update<br />
pkg upgrade -y<br />
<br />
pkg install -y \<br />
git \<br />
openssh \<br />
tmux \<br />
neovim \<br />
nodejs \<br />
golang \<br />
python \<br />
clang \<br />
make \<br />
cmake \<br />
pkg-config \<br />
rust \<br />
ripgrep \<br />
fd \<br />
fzf \<br />
jq \<br />
tree \<br />
rsync \<br />
curl \<br />
wget \<br />
zip \<br />
unzip \<br />
procps \<br />
htop</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

**Camada visual recomendada**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>pkg install -y \<br />
vtm \<br />
lf \<br />
lazygit</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

Alguns pacotes podem mudar de nome, migrar entre repositórios ou exigir
compilação. Quando um pacote não estiver disponível, confirme a
arquitetura \`aarch64\` e procure uma versão específica para
Termux/Android.

## 4.2 Validação do ambiente

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>uname -m<br />
git --version<br />
ssh -V<br />
tmux -V<br />
nvim --version<br />
node --version<br />
npm --version<br />
go version<br />
python --version<br />
clang --version</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th><p><strong>Resultado esperado</strong></p>
<p>A arquitetura deve ser ARM64/aarch64. Go, Node e Python executarão
nativamente no Android, sem emulação de CPU.</p></th>
</tr>
</thead>
<tbody>
</tbody>
</table>

## 4.3 Estrutura de diretórios

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>mkdir -p ~/code/{go,node,python}<br />
mkdir -p ~/workspace<br />
mkdir -p ~/.config<br />
mkdir -p ~/.local/bin</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

**Estrutura sugerida**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>/data/data/com.termux/files/home/<br />
├── code/<br />
│ ├── go/<br />
│ ├── node/<br />
│ └── python/<br />
├── workspace/<br />
├── .ssh/<br />
├── .config/<br />
└── .local/bin/</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th><p><strong>Não trabalhe diretamente em /sdcard</strong></p>
<p>Mantenha projetos, repositórios Git e executáveis dentro de `$HOME`.
O armazenamento compartilhado do Android pode usar `noexec`, limitar
permissões Unix e causar problemas com Git, Node e scripts.</p></th>
</tr>
</thead>
<tbody>
</tbody>
</table>

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>termux-setup-storage<br />
<br />
# Use apenas para importação/exportação<br />
cp ~/code/projeto/build.zip ~/storage/downloads/<br />
cp ~/storage/downloads/projeto.zip ~/code/</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

# 5. Transformando o POCO em servidor SSH

## 5.1 Inicialização simples

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>passwd<br />
whoami<br />
sshd<br />
ss -tln | grep 8022</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

O servidor SSH do Termux normalmente escuta na porta 8022. O usuário
será semelhante a \`u0_a123\`.

**Conexão pela rede local**

| ssh -p 8022 u0_a123@192.168.1.80 |
|----------------------------------|

## 5.2 Autenticação por chave

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th># No computador cliente<br />
ssh-keygen -t ed25519<br />
<br />
ssh-copy-id -p 8022 u0_a123@192.168.1.80<br />
ssh -p 8022 u0_a123@192.168.1.80</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

Depois de validar a chave, desative senha para reduzir risco de acesso
indevido.

| nvim "\$PREFIX/etc/ssh/sshd_config" |
|-------------------------------------|

**Trecho do sshd_config**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>PubkeyAuthentication yes<br />
PasswordAuthentication no</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

**Reinício do serviço**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>pkill sshd<br />
sshd</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

## 5.3 Configuração do cliente SSH

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th># ~/.ssh/config no computador cliente<br />
Host poco<br />
HostName 192.168.1.80<br />
User u0_a123<br />
Port 8022<br />
IdentityFile ~/.ssh/id_ed25519<br />
ServerAliveInterval 30<br />
ServerAliveCountMax 3</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

**Conexão simplificada**

| ssh poco |
|----------|

## 5.4 Túnel HTTP pelo SSH

O túnel permite que os servidores escutem apenas em \`127.0.0.1\` no
celular. O navegador do computador os acessa como se estivessem
executando localmente.

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>ssh poco \<br />
-L 5173:127.0.0.1:5173 \<br />
-L 8080:127.0.0.1:8080 \<br />
-L 8000:127.0.0.1:8000</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

| **No computador**     | **No POCO**    | **Uso**        |
|-----------------------|----------------|----------------|
| http://localhost:5173 | 127.0.0.1:5173 | Vite/Vue/React |
| http://localhost:8080 | 127.0.0.1:8080 | API Go         |
| http://localhost:8000 | 127.0.0.1:8000 | Python/FastAPI |

## 5.5 Acesso fora de casa

Use o aplicativo Android do Tailscale como VPN privada. Depois, conecte
pelo IP privado ou nome MagicDNS do celular. Evite encaminhar a porta
8022 diretamente no roteador para a internet.

| ssh -p 8022 u0_a123@100.x.y.z |
|-------------------------------|

# 6. Persistência com tmux

O \`tmux\` é indispensável porque mantém editor, shells e servidores
ativos mesmo quando a conexão SSH cai.

**Criar sessão**

| tmux new -s dev |
|-----------------|

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th># Desanexar sem encerrar<br />
Ctrl+B, depois D<br />
<br />
# Reconectar<br />
tmux attach -t dev<br />
<br />
# Listar sessões<br />
tmux ls</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

| **Janela** | **Comando**          | **Função**       |
|------------|----------------------|------------------|
| 1          | nvim .               | Editor principal |
| 2          | npm run dev          | Frontend         |
| 3          | go run ./cmd/api     | Backend Go       |
| 4          | lazygit              | Git              |
| 5          | htop ou btop         | Monitoramento    |
| 6          | tail -f logs/app.log | Logs             |

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th><p><strong>VTM ou tmux?</strong></p>
<p>O `tmux` é a base confiável de persistência. O VTM é a camada visual.
Você pode usar apenas tmux, apenas VTM, ou combinar os dois; para
produção pessoal, mantenha os serviços importantes dentro de
tmux.</p></th>
</tr>
</thead>
<tbody>
</tbody>
</table>

# 7. Linguagens e servidores HTTP

## 7.1 Go

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>mkdir -p ~/code/go/hello<br />
cd ~/code/go/hello<br />
go mod init hello</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

**main.go**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>package main<br />
<br />
import (<br />
"fmt"<br />
"net/http"<br />
)<br />
<br />
func main() {<br />
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request)
{<br />
fmt.Fprintln(w, "Go rodando no POCO F6")<br />
})<br />
<br />
fmt.Println("Servidor em http://127.0.0.1:8080")<br />
_ = http.ListenAndServe("127.0.0.1:8080", nil)<br />
}</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

| go run . |
|----------|

## 7.2 Node.js e Vite

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>mkdir -p ~/code/node/app<br />
cd ~/code/node/app<br />
npm create vite@latest .<br />
npm install<br />
npm run dev -- --host 127.0.0.1</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

Acesse o Vite pelo túnel SSH em \`http://localhost:5173\`. Para acesso
direto pela LAN, use \`--host 0.0.0.0\`, sabendo que o serviço ficará
exposto na rede local.

## 7.3 Python

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>mkdir -p ~/code/python/api<br />
cd ~/code/python/api<br />
<br />
python -m venv .venv<br />
source .venv/bin/activate<br />
<br />
pip install fastapi uvicorn</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

**main.py**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>from fastapi import FastAPI<br />
<br />
app = FastAPI()<br />
<br />
@app.get("/")<br />
def home():<br />
return {"message": "Python rodando no POCO F6"}</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

| uvicorn main:app --host 127.0.0.1 --port 8000 |
|-----------------------------------------------|

## 7.4 Compatibilidade no Termux

| **Stack**  | **Funciona bem**                                         | **Possíveis problemas**                         |
|------------|----------------------------------------------------------|-------------------------------------------------|
| Go         | APIs, CLIs, Templ, testes, SQLite e builds ARM64.        | CGO exige bibliotecas Android compatíveis.      |
| Node.js    | Vite, Vue, React, TypeScript, Express e ferramentas npm. | Módulos nativos que só fornecem binários glibc. |
| Python     | Scripts, FastAPI, automações e pacotes Python puros.     | Wheels manylinux podem não funcionar em Bionic. |
| Rust/C/C++ | Compilação nativa ARM64 com Clang/Rust.                  | Projetos podem assumir paths ou libc GNU/Linux. |

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th><p><strong>Fallback para incompatibilidades</strong></p>
<p>Quando uma ferramenta exigir `linux-arm64` com glibc, execute-a
dentro do Debian PRoot em vez de tentar forçar o binário diretamente no
Termux.</p></th>
</tr>
</thead>
<tbody>
</tbody>
</table>

# 8. Debian via PRoot

O PRoot cria um ambiente Debian ARM64 com glibc sem exigir root. Ele é
útil para compatibilidade, mas não deve substituir o Termux nativo
quando desempenho e estabilidade forem prioridade.

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>pkg install proot-distro<br />
<br />
proot-distro install debian<br />
proot-distro login debian</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>apt update<br />
apt install -y \<br />
git \<br />
neovim \<br />
nodejs \<br />
npm \<br />
golang \<br />
python3 \<br />
python3-venv \<br />
build-essential</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

| **Use o Termux nativo para**       | **Use o Debian PRoot para**                   |
|------------------------------------|-----------------------------------------------|
| Git, SSH, tmux e Neovim.           | Instaladores que assumem Debian/Ubuntu.       |
| Go, Node e Python compatíveis.     | Binários GNU/Linux ARM64/glibc.               |
| Servidores HTTP e builds rápidos.  | Ferramentas sem pacote Termux.                |
| Projetos e arquivos em \`\$HOME\`. | Testes de compatibilidade Linux convencional. |

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th><p><strong>Limitação estrutural</strong></p>
<p>PRoot não oferece root verdadeiro, cgroups, namespaces completos ou
Docker. Ele também adiciona custo em operações de filesystem e criação
de processos.</p></th>
</tr>
</thead>
<tbody>
</tbody>
</table>

# 9. Estabilidade no HyperOS

O maior risco operacional não é falta de CPU ou RAM. É o Android/HyperOS
encerrar processos em segundo plano, especialmente ambientes com muitos
processos filhos.

## 9.1 Ajustes obrigatórios

- Definir o Termux como aplicativo sem restrições de bateria.

- Habilitar inicialização automática/autostart.

- Bloquear o Termux na tela de aplicativos recentes.

- Permitir notificações e atividade em segundo plano.

- Evitar economia extrema de bateria durante desenvolvimento.

**Evitar suspensão agressiva durante a sessão**

| termux-wake-lock |
|------------------|

**Liberar quando terminar**

| termux-wake-unlock |
|--------------------|

## 9.2 Por que desktops gráficos travavam

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>Termux<br />
└── PRoot<br />
└── Debian/Ubuntu<br />
└── LXDE/XFCE<br />
├── servidor X<br />
├── DBus<br />
├── gerenciador de janelas<br />
├── VNC<br />
├── websockify<br />
└── noVNC no navegador</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

A lentidão vinha da combinação de interceptação de syscalls pelo PRoot,
muitos processos do desktop, captura/compressão de tela pelo VNC e
decodificação no navegador. O encerramento repentino podia ocorrer mesmo
com RAM livre, por políticas de processos em segundo plano e phantom
processes do Android.

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th><p><strong>Decisão recomendada</strong></p>
<p>Não use LXDE/XFCE + VNC/noVNC como ambiente principal. Para uma
aplicação gráfica isolada, considere Termux:X11. Para trabalho diário,
use TUI por SSH.</p></th>
</tr>
</thead>
<tbody>
</tbody>
</table>

## 9.3 Monitoramento

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>watch -n 2 '<br />
echo "=== MEMÓRIA ==="<br />
free -h<br />
echo<br />
echo "=== LOAD ==="<br />
uptime<br />
echo<br />
echo "=== PROCESSOS ==="<br />
ps -e --no-headers | wc -l<br />
echo<br />
echo "=== MAIORES CONSUMIDORES ==="<br />
ps -eo pid,ppid,%cpu,%mem,cmd --sort=-%cpu | head -15<br />
'</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

**Diagnóstico pelo computador**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>adb logcat -d |<br />
grep -Ei "termux|phantom|lmkd|lowmemory|kill|signal 9"</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

Procure por \`PhantomProcess\`, \`lmkd\`, \`Killing\`, \`excessive
cpu\`, \`signal 9\` ou referências ao pacote \`com.termux\`.

# 10. Inicialização automática e serviços

## 10.1 Termux:Boot

Instale o Termux:Boot pela mesma fonte do Termux e abra o aplicativo
pelo menos uma vez.

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>mkdir -p ~/.termux/boot<br />
nvim ~/.termux/boot/00-dev-server</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

**~/.termux/boot/00-dev-server**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>#!/data/data/com.termux/files/usr/bin/sh<br />
<br />
termux-wake-lock<br />
<br />
sshd<br />
<br />
tmux has-session -t dev 2&gt;/dev/null ||<br />
tmux new-session -d -s dev</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

| chmod +x ~/.termux/boot/00-dev-server |
|---------------------------------------|

## 10.2 termux-services

| pkg install termux-services |
|-----------------------------|

Reabra o shell depois da instalação.

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>sv-enable sshd<br />
sv up sshd<br />
sv status sshd</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

O mesmo mecanismo pode ser usado para APIs, workers e túneis. Não
inicialize automaticamente todos os projetos: mantenha apenas a
infraestrutura essencial.

# 11. Setup unificado recomendado

A sequência abaixo instala o núcleo, cria diretórios, configura uma
sessão persistente e deixa o aparelho pronto para receber conexões SSH.

**Bootstrap inicial**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th># 1. Atualização<br />
pkg update<br />
pkg upgrade -y<br />
<br />
# 2. Núcleo<br />
pkg install -y \<br />
git openssh tmux neovim \<br />
nodejs golang python clang make cmake pkg-config \<br />
ripgrep fd fzf jq tree rsync curl wget zip unzip \<br />
procps htop<br />
<br />
# 3. Camada TUI<br />
pkg install -y vtm lf lazygit<br />
<br />
# 4. Estrutura<br />
mkdir -p ~/code/{go,node,python}<br />
mkdir -p ~/workspace ~/.config ~/.local/bin<br />
<br />
# 5. Acesso<br />
passwd<br />
sshd<br />
<br />
# 6. Persistência<br />
termux-wake-lock<br />
tmux new-session -d -s dev</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

## 11.1 Script de inicialização manual

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>mkdir -p ~/.local/bin<br />
nvim ~/.local/bin/dev-start</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

**~/.local/bin/dev-start**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>#!/data/data/com.termux/files/usr/bin/sh<br />
set -eu<br />
<br />
termux-wake-lock<br />
<br />
if ! pgrep -x sshd &gt;/dev/null 2&gt;&amp;1; then<br />
sshd<br />
fi<br />
<br />
if ! tmux has-session -t dev 2&gt;/dev/null; then<br />
tmux new-session -d -s dev<br />
fi<br />
<br />
echo "Ambiente ativo."<br />
echo "SSH: porta 8022"<br />
echo "tmux: tmux attach -t dev"</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>chmod +x ~/.local/bin/dev-start<br />
~/.local/bin/dev-start</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

## 11.2 Workspace visual diário

1.  Conecte ao POCO com \`ssh poco\`.

2.  Anexe à sessão com \`tmux attach -t dev\`.

3.  Inicie \`vtm\` quando quiser a experiência de desktop textual.

4.  Abra o explorador (\`lf\`), o editor (\`nvim\`), o Git (\`lazygit\`)
    e os servidores.

5.  Use túneis SSH para visualizar os projetos no navegador do
    computador.

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th># Terminal 1 — sessão e desktop<br />
ssh poco<br />
tmux attach -t dev<br />
vtm<br />
<br />
# Terminal 2 — túneis HTTP<br />
ssh poco \<br />
-N \<br />
-L 5173:127.0.0.1:5173 \<br />
-L 8080:127.0.0.1:8080 \<br />
-L 8000:127.0.0.1:8000</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

## 11.3 Layout operacional sugerido

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>┌─ Arquivos: lf ─────────┐ ┌─ Editor: Neovim
───────────────────┐<br />
│ ~/code/tcc │ │ package main │<br />
│ ├── cmd │ │ │<br />
│ ├── internal │ │ func main() { │<br />
│ ├── web │ │ ... │<br />
│ └── go.mod │ │ } │<br />
└────────────────────────┘ └────────────────────────────────────┘<br />
<br />
┌─ Git: lazygit ─────────┐ ┌─ Processos: htop ──────────────────┐<br />
│ staged / commits │ │ CPU / RAM / processos │<br />
└────────────────────────┘ └────────────────────────────────────┘<br />
<br />
┌─ Servidores ───────────────────────────────────────────────────┐<br />
│ Vite :5173 | Go :8080 | Python :8000 │<br />
└────────────────────────────────────────────────────────────────┘</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

# 12. Alternativas avançadas

## 12.1 Termux:X11

Use quando precisar executar uma aplicação gráfica isolada. É preferível
a transmitir um desktop inteiro por VNC/noVNC, mas ainda adiciona uma
camada gráfica e não deve ser o núcleo do setup.

## 12.2 Podroid ou VM para Docker

Uma VM com kernel próprio pode executar Docker real. No POCO F6, sem
confirmação de AVF/pKVM utilizável, o fallback tende a ser QEMU, com
desempenho inferior. Trate-a como laboratório para Compose e containers.

## 12.3 AVF e Terminal Linux oficial

A presença do Android Virtualization Framework depende do firmware da
Xiaomi. Não baseie o ambiente nele até confirmar no aparelho.

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th>adb shell pm list features | grep -i virtualization<br />
<br />
adb shell pm list packages |<br />
grep -Ei "virtualization|virtualmachine|terminal"<br />
<br />
adb shell getprop ro.build.version.release<br />
adb shell getprop ro.build.version.incremental</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

O resultado \`feature:android.software.virtualization_framework\` indica
a presença da feature, mas não garante que o Terminal Linux esteja
disponível ou que VMs não protegidas estejam liberadas.

## 12.4 Root e chroot

Root e ROM customizada permitem ambientes mais próximos de um Linux
tradicional, porém adicionam risco de perda de dados, manutenção,
problemas bancários e falhas após atualizações. Não são necessários para
SSH, Go, Node, Python, Neovim ou servidores HTTP.

# 13. Troubleshooting

| **Sintoma**                     | **Causa provável**                                | **Ação**                                                |
|---------------------------------|---------------------------------------------------|---------------------------------------------------------|
| SSH cai, mas tmux continua      | Rede ou suspensão temporária.                     | Reconectar e executar \`tmux attach -t dev\`.           |
| SSH e processos somem           | Android/HyperOS encerrou o Termux.                | Revisar bateria, autostart, wake lock e logs.           |
| Vite abre só no celular         | Servidor escutando apenas em localhost.           | Usar túnel SSH ou \`--host 0.0.0.0\`.                   |
| Git/Node falha em /sdcard       | Permissões/noexec do armazenamento compartilhado. | Mover o projeto para \`\$HOME\`.                        |
| Pacote npm nativo falha         | Binário pré-compilado para glibc.                 | Compilar, trocar pacote ou usar Debian PRoot.           |
| Wheel Python não instala        | Wheel manylinux incompatível com Bionic.          | Compilar fonte ou usar PRoot.                           |
| VTM/TUI quebra caracteres       | Fonte/locale/terminal incompatível.               | Usar UTF-8 e terminal com boa fonte monoespaçada.       |
| Desktop gráfico fica lento      | PRoot + X11/VNC + compressão.                     | Migrar para TUI por SSH.                                |
| Servidor morre com tela apagada | Política de energia.                              | Wake lock, sem restrições e app bloqueado nos recentes. |

## 13.1 Checklist de diagnóstico

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th># Rede e SSH<br />
ip addr<br />
ss -tln<br />
pgrep -a sshd<br />
<br />
# Sessões<br />
tmux ls<br />
<br />
# Recursos<br />
free -h<br />
uptime<br />
ps -eo pid,ppid,%cpu,%mem,cmd --sort=-%cpu | head -20<br />
<br />
# Servidores<br />
ss -tln | grep -E '5173|8080|8000'<br />
<br />
# Armazenamento<br />
pwd<br />
mount | grep emulated</th>
</tr>
</thead>
<tbody>
</tbody>
</table>

# 14. Segurança

- Prefira autenticação SSH por chave e desative senha depois de validar
  o acesso.

- Não exponha a porta 8022 diretamente na internet.

- Use Tailscale ou outro túnel privado para acesso externo.

- Faça os servidores escutarem em \`127.0.0.1\` e encaminhe portas pelo
  SSH.

- Não execute scripts de instalação desconhecidos como root — e no
  Termux, revise scripts remotos antes de executar.

- Mantenha repositórios no Git e backups fora do celular.

- Considere o celular uma máquina de desenvolvimento, não o único local
  de armazenamento dos projetos.

# 15. Roteiro de implementação

| **Fase**            | **Objetivo**                      | **Critério de conclusão**                          |
|---------------------|-----------------------------------|----------------------------------------------------|
| 1\. Núcleo          | Termux, Git, SSH, tmux e Neovim.  | Entrar por SSH e editar um arquivo no \`\$HOME\`.  |
| 2\. HTTP            | Go, Node e Python.                | Abrir três serviços por túnel SSH.                 |
| 3\. Persistência    | Wake lock, HyperOS e Termux:Boot. | Reiniciar e recuperar o ambiente.                  |
| 4\. Desktop TUI     | VTM, lf e lazygit.                | Operar arquivos, editor e Git em janelas textuais. |
| 5\. Compatibilidade | Debian PRoot.                     | Executar uma ferramenta que exige glibc.           |
| 6\. Acesso externo  | Tailscale.                        | Acessar o POCO fora da LAN sem abrir portas.       |
| 7\. Containers      | Podroid/VM opcional.              | Executar um Compose de laboratório.                |

# 16. Conclusão

O POCO F6 não precisa executar um desktop Linux gráfico para se tornar
uma workstation. A combinação \`Termux + SSH + tmux + Neovim\` já
oferece um ambiente de desenvolvimento real, persistente e acessível
remotamente. VTM e os demais projetos TUI adicionam a sensação de um
sistema operacional visual sem o custo de transmitir uma interface
gráfica.

A estratégia mais robusta é construir de baixo para cima: primeiro o
servidor Termux estável; depois persistência, túneis e linguagens; por
fim a camada visual. PRoot e VM permanecem ferramentas auxiliares, não o
centro do ambiente.

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th><p><strong>Stack final</strong></p>
<p>Termux nativo + OpenSSH + tmux + Neovim + VTM + lf + lazygit +
Go/Node/Python + túneis SSH.</p></th>
</tr>
</thead>
<tbody>
</tbody>
</table>

# 17. Projetos e referências úteis

**Termux:** [<u>https://termux.dev/</u>](https://termux.dev/)

**Termux packages:**
[<u>https://github.com/termux/termux-packages</u>](https://github.com/termux/termux-packages)

**Termux:Boot:**
[<u>https://github.com/termux/termux-boot</u>](https://github.com/termux/termux-boot)

**termux-services:**
[<u>https://github.com/termux/termux-services</u>](https://github.com/termux/termux-services)

**PRoot-Distro:**
[<u>https://github.com/termux/proot-distro</u>](https://github.com/termux/proot-distro)

**Termux:X11:**
[<u>https://github.com/termux/termux-x11</u>](https://github.com/termux/termux-x11)

**VTM:**
[<u>https://github.com/directvt/vtm</u>](https://github.com/directvt/vtm)

**Twin:**
[<u>https://github.com/cosmos72/twin</u>](https://github.com/cosmos72/twin)

**Desktop-TUI:**
[<u>https://github.com/Julien-cpsn/desktop-tui</u>](https://github.com/Julien-cpsn/desktop-tui)

**TermOS:**
[<u>https://github.com/ThatOtherAndrew/TermOS</u>](https://github.com/ThatOtherAndrew/TermOS)

**WibWob-DOS:**
[<u>https://github.com/j-greig/wibandwob-dos</u>](https://github.com/j-greig/wibandwob-dos)

**TUIFI Manager:**
[<u>https://github.com/GiorgosXou/TUIFIManager</u>](https://github.com/GiorgosXou/TUIFIManager)

**code-server no Termux:**
[<u>https://coder.com/docs/code-server/termux</u>](https://coder.com/docs/code-server/termux)

**Tailscale:**
[<u>https://tailscale.com/kb/1017/install</u>](https://tailscale.com/kb/1017/install)

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<thead>
<tr class="header">
<th><p><strong>Nota de atualização</strong></p>
<p>Projetos experimentais, pacotes e compatibilidade com Android mudam
rapidamente. Antes de instalar uma ferramenta menos conhecida, confira
releases recentes, issues abertas, arquitetura ARM64 e instruções
específicas para Termux.</p></th>
</tr>
</thead>
<tbody>
</tbody>
</table>
