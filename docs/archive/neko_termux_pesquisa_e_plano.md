# Neko no Termux: pesquisa, viabilidade e plano de automação

> Estudo consolidado sobre executar o Neko no POCO F6 sem Docker, usando Termux, PRoot-Distro e a imagem oficial ARM64 do projeto.

## 1. Conclusão principal

Dá para facilitar **muito** a instalação e o uso do Neko no Android.

O caminho mais sensato não é começar compilando o Neko diretamente no POCO F6 nem tentar portá-lo imediatamente para a libc Bionic do Termux.

A estratégia mais promissora é:

```text
Imagem oficial Neko Firefox ARM64
        ↓
PRoot-Distro instala a imagem OCI diretamente
        ↓
Executa o CMD/Entrypoint sem Docker daemon
        ↓
Um script neko-termux controla tudo
```

A versão atual do PRoot-Distro consegue:

- baixar imagens Docker/OCI diretamente de registries;
- selecionar a arquitetura ARM64;
- aplicar as camadas da imagem;
- registrar `Entrypoint`, `Cmd`, `Env` e `WorkingDir`;
- executar a aplicação sem Docker daemon;
- iniciar sessões em segundo plano;
- listar e encerrar a árvore completa de processos;
- fazer backup, restore e reset do ambiente.

Isso elimina:

- Docker daemon;
- containerd;
- cgroups;
- criação manual de rootfs;
- extração manual das camadas OCI no computador;
- grande parte do processo de compilação.

Ainda é necessário validar especificamente se a imagem oficial do Neko funciona sem alterações no PRoot do POCO F6.

O PRoot não implementa completamente:

- namespaces reais;
- cgroups;
- seccomp;
- isolamento de rede;
- mounts reais;
- recursos de kernel equivalentes ao Docker.

Portanto, alguns pressupostos do runtime do Neko podem precisar de ajustes.

Mesmo assim, essa abordagem reduz bastante a complexidade.

---

## 2. O que o Neko já oferece pronto

Você não precisa construir a maior parte do sistema do zero.

### 2.1 Imagens oficiais ARM64

O Neko publica imagens no GitHub Container Registry.

Exemplos:

```text
ghcr.io/m1k1o/neko/firefox:3
ghcr.io/m1k1o/neko/firefox:3.1
ghcr.io/m1k1o/neko/firefox:3.1.0
ghcr.io/m1k1o/neko/firefox:latest
```

As imagens oficiais possuem builds para `linux/arm64`.

Existem variantes com:

- Firefox;
- Chromium;
- Brave;
- Vivaldi;
- XFCE;
- KDE;
- Remmina;
- VLC.

As variantes ARM64 de navegadores não oferecem DRM completo.

Para estabilidade, é melhor fixar uma versão:

```text
3.1.0
```

em vez de usar:

```text
latest
```

Assim, uma atualização publicada pelo projeto não altera seu ambiente inesperadamente.

### 2.2 Firefox ARM64 preparado

A imagem oficial do Firefox já:

- identifica a arquitetura ARM64;
- baixa o Firefox AArch64 oficial da Mozilla;
- instala GTK e bibliotecas de DBus;
- instala Openbox;
- cria o perfil do usuário;
- adiciona políticas;
- configura extensões;
- configura o supervisor de processos;
- prepara diretórios e permissões.

O supervisor inicia:

```text
Firefox
Openbox
```

Também reinicia o Firefox automaticamente se o processo cair.

### 2.3 Runtime completo

A imagem final do Neko contém mais que o servidor Go.

Ela combina:

```text
servidor Neko
frontend web
plugins
runtime Linux
drivers Xorg
configuração
aplicativo escolhido
supervisor
GStreamer
PulseAudio
```

A imagem oficial coloca, entre outros componentes:

```text
/usr/bin/neko
/var/www
/etc/neko/plugins
/etc/neko/neko.yaml
drivers do Xorg
configurações do navegador
```

Portanto, a própria imagem já é muito próxima da distribuição portátil que seria necessário construir manualmente.

