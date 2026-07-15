---
title: "POCO F6 como Workstation Móvel"
subtitle: "Termux, SSH, desenvolvimento, aplicações TUI/CLI, PRoot e ambientes reproduzíveis com Nix-on-Droid"
author: "Érick Lúcio"
date: "Julho de 2026"
lang: pt-BR
---

# Sumário

1. Apresentação e escopo
2. Resumo executivo
3. O POCO F6 como nó ARM64 de desenvolvimento
4. Como o Termux executa programas
5. Arquitetura recomendada
6. Termux nativo versus Ubuntu/Debian via PRoot
7. Instalação base do Termux
8. Servidor SSH, acesso remoto, túneis e persistência
9. Desenvolvimento com Go, Node.js, Python, Rust, Java e Kotlin
10. Ecossistema de aplicações CLI e TUI
11. tmux, Zellij e VTM
12. Configuração do VTM
13. Neovim e LazyVim com atalhos híbridos
14. Nix, NixOS e Nix-on-Droid
15. Repositório declarativo com Home Manager
16. Ambientes por projeto com Nix
17. Setup reproduzível sem Nix
18. Automação, CLI, TUI e aplicativo Android complementar
19. Atalhos e integração com a tela inicial
20. Interface gráfica, X11 e desktop Linux
21. Docker, containers e máquinas virtuais
22. Segurança, estabilidade e desempenho
23. Troubleshooting e diagnóstico
24. Roteiro de implementação
25. Decisão final por perfil de uso
26. Arquitetura final recomendada
27. Apêndices: bootstrap, diagnóstico, SSH, LazyVim, VTM e comandos
28. Referências oficiais e encerramento

# Apresentação

Este documento consolida as decisões, comparações, dúvidas e experimentos discutidos nos últimos dias sobre transformar um **POCO F6** em uma estação móvel de desenvolvimento. O objetivo não é fingir que o Android é um notebook Linux tradicional, mas explorar corretamente o que o aparelho já oferece: processador ARM64 potente, kernel Linux, conectividade, bateria, armazenamento e capacidade de executar ferramentas de desenvolvimento diretamente pelo Termux.

O projeto parte de uma ideia simples:

> O celular executa os processos; o notebook, computador ou tablet pode funcionar apenas como terminal, teclado, monitor e navegador.

Com essa arquitetura, o POCO F6 pode hospedar repositórios Git, editores de terminal, compiladores, servidores HTTP, ferramentas TUI e sessões persistentes. Quando uma ferramenta não é compatível com o ambiente Android/Bionic, entra uma distribuição GNU/Linux por PRoot. Quando a prioridade passa a ser configuração declarativa, rollback e reprodução do setup, entra o Nix-on-Droid.

## Escopo

Este guia cobre:

- Termux nativo e seu modelo de execução;
- servidor SSH no celular;
- acesso local e remoto;
- persistência com tmux;
- desenvolvimento e execução de projetos Node.js, Go, Python, Rust, Java e Kotlin;
- compatibilidade de aplicações CLI e TUI da comunidade;
- VTM como desktop textual;
- Neovim e LazyVim com atalhos mais convencionais;
- Ubuntu/Debian via PRoot;
- Nix e Nix-on-Droid para declarar o setup completo;
- configuração de plugins, temas, cores, dotfiles e launchers;
- automação por scripts, TUI própria ou aplicativo Android complementar;
- limitações de Docker, máquinas virtuais, X11 e desktop gráfico;
- segurança, estabilidade no HyperOS e troubleshooting.

O experimento de navegador remoto foi deliberadamente deixado fora do núcleo deste documento. Ele é uma frente separada e não deve complicar a primeira versão da workstation.

# 1. Resumo executivo

## 1.1 Recomendação principal

A arquitetura mais equilibrada para o seu caso é:

```text
POCO F6 / HyperOS
│
├── Termux nativo
│   ├── OpenSSH Server
│   ├── tmux
│   ├── Neovim ou LazyVim
│   ├── Git + lazygit
│   ├── Go, Node.js, Python, Rust, Java e Kotlin
│   ├── VTM e outras aplicações TUI
│   └── servidores HTTP locais
│
├── PRoot Ubuntu/Debian, sob demanda
│   └── ferramentas que exigem glibc ou layout GNU/Linux tradicional
│
└── Nix-on-Droid, como alternativa de ambiente principal
    ├── pacotes declarados
    ├── configurações declaradas
    ├── Home Manager
    ├── switch e rollback
    └── repositório reproduzível do setup
```

A decisão mais importante é não colocar todas as camadas ao mesmo tempo. O caminho recomendado é incremental:

1. tornar o Termux nativo estável;
2. ativar SSH e tmux;
3. instalar linguagens e ferramentas;
4. adicionar Neovim/LazyVim e aplicações TUI;
5. usar PRoot somente para incompatibilidades reais;
6. testar Nix-on-Droid quando a configuração declarativa passar a valer mais que a simplicidade;
7. tratar VM, Docker e desktop gráfico apenas como laboratórios opcionais.

## 1.2 Decisão em uma frase

**Para programar e rodar a maioria dos seus projetos, use Termux nativo. Para binários e ferramentas presos à glibc, use PRoot. Para reconstruir todo o setup a partir de código, use Nix-on-Droid.**

## 1.3 Stack final sugerida

```text
Termux nativo
+ OpenSSH
+ tmux
+ Neovim/LazyVim
+ Git/lazygit
+ VTM
+ lf ou Yazi
+ Go/Node/Python/Rust/JDK/Kotlin
+ Tailscale
+ scripts de bootstrap e backup
```

Alternativa declarativa:

```text
Nix-on-Droid
+ Home Manager
+ repositório de configuração
+ VTM configurado
+ LazyVim configurado
+ linguagens e CLIs declaradas
```

# 2. O que o POCO F6 será nesse projeto

O POCO F6 não precisa ser tratado como um “desktop Linux gráfico em miniatura”. Esse modelo costuma levar rapidamente a PRoot, X11, VNC, desktop completo e consumo excessivo. A ideia mais eficiente é considerá-lo um **nó ARM64 de desenvolvimento**.

Ele pode cumprir quatro papéis ao mesmo tempo:

1. **Servidor SSH:** recebe conexões de outro computador.
2. **Máquina de desenvolvimento:** mantém código, Git, compiladores e dependências.
3. **Servidor HTTP:** executa APIs, front-ends e ferramentas web.
4. **Workstation TUI:** apresenta editor, arquivos, Git, processos e logs dentro do terminal.

## 2.1 O que acontece no celular

- compilação;
- execução dos programas;
- armazenamento dos repositórios;
- servidores HTTP;
- Git;
- editor e LSP;
- shells e sessões tmux;
- aplicações TUI;
- tarefas agendadas e scripts.

## 2.2 O que o computador cliente fornece

- teclado físico;
- monitor maior;
- terminal com fonte melhor;
- mouse;
- navegador para acessar aplicações HTTP;
- cliente SSH.

O computador não precisa possuir as linguagens instaladas. Ele pode funcionar apenas como interface para os processos que estão no POCO.

# 3. Termux não é uma distribuição Linux tradicional

O Termux executa programas **nativamente no Android**, usando o kernel Linux do aparelho. Ele não inicia uma máquina virtual e não coloca uma distribuição Ubuntu inteira por baixo de cada comando.

A diferença principal está no userland:

```text
Linux tradicional
└── kernel Linux + glibc + /usr/bin + filesystem convencional

Termux
└── kernel Linux do Android + Bionic + prefixo privado do aplicativo
```

Os pacotes do Termux são compilados para:

```text
/data/data/com.termux/files/usr
```

Esse diretório funciona como o prefixo do ambiente. Em vez de depender de `/usr/bin`, `/etc` e `/var` do sistema Android, o Termux mantém seus próprios binários, bibliotecas e configurações dentro do espaço privado do aplicativo.

## 3.1 Vantagem dessa arquitetura

Os programas rodam diretamente na CPU ARM64, sem emulação de processador. Isso favorece:

- Go;
- Node.js;
- Python;
- Rust;
- Java;
- Kotlin/JVM;
- Git;
- SSH;
- Neovim;
- CLIs e TUIs compiladas para Android/Termux.

## 3.2 Principal incompatibilidade

O Android usa **Bionic**, enquanto a maioria das distribuições GNU/Linux usa **glibc**. Portanto, um arquivo anunciado como:

```text
linux-arm64
```

não é automaticamente compatível com o Termux. Ele pode ter sido compilado para Linux ARM64 com glibc e falhar no Android, mesmo que a arquitetura da CPU esteja correta.

A ordem correta de preferência é:

1. pacote oficial do Termux;
2. compilação do código-fonte dentro do Termux;
3. versão específica para Android/Termux;
4. PRoot Ubuntu/Debian para binários glibc;
5. VM somente quando houver dependência real de kernel ou container.

