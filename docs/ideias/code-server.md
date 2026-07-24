# Code-server no Ubuntu PRoot do Android

## Arquitetura final

O ambiente ficará assim:

```text
Poco F6
└── Termux
    ├── servidor SSH na porta 8022
    ├── tmux
    └── Ubuntu PRoot
        ├── code-server na porta 1212
        ├── /root/workspace
        ├── extensões
        ├── configurações
        └── ferramentas de desenvolvimento
```

O processamento acontece no celular. O navegador do computador é apenas o cliente visual.

O PRoot-Distro executa um ambiente Linux sem exigir root real no Android. Portanto, o usuário `root` dentro do Ubuntu não significa que seu Android esteja rooteado.

---

# 1. Preparar o Termux

Execute estes comandos diretamente no Termux, fora do Ubuntu:

```bash
pkg update
pkg upgrade -y

pkg install -y \
  proot-distro \
  openssh \
  tmux
```

O `proot-distro` é distribuído oficialmente pelo gerenciador de pacotes do Termux.

Confira os ambientes instalados:

```bash
proot-distro list
```

Caso ainda não tenha instalado o Ubuntu:

```bash
proot-distro install ubuntu:24.04
```

Em instalações anteriores, o comando também pode ter sido:

```bash
proot-distro install ubuntu
```

Como o seu Ubuntu já está instalado e funcionando, você deve apenas entrar nele:

```bash
proot-distro login ubuntu
```

---

# 2. Preparar o Ubuntu

Agora os comandos são executados dentro do Ubuntu PRoot:

```bash
apt update
apt upgrade -y
```

Instale as ferramentas básicas:

```bash
apt install -y \
  curl \
  ca-certificates \
  git \
  openssl \
  nano
```

Confira a arquitetura:

```bash
uname -m
```

No Poco F6, provavelmente aparecerá:

```text
aarch64
```

O projeto code-server publica builds oficiais para `amd64` e `arm64`.

---

# 3. Instalar o code-server

Primeiro, visualize o que o instalador pretende fazer:

```bash
curl -fsSL https://code-server.dev/install.sh |
  sh -s -- --dry-run
```

Depois instale:

```bash
curl -fsSL https://code-server.dev/install.sh | sh
```

No Ubuntu, o instalador detecta uma distribuição baseada em Debian e normalmente instala o pacote correspondente.

Confira:

```bash
code-server --version
```

No seu caso atual:

```text
code-server 4.129.0
```

Caso apareça `code-server: command not found`, adicione o diretório local ao `PATH`:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

E teste novamente:

```bash
code-server --version
```

---

# 4. Criar o workspace

Ainda dentro do Ubuntu:

```bash
mkdir -p /root/workspace
```

Esse será o diretório principal para seus projetos:

```text
/root/workspace
```

Exemplo:

```text
/root/workspace/
├── mobdesk/
├── estudos-go/
├── ecommerce-tcc/
└── testes/
```

Para clonar um projeto:

```bash
cd /root/workspace
git clone URL_DO_REPOSITORIO
```

---

# 5. Gerar a configuração inicial

Execute o code-server pela primeira vez:

```bash
code-server /root/workspace
```

Na primeira execução, ele cria automaticamente:

```text
/root/.config/code-server/config.yaml
```

A configuração padrão usa:

```yaml
bind-addr: 127.0.0.1:8080
auth: password
password: senha-gerada
cert: false
```

O padrão é escutar apenas em `localhost`, usar autenticação por senha e não fornecer HTTPS diretamente.

Pare com:

```text
Ctrl+C
```

---

# 6. Configurar a porta 1212

Abra o arquivo:

```bash
nano /root/.config/code-server/config.yaml
```

Configure assim:

```yaml
bind-addr: 127.0.0.1:1212
auth: password
password: SUA_SENHA
cert: false
```

Para gerar uma senha forte:

```bash
openssl rand -hex 24
```

Copie o resultado e coloque no arquivo:

```yaml
password: resultado-gerado
```

Exemplo:

```yaml
bind-addr: 127.0.0.1:1212
auth: password
password: 79dd05f99af6762f49aa27387178f842b212906d930dc001
cert: false
```