### 2.4 Execução sem Docker

O Neko não depende conceitualmente do Docker.

É possível executá-lo em um sistema Debian ou Ubuntu configurado com:

- Xorg;
- PulseAudio;
- GStreamer;
- navegador;
- servidor Neko;
- frontend;
- supervisor.

A documentação permite copiar o binário compilado da imagem oficial em vez de obrigatoriamente recompilar o servidor.

O Docker é principalmente o mecanismo oficial de distribuição, isolamento e orquestração.

### 2.5 Configuração modular

O Neko aceita configuração por:

- YAML;
- JSON;
- TOML;
- variáveis de ambiente;
- argumentos da linha de comando.

A precedência é:

```text
valores padrão
    ↓
arquivo de configuração
    ↓
variáveis de ambiente
    ↓
argumentos CLI
```

Isso facilita criar um wrapper.

O wrapper pode manter quase tudo em um arquivo:

```text
neko.yaml
```

e alterar apenas:

- endereço;
- porta;
- resolução;
- FPS;
- WebRTC;
- autenticação;
- perfil;
- logs.

### 2.6 Autenticação

O Neko oferece:

- autenticação por senha;
- usuários armazenados em arquivo JSON;
- hashes SHA-256 em Base64;
- sessões persistentes;
- cookies HttpOnly;
- usuários específicos para API;
- permissões por usuário;
- modo sem autenticação para testes.

### 2.7 Perfil persistente do navegador

O Firefox utiliza normalmente:

```text
/home/neko/.mozilla/firefox/profile.default
```

Essa pasta pode ser vinculada a um diretório persistente do Termux.

Isso permite preservar:

- logins;
- cookies;
- histórico;
- favoritos;
- extensões;
- preferências;
- abas;
- sessão anterior.

As políticas padrão do Neko podem limpar dados ao encerrar.

Para um navegador pessoal, é necessário alterar as políticas para impedir a limpeza.

Exemplo:

```json
{
  "policies": {
    "SanitizeOnShutdown": false,
    "Homepage": {
      "StartPage": "previous-session"
    },
    "ExtensionSettings": {
      "*": {
        "installation_mode": "allowed"
      }
    }
  }
}
```

### 2.8 WebRTC simplificado para SSH

O Neko pode multiplexar o tráfego WebRTC em uma única porta TCP.

Configuração:

```text
NEKO_WEBRTC_TCPMUX=52000
NEKO_WEBRTC_ICELITE=1
NEKO_WEBRTC_NAT1TO1=127.0.0.1
```

Assim, basta encaminhar:

```text
8080 → interface HTTP
52000 → WebRTC
```

Exemplo:

```bash
ssh \
  -p 8022 \
  -L 8080:127.0.0.1:8080 \
  -L 52000:127.0.0.1:52000 \
  usuario@ip-do-poco
```

Depois:

```text
http://localhost:8080
```

---

## 3. Arquitetura proposta

```text
Termux
│
├── neko-termux
│   ├── install
│   ├── start
│   ├── stop
│   ├── restart
│   ├── status
│   ├── logs
│   ├── doctor
│   ├── backup
│   ├── update
│   └── uninstall
│
├── PRoot-Distro
│   └── imagem oficial Neko Firefox ARM64
│
└── ~/.local/share/neko-termux/
    ├── config/
    │   ├── neko.yaml
    │   ├── members.json
    │   └── policies.json
    ├── profile/
    │   └── perfil real do Firefox
    ├── state/
    │   └── sessões e usuários
    ├── logs/
    └── backups/
```

O script `neko-termux` seria um:

- launcher;
- configurador;
- supervisor;
- gerenciador;
- sistema de diagnóstico.

Ele não precisaria ser um fork completo do Neko.

---

## 4. Primeira prova de conceito

Esta etapa deve evitar qualquer compilação.

### 4.1 Instalar dependências

```bash
pkg update

pkg install -y \
  proot-distro \
  curl \
  jq \
  openssl
```