## 3.3 Por que os projetos devem ficar em `$HOME`

O armazenamento compartilhado do Android, normalmente acessado por `/sdcard` ou `~/storage`, possui diferenças importantes:

- pode ser montado com `noexec`;
- não preserva todas as permissões Unix;
- pode não suportar links simbólicos corretamente;
- pode quebrar ferramentas de build;
- pode causar problemas em repositórios Git;
- permite acesso por outros aplicativos.

Use:

```text
/data/data/com.termux/files/home
```

como área real de desenvolvimento. Use `~/storage/downloads` apenas para importar e exportar arquivos.

# 4. Arquitetura recomendada

## 4.1 Camadas

| Camada | Ferramenta | Responsabilidade |
|---|---|---|
| Sistema hospedeiro | Android/HyperOS | Kernel, rede, energia, armazenamento e permissões |
| Userland principal | Termux | Desenvolvimento e execução nativa |
| Acesso | OpenSSH | Terminal remoto e encaminhamento de portas |
| Persistência | tmux | Manter sessões após desconexões |
| Editor | Neovim/LazyVim | Código, LSP, Git e terminal integrado |
| Desktop textual | VTM | Organizar aplicações TUI em janelas |
| Arquivos | lf, Yazi, ranger, TUIFI ou Superfile | Navegação visual |
| Git | lazygit | Staging, commits, branches e diffs |
| Compatibilidade | Ubuntu/Debian PRoot | glibc e estrutura Linux convencional |
| Reprodutibilidade | Nix-on-Droid + Home Manager | Pacotes e configurações declarativas |
| Acesso externo | Tailscale | Rede privada sem abrir a porta SSH na internet |

## 4.2 Regra de ouro

> Tudo que funciona bem no Termux deve permanecer no Termux.

Cada camada adicional aumenta:

- consumo de armazenamento;
- uso de memória;
- tempo de inicialização;
- complexidade de diagnóstico;
- quantidade de incompatibilidades;
- dificuldade de integrar com o Android.

# 5. Termux nativo versus Ubuntu/Debian via PRoot

## 5.1 Termux nativo

### Vantagens

- execução direta no Android;
- melhor desempenho;
- menor consumo de armazenamento;
- integração com Termux:API, Termux:Boot e Termux:Widget;
- acesso simples ao armazenamento do aplicativo;
- pacotes adaptados ao Android;
- ótimo para servidores, scripts, CLIs e TUIs;
- ideal para o uso diário.

### Desvantagens

- Bionic em vez de glibc;
- caminhos diferentes de uma distribuição tradicional;
- scripts que assumem `/bin/bash` podem falhar;
- binários Linux ARM64 comuns podem não executar;
- alguns módulos nativos de npm e wheels Python não possuem build Android;
- certas ferramentas esperam recursos de uma distribuição completa.

## 5.2 Ubuntu ou Debian via PRoot

O PRoot oferece um filesystem e um userland semelhantes a uma distribuição Linux sem exigir root. Ele intercepta chamadas de sistema e reescreve caminhos para simular operações parecidas com `chroot` e bind mounts.

### Vantagens

- `apt` e estrutura familiar;
- glibc;
- maior compatibilidade com binários Linux ARM64;
- tutoriais de Ubuntu/Debian funcionam com menos adaptações;
- ambiente isolado para ferramentas específicas;
- boa solução de fallback.

### Desvantagens

- sobrecarga de execução;
- mais armazenamento;
- inicialização e builds mais lentos;
- não fornece root real;
- não fornece namespaces, cgroups, seccomp ou isolamento equivalente a Docker;
- `systemd` não funciona como em uma máquina tradicional;
- acesso a dispositivos e recursos do kernel continua limitado pelo Android;
- integração com Termux e Android fica menos direta.

## 5.3 Matriz de decisão

| Situação | Melhor escolha |
|---|---|
| Desenvolver Go, Node, Python ou Rust | Termux nativo |
| Rodar seus próprios programas | Termux nativo |
| Java e Kotlin CLI/backend | Termux nativo |
| Aplicações TUI disponíveis no `pkg` | Termux nativo |
| Compilar projeto open source | Primeiro Termux nativo |
| Binário `linux-arm64` ligado à glibc | PRoot |
| Script que depende fortemente de Ubuntu | PRoot |
| Ferramenta que exige `/usr`, `/bin/bash` e pacotes Debian | PRoot |
| Docker real ou recursos avançados de kernel | VM/root, não PRoot |

## 5.4 Estratégia híbrida

A melhor estratégia não é escolher um ambiente e apagar o outro. É usar:

```text
Termux = ambiente principal
PRoot  = caixa de compatibilidade
```

Os repositórios podem continuar no `$HOME` do Termux e ser expostos ao PRoot somente quando necessário.

# 6. Instalação base do Termux

## 6.1 Fonte do aplicativo

Instale Termux e seus complementos a partir de fontes mantidas pelo projeto, mantendo todos os APKs na mesma família de assinatura. Evite versões antigas ou forks aleatórios.

## 6.2 Atualização inicial

```bash
pkg update
pkg upgrade -y
```

## 6.3 Núcleo recomendado

```bash
pkg install -y \
  git \
  openssh \
  tmux \
  neovim \
  nodejs \
  golang \
  python \
  rust \
  clang \
  make \
  cmake \
  pkg-config \
  openjdk-21 \
  ripgrep \
  fd \
  fzf \
  jq \
  tree \
  rsync \
  curl \
  wget \
  zip \
  unzip \
  procps \
  htop
```

O nome exato do pacote JDK ou de outras ferramentas pode mudar entre repositórios. Antes de concluir que uma tecnologia não está disponível, pesquise:

```bash
pkg search openjdk
pkg search kotlin
pkg search yazi
pkg search zellij
```

## 6.4 Camada TUI inicial

```bash
pkg install -y vtm lf lazygit
```

Instale os demais conforme a necessidade, evitando transformar o setup inicial em uma coleção de ferramentas que você ainda não usa.

## 6.5 Estrutura de diretórios

```bash
mkdir -p ~/code/{go,node,python,rust,java,kotlin}
mkdir -p ~/workspace
mkdir -p ~/.config
mkdir -p ~/.local/bin
mkdir -p ~/backups
```

Estrutura sugerida:

```text
~
├── code/
│   ├── go/
│   ├── node/
│   ├── python/
│   ├── rust/
│   ├── java/
│   └── kotlin/
├── workspace/
├── .ssh/
├── .config/
├── .local/bin/
└── backups/
```

## 6.6 Validação

```bash
uname -m
git --version
ssh -V
tmux -V
nvim --version
node --version
npm --version
go version
python --version
rustc --version
cargo --version
java --version
```

O resultado de `uname -m` deve indicar `aarch64`.

# 7. Transformando o POCO F6 em servidor SSH

## 7.1 Iniciar o servidor

```bash
passwd
whoami
sshd
ss -tln | grep 8022
```

No Termux, o OpenSSH normalmente utiliza a porta `8022`, porque aplicativos Android comuns não podem abrir portas privilegiadas como `22` sem permissões especiais.

Descubra o endereço do aparelho na rede local:

```bash
ip addr show wlan0
```

Conexão a partir do computador:

```bash
ssh -p 8022 usuario_do_termux@IP_DO_POCO
```

O usuário costuma ter formato semelhante a:

```text
u0_a123
```

Use o resultado real de `whoami`.

## 7.2 Configuração conveniente no computador

No computador cliente, crie ou edite:

```text
~/.ssh/config
```

Exemplo:

```sshconfig
Host poco
    HostName 192.168.1.80
    User u0_a123
    Port 8022
    ServerAliveInterval 30
    ServerAliveCountMax 3
```

Depois:

```bash
ssh poco
```

## 7.3 Autenticação por chave

No computador:

```bash
ssh-keygen -t ed25519
ssh-copy-id -p 8022 usuario_do_termux@IP_DO_POCO
```

Teste:

```bash
ssh poco
```

Depois de confirmar que a chave funciona, endureça a configuração em:

```text
$PREFIX/etc/ssh/sshd_config
```

Configuração recomendada:

```text
PubkeyAuthentication yes
PasswordAuthentication no
PermitEmptyPasswords no
```

Reinicie o servidor:

```bash
pkill sshd
sshd
```

Mantenha uma sessão já conectada durante a alteração para não se bloquear por um erro de configuração.

## 7.4 Encaminhamento de portas

A forma mais segura de acessar servidores de desenvolvimento é mantê-los em `127.0.0.1` no celular e encaminhar as portas pelo SSH.

```bash
ssh poco \
  -L 5173:127.0.0.1:5173 \
  -L 8080:127.0.0.1:8080 \
  -L 8000:127.0.0.1:8000
```

No navegador do computador:

```text
http://localhost:5173
http://localhost:8080
http://localhost:8000
```