Proteja o arquivo:

```bash
chmod 600 /root/.config/code-server/config.yaml
```

Confira:

```bash
cat /root/.config/code-server/config.yaml
```

## Por que manter `127.0.0.1`

Esta configuração:

```yaml
bind-addr: 127.0.0.1:1212
```

faz o code-server aceitar conexões apenas do próprio celular ou de túneis locais.

Isso é mais seguro que:

```yaml
bind-addr: 0.0.0.0:1212
```

O code-server recomenda usar encaminhamento SSH ou outra solução segura ao acessá-lo de outra máquina.

---

# 7. Rodar o code-server

Dentro do Ubuntu:

```bash
code-server /root/workspace
```

O log correto será semelhante a:

```text
HTTP server listening on http://127.0.0.1:1212/
Authentication is enabled
Using password from /root/.config/code-server/config.yaml
Not serving HTTPS
```

O significado é:

```text
127.0.0.1:1212
```

O serviço está disponível localmente na porta `1212`.

```text
Authentication is enabled
```

Será necessário informar a senha.

```text
Not serving HTTPS
```

Você deve usar:

```text
http://
```

e não:

```text
https://
```

---

# 8. Testar no celular

Com o code-server rodando, abra no navegador do celular:

```text
http://127.0.0.1:1212
```

Digite a senha configurada em:

```text
/root/.config/code-server/config.yaml
```

Também é possível testar pelo terminal, usando outra sessão do Termux:

```bash
curl http://127.0.0.1:1212/healthz
```

A resposta deve ser semelhante a:

```json
{
  "status": "alive",
  "lastHeartbeat": 0
}
```

O endpoint `/healthz` serve para verificar se o code-server está ativo e não exige autenticação.

---

# 9. Configurar o servidor SSH no Termux

O servidor SSH deve ser iniciado no Termux, fora do Ubuntu PRoot.

Abra uma nova sessão do Termux ou saia temporariamente do Ubuntu:

```bash
exit
```

Defina uma senha para o usuário do Termux:

```bash
passwd
```

Inicie o servidor SSH:

```bash
sshd
```

Descubra o usuário:

```bash
whoami
```

O resultado será semelhante a:

```text
u0_a384
```

Esse é o usuário que deve ser usado pelo computador.

O Termux não possui senha configurada por padrão para login SSH; o comando `passwd` define essa senha.

## Descobrir o IP do celular

Como o Android pode bloquear comandos como `ip a`, veja o IP em:

```text
Configurações
→ Wi-Fi
→ Rede conectada
→ Endereço IP
```

No seu caso atual:

```text
192.168.3.228
```

Esse endereço pode mudar ao reconectar ao Wi-Fi.

---

# 10. Criar o túnel no computador

No computador, execute:

```bash
ssh -p 8022 -N \
  -o ServerAliveInterval=15 \
  -o ServerAliveCountMax=3 \
  -o ExitOnForwardFailure=yes \
  -L 1212:127.0.0.1:1212 \
  USUARIO_TERMUX@192.168.3.228
```

Exemplo:

```bash
ssh -p 8022 -N \
  -o ServerAliveInterval=15 \
  -o ServerAliveCountMax=3 \
  -o ExitOnForwardFailure=yes \
  -L 1212:127.0.0.1:1212 \
  u0_a384@192.168.3.228
```

Digite a senha criada com `passwd`.

Enquanto esse comando estiver aberto, acesse no computador:

```text
http://127.0.0.1:1212
```

O fluxo será:

```text
Navegador do computador
        │
        ▼
localhost:1212 do computador
        │
        ▼
Túnel SSH criptografado
        │
        ▼
Termux no celular
        │
        ▼
127.0.0.1:1212 do celular
        │
        ▼
code-server dentro do Ubuntu PRoot
```

O SSH pode encaminhar portas TCP arbitrárias através do canal criptografado.

---

# 11. Encaminhar várias portas

Você pode repetir a opção `-L` quantas vezes precisar:

```bash
ssh -p 8022 -N \
  -o ServerAliveInterval=15 \
  -o ServerAliveCountMax=3 \
  -o ExitOnForwardFailure=yes \
  -L 1212:127.0.0.1:1212 \
  -L 3000:127.0.0.1:3000 \
  -L 5173:127.0.0.1:5173 \
  -L 8080:127.0.0.1:8080 \
  USUARIO_TERMUX@192.168.3.228
```

Isso disponibiliza no computador:

```text
http://127.0.0.1:1212 → code-server
http://127.0.0.1:3000 → aplicação na porta 3000
http://127.0.0.1:5173 → Vite
http://127.0.0.1:8080 → aplicação na porta 8080
```

A sintaxe é:

```text
-L PORTA_NO_PC:ENDEREÇO_NO_CELULAR:PORTA_NO_CELULAR
```

Exemplo:

```bash
-L 5174:127.0.0.1:5173
```

Significa:

```text
PC:5174 → celular:5173
```

Você acessaria:

```text
http://127.0.0.1:5174
```

enquanto o servidor continua rodando na porta `5173` do celular.

## Adicionar portas depois

Se o primeiro túnel já estiver rodando, você pode abrir outro terminal no computador:

```bash
ssh -p 8022 -N \
  -L 4000:127.0.0.1:4000 \
  -L 9000:127.0.0.1:9000 \
  USUARIO_TERMUX@192.168.3.228
```

Os dois túneis podem permanecer ativos simultaneamente.

Apenas não tente usar a mesma porta local duas vezes. Por exemplo, duas conexões não podem escutar simultaneamente em:

```text
127.0.0.1:1212
```

A outra alternativa é encerrar o túnel atual com `Ctrl+C` e abri-lo novamente com todas as portas.

---

# 12. Salvar a conexão no SSH config

No computador, abra:

```bash
nano ~/.ssh/config
```

Adicione:

```sshconfig
Host poco
    HostName 192.168.3.228
    User USUARIO_TERMUX
    Port 8022

    ServerAliveInterval 15
    ServerAliveCountMax 3
    ExitOnForwardFailure yes

    LocalForward 1212 127.0.0.1:1212
    LocalForward 3000 127.0.0.1:3000
    LocalForward 5173 127.0.0.1:5173
```

Substitua:

```text
USUARIO_TERMUX
```

por algo como:

```text
u0_a384
```

Depois, o túnel completo será iniciado com:

```bash
ssh -N poco
```

Se quiser entrar normalmente no terminal do celular:

```bash
ssh poco
```

---

# 13. Configurar autenticação SSH por chave

No computador:

```bash
ssh-keygen -t ed25519
```

Pressione Enter para aceitar o caminho padrão.

Copie a chave para o Termux:

```bash
ssh-copy-id -p 8022 USUARIO_TERMUX@192.168.3.228
```

Exemplo:

```bash
ssh-copy-id -p 8022 u0_a384@192.168.3.228
```

Depois, teste:

```bash
ssh -p 8022 u0_a384@192.168.3.228
```

O SSH oferece autenticação por chave pública e canais criptografados para login e encaminhamento de portas.

---

# 14. Manter o code-server rodando com tmux

O Ubuntu PRoot não executa normalmente um sistema completo de inicialização como `systemd`. O PRoot-Distro permite processos longos individuais, mas supervisores completos como `systemd` não são geralmente suportados.

Portanto, não use:

```bash
systemctl enable --now code-server
```

Use `tmux`.

## Forma manual

No Termux:

```bash
tmux new -s code-server
```

Dentro do tmux:

```bash
proot-distro login ubuntu
```

Dentro do Ubuntu:

```bash
code-server /root/workspace
```

Para sair da tela sem encerrar o servidor:

```text
Ctrl+B
D
```

Primeiro pressione `Ctrl+B`, solte e pressione `D`.

Para voltar:

```bash
tmux attach -t code-server
```

Para listar as sessões:

```bash
tmux ls
```

Para encerrar completamente:

```bash
tmux kill-session -t code-server
```

## Iniciar diretamente em segundo plano

No Termux:

```bash
tmux new-session -d \
  -s code-server \
  'proot-distro login ubuntu -- bash -lc "exec code-server /root/workspace"'
```