Verifique se a versão instalada possui os comandos modernos:

```bash
proot-distro install --help
proot-distro run --help
proot-distro ps --help
proot-distro kill --help
```

### 4.2 Instalar a imagem oficial

Exemplo com versão fixada:

```bash
proot-distro install \
  ghcr.io/m1k1o/neko/firefox:3.1.0 \
  --name neko-firefox
```

O PRoot-Distro deve:

1. detectar a arquitetura ARM64;
2. selecionar o manifesto `linux/arm64`;
3. baixar as camadas;
4. montar a rootfs;
5. preservar metadados da imagem;
6. registrar o ambiente localmente.

### 4.3 Criar diretórios persistentes

```bash
BASE="$HOME/.local/share/neko-termux"

mkdir -p \
  "$BASE/config" \
  "$BASE/profile" \
  "$BASE/state" \
  "$BASE/logs" \
  "$BASE/backups"
```

### 4.4 Criar a configuração

Arquivo:

```text
~/.local/share/neko-termux/config/neko.yaml
```

Exemplo:

```yaml
server:
  bind: "127.0.0.1:8080"
  metrics: true
  pprof: false

desktop:
  screen: "1280x720@20"

member:
  provider: "file"
  file:
    path: "/data/members.json"
    hash: true

session:
  file: "/data/sessions.json"
  merciful_reconnect: true
  cookie:
    enabled: true
    secure: false
    http_only: true
    expiration: "168h"

webrtc:
  icelite: true
  tcpmux: 52000
  nat1to1: "127.0.0.1"

capture:
  microphone:
    enabled: false
  webcam:
    enabled: false

log:
  dir: "/var/log/neko"
  level: "info"
  nocolor: true
```

`secure: false` deve ser usado apenas em uma prova de conceito protegida por túnel SSH.

Para uso permanente, o ideal é HTTPS e cookie seguro.

### 4.5 Criar o usuário

```bash
read -rsp "Senha do Neko: " PASSWORD
echo

HASH="$(
  printf '%s' "$PASSWORD" |
  openssl sha256 -binary |
  base64
)"

unset PASSWORD
```

Crie o arquivo:

```text
~/.local/share/neko-termux/state/members.json
```

Exemplo:

```bash
cat > "$BASE/state/members.json" <<EOF
{
  "erick": {
    "password": "$HASH",
    "profile": {
      "name": "Érick",
      "is_admin": true,
      "can_login": true,
      "can_connect": true,
      "can_watch": true,
      "can_host": true,
      "can_share_media": true,
      "can_access_clipboard": true,
      "sends_inactive_cursor": true,
      "can_see_inactive_cursors": true,
      "plugins": {}
    }
  }
}
EOF

chmod 600 "$BASE/state/members.json"
```

### 4.6 Criar política persistente do Firefox

Arquivo:

```text
~/.local/share/neko-termux/config/policies.json
```

Conteúdo:

```json
{
  "policies": {
    "SanitizeOnShutdown": false,
    "Homepage": {
      "StartPage": "previous-session"
    },
    "ExtensionSettings": {
      "*": {
        "installation_mode": "allowed"
      }
    }
  }
}
```

### 4.7 Iniciar a imagem

O comando abaixo é uma proposta experimental.

Ele ainda precisa ser validado no POCO F6.

```bash
BASE="$HOME/.local/share/neko-termux"

termux-wake-lock

proot-distro run \
  --detach \
  --bind "$BASE/config/neko.yaml:/etc/neko/neko.yaml" \
  --bind "$BASE/config/policies.json:/usr/lib/firefox/distribution/policies.json" \
  --bind "$BASE/profile:/home/neko/.mozilla/firefox/profile.default" \
  --bind "$BASE/state:/data" \
  --bind "$BASE/logs:/var/log/neko" \
  --env PULSE_SERVER=unix:/tmp/pulseaudio.socket \
  --env NEKO_SERVER_BIND=127.0.0.1:8080 \
  --env NEKO_DESKTOP_SCREEN=1280x720@20 \
  --env NEKO_WEBRTC_TCPMUX=52000 \
  --env NEKO_WEBRTC_ICELITE=1 \
  --env NEKO_WEBRTC_NAT1TO1=127.0.0.1 \
  neko-firefox
```

