Faça **dentro do Debian/Ubuntu do PRoot**, não no shell nativo do Termux. A Microsoft disponibiliza oficialmente o CLI standalone para Linux ARM64, que é a arquitetura do Poco F6. ([Visual Studio Code][1])

## 1. Entrar no Debian

No Termux:

```bash
pkg update
pkg install proot-distro
```

Caso ainda não tenha instalado:

```bash
proot-distro install debian
```

Entre nele:

```bash
proot-distro login debian
```

O prompt deve mudar para algo parecido com:

```text
root@localhost:~#
```

## 2. Confirmar a arquitetura

Dentro do Debian:

```bash
uname -m
```

O resultado esperado é:

```text
aarch64
```

Instale as dependências básicas:

```bash
apt update
apt install -y curl ca-certificates tar
```

O VS Code Server requer um Linux baseado em glibc, como Debian 10+ ou Ubuntu 20.04+, inclusive em ARM64. O Debian atual do `proot-distro` normalmente atende à parte de bibliotecas, embora o ambiente PRoot em Android não seja oficialmente certificado pela Microsoft. ([Visual Studio Code][2])

## 3. Baixar o VS Code CLI oficial ARM64

```bash
cd /tmp

curl -L \
  "https://update.code.visualstudio.com/latest/cli-linux-arm64/stable" \
  -o vscode_cli.tar.gz
```

Extraia:

```bash
tar -xzf vscode_cli.tar.gz
```

O pacote deve criar um executável chamado:

```text
code
```

Confirme:

```bash
ls -lh code
```

## 4. Instalar no PATH

```bash
install -m 0755 code /usr/local/bin/code
```

Agora teste:

```bash
code --version
```

A Microsoft documenta o endereço `cli-linux-arm64` como o pacote oficial do CLI para Linux ARM64, e aceita `latest` para baixar a versão estável mais recente. ([Visual Studio Code][1])

## 5. Iniciar o túnel

Ainda dentro do Debian:

```bash
code tunnel
```

Na primeira execução, ele pedirá para aceitar os termos do VS Code Server. Você também pode aceitar diretamente:

```bash
code tunnel --accept-server-license-terms
```

O terminal apresentará um processo de autenticação. Siga o endereço e o código mostrados, entrando com sua conta GitHub ou Microsoft.

Depois, ele exibirá uma URL semelhante a:

```text
https://vscode.dev/tunnel/nome-da-maquina
```

Abra essa URL no navegador do notebook. O comando baixa e inicia o VS Code Server no celular e cria o túnel oficial da Microsoft. ([Visual Studio Code][3])

## Comandos completos

No Termux:

```bash
pkg update
pkg install -y proot-distro

proot-distro install debian
proot-distro login debian
```

No Debian:

```bash
apt update
apt install -y curl ca-certificates tar

cd /tmp

curl -L \
  "https://update.code.visualstudio.com/latest/cli-linux-arm64/stable" \
  -o vscode_cli.tar.gz

tar -xzf vscode_cli.tar.gz

install -m 0755 code /usr/local/bin/code

code --version
code tunnel --accept-server-license-terms
```

## Para executar novamente depois

Sempre que fechar tudo:

```bash
proot-distro login debian
code tunnel
```

O túnel só fica acessível enquanto o processo `code tunnel` estiver executando. A própria documentação recomenda deixar o comando ativo ou usar mecanismos de serviço; dentro do Android/PRoot, primeiro teste em primeiro plano antes de tentar automatizar. ([Visual Studio Code][3])

Caso apareça algum erro, os três comandos mais úteis para diagnóstico são:

```bash
uname -m
ldd --version
code --version
```

[1]: https://code.visualstudio.com/docs/supporting/faq "Visual Studio Code FAQ"
[2]: https://code.visualstudio.com/docs/remote/linux "Remote Development with Linux"
[3]: https://code.visualstudio.com/docs/remote/tunnels "Developing with Remote Tunnels"
----------------------------------------------------


Certo, **é Ubuntu via PRoot**. Os comandos de instalação continuam praticamente iguais. O erro atual **não aconteceu por ser Ubuntu em vez de Debian**: o CLI foi instalado corretamente e falhou ao tentar acessar o endpoint HTTPS de autenticação do GitHub.

Dentro do Ubuntu, rode:

```bash
apt update
apt install --reinstall -y ca-certificates curl openssl
update-ca-certificates
```

Depois confira data, DNS e HTTPS:

```bash
date
getent hosts github.com
curl -Iv https://github.com/login/device
```

O `curl` pode retornar `200`, `301`, `302` ou até outro status HTTP; o importante é não aparecer erro de DNS, certificado ou conexão.

Também elimine possíveis variáveis de proxy, pois há casos registrados em que `HTTPS_PROXY` interfere no `code tunnel`: ([GitHub][1])

```bash
env | grep -i proxy

unset HTTP_PROXY HTTPS_PROXY ALL_PROXY
unset http_proxy https_proxy all_proxy
```

Tente novamente mostrando o erro completo:

```bash
code tunnel \
  --accept-server-license-terms \
  --verbose \
  --log trace
```

O Remote Tunnel oficialmente precisa acessar GitHub ou Microsoft para autenticar e depois cria conexões de saída com o serviço de túneis; não é necessário abrir portas no Android. ([Visual Studio Code][2])

## Se o `curl` apresentar erro de DNS

Veja o DNS atual:

```bash
cat /etc/resolv.conf
```

Você pode testar temporariamente:

```bash
printf 'nameserver 1.1.1.1\nnameserver 8.8.8.8\n' > /etc/resolv.conf
```

Depois:

```bash
getent hosts github.com
curl -Iv https://github.com/login/device
```

## Compare com o Termux nativo

Saia temporariamente do Ubuntu:

```bash
exit
```

No Termux:

```bash
pkg install curl
curl -Iv https://github.com/login/device
```

Interpretação:

```text
Funciona no Termux, mas não no Ubuntu
→ problema de DNS, certificado ou proxy dentro do PRoot

Não funciona nem no Termux
→ problema da rede Android, VPN, DNS privado, firewall ou bloqueador

Funciona nos dois, mas o code tunnel falha
→ problema específico do CLI; o log trace mostrará a causa
```

Você também pode testar entrar com **Microsoft Account** em vez de GitHub. Isso não impede usar GitHub, Git, repositórios ou Copilot posteriormente dentro do VS Code:

```bash
code tunnel --accept-server-license-terms
```

E selecione:

```text
Microsoft Account
```

A instalação está certa:

```text
code 1.129.0
Linux ARM64
Ubuntu PRoot
```

Agora o ponto a corrigir é exclusivamente a comunicação HTTPS durante a autenticação.

[1]: https://github.com/microsoft/vscode/issues/235874?utm_source=chatgpt.com "Visual Studio Code Server \"code tunnel\" doesn't use ..."
[2]: https://code.visualstudio.com/docs/remote/tunnels?utm_source=chatgpt.com "Developing with Remote Tunnels"


deu certo, mas ainda falta validar como produto