Isso evita expor cada servidor para todos os dispositivos da rede local.

## 7.5 Acesso fora de casa

Não encaminhe a porta `8022` diretamente no roteador para a internet. Use uma rede privada como o Tailscale.

Arquitetura:

```text
Computador ── rede privada Tailscale ── POCO F6
```

O Tailscale pode rodar como aplicativo Android. O SSH continua rodando no Termux, mas o computador utiliza o endereço privado fornecido pela VPN.

## 7.6 Persistência com tmux

O SSH é apenas a conexão. Se ela cair, os processos iniciados diretamente naquele terminal podem receber sinais de encerramento. O tmux cria sessões persistentes.

```bash
tmux new -s dev
```

Dentro da sessão, execute editor, servidores e logs. Para sair sem encerrar:

```text
Ctrl+B, depois D
```

Para retornar:

```bash
tmux attach -t dev
```

Lista de sessões:

```bash
tmux ls
```

## 7.7 Inicialização automática

### Opção A: Termux:Boot

Crie:

```text
~/.termux/boot/start-dev.sh
```

```bash
#!/data/data/com.termux/files/usr/bin/bash

termux-wake-lock
sshd

tmux has-session -t dev 2>/dev/null || \
  tmux new-session -d -s dev
```

Permissão:

```bash
chmod +x ~/.termux/boot/start-dev.sh
```

Abra o Termux:Boot ao menos uma vez após instalar e configure o HyperOS para permitir inicialização e execução em segundo plano.

### Opção B: termux-services

```bash
pkg install termux-services
```

Reinicie a sessão do Termux e habilite:

```bash
sv-enable sshd
sv up sshd
```

O nome e a disponibilidade do serviço devem ser confirmados na instalação atual.

## 7.8 Sobre o HyperOS matar processos

O maior risco prático não é o SSH. É o gerenciamento agressivo de energia do Android/HyperOS.

Revise:

- bateria sem restrições para Termux;
- inicialização automática;
- execução em segundo plano;
- aplicativo bloqueado na tela de recentes;
- wake lock enquanto o aparelho estiver servindo;
- notificações do Termux habilitadas;
- uso de tmux para recuperar o trabalho após queda de rede.

`termux-wake-lock` reduz a chance de suspensão, mas não substitui as configurações do sistema.

# 8. Desenvolvimento por linguagem

## 8.1 Visão geral

| Tecnologia | Desenvolvimento no Termux | Execução no Termux | Principais cuidados |
|---|---|---|---|
| Node.js | Muito bom | Muito bom | módulos nativos e binários glibc |
| Go | Excelente | Excelente | CGO e bibliotecas específicas |
| Python | Muito bom | Muito bom | wheels manylinux e pacotes nativos |
| Rust | Muito bom | Excelente | crates com dependências de sistema |
| Java | Bom | Muito bom | uso de RAM, Gradle e ausência de GUI desktop |
| Kotlin | Bom | Muito bom na JVM | Gradle, SDK Android e builds pesados |

## 8.2 Go

Go combina muito bem com o celular porque produz binários, possui toolchain simples e grande parte do ecossistema pode trabalhar sem CGO.

```bash
mkdir -p ~/code/go/hello
cd ~/code/go/hello

go mod init example.com/hello
cat > main.go <<'GO'
package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Go rodando no POCO F6")
    })

    if err := http.ListenAndServe("127.0.0.1:8080", nil); err != nil {
        panic(err)
    }
}
GO

go run .
```

Acesse por túnel SSH em `http://localhost:8080`.

### Pontos fortes

- APIs HTTP com `net/http`;
- CLIs;
- aplicações TUI com Bubble Tea;
- Templ e SSR;
- testes;
- ferramentas de automação;
- build de binários ARM64.

### Cuidados

- bibliotecas com CGO precisam de Clang e dependências compatíveis;
- uma dependência que assume glibc pode falhar;
- banco SQLite por CGO pode exigir preparação adicional, enquanto implementações puras em Go tendem a ser mais simples.

## 8.3 Node.js e TypeScript

```bash
mkdir -p ~/code/node/app
cd ~/code/node/app
npm create vite@latest .
npm install
npm run dev -- --host 127.0.0.1
```

Túnel:

```bash
ssh poco -L 5173:127.0.0.1:5173
```

### Funciona bem

- TypeScript;
- Vite;
- Vue;
- React;
- Express;
- Fastify;
- ferramentas npm escritas em JavaScript;
- CLIs Node.

### Pode falhar

- pacotes que baixam binários precompilados apenas para glibc;
- addons N-API sem variante Android;
- ferramentas que dependem de navegador desktop ou sandbox específica;
- Bun e runtimes alternativos sem suporte Android adequado.

Quando um pacote nativo falhar, tente:

1. verificar se existe compilação local;
2. instalar Clang, Make e `pkg-config`;
3. trocar por biblioteca pura em JavaScript;
4. usar Ubuntu/Debian PRoot.

## 8.4 Python

```bash
mkdir -p ~/code/python/api
cd ~/code/python/api
python -m venv .venv
source .venv/bin/activate
pip install fastapi uvicorn
```

`main.py`:

```python
from fastapi import FastAPI

app = FastAPI()


@app.get("/")
def home() -> dict[str, str]:
    return {"message": "Python rodando no POCO F6"}
```

Execução:

```bash
uvicorn main:app --host 127.0.0.1 --port 8000
```

### Funciona bem

- scripts;
- FastAPI;
- automações;
- clientes HTTP;
- parsing;
- ferramentas de dados leves;
- TUIs com Textual, Rich ou curses.

### Limitação comum

O PyPI distribui muitos pacotes como wheels `manylinux`, criados para glibc. O pip pode precisar compilar o pacote no aparelho ou pode não conseguir instalar a dependência.

Alternativas:

- pacote do repositório Termux;
- compilação com dependências locais;
- biblioteca Python pura;
- PRoot para o ecossistema manylinux.

## 8.5 Rust

```bash
cargo new ~/code/rust/hello
cd ~/code/rust/hello
cargo run
```

Rust é adequado para:

- CLIs rápidas;
- exploradores de arquivos;
- parsers;
- ferramentas de sistema em userland;
- TUIs;
- servidores HTTP.

Crates puramente Rust tendem a funcionar melhor. Crates que chamam bibliotecas C, assumem caminhos de Linux desktop ou dependem de recursos específicos de kernel podem exigir ajustes.

## 8.6 Java

```bash
mkdir -p ~/code/java/hello
cd ~/code/java/hello

cat > Main.java <<'JAVA'
public class Main {
    public static void main(String[] args) {
        System.out.println("Java rodando no POCO F6");
    }
}
JAVA

javac Main.java
java Main
```

Java funciona bem para:

- aplicações CLI;
- backends;
- ferramentas Gradle/Maven;
- servidores HTTP;
- utilitários JVM.

Limitações práticas:

- Gradle pode consumir bastante memória e armazenamento;
- frameworks grandes são mais pesados no celular;
- aplicações Swing/JavaFX exigem ambiente gráfico e não são o foco do setup TUI;
- o daemon do Gradle pode permanecer consumindo memória.

Para economizar recursos:

```bash
./gradlew --no-daemon build
```

## 8.7 Kotlin

Kotlin pode ser utilizado de duas formas principais:

1. compilador Kotlin CLI;
2. projeto JVM com Gradle.

Exemplo conceitual:

```kotlin
fun main() {
    println("Kotlin rodando no POCO F6")
}
```

Kotlin/JVM para CLI e backend é viável. Desenvolver aplicativos Android completos diretamente no celular também é tecnicamente possível, mas envolve Android SDK, Gradle, ferramentas de build e alto consumo. Para seu objetivo atual, Kotlin no Termux é mais interessante para:

- estudo da linguagem;
- CLIs;
- Ktor;
- bibliotecas JVM;
- scripts e automações.

## 8.8 Separar dependências globais e dependências de projeto

Evite instalar tudo globalmente. Use:

- `go.mod` para Go;
- `package.json` e lockfile para Node;
- `.venv` para Python;
- `Cargo.toml` para Rust;
- Gradle/Maven para Java e Kotlin;
- `nix develop` quando estiver no Nix-on-Droid.

# 9. Ecossistema CLI e TUI

## 9.1 O que determina se uma TUI funciona

Uma aplicação de terminal não depende de interface gráfica tradicional, mas ainda pode depender de:

- arquitetura ARM64;
- libc;
- PTY;
- chamadas de sistema Unix;
- clipboard;
- fontes e ícones;
- protocolo de mouse do terminal;
- shell e caminhos;
- bibliotecas nativas;
- runtime como Python, Go, Rust, Node ou Bun.

A linguagem do projeto é um sinal, não uma garantia. Um programa em Go pode compilar perfeitamente; outro pode usar chamadas Linux específicas. Um projeto em Python pode funcionar imediatamente; outro pode depender de bibliotecas nativas.