O `PULSE_SERVER` pode precisar de adaptação porque o PRoot-Distro e a imagem do Neko podem assumir sockets diferentes.

Essa é uma das primeiras áreas que precisam ser testadas.

### 4.8 Verificar o estado

```bash
proot-distro ps
```

Teste HTTP:

```bash
curl -f http://127.0.0.1:8080/health
```

Ver logs:

```bash
proot-distro login neko-firefox -- \
  sh -lc 'ls -lah /var/log/neko && tail -F /var/log/neko/*.log'
```

Parar:

```bash
proot-distro kill neko-firefox
```

O comando `kill` do PRoot-Distro encerra a árvore de processos do ambiente, não apenas o processo pai.

---

## 5. CLI `neko-termux`

Depois que a prova de conceito funcionar, toda a complexidade pode ser escondida por uma CLI.

Interface sugerida:

```bash
neko-termux install
neko-termux start
neko-termux stop
neko-termux restart
neko-termux status
neko-termux logs
neko-termux doctor
neko-termux backup
neko-termux update
neko-termux uninstall
```

### 5.1 `install`

Responsabilidades:

1. confirmar arquitetura `aarch64`;
2. verificar espaço disponível;
3. instalar dependências;
4. verificar suporte OCI do PRoot-Distro;
5. criar diretórios;
6. baixar uma versão fixada da imagem;
7. gerar `neko.yaml`;
8. solicitar senha;
9. gerar `members.json`;
10. gerar `policies.json`;
11. validar permissões;
12. executar diagnóstico inicial.

### 5.2 `start`

Responsabilidades:

1. verificar se já está ativo;
2. executar `termux-wake-lock`;
3. verificar se portas estão livres;
4. montar configuração;
5. montar perfil;
6. montar estado;
7. montar logs;
8. iniciar em background;
9. esperar o endpoint de saúde;
10. validar HTTP;
11. validar WebRTC TCP mux;
12. imprimir o comando SSH.

Saída esperada:

```text
Neko iniciado.

Interface: 127.0.0.1:8080
WebRTC:   127.0.0.1:52000
Perfil:   ~/.local/share/neko-termux/profile

No computador:

ssh poco \
  -L 8080:127.0.0.1:8080 \
  -L 52000:127.0.0.1:52000

Abra:

http://localhost:8080
```

### 5.3 `status`

Exemplo:

```text
Container: ativo
PID PRoot: 18452
HTTP: saudável
WebRTC TCP mux: escutando
Firefox: ativo
Xorg: ativo
PulseAudio: ativo
Uso de RAM: 1,4 GB
Perfil: 382 MB
Tempo ativo: 01:42:17
```

### 5.4 `logs`

Deve reunir:

- logs do servidor Neko;
- Firefox;
- Openbox;
- Xorg;
- PulseAudio;
- supervisor;
- processo PRoot.

Opções possíveis:

```bash
neko-termux logs
neko-termux logs --follow
neko-termux logs firefox
neko-termux logs xorg
```

### 5.5 `doctor`

O comando `doctor` seria uma das partes mais importantes.

Exemplo:

```text
[OK] Arquitetura aarch64
[OK] PRoot-Distro com suporte OCI
[OK] Imagem instalada
[OK] Espaço disponível
[OK] Configuração YAML
[OK] Perfil gravável
[OK] Porta 8080 livre
[OK] Porta 52000 livre
[ERRO] Xorg encerrou
[AVISO] HyperOS pode limitar processos em segundo plano
```

Informações coletadas:

```bash
proot-distro ps
free -h
df -h
ss -tln
ps -eo pid,%cpu,%mem,cmd
tail -n 200 "$BASE/logs"/*.log
```

Também deve verificar:

- `termux-wake-lock`;
- permissão de gravação;
- versão da imagem;
- versão do PRoot-Distro;
- presença do Firefox;
- saúde da API;
- portas;
- espaço;
- arquitetura;
- processos filhos;
- política de bateria do Android, quando possível.

### 5.6 `backup`

Duas estratégias:

#### Backup leve

Inclui:

```text
configuração
perfil
usuários
sessões
políticas
```

Exemplo:

```bash
tar -czf \
  "$BASE/backups/neko-data-$(date +%F).tar.gz" \
  -C "$BASE" \
  config profile state
```

#### Backup completo

Inclui todo o ambiente PRoot:

```bash
proot-distro backup \
  neko-firefox \
  --output "$BASE/backups/neko-container-$(date +%F).tar.xz"
```

### 5.7 `update`

Não deve atualizar diretamente por cima da instalação existente.

Fluxo seguro:

```text
1. Fazer backup do perfil.
2. Baixar a nova imagem com outro nome.
3. Iniciar com um perfil temporário.
4. Validar o health check.
5. Validar browser e WebRTC.
6. Parar a versão antiga.
7. Anexar o perfil real.
8. Manter rollback disponível.
```

Exemplo:

```text
neko-firefox-3.1.0
neko-firefox-3.2.0-test
```

### 5.8 `uninstall`

Deve:

- parar o ambiente;
- oferecer backup;
- remover o container;
- manter ou remover o perfil;
- remover configuração;
- liberar wake lock.

Nunca deve apagar o perfil sem confirmação explícita.

---

## 6. Quando criar uma imagem personalizada

Talvez a imagem oficial funcione diretamente.

Caso contrário, não é necessário começar um port nativo para Termux.

É melhor criar uma imagem derivada.

Exemplo:

```dockerfile
FROM ghcr.io/m1k1o/neko/firefox:3.1.0

COPY neko.yaml /etc/neko/neko.yaml
COPY policies.json /usr/lib/firefox/distribution/policies.json
COPY supervisord-termux.conf /etc/neko/supervisord.conf
COPY start-termux.sh /usr/local/bin/start-termux

CMD ["/usr/local/bin/start-termux"]
```

Essa imagem pode corrigir:

- PulseAudio dentro do PRoot;
- parâmetros do Xorg;
- diretórios temporários;
- driver de entrada;
- permissões do perfil;
- logs;
- encerramento correto;
- resolução;
- FPS;
- sockets;
- comportamento do supervisor.

O build pode ocorrer no computador ou no GitHub Actions:

```text
GitHub Actions
    ↓
build linux/arm64
    ↓
publica no GHCR
    ↓
POCO baixa pelo PRoot-Distro
```

No celular:

```bash
proot-distro install \
  ghcr.io/ericklucioh/neko-termux-firefox:0.1.0 \
  --name neko-firefox
```

Uma imagem derivada é muito mais simples que recompilar o Neko inteiro.

---

## 7. Possível estrutura do repositório

```text
neko-termux/
├── cmd/
│   └── neko-termux/
│       └── main.go
│
├── internal/
│   ├── install/
│   ├── runtime/
│   ├── config/
│   ├── doctor/
│   ├── backup/
│   ├── update/
│   └── logs/
│
├── image/
│   ├── Dockerfile
│   ├── neko.yaml
│   ├── policies.json
│   ├── supervisord-termux.conf
│   └── start-termux.sh
│
├── scripts/
│   ├── install.sh
│   ├── test-arm64.sh
│   └── release.sh
│
├── .github/
│   └── workflows/
│       ├── build-arm64.yml
│       └── release.yml
│
└── README.md
```

A CLI pode ser escrita em Go.

Isso combina bem com:

- binário único do launcher;
- comandos claros;
- testes;
- logs;
- geração de configuração;
- download de releases;
- atualização;
- diagnóstico;
- suporte futuro a outros browsers.

---

## 8. Firefox antes de Chromium

A primeira versão deve usar Firefox.

Motivos:

- menos dependência de sandbox;
- menor exigência de `/dev/shm`;
- imagem oficial ARM64 disponível;
- perfil mais simples;
- melhor compatibilidade com PRoot;
- menos risco de crash causado por sandbox e memória compartilhada.

Chromium normalmente precisa de:

```text
--no-sandbox
```

e pode precisar de uma área `/dev/shm` grande.

Esses são pontos delicados em PRoot.

Ordem recomendada:

```text
1. Firefox
2. Chromium
3. Brave/Vivaldi
```

---

## 9. Metas de teste

Não comece tentando executar YouTube em 1080p/60.

Primeiro teste:

```text
Neko inicia
Firefox inicia
interface HTTP abre
mouse funciona
teclado funciona
login persiste
cookies persistem
página simples abre
áudio funciona
```

Depois:

```text
YouTube 480p
YouTube 720p
vídeo em tela cheia
múltiplas abas
download
upload
clipboard
reconexão
sessão persistente
```

Depois:

```text
uso prolongado
temperatura
consumo de bateria
uso de CPU
uso de RAM
estabilidade com tela apagada
comportamento no HyperOS
```

---

## 10. Limitações prováveis

### 10.1 Codificação de vídeo por CPU

Sem integração específica com:

- GPU Adreno;
- MediaCodec;
- aceleração de hardware Android;

o Neko provavelmente codificará a tela principalmente pela CPU.

Para estes usos, deve ser mais aceitável:

- documentação;
- páginas administrativas;
- Gmail;
- GitHub;
- sites comuns;
- baixa taxa de atualização;
- poucas abas.

Para estes usos, pode ficar pesado:

- YouTube 1080p/60;
- animações complexas;
- múltiplas abas de vídeo;
- longas sessões;
- alta resolução;
- 60 FPS.

### 10.2 DRM

As imagens ARM64 não oferecem DRM completo.

Serviços protegidos podem não funcionar:

- Netflix;
- Prime Video;
- algumas transmissões;
- conteúdo Widevine;
- alguns serviços de streaming.

### 10.3 PRoot

PRoot adiciona:

- interceptação de syscalls;
- custo de filesystem;
- custo de criação de processos;
- limitações de kernel;
- ausência de namespaces reais;
- diferenças em IPC;
- diferenças em sockets e mounts.

### 10.4 HyperOS

O HyperOS pode:

- suspender processos;
- encerrar o Termux;
- limitar CPU em background;
- limitar rede;
- encerrar processos filhos;
- interferir quando a tela apaga.

Será necessário:

```bash
termux-wake-lock
```

E no sistema:

- bateria sem restrições;
- autostart;
- aplicativo bloqueado nos recentes;
- atividade em segundo plano permitida.

---

## 11. O que não fazer

### 11.1 Não começar compilando Neko nativamente para Termux

Isso exigiria adaptar:

- CGO;
- Bionic;
- X11;
- GStreamer;
- GTK;
- paths do Termux;
- drivers;
- plugins Go;
- scripts;
- supervisor;
- navegador;
- áudio.

Dificuldade estimada:

```text
9/10
```

### 11.2 Não montar manualmente todo o Debian

A imagem oficial já contém o runtime preparado.

Use a imagem OCI como base.

### 11.3 Não usar `latest`

Use uma versão fixa.

### 11.4 Não apagar o perfil em updates

O perfil deve ficar fora do container.

### 11.5 Não expor diretamente as portas na internet

Use:

- SSH;
- Tailscale;
- VPN privada;
- HTTPS;
- autenticação.

---

## 12. Nível estimado de dificuldade

| Etapa | Dificuldade |
|---|---:|
| Compilar Neko nativo para Termux/Bionic | 9/10 |
| Montar Debian e dependências manualmente | 8/10 |
| Extrair e adaptar OCI manualmente | 7/10 |
| Instalar imagem pelo PRoot-Distro | 5/10 |
| Criar wrapper após funcionar | 4/10 |
| Usar diariamente após estabilizado | 2/10 |

O maior trabalho não deve ser a instalação final.