Confira:

```bash
tmux ls
```

Veja os logs:

```bash
tmux attach -t code-server
```

---

# 15. Evitar que o Android encerre o Termux

No Termux:

```bash
termux-wake-lock
```

Nas configurações do Android, coloque o Termux como aplicativo sem restrições de bateria:

```text
Configurações
→ Aplicativos
→ Termux
→ Bateria
→ Sem restrições
```

Mesmo assim, o Android ainda pode encerrar processos em situações de forte pressão de memória, reinicialização ou encerramento manual do aplicativo.

Para liberar o wake lock:

```bash
termux-wake-unlock
```

---

# 16. Instalar extensões

As extensões podem ser instaladas pelo botão de extensões:

```text
Ctrl+Shift+X
```

Ou pelo terminal:

```bash
code-server --install-extension ID_DA_EXTENSAO
```

Exemplo:

```bash
code-server --install-extension eamodio.gitlens
```

Listar extensões:

```bash
code-server --list-extensions
```

Ver versões:

```bash
code-server --list-extensions --show-versions
```

As extensões ficam em:

```text
/root/.local/share/code-server/extensions
```

As configurações ficam dentro de:

```text
/root/.local/share/code-server
```

Esses são os diretórios padrão documentados pelo code-server.

## Celular e computador já compartilham extensões

Quando você abre:

```text
http://127.0.0.1:1212
```

no celular e no computador, ambos acessam a mesma instalação do code-server.

Portanto:

```text
navegador do celular
        │
        ├── mesma instalação
        │
navegador do PC
        │
        ▼
/root/.local/share/code-server/extensions
```

Você não precisa sincronizar extensões entre o navegador do PC e o navegador do celular.

---

# 17. Por que não sincroniza automaticamente com o VS Code desktop

O VS Code instalado no seu notebook e o code-server instalado no Ubuntu são duas instalações separadas.

Além disso, o code-server não usa oficialmente o Microsoft Visual Studio Marketplace. Ele utiliza o Open VSX porque os termos do marketplace da Microsoft restringem suas extensões aos produtos Visual Studio.

Isso significa que:

```text
VS Code do notebook
├── extensões próprias
├── configurações próprias
└── Microsoft Marketplace

code-server do celular
├── extensões próprias
├── configurações próprias
└── Open VSX
```

Algumas extensões podem:

* não existir no Open VSX;
* ter nome ou versão diferente;
* exigir instalação manual por VSIX;
* não funcionar corretamente no navegador;
* depender de binários incompatíveis com Linux ARM64.

---

# 18. Copiar a lista de extensões do VS Code desktop

No computador:

```bash
code --list-extensions |
  sort -u > extensions.txt
```

Confira:

```bash
cat extensions.txt
```

Envie para o Termux:

```bash
scp -P 8022 \
  extensions.txt \
  USUARIO_TERMUX@192.168.3.228:~/
```

No Termux, copie para dentro do Ubuntu:

```bash
proot-distro copy \
  ~/extensions.txt \
  ubuntu:/root/extensions.txt
```

O PRoot-Distro possui um comando próprio para copiar arquivos entre o Termux e o filesystem interno do container.

Entre no Ubuntu:

```bash
proot-distro login ubuntu
```

Instale as extensões:

```bash
: > /root/extensions-failed.txt

while IFS= read -r extension; do
  [ -z "$extension" ] && continue

  echo "Instalando: $extension"

  code-server --install-extension "$extension" ||
    echo "$extension" >> /root/extensions-failed.txt
done < /root/extensions.txt
```

Veja as que falharam:

```bash
cat /root/extensions-failed.txt
```

Falhas não significam necessariamente problema no code-server. A extensão pode simplesmente não estar disponível no Open VSX.

---

# 19. Instalar extensão manualmente por VSIX

Quando uma extensão não estiver disponível, procure um arquivo oficial `.vsix` na página de releases do projeto.

Depois instale:

```bash
code-server --install-extension extensao.vsix
```

Ou, dentro do code-server:

```text
Ctrl+Shift+P
Extensions: Install from VSIX...
```