## 9.2 Compatibilidade dos projetos discutidos

| Projeto | Tecnologia | Situação esperada no Termux | Recomendação |
|---|---|---|---|
| LazyVim | Lua/Neovim | Funciona | Recomendado, com ajustes de LSP e clipboard |
| broot | Rust | Funciona bem | Recomendado |
| ranger | Python | Funciona bem | Recomendado |
| Superfile | Go | Provável; pode exigir build | Testar depois de lf/Yazi |
| Yazi | Rust | Funciona bem quando pacote/build disponível | Recomendado |
| openmux | TypeScript/OpenTUI | Difícil devido a Bun e PTY nativo | Melhor em PRoot ou aguardar suporte |
| procmux | Python | Provável | Experimental |
| tuios | Go | Pode compilar; PTY e processos exigem teste | Experimental |
| Zellij | Rust | Funciona quando disponível no ecossistema Termux | Boa alternativa ao tmux |
| lf | Go | Funciona bem | Melhor ponto de partida |
| TUIFI Manager | Python | Focado em Termux | Bom para visual de “desktop” |
| desktop-tui | Rust | Pode compilar; suporte Android não é garantido | Projeto de estudo |
| VTM | C++ | Possui caminho no ecossistema Termux | Principal desktop textual |

A disponibilidade exata dos pacotes muda. Antes de compilar:

```bash
pkg search nome-do-projeto
```

## 9.3 Categorias diferentes

Não trate todas essas ferramentas como concorrentes diretas.

### Multiplexadores e desktops

- tmux;
- Zellij;
- VTM;
- tuios;
- openmux;
- desktop-tui.

### Gerenciadores de arquivos

- lf;
- Yazi;
- ranger;
- Superfile;
- TUIFI Manager;
- Midnight Commander;
- broot, que combina navegação e árvore.

### Desenvolvimento

- Neovim/LazyVim;
- lazygit;
- btop/htop;
- ripgrep;
- fzf;
- GitHub CLI;
- clientes de banco e HTTP em terminal.

## 9.4 Touchscreen

Uma TUI pode funcionar em uma tela touch, mas o modelo de interação continua sendo o de terminal.

Possibilidades:

- toques usados pelo próprio emulador para posicionar cursor, selecionar e rolar;
- gestos convertidos em eventos de mouse quando o terminal e a aplicação suportam protocolo de mouse;
- teclado virtual para atalhos;
- mouse Bluetooth ou USB;
- acesso por SSH a partir de um computador, onde o terminal cliente oferece mouse e teclado completos.

Limitações:

- nem toda TUI entende touch diretamente;
- combinações como `Ctrl`, `Alt`, `Esc` e teclas de função são desconfortáveis no teclado virtual;
- múltiplos gestos podem ser interceptados pelo Termux em vez do programa;
- TUIs desenhadas para telas grandes podem ficar comprimidas.

Para produtividade real, o celular pode ficar como servidor e o notebook como interface. Para uso local, um teclado Bluetooth transforma bastante a experiência.

# 10. tmux, Zellij e VTM: papéis diferentes

## 10.1 tmux

O tmux prioriza:

- persistência;
- sessões;
- painéis;
- confiabilidade;
- reconexão após queda do SSH.

Ele deve permanecer como camada de segurança, mesmo que você use VTM.

## 10.2 Zellij

Zellij oferece layout mais amigável, plugins e descoberta de atalhos. Pode substituir o tmux em algumas rotinas, mas o tmux ainda é mais universal e previsível.

## 10.3 VTM

VTM tenta criar um desktop textual com:

- janelas;
- barra de tarefas;
- aplicações CUI dentro de janelas;
- mouse;
- áreas de trabalho visuais;
- configurações de cores, terminal e launchers.

## 10.4 Arquitetura recomendada de sessão

```text
SSH
└── tmux session: dev
    └── VTM
        ├── Neovim
        ├── lf/Yazi
        ├── lazygit
        ├── servidor Go
        ├── servidor Node
        └── logs/htop
```

Isso pode gerar alguns conflitos de atalhos. A alternativa é manter duas janelas tmux:

```text
tmux
├── janela 1: VTM
└── janela 2: shell de recuperação
```

Assim, se o VTM travar, o tmux continua acessível.

# 11. VTM como desktop textual configurável

## 11.1 Arquivo de configuração

O VTM utiliza configuração de usuário em:

```text
~/.config/vtm/settings.xml
```

A configuração pode controlar:

- cores de janelas focadas e desfocadas;
- background;
- barra de tarefas;
- paleta do terminal;
- scrollback;
- launchers;
- aplicações iniciadas automaticamente;
- layouts de janelas.

## 11.2 Configuração mínima

```xml
<config>
  <desktop>
    <taskbar selected="Term" item*>
      <item id="Term"/>
    </taskbar>
  </desktop>
</config>
```

## 11.3 Exemplo de tema e launchers

```xml
<config>
  <colors>
    <window fgc=#D8DEE9 bgc=#2E3440/>
    <focus  fgc=#ECEFF4 bgc=#5E81AC/>
  </colors>

  <desktop>
    <background>
      <color fgc=#D8DEE9 bgc=#1B1F2AFF/>
    </background>

    <taskbar selected="Shell" item*>
      <item id="Shell" label="shell" type="term" cmd="$SHELL"/>
      <item id="Editor" label="nvim" type="term" cmd="nvim"/>
      <item id="Files" label="files" type="term" cmd="lf"/>
      <item id="Git" label="git" type="term" cmd="lazygit"/>
      <item id="Processes" label="proc" type="term" cmd="htop"/>

      <colors>
        <bground  fgc=#D8DEE9 bgc=#2E3440FF/>
        <focused  fgc=#88C0D0/>
        <selected fgc=#ECEFF4/>
        <active   fgc=#A3BE8C/>
        <inactive fgc=#4C566A/>
      </colors>
    </taskbar>
  </desktop>

  <terminal>
    <scrollback>
      <size=50000/>
      <wrap=true/>
    </scrollback>
    <colors>
      <default fgc=#D8DEE9 bgc=#1B1F2A/>
    </colors>
  </terminal>
</config>
```

Os nomes, comandos e layouts podem ser adaptados à versão instalada. A vantagem do VTM é que os itens da barra podem executar aplicações CLI comuns dentro de janelas textuais.

## 11.4 Layout operacional

```text
┌─ Arquivos ────────────────┐ ┌─ Neovim ──────────────────────┐
│ ~/code/tcc                │ │ package main                  │
│ ├── cmd                   │ │                               │
│ ├── internal              │ │ func main() {                 │
│ ├── web                   │ │     ...                       │
│ └── go.mod                │ │ }                             │
└───────────────────────────┘ └───────────────────────────────┘

┌─ lazygit ─────────────────┐ ┌─ htop/logs ───────────────────┐
│ status / diff / commits   │ │ CPU / RAM / processos        │
└───────────────────────────┘ └───────────────────────────────┘

┌─ Servidores ─────────────────────────────────────────────────┐
│ Vite :5173 | Go :8080 | Python :8000                         │
└───────────────────────────────────────────────────────────────┘
```

# 12. Neovim e LazyVim sem abandonar completamente atalhos convencionais

## 12.1 A dúvida: é possível “desfazer o Vim”?

É possível usar LazyVim e substituir vários atalhos por combinações comuns de editores gráficos:

```text
Ctrl+S  salvar
Ctrl+Z  desfazer
Ctrl+Y  refazer
Ctrl+A  selecionar tudo
Ctrl+C  copiar
Ctrl+V  colar
```

O LazyVim permite configurar mapeamentos globais em:

```text
~/.config/nvim/lua/config/keymaps.lua
```

Também permite remover mapeamentos existentes antes de substituí-los.

## 12.2 Configuração híbrida recomendada

```lua
local map = vim.keymap.set

-- Salvar.
map({ "n", "i", "v" }, "<C-s>", "<cmd>w<cr>", {
  desc = "Salvar arquivo",
})

-- Desfazer.
map("n", "<C-z>", "u", {
  desc = "Desfazer",
})

map("i", "<C-z>", "<C-o>u", {
  desc = "Desfazer",
})

-- Refazer.
map("n", "<C-y>", "<C-r>", {
  desc = "Refazer",
})

map("i", "<C-y>", "<C-o><C-r>", {
  desc = "Refazer",
})

-- Selecionar tudo.
map("n", "<C-a>", "ggVG", {
  desc = "Selecionar tudo",
})

-- Clipboard do sistema/terminal.
map("v", "<C-c>", '"+y', {
  desc = "Copiar seleção",
})

map({ "n", "v" }, "<C-v>", '"+p', {
  desc = "Colar",
})

map("i", "<C-v>", "<C-r>+", {
  desc = "Colar",
})
```

## 12.3 O que é perdido ao sobrescrever