O principal desafio será descobrir quais componentes da imagem oficial precisam de adaptação para funcionar sob PRoot.

---

## 13. Plano de desenvolvimento

### Fase 1 — validar a imagem oficial

```text
Imagem oficial ARM64
+ proot-distro install
+ proot-distro run
```

Objetivos:

- imagem instala;
- entrypoint inicia;
- logs aparecem;
- primeira falha é identificada.

### Fase 2 — corrigir sem criar imagem derivada

Tentar corrigir com:

- `--bind`;
- `--env`;
- arquivo de configuração;
- política do Firefox;
- diretórios persistentes;
- parâmetros do runtime.

### Fase 3 — criar imagem derivada

Somente se necessário.

Objetivo:

- incluir patches estáveis;
- simplificar startup;
- remover dependências desnecessárias;
- adaptar supervisor;
- definir defaults para Termux.

### Fase 4 — criar `neko-termux`

Encapsular:

- instalação;
- start;
- stop;
- logs;
- status;
- doctor;
- backup;
- update.

### Fase 5 — integração com Termux

Adicionar:

- Termux:Boot;
- wake lock;
- atalhos;
- serviço persistente;
- logs locais;
- notificações.

### Fase 6 — benchmarks

Medir:

- tempo de inicialização;
- uso de RAM;
- uso de CPU;
- temperatura;
- autonomia;
- latência;
- FPS;
- estabilidade;
- YouTube 480p e 720p.

---

## 14. Descoberta central

A descoberta mais importante da pesquisa é:

> Não é necessário começar construindo uma rootfs nem compilando o Neko no celular.

A primeira tentativa deve ser:

```text
imagem oficial Neko Firefox ARM64
+
suporte OCI do PRoot-Distro
+
wrapper próprio
```

O projeto `neko-termux` pode ser principalmente um:

```text
launcher
configurador
supervisor
diagnosticador
sistema de backup
sistema de atualização
```

e não um fork completo do Neko.

---

## 15. Stack final imaginada

```text
POCO F6
└── Termux
    ├── OpenSSH
    ├── PRoot-Distro
    │   └── Neko Firefox ARM64
    │       ├── Firefox
    │       ├── Openbox
    │       ├── Xorg
    │       ├── PulseAudio
    │       ├── GStreamer
    │       ├── servidor Neko
    │       └── frontend web
    │
    └── neko-termux
        ├── install
        ├── start
        ├── stop
        ├── status
        ├── logs
        ├── doctor
        ├── backup
        └── update
```

Acesso:

```text
Computador
└── navegador
    └── http://localhost:8080
            │
            │ túnel SSH
            ▼
        POCO F6
        Neko + Firefox real
```

---

## 16. Referências principais

- Neko: https://github.com/m1k1o/neko
- Documentação do Neko: https://neko.m1k1o.net/
- Imagens do Neko: https://neko.m1k1o.net/docs/v3/installation/docker-images
- Configuração do Neko: https://neko.m1k1o.net/docs/v3/configuration
- Autenticação: https://neko.m1k1o.net/docs/v3/configuration/authentication
- Networking e SSH: https://neko.m1k1o.net/docs/v3/customization/networking
- Customização de browsers: https://neko.m1k1o.net/docs/v3/customization/browsers
- PRoot-Distro: https://github.com/termux/proot-distro
- Termux: https://termux.dev/

---

## 17. Próximo passo quando o projeto for retomado

O primeiro experimento deve ser o menor possível:

```bash
pkg install proot-distro

proot-distro install \
  ghcr.io/m1k1o/neko/firefox:3.1.0 \
  --name neko-firefox

proot-distro run neko-firefox
```

O objetivo não é esperar que funcione perfeitamente.

O objetivo é obter o primeiro erro real.

Depois, cada erro deve ser resolvido na seguinte ordem:

```text
configuração
variável de ambiente
bind mount
script de inicialização
imagem derivada
port nativo
```

O port nativo para Termux deve ser considerado somente como último estágio.