A instalação manual por VSIX é oficialmente suportada pelo code-server.

---

# 20. Copiar configurações do VS Code desktop

No Linux, as configurações do VS Code desktop normalmente ficam em:

```text
~/.config/Code/User/settings.json
~/.config/Code/User/keybindings.json
~/.config/Code/User/snippets/
```

No computador, envie os arquivos para o Termux:

```bash
scp -P 8022 \
  ~/.config/Code/User/settings.json \
  USUARIO_TERMUX@192.168.3.228:~/
```

Caso tenha atalhos personalizados:

```bash
scp -P 8022 \
  ~/.config/Code/User/keybindings.json \
  USUARIO_TERMUX@192.168.3.228:~/
```

No Termux:

```bash
proot-distro copy \
  ~/settings.json \
  ubuntu:/root/.local/share/code-server/User/settings.json
```

Para os atalhos:

```bash
proot-distro copy \
  ~/keybindings.json \
  ubuntu:/root/.local/share/code-server/User/keybindings.json
```

Antes, pode ser necessário criar o diretório:

```bash
proot-distro login ubuntu -- \
  mkdir -p /root/.local/share/code-server/User
```

Reinicie o code-server depois da cópia.

O code-server também documenta o uso de uma extensão de Settings Sync ou a reutilização manual do diretório de configurações.

Para um ambiente reproduzível, o método mais previsível é manter configurações e listas de extensões em um repositório privado de dotfiles.

---

# 21. Entender os dados por usuário

Você está executando o code-server como:

```text
root
```

Por isso os dados ficam em:

```text
/root/.config/code-server
/root/.local/share/code-server
/root/workspace
```

Se posteriormente criar outro usuário, por exemplo `erick`, ele terá seus próprios diretórios:

```text
/home/erick/.config/code-server
/home/erick/.local/share/code-server
/home/erick/workspace
```

Nesse caso, as extensões parecerão ter desaparecido, mas estarão apenas armazenadas no diretório do usuário anterior.

Procure sempre iniciar o code-server com o mesmo usuário.

---

# 22. Sincronizar os projetos

Ao acessar o code-server pelo PC, você edita diretamente os arquivos armazenados no celular:

```text
/root/workspace
```

Não existe uma cópia automática no disco do computador.

Para sincronizar projetos entre celular, computador e GitHub, use Git:

```bash
cd /root/workspace/projeto

git status
git add .
git commit -m "Descrição da alteração"
git push
```

No computador:

```bash
git pull
```

O Git deve ser o mecanismo de sincronização dos projetos. O code-server oferece acesso remoto ao mesmo filesystem, mas não funciona como Dropbox ou OneDrive.

---

# 23. Acessar diretamente pela rede local

Existe uma alternativa sem túnel SSH.

Edite:

```bash
nano /root/.config/code-server/config.yaml
```

Use:

```yaml
bind-addr: 0.0.0.0:1212
auth: password
password: SUA_SENHA
cert: false
```

Reinicie:

```bash
code-server /root/workspace
```

O log mostrará:

```text
HTTP server listening on http://0.0.0.0:1212/
```

No computador:

```text
http://192.168.3.228:1212
```

Porém, essa alternativa expõe o serviço para outros dispositivos da mesma rede.

Como `cert: false`, continue usando:

```text
http://
```

e não:

```text
https://
```

Para o uso diário, mantenha:

```yaml
bind-addr: 127.0.0.1:1212
```

e utilize o túnel SSH.

---

# 24. Atualizar o code-server

Dentro do Ubuntu:

```bash
curl -fsSL https://code-server.dev/install.sh | sh
```

Depois confira:

```bash
code-server --version
```

Reinicie a sessão do tmux:

```bash
exit
tmux kill-session -t code-server
```

E inicie novamente:

```bash
tmux new-session -d \
  -s code-server \
  'proot-distro login ubuntu -- bash -lc "exec code-server /root/workspace"'
```

---

# 25. Fazer backup completo do Ubuntu

Primeiro, pare o code-server:

```bash
tmux kill-session -t code-server
```

No Termux:

```bash
proot-distro backup ubuntu \
  --output ~/ubuntu-code-server-$(date +%Y-%m-%d).tar.xz
```