| Atalho | Comportamento tradicional no Vim |
|---|---|
| `Ctrl+V` | seleção visual em bloco |
| `Ctrl+A` | incrementar número |
| `Ctrl+Z` | suspender o processo |
| `Ctrl+Y` | rolagem ou comando dependente do modo |

Esses comandos continuam acessíveis por outros mapeamentos, mas você deixa de seguir parte da ergonomia padrão do Vim.

## 12.4 Recomendação prática

Não tente transformar o Neovim completamente em VS Code. Use uma camada híbrida:

- atalhos convencionais para salvar, copiar, colar, desfazer e selecionar;
- `Esc` para modo normal;
- `i` para inserir;
- `dd`, `yy`, `p`, movimentos e text objects para edição rápida;
- comandos e menus do LazyVim para plugins;
- terminal integrado apenas quando fizer sentido.

Assim, a curva inicial fica menor sem remover o principal diferencial da edição modal.

## 12.5 Clipboard no Termux e por SSH

O registrador `+` depende de suporte a clipboard. Em sessões SSH, o clipboard do servidor não é automaticamente o clipboard do computador cliente.

Alternativas:

- copiar pelo próprio terminal cliente;
- usar OSC 52 quando o terminal e o Neovim estiverem configurados para isso;
- usar Termux:API localmente;
- manter os mapeamentos `y` e `p` para clipboard interno do Neovim;
- evitar depender de `Ctrl+C`/`Ctrl+V` em todas as situações.

## 12.6 LazyVim, Mason e LSP no Android

LazyVim em si é uma configuração do Neovim e funciona. O ponto delicado é o ecossistema de servidores de linguagem e ferramentas instaladas automaticamente.

O Mason pode tentar baixar binários Linux precompilados para glibc. Quando isso falhar:

1. instale o LSP pelo `pkg`, `npm`, `pip`, `cargo` ou `go install`;
2. configure o LazyVim para usar o executável já presente no `PATH`;
3. desative a instalação automática daquela ferramenta no Mason;
4. use PRoot somente para o LSP incompatível, se realmente necessário.

Exemplos de instalação externa:

```bash
# Go
go install golang.org/x/tools/gopls@latest

# TypeScript
npm install -g typescript typescript-language-server

# Python
pip install basedpyright ruff

# Rust
pkg search rust-analyzer
```

# 13. Nix, NixOS e Nix-on-Droid

## 13.1 Três conceitos diferentes

### Nix

Gerenciador de pacotes e ferramenta para builds e ambientes reproduzíveis.

### NixOS

Distribuição Linux cujo sistema inteiro é configurado com módulos Nix.

### Nix-on-Droid

Aplicativo e ambiente Nix adaptado ao Android. Ele não transforma o Android em NixOS, mas permite declarar pacotes e configurações do ambiente de terminal.

## 13.2 O que você quer configurar

O objetivo discutido não é substituir o Android. É declarar o setup de desenvolvimento:

- shell;
- Git;
- SSH do usuário;
- Neovim/LazyVim;
- plugins;
- keymaps;
- temas;
- VTM;
- launchers e cores;
- Go, Node, Python, Rust, Java e Kotlin;
- LSPs e formatadores;
- CLIs e TUIs;
- aliases;
- variáveis de ambiente;
- dotfiles;
- ambientes específicos por projeto.

Nix-on-Droid com Home Manager atende exatamente a essa categoria.

## 13.3 Importante: não é o mesmo Termux instalado anteriormente

O Nix-on-Droid utiliza um aplicativo baseado em um fork do emulador de terminal do Termux e adapta o acesso ao `/nix/store` com PRoot e outros mecanismos.

Na prática, você escolhe entre:

```text
Termux tradicional
ou
Nix-on-Droid como ambiente Nix principal
```

Não é necessário tentar encaixar Nix à força dentro do seu Termux já configurado.

## 13.4 Benefícios

- configuração declarativa;
- pacotes versionados;
- rollback;
- reconstrução após reinstalação;
- separação de ambientes por projeto;
- compartilhamento do setup em Git;
- menor dependência de scripts imperativos;
- dotfiles gerenciados pelo Home Manager;
- possibilidade de testar ferramentas sem poluir o ambiente global.

## 13.5 Limitações

- depende de PRoot e adaptações;
- não possui controle do kernel Android;
- ocupa mais armazenamento;
- builds locais podem ser pesados;
- nem todo pacote do nixpkgs funciona corretamente no Android/PRoot;
- integração com APIs do Android é menos direta que no Termux tradicional;
- o próprio projeto se apresenta como um ambiente ainda experimental/protótipo;
- o setup é mais complexo de compreender e diagnosticar.

## 13.6 Quando escolher Nix-on-Droid

Escolha quando estas propriedades forem importantes:

- “quero apagar o celular e reconstruir o ambiente”;
- “quero o mesmo shell e editor no notebook e no celular”;
- “quero declarar plugins e configurações junto com os pacotes”;
- “quero testar versões sem quebrar o ambiente principal”;
- “quero aprender Nix construindo algo útil”.

Continue no Termux tradicional quando quiser:

- máxima simplicidade;
- melhor integração com Android;
- menor consumo;
- comandos `pkg` diretos;
- menos camadas;
- um servidor SSH confiável rapidamente.

# 14. Estrutura de um repositório Nix-on-Droid

```text
mobile-dev-env/
├── nix-on-droid.nix
├── home.nix
├── packages.nix
├── modules/
│   ├── shell.nix
│   ├── git.nix
│   ├── neovim.nix
│   └── vtm.nix
├── configs/
│   ├── nvim/
│   │   └── lua/config/keymaps.lua
│   ├── vtm/
│   │   └── settings.xml
│   └── starship.toml
└── scripts/
    ├── doctor.sh
    ├── backup.sh
    └── project-new.sh
```

## 14.1 Configuração principal

`nix-on-droid.nix`:

```nix
{ pkgs, ... }:

{
  environment.packages = with pkgs; [
    git
    openssh
    tmux
    neovim

    go
    nodejs
    python3
    rustc
    cargo
    jdk
    kotlin

    ripgrep
    fd
    fzf
    jq
    tree
    rsync
    curl
    wget
    unzip

    lazygit
    yazi
    zellij
    broot
    ranger
    lf
    vtm
  ];

  home-manager.config = ./home.nix;

  # Use o valor criado pela sua instalação e leia o changelog
  # antes de alterá-lo.
  system.stateVersion = "24.05";
}
```

A lista é uma intenção de setup. Um pacote pode não estar disponível ou não funcionar no ambiente atual. A configuração deve ser testada incrementalmente.

## 14.2 Home Manager

`home.nix`:

```nix
{ config, pkgs, ... }:

{
  home.stateVersion = "24.05";

  programs.git = {
    enable = true;
    userName = "Érick Lúcio";
    userEmail = "seu-email-vinculado-ao-github@example.com";
    extraConfig = {
      init.defaultBranch = "main";
      pull.rebase = false;
    };
  };

  programs.bash = {
    enable = true;
    shellAliases = {
      ll = "ls -lah";
      gs = "git status";
      ga = "git add";
      gc = "git commit";
      dev = "tmux attach -t dev || tmux new -s dev";
    };
  };

  programs.neovim = {
    enable = true;
    defaultEditor = true;
    viAlias = true;
    vimAlias = true;
  };

  xdg.configFile."nvim/lua/config/keymaps.lua".source =
    ./configs/nvim/lua/config/keymaps.lua;

  xdg.configFile."vtm/settings.xml".source =
    ./configs/vtm/settings.xml;
}
```

Não coloque um endereço fictício no Git em uso real. Use exatamente um e-mail vinculado à sua conta para que commits possam ser associados corretamente ao perfil.

## 14.3 Aplicar e voltar

```bash
nix-on-droid switch
```

Rollback:

```bash
nix-on-droid rollback
```

Com flakes, a ativação pode apontar para uma configuração específica do repositório. Como a interface e os exemplos evoluem, use a documentação da versão instalada ao fixar os inputs.

## 14.4 Declarar plugins e temas

Existem duas formas principais.

### Módulo específico

Quando Home Manager possui módulo para a ferramenta, use as opções declarativas do módulo.

### Arquivos de configuração

Quando não existe módulo específico, declare o arquivo inteiro:

```nix
xdg.configFile."vtm/settings.xml".source = ./configs/vtm/settings.xml;
```

ou:

```nix
xdg.configFile."algum-app/config.toml".text = ''
  theme = "nord"
  mouse = true
'';
```

A mesma estratégia funciona para:

- Lua do LazyVim;
- XML do VTM;
- TOML do Yazi e Starship;
- configurações do Zellij;
- aliases do shell;
- SSH config;
- scripts em `~/.local/bin`.

## 14.5 O papel do Nix e do Home Manager

```text
Nix
└── instala e versiona programas

Home Manager
└── cria e mantém configurações do usuário

Aplicativo
└── interpreta a própria configuração
```

O Nix não “entende” semanticamente todas as cores do VTM. Ele garante que o arquivo `settings.xml` correto exista no lugar correto e possa ser reconstruído.

# 15. Ambientes por projeto com Nix

O setup global resolve editor, shell e utilitários. Cada projeto pode declarar sua própria toolchain.

Exemplo conceitual para Go, Node e PostgreSQL client:

```nix
{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  packages = with pkgs; [
    go
    nodejs
    postgresql
    gnumake
  ];

  shellHook = ''
    export APP_ENV=development
    echo "Ambiente do projeto carregado"
  '';
}
```

Entrada:

```bash
nix-shell
```

Ou, usando flakes:

```bash
nix develop
```

Benefícios:

- o projeto declara a versão das ferramentas;
- o ambiente global não precisa conter tudo;
- colaboradores podem reproduzir a toolchain;
- um projeto Java não interfere no projeto Go;
- atualizações podem ser testadas por branch.

# 16. Setup reproduzível sem Nix

Nix não é obrigatório para obter boa reprodutibilidade. Uma alternativa mais simples é:

```text
dotfiles/
├── bootstrap.sh
├── packages.txt
├── config/
│   ├── nvim/
│   ├── vtm/
│   └── tmux/
└── scripts/
```

## 16.1 Exemplo de bootstrap

```bash
#!/data/data/com.termux/files/usr/bin/bash
set -Eeuo pipefail

pkg update
pkg upgrade -y

pkg install -y \
  git openssh tmux neovim \
  nodejs golang python rust clang make \
  ripgrep fd fzf jq tree rsync curl wget \
  vtm lf lazygit

mkdir -p ~/.config ~/.local/bin ~/code

cp -r ./config/nvim ~/.config/
cp -r ./config/vtm ~/.config/
cp -r ./config/tmux ~/.config/

printf '%s\n' "Setup aplicado. Revise SSH e dados pessoais do Git."
```

## 16.2 Comparação

| Critério | Bootstrap + dotfiles | Nix-on-Droid |
|---|---|---|
| Curva de aprendizado | Menor | Maior |
| Integração Termux | Melhor | Ambiente próprio |
| Rollback | Manual | Nativo |
| Versões reproduzíveis | Parcial | Forte |
| Configuração declarativa | Parcial | Forte |
| Debug | Shell tradicional | Nix + PRoot |
| Armazenamento | Menor | Maior |
| Melhor para começar | Sim | Depois do núcleo estável |

# 17. Automatização e aplicativo Android próprio

## 17.1 O objetivo

Você cogitou criar uma experiência em que o usuário instala algo e recebe:

- servidor SSH configurado;
- linguagens instaladas;
- ferramentas TUI;
- botões em vez de apenas comandos;
- scripts para iniciar e parar serviços;
- ambiente padronizado.

Isso pode ser implementado em níveis.

## 17.2 Nível 1: repositório de instalação

```bash
git clone https://github.com/usuario/mobile-dev-env
cd mobile-dev-env
./install.sh
```

É o caminho mais barato para validar o produto.

## 17.3 Nível 2: CLI própria

Exemplo:

```text
poco-dev install
poco-dev start
poco-dev stop
poco-dev status
poco-dev doctor
poco-dev backup
poco-dev update
```

Pode ser escrita em Go e distribuída como binário Termux/Android ARM64.

## 17.4 Nível 3: TUI de configuração

Uma TUI com Bubble Tea pode oferecer:

```text
[✓] OpenSSH
[✓] tmux
[✓] Neovim
[ ] LazyVim
[✓] Go
[✓] Node.js
[ ] Python Data
[✓] VTM

[ Instalar ] [ Diagnóstico ] [ Backup ]
```

Essa abordagem mantém tudo no terminal, mas elimina a necessidade de memorizar comandos.

## 17.5 Nível 4: aplicativo Android complementar

Um aplicativo Android pode apresentar botões como:

- iniciar servidor;
- abrir terminal;
- mostrar IP e porta;
- iniciar projeto;
- abrir URL local;
- executar backup;
- mostrar logs.

O aplicativo pode acionar comandos no Termux através das integrações permitidas, desde que permissões, assinatura e políticas sejam tratadas corretamente.

## 17.6 Nível 5: fork do Termux

Um fork completo é tecnicamente possível, mas é muito mais caro que parece.

Problemas:

- os pacotes do Termux são compilados para caminhos ligados ao nome do pacote Android;
- mudar o `applicationId` pode exigir reconstruir bootstrap e pacotes;
- plugins precisam ser compatíveis com a assinatura;
- é necessário manter patches, atualizações e segurança;
- não basta trocar ícone e adicionar botões;
- publicação em loja exige revisão separada de licença e políticas vigentes.

## 17.7 Recomendação de produto

Construa nesta ordem:

```text
script instalador
    ↓
CLI em Go
    ↓
TUI de gerenciamento
    ↓
app Android complementar
    ↓
fork completo somente se o produto exigir
```

Assim, cada etapa produz algo utilizável e reduz o risco de investir cedo em manutenção de um fork.

# 18. Ícones, atalhos e integração com a tela inicial

Você não precisa criar um fork apenas para ter ações visuais no Android.

Possibilidades:

- Termux:Widget com scripts na tela inicial;
- atalhos do launcher;
- aplicativo complementar;
- notificações com ações;
- TUI aberta por um script específico;
- URL local que abre uma interface web de gerenciamento.

Estrutura:

```text
~/.shortcuts/
├── iniciar-servidor
├── parar-servidor
├── abrir-vtm
├── abrir-projeto-tcc
└── status
```

Exemplo:

```bash
#!/data/data/com.termux/files/usr/bin/bash

termux-wake-lock
sshd
exec tmux attach -t dev || tmux new -s dev
```

Com um widget ou atalho, o usuário toca em um ícone e o script executa.

# 19. Interface gráfica, X11, VNC e desktop Linux

## 19.1 Termux:X11

Termux:X11 é útil quando você precisa executar uma aplicação gráfica específica. Ele evita parte da sobrecarga de transmitir um desktop inteiro por VNC, mas ainda adiciona uma camada gráfica.

Use para:

- testar uma ferramenta GUI;
- abrir aplicação X11 isolada;
- experimentar um window manager leve;
- acessar recursos gráficos que uma TUI não oferece.

Não use como núcleo inicial da workstation.

## 19.2 Desktop Ubuntu via PRoot

É possível instalar Ubuntu/Debian e abrir XFCE, LXQt ou outro desktop por X11/VNC. Porém, isso soma:

```text
Android
+ Termux
+ PRoot
+ distribuição
+ desktop
+ servidor gráfico
+ cliente gráfico
```

O resultado tende a consumir mais RAM, bateria e armazenamento, além de apresentar latência e problemas de suspensão.

## 19.3 Por que a abordagem TUI é melhor para este projeto

A TUI transmite caracteres e eventos do terminal, não quadros de vídeo completos. Isso oferece:

- menor latência;
- uso eficiente por SSH;
- reconexão simples;
- funcionamento em redes ruins;
- menor consumo;
- integração natural com compiladores e servidores;
- facilidade de automatizar.

# 20. Docker, containers e máquinas virtuais

## 20.1 Docker no Termux nativo

Docker não é apenas um executável. Ele depende de recursos do kernel como:

- namespaces;
- cgroups;
- mounts;
- redes virtuais;
- capabilities;
- seccomp;
- daemon privilegiado.

O Termux sem root não controla esses recursos como uma distribuição Linux tradicional.

## 20.2 PRoot não é Docker

PRoot cria um ambiente semelhante a `chroot` em user space, mas não fornece isolamento real equivalente a containers.

Portanto:

```text
Ubuntu no PRoot ≠ máquina virtual
Ubuntu no PRoot ≠ Docker host completo
```

Ele é adequado para compatibilidade de userland, não para reproduzir toda a infraestrutura de container.

## 20.3 Quando usar VM

Uma VM com kernel próprio pode executar Docker real. O custo é:

- mais memória;
- emulação ou virtualização;
- desempenho inferior;
- maior complexidade;
- consumo de bateria;
- dependência do suporte do firmware.

Use como laboratório para Compose ou testes específicos, não como ambiente principal de edição e desenvolvimento.

## 20.4 AVF e terminal Linux do Android

Algumas versões e fabricantes podem disponibilizar recursos baseados no Android Virtualization Framework. A presença de uma feature no aparelho não garante que o terminal Linux oficial ou VMs não protegidas estejam liberados no firmware.

Comandos de investigação por ADB:

```bash
adb shell pm list features | grep -i virtualization
adb shell pm list packages | grep -Ei "virtualization|virtualmachine|terminal"
adb shell getprop ro.build.version.release
adb shell getprop ro.build.version.incremental
```