O backup do PRoot-Distro inclui o filesystem completo do Ubuntu.

Confira:

```bash
ls -lh ~/ubuntu-code-server-*.tar.xz
```

Esse arquivo inclui:

```text
/root/workspace
/root/.config/code-server
/root/.local/share/code-server
pacotes instalados
configurações do Ubuntu
extensões
projetos
```

Para restaurar:

```bash
proot-distro restore ubuntu-code-server-DATA.tar.xz
```

Não execute:

```bash
proot-distro reset ubuntu
```

sem backup. O comando `reset` reinstala o container e remove todos os dados internos.

---

# 26. Diagnóstico de problemas

## code-server não abre no celular

Confira se está rodando:

```bash
curl http://127.0.0.1:1212/healthz
```

Se falhar, inicie:

```bash
proot-distro login ubuntu
code-server /root/workspace
```

## code-server abre no celular, mas não no PC

Teste o SSH:

```bash
ssh -p 8022 USUARIO_TERMUX@192.168.3.228
```

Depois teste o túnel:

```bash
ssh -v -p 8022 -N \
  -L 1212:127.0.0.1:1212 \
  USUARIO_TERMUX@192.168.3.228
```

## Aparece erro de porta ocupada no PC

Use outra porta local:

```bash
ssh -p 8022 -N \
  -L 1213:127.0.0.1:1212 \
  USUARIO_TERMUX@192.168.3.228
```

Abra:

```text
http://127.0.0.1:1213
```

## O navegador mostra erro de HTTPS

Não use:

```text
https://127.0.0.1:1212
```

Use:

```text
http://127.0.0.1:1212
```

Seu arquivo possui:

```yaml
cert: false
```

## As extensões desapareceram

Confira o usuário:

```bash
whoami
```

Confira o diretório:

```bash
ls /root/.local/share/code-server/extensions
```

Se você iniciou o code-server com outro usuário ou outro `--user-data-dir`, os dados estarão em outro caminho.

## `systemctl` não funciona

Isso é esperado no PRoot. Use `tmux` para manter o code-server ativo. O PRoot-Distro não fornece um sistema init completo.

## O PC não alcança o celular

Verifique:

```text
1. PC e celular estão na mesma rede Wi-Fi
2. O IP do celular ainda é 192.168.3.228
3. O sshd está rodando
4. A porta utilizada é 8022
5. O Android não encerrou o Termux
6. A rede não possui isolamento entre dispositivos
```

---

# 27. Rotina diária recomendada

## No celular

Abra o Termux:

```bash
termux-wake-lock
```

Inicie o SSH:

```bash
sshd
```

Inicie o code-server:

```bash
tmux new-session -d \
  -s code-server \
  'proot-distro login ubuntu -- bash -lc "exec code-server /root/workspace"'
```

Se a sessão já existir, não crie outra. Confira:

```bash
tmux ls
```

## No computador

Com o arquivo `~/.ssh/config` configurado:

```bash
ssh -N poco
```

Abra:

```text
http://127.0.0.1:1212
```

## Para parar

No computador:

```text
Ctrl+C
```

No celular:

```bash
tmux kill-session -t code-server
termux-wake-unlock
```

---

# Configuração final recomendada

```text
Ubuntu:
  ambiente: PRoot-Distro
  usuário: root
  workspace: /root/workspace

code-server:
  porta: 1212
  bind: 127.0.0.1
  autenticação: senha
  HTTPS próprio: desativado
  execução: tmux

Termux:
  SSH: porta 8022
  usuário: u0_a...
  persistência: tmux + wake lock

Computador:
  acesso: túnel SSH
  editor: http://127.0.0.1:1212
  portas extras: múltiplas opções -L

Extensões:
  diretório: /root/.local/share/code-server/extensions
  marketplace: Open VSX
  sincronização: lista exportada ou dotfiles
  extensão ausente: instalar por VSIX

Projetos:
  armazenamento real: celular
  acesso pelo PC: remoto pelo navegador
  sincronização entre máquinas: Git

Backup:
  proot-distro backup ubuntu
```