Não baseie a arquitetura principal em um recurso que ainda não foi confirmado no POCO F6 específico.

# 21. Segurança

## 21.1 SSH

- use chave Ed25519;
- desative senha após validar a chave;
- não exponha `8022` diretamente na internet;
- use Tailscale ou túnel privado;
- mantenha o OpenSSH atualizado;
- revise `authorized_keys`;
- remova chaves antigas;
- use uma configuração separada para o aparelho.

## 21.2 Servidores HTTP

Prefira:

```text
127.0.0.1
```

em vez de:

```text
0.0.0.0
```

Encaminhe portas pelo SSH. Use `0.0.0.0` apenas quando quiser conscientemente expor o serviço na LAN.

## 21.3 Scripts remotos

Evite executar automaticamente:

```bash
curl URL | bash
```

Baixe, leia e versiona o instalador antes de executar.

## 21.4 Segredos

- não coloque tokens em repositórios;
- use arquivos fora do Git;
- proteja chaves SSH;
- aplique `chmod 600` em arquivos sensíveis;
- não misture configuração pública com credenciais privadas;
- mantenha `.env` no `.gitignore`.

## 21.5 Backup

O celular não deve ser o único local dos projetos.

Use:

- Git remoto;
- `rsync` para o computador;
- arquivos compactados em armazenamento externo;
- backup das configurações;
- exportação periódica das chaves, com proteção adequada;
- repositório separado para dotfiles/Nix.

Exemplo:

```bash
rsync -av --delete ~/code/ computador:~/backup-poco/code/
```

# 22. Estabilidade e desempenho

## 22.1 Processos em segundo plano

O Android pode encerrar processos por:

- economia de bateria;
- falta de memória;
- políticas do fabricante;
- limite de processos filhos;
- aplicativo removido dos recentes;
- atualização ou reinicialização.

Mitigações:

- wake lock;
- tmux;
- Termux:Boot;
- bateria sem restrições;
- autostart;
- logs;
- serviços supervisionados;
- scripts idempotentes de recuperação.

## 22.2 Temperatura

Builds Rust, Java/Gradle, compilação C/C++ e vários servidores simultâneos podem aquecer o aparelho.

Boas práticas:

- evitar carregar e compilar pesado em local quente;
- limitar processos paralelos;
- desligar daemons que não estão sendo usados;
- reduzir LSPs simultâneos;
- monitorar com `htop`, `free -h` e `uptime`;
- preferir builds incrementais;
- não executar desktop gráfico sem necessidade.

## 22.3 Armazenamento

Nix, PRoot, Gradle, npm, Cargo e caches de compilação crescem rapidamente.

Monitore:

```bash
df -h
du -sh ~/.cache/* 2>/dev/null | sort -h
du -sh ~/code/* 2>/dev/null | sort -h
```

Limpezas possíveis:

```bash
npm cache verify
pip cache purge
cargo cache --help
```

Não execute comandos de limpeza agressiva sem revisar os caminhos.

# 23. Troubleshooting

| Sintoma | Causa provável | Ação |
|---|---|---|
| SSH cai, mas tmux continua | rede ou suspensão temporária | reconectar e usar `tmux attach -t dev` |
| SSH e processos desaparecem | HyperOS encerrou o Termux | bateria sem restrições, autostart, wake lock e logs |
| `Connection refused` na porta 8022 | `sshd` não está rodando ou IP mudou | executar `sshd`, verificar `ss -tln` e IP |
| chave SSH recusada | usuário, permissão ou `authorized_keys` incorreto | verificar `whoami`, `~/.ssh` e logs |
| servidor abre no celular, mas não no PC | bind ou túnel incorreto | usar `127.0.0.1` + `ssh -L` ou `0.0.0.0` na LAN |
| Git falha em `/sdcard` | filesystem compartilhado/noexec | mover repositório para `$HOME` |
| npm falha ao instalar binário | pacote glibc sem build Android | compilar, trocar dependência ou usar PRoot |
| pip não encontra wheel compatível | wheel manylinux | compilar fonte, usar pacote Termux ou PRoot |
| crate Rust falha no linker | biblioteca de sistema ausente | instalar dependência, revisar `pkg-config` ou usar PRoot |
| Java/Gradle trava | RAM, armazenamento ou daemon | `--no-daemon`, reduzir paralelismo e limpar caches |
| LazyVim abre, mas LSP não instala | Mason baixou binário incompatível | instalar LSP por outro gerenciador e usar o `PATH` |
| clipboard do Neovim não funciona por SSH | clipboard remoto não é o local | usar terminal, OSC 52 ou clipboard interno |
| caracteres e ícones quebrados | fonte/locale | UTF-8 e fonte Nerd Font no terminal cliente |
| VTM fica estranho | tamanho, fonte, mouse ou config | testar config mínima e terminal maior |
| TUI não responde ao toque | app não recebe protocolo de mouse | usar teclado/mouse ou acessar por SSH |
| PRoot está muito lento | interceptação de syscalls | mover a ferramenta compatível de volta ao Termux |
| `systemctl` falha no PRoot | não há systemd real | iniciar processo diretamente ou usar supervisor simples |
| Nix-on-Droid ocupa muito espaço | store, gerações e caches | revisar gerações e coleta de lixo com cuidado |

## 23.1 Checklist de diagnóstico

```bash
# Arquitetura e sistema
uname -a
uname -m

# Rede e SSH
ip addr
ss -tln
pgrep -a sshd

# Sessões
tmux ls

# Recursos
free -h
uptime
ps -eo pid,ppid,%cpu,%mem,cmd --sort=-%cpu | head -20

# Portas de desenvolvimento
ss -tln | grep -E '5173|8080|8000'

# Disco
pwd
df -h
du -sh ~/.cache 2>/dev/null

# Toolchains
go version
node --version
python --version
rustc --version
java --version
```


# 24. Roteiro de implementação

## Fase 1 — Núcleo

Objetivo:

- instalar Termux;
- criar estrutura no `$HOME`;
- instalar Git, SSH, tmux e Neovim;
- acessar por SSH.

Critério:

```text
Entrar no POCO pelo computador e editar um arquivo dentro de ~/code.
```

## Fase 2 — Persistência

Objetivo:

- sessão tmux;
- wake lock;
- Termux:Boot ou services;
- ajustes do HyperOS.

Critério:

```text
Desconectar o SSH, reconectar e recuperar editor/processos.
```

## Fase 3 — Linguagens

Objetivo:

- Go;
- Node;
- Python;
- Rust;
- Java;
- Kotlin.

Critério:

```text
Executar um Hello World e um servidor HTTP por tecnologia relevante.
```

## Fase 4 — TUI

Objetivo:

- lf ou Yazi;
- lazygit;
- htop;
- VTM;
- LazyVim.

Critério:

```text
Editar, navegar, usar Git e acompanhar servidor dentro do terminal.
```

## Fase 5 — Acesso externo

Objetivo:

- Tailscale;
- SSH por chave;
- túneis HTTP.

Critério:

```text
Acessar o POCO fora da LAN sem abrir portas no roteador.
```

## Fase 6 — Compatibilidade

Objetivo:

- instalar PRoot-Distro;
- criar Ubuntu ou Debian;
- executar uma ferramenta glibc incompatível com o Termux.

Critério:

```text
Usar PRoot apenas para o caso que justificou sua instalação.
```

## Fase 7 — Reprodutibilidade

Escolha uma trilha:

### Trilha simples

- bootstrap;
- dotfiles;
- backups;
- CLI de gerenciamento.

### Trilha Nix

- Nix-on-Droid;
- Home Manager;
- VTM e LazyVim declarados;
- ambientes por projeto;
- rollback.

Critério:

```text
Reconstruir o setup a partir do repositório.
```

## Fase 8 — Produto próprio

- CLI em Go;
- TUI de configuração;
- widgets/atalhos;
- aplicativo Android complementar.

Critério:

```text
Outra pessoa consegue instalar e operar o ambiente sem conhecer todos os comandos.
```

# 25. Decisão final por perfil de uso

## Quero começar a trabalhar hoje

```text
Termux + SSH + tmux + Neovim + Git
```

## Quero um desktop textual

```text
Termux + tmux + VTM + lf/Yazi + lazygit + htop
```

## Quero compatibilidade com ferramentas Ubuntu

```text
Termux principal + Debian/Ubuntu PRoot sob demanda
```

## Quero declarar e reconstruir tudo

```text
Nix-on-Droid + Home Manager + repositório de configuração
```

## Quero distribuir para outros usuários

```text
bootstrap → CLI → TUI → app Android complementar
```

## Quero Docker real

```text
VM/root/dispositivo com suporte adequado; não confundir PRoot com Docker
```

# 26. Arquitetura final recomendada para Érick

```text
POCO F6 / HyperOS
│
├── Tailscale Android
│   └── acesso privado externo
│
├── Termux nativo
│   ├── sshd :8022
│   ├── tmux session "dev"
│   ├── Neovim/LazyVim híbrido
│   ├── VTM
│   │   ├── lf ou Yazi
│   │   ├── Neovim
│   │   ├── lazygit
│   │   ├── htop
│   │   └── servidores/logs
│   ├── Go
│   ├── Node.js/TypeScript
│   ├── Python
│   ├── Rust
│   ├── Java/Kotlin
│   └── ~/code no filesystem privado
│
├── Debian PRoot
│   └── somente para glibc/incompatibilidades
│
└── Repositório mobile-dev-env
    ├── bootstrap.sh
    ├── dotfiles
    ├── keymaps LazyVim
    ├── settings VTM
    ├── scripts de diagnóstico
    └── futura migração para Nix-on-Droid
```

A evolução mais racional é começar com a arquitetura acima em Termux e manter a configuração versionada. Depois, recriar a mesma experiência em Nix-on-Droid se o aprendizado e a reprodutibilidade justificarem a troca.

# Apêndice A — Bootstrap inicial do Termux

```bash
#!/data/data/com.termux/files/usr/bin/bash
set -Eeuo pipefail

log() {
  printf '[setup] %s\n' "$*"
}

log "Atualizando pacotes"
pkg update
pkg upgrade -y

log "Instalando núcleo"
pkg install -y \
  git openssh tmux neovim \
  nodejs golang python rust clang make cmake pkg-config \
  ripgrep fd fzf jq tree rsync curl wget zip unzip \
  procps htop vtm lf lazygit

log "Criando diretórios"
mkdir -p \
  ~/code/{go,node,python,rust,java,kotlin} \
  ~/workspace \
  ~/.config \
  ~/.local/bin \
  ~/backups

log "Validando comandos"
for command in git ssh tmux nvim node npm go python rustc cargo; do
  if command -v "$command" >/dev/null 2>&1; then
    printf '  OK  %s\n' "$command"
  else
    printf '  ERRO %s não encontrado\n' "$command" >&2
  fi
done

log "Concluído. Configure senha, chave SSH e HyperOS manualmente."
```

# Apêndice B — Script de diagnóstico

```bash
#!/data/data/com.termux/files/usr/bin/bash
set -u

section() {
  printf '\n=== %s ===\n' "$1"
}

section "Sistema"
uname -a
printf 'HOME=%s\n' "$HOME"
printf 'PREFIX=%s\n' "$PREFIX"

section "Armazenamento"
df -h "$HOME"

section "Rede"
ip addr show wlan0 2>/dev/null || true
ss -tln 2>/dev/null || true

section "SSH"
pgrep -a sshd || printf 'sshd não está rodando\n'

section "tmux"
tmux ls 2>/dev/null || printf 'nenhuma sessão tmux\n'

section "Toolchains"
for command in git nvim node npm go python rustc cargo java; do
  if command -v "$command" >/dev/null 2>&1; then
    printf '\n[%s]\n' "$command"
    "$command" --version 2>&1 | head -3
  else
    printf '\n[%s] ausente\n' "$command"
  fi
done

section "Processos pesados"
ps -eo pid,ppid,%cpu,%mem,cmd --sort=-%cpu 2>/dev/null | head -20
```

# Apêndice C — Configuração SSH do computador

```sshconfig
Host poco
    HostName 192.168.1.80
    User u0_a123
    Port 8022
    IdentityFile ~/.ssh/id_ed25519
    ServerAliveInterval 30
    ServerAliveCountMax 3

Host poco-dev
    HostName 192.168.1.80
    User u0_a123
    Port 8022
    IdentityFile ~/.ssh/id_ed25519
    LocalForward 5173 127.0.0.1:5173
    LocalForward 8080 127.0.0.1:8080
    LocalForward 8000 127.0.0.1:8000
```

Uso:

```bash
ssh poco-dev
```

# Apêndice D — Keymaps LazyVim

Arquivo:

```text
~/.config/nvim/lua/config/keymaps.lua
```

```lua
local map = vim.keymap.set

map({ "n", "i", "v" }, "<C-s>", "<cmd>w<cr>", {
  desc = "Salvar",
})

map("n", "<C-z>", "u", { desc = "Desfazer" })
map("i", "<C-z>", "<C-o>u", { desc = "Desfazer" })

map("n", "<C-y>", "<C-r>", { desc = "Refazer" })
map("i", "<C-y>", "<C-o><C-r>", { desc = "Refazer" })

map("n", "<C-a>", "ggVG", { desc = "Selecionar tudo" })
map("v", "<C-c>", '"+y', { desc = "Copiar" })
map({ "n", "v" }, "<C-v>", '"+p', { desc = "Colar" })
map("i", "<C-v>", "<C-r>+", { desc = "Colar" })
```

# Apêndice E — Configuração VTM

Arquivo:

```text
~/.config/vtm/settings.xml
```

```xml
<config>
  <colors>
    <window fgc=#D8DEE9 bgc=#2E3440/>
    <focus  fgc=#ECEFF4 bgc=#5E81AC/>
  </colors>

  <desktop>
    <background>
      <color fgc=#D8DEE9 bgc=#1B1F2AFF/>
    </background>

    <taskbar selected="Shell" item*>
      <item id="Shell" label="shell" type="term" cmd="$SHELL"/>
      <item id="Editor" label="nvim" type="term" cmd="nvim"/>
      <item id="Files" label="files" type="term" cmd="lf"/>
      <item id="Git" label="git" type="term" cmd="lazygit"/>
      <item id="Processes" label="proc" type="term" cmd="htop"/>

      <colors>
        <bground  fgc=#D8DEE9 bgc=#2E3440FF/>
        <focused  fgc=#88C0D0/>
        <selected fgc=#ECEFF4/>
        <active   fgc=#A3BE8C/>
        <inactive fgc=#4C566A/>
      </colors>
    </taskbar>
  </desktop>

  <terminal>
    <scrollback>
      <size=50000/>
      <wrap=true/>
    </scrollback>
    <colors>
      <default fgc=#D8DEE9 bgc=#1B1F2A/>
    </colors>
  </terminal>
</config>
```

# Apêndice F — Comandos de bolso

## Termux

```bash
pkg update
pkg upgrade
pkg search NOME
pkg install NOME
```

## SSH

```bash
sshd
pkill sshd
ss -tln | grep 8022
ssh -p 8022 usuario@ip
```

## tmux

```bash
tmux new -s dev
tmux attach -t dev
tmux ls
tmux kill-session -t dev
```

## Port forwarding

```bash
ssh poco -L 8080:127.0.0.1:8080
```

## PRoot-Distro

```bash
pkg install proot-distro
proot-distro list
proot-distro install ubuntu
proot-distro login ubuntu
proot-distro remove ubuntu
```

## Nix-on-Droid

```bash
nix-on-droid switch
nix-on-droid rollback
nix-shell
nix develop
```

## Recursos

```bash
free -h
htop
df -h
du -sh ~/.cache/* 2>/dev/null | sort -h
```

# Apêndice G — Referências oficiais consultadas

- Termux — página oficial e documentação do ambiente de execução.
- Termux Packages Wiki — diferenças de Bionic, paths, execução nativa e armazenamento externo.
- PRoot-Distro — gerenciamento de distribuições e limitações de isolamento.
- Nix-on-Droid — módulo de configuração, Home Manager, `switch` e `rollback`.
- LazyVim — configuração e sobrescrita de keymaps.
- VTM — `settings.xml`, cores, taskbar, tipos de janela e comandos.
- OpenSSH — autenticação por chave e configuração do servidor.
- tmux — sessões persistentes e reconexão.

Links principais:

- https://termux.dev/
- https://github.com/termux/termux-packages/wiki/Termux-execution-environment
- https://github.com/termux/proot-distro
- https://github.com/nix-community/nix-on-droid
- https://www.lazyvim.org/configuration/keymaps
- https://github.com/directvt/vtm/blob/master/doc/settings.md
- https://github.com/tmux/tmux
- https://tailscale.com/

# Encerramento

O ponto mais importante de toda a discussão é que o POCO F6 já pode ser uma máquina de desenvolvimento útil sem root, sem Docker e sem desktop gráfico completo. A base técnica é relativamente simples: **Termux nativo, SSH, tmux e ferramentas de terminal**.

PRoot resolve incompatibilidades específicas. Nix-on-Droid resolve outra categoria de problema: não compatibilidade, mas **reprodutibilidade e configuração declarativa**. VTM e os demais projetos TUI resolvem a camada visual. LazyVim pode ser adaptado para atalhos mais familiares sem abandonar inteiramente a edição modal.

A arquitetura correta não é escolher a ferramenta mais sofisticada desde o começo. É criar um núcleo confiável e adicionar camadas apenas quando cada uma resolver um problema concreto.
