# Conclusão: melhor opção para rodar um navegador no POCO F6 e acessá-lo pelo navegador de outro PC

## Minha conclusão final

A melhor opção para o objetivo proposto é:

```text
Neko + Firefox ARM64
rodando pelo PRoot-Distro no Termux
com acesso externo pelo Tailscale
```

A arquitetura ficaria:

```text
POCO F6
└── Android
    └── Termux
        └── PRoot-Distro
            └── imagem OCI do Neko Firefox ARM64
                ├── Firefox real
                ├── perfil persistente
                ├── Xorg/Openbox
                ├── áudio
                ├── servidor Neko
                └── WebRTC
                     │
                     │ Tailscale
                     ▼
              PC do João
              └── navegador comum
                  └── http://poco-f6:8080
```

Essa é a solução que mais corresponde ao objetivo:

> Um navegador real rodando no celular, mantendo logins, cookies e sessões, cuja interface aparece remotamente dentro do navegador de outro computador.

O Neko foi criado especificamente para transmitir um navegador ou aplicação Linux por WebRTC. Ele envia vídeo e áudio e recebe mouse e teclado, em vez de tentar incorporar os sites em um `iframe`.

Referência:

- https://neko.m1k1o.net/

---

## Por que Neko e não Xpra?

O Xpra é um bom plano B. Ele consegue exportar uma aplicação gráfica individual, possui cliente HTML5 integrado e suporta HTTP, WebSocket, áudio e clipboard.

Referência:

- https://xpra.org/

Mas o Neko é mais adequado porque foi projetado especificamente para:

- navegador remoto;
- vídeo e áudio;
- baixa latência via WebRTC;
- controle por mouse e teclado;
- múltiplos participantes;
- sessão persistente;
- uso diretamente dentro de uma página web.

O Neko também utiliza uma arquitetura mais apropriada para vídeo e áudio do que abordagens tradicionais baseadas em VNC ou transmissão de imagens por WebSocket.

Comparação:

| Objetivo | Melhor escolha |
|---|---|
| Exportar qualquer aplicação Linux | Xpra |
| Abrir um navegador remotamente | **Neko** |
| Vídeo, YouTube e áudio | **Neko** |
| Sessão compartilhada com João | **Neko** |
| Solução genérica e mais tradicional | Xpra |
| Plano B caso o Neko não rode no PRoot | Xpra |

---

## Por que Firefox?

A imagem recomendada para começar é:

```text
ghcr.io/m1k1o/neko/firefox:3.1.0
```

O Neko publica oficialmente Firefox para `linux/arm64`, arquitetura compatível com o POCO F6.

Referência:

- https://neko.m1k1o.net/docs/v3/installation/docker-images

Firefox é preferível ao Chromium no primeiro teste porque:

- não exige as mesmas capacidades adicionais;
- não depende tanto de uma área `/dev/shm` grande;
- evita parte dos problemas de sandbox;
- tende a ser mais simples dentro de PRoot.

Chromium normalmente é executado com:

```text
--no-sandbox
```

e pode precisar de uma área de memória compartilhada maior para não travar.

Ordem recomendada:

```text
1. Neko Firefox
2. Neko Chromium
3. Brave ou Vivaldi, se houver necessidade específica
```

---

## Por que PRoot-Distro?

O PRoot-Distro resolve a parte de executar a imagem sem Docker.

Ele pode:

- baixar uma imagem Docker/OCI diretamente do GHCR;
- escolher a variante ARM64;
- reconstruir as camadas da imagem;
- armazenar o filesystem no Android;
- preservar `ENTRYPOINT`, `CMD`, `ENV` e `WORKDIR`;
- executar o comando original da imagem;
- funcionar sem root;
- funcionar sem módulo de kernel;
- funcionar sem Docker daemon.

Referência:

- https://github.com/termux/proot-distro

A primeira tentativa seria:

```bash
pkg update
pkg install proot-distro
```

Depois:

```bash
proot-distro install   ghcr.io/m1k1o/neko/firefox:3.1.0   --name neko-firefox
```

E:

```bash
proot-distro run neko-firefox
```

O PRoot-Distro também aceita imagens armazenadas em arquivos OCI ou TAR exportados por ferramentas como `docker save`.

---

## O ponto de incerteza

Essa é a melhor opção arquitetural, mas ainda é experimental no POCO F6.

A documentação oficial do Neko recomenda Docker e não oferece uma receita pronta para:

```text
Neko v3
+
imagem oficial
+
PRoot-Distro
+
Android
```

Portanto, não é correto afirmar:

> É só executar e certamente funcionará.

A conclusão correta é:

> É o caminho com melhor relação entre adequação ao problema, reutilização de software pronto e possibilidade real de execução no Android sem Docker.

As falhas mais prováveis estarão no runtime gráfico:

```text
Xorg
PulseAudio
DBus
/tmp
/dev/shm
sockets
permissões
supervisord
drivers virtuais
```

O PRoot fornece um userland Linux, mas não reproduz completamente:

- namespaces;
- cgroups;
- seccomp;
- mounts reais;
- alguns recursos de IPC;
- acesso direto a dispositivos gráficos.

---

## O resultado provavelmente precisará de uma imagem derivada

O fluxo mais realista é:

```text
Imagem oficial Neko Firefox
          ↓
Primeiro teste no PRoot
          ↓
Identificar falhas
          ↓
Criar pequena imagem derivada
          ↓
Publicar neko-termux-firefox no GHCR
```

Exemplo:

```dockerfile
FROM ghcr.io/m1k1o/neko/firefox:3.1.0

COPY start-proot.sh /usr/local/bin/start-proot
COPY neko.yaml /etc/neko/neko.yaml
COPY policies.json /usr/lib/firefox/distribution/policies.json

ENTRYPOINT ["/usr/local/bin/start-proot"]
```

O script corrigiria apenas o necessário:

```text
diretórios temporários
permissões
PulseAudio
DBus
Xorg
logs
perfil persistente
encerramento correto
```

No celular:

```bash
proot-distro install   ghcr.io/ericklucioh/neko-termux-firefox:0.1.0   --name neko
```

Depois:

```bash
proot-distro run --detach neko
```

---

## Acesso externo: Tailscale

Para o João acessar de outra rede, a melhor opção é Tailscale.

Não é recomendável expor diretamente as portas do celular na internet.

O Neko permite configurar o endereço usado pelo WebRTC dentro de uma VPN privada.

O Tailscale conecta celular e computador em uma rede privada, mesmo quando:

- estão em redes diferentes;
- estão atrás de NAT;
- mudam de Wi-Fi;
- usam internet móvel;
- não possuem IP público acessível.

Referência:

- https://tailscale.com/kb/1017/install

Exemplo:

```text
POCO F6:
100.x.y.10

PC do João:
100.x.y.20
```

Configuração conceitual do Neko:

```bash
NEKO_SERVER_BIND=0.0.0.0:8080
NEKO_WEBRTC_EPR=52000-52100
NEKO_WEBRTC_ICELITE=1
NEKO_WEBRTC_NAT1TO1=100.x.y.10
```

João abriria:

```text
http://100.x.y.10:8080
```

Ou, usando MagicDNS:

```text
http://poco-f6:8080
```

Assim, somente dispositivos autorizados na rede Tailscale conseguem acessar o serviço.

---

## Perfil persistente

O perfil do Firefox deve ficar fora da rootfs:

```text
~/.local/share/neko-termux/profile
```

Montado dentro do ambiente em:

```text
/home/neko/.mozilla/firefox/profile.default
```

Isso permite preservar:

- logins;
- cookies;
- histórico;
- favoritos;
- extensões;
- preferências;
- abas;
- sessão anterior.

Também é necessário substituir a política padrão do Neko, porque ela pode limpar cookies e histórico quando o navegador fecha.

A política deve desabilitar a limpeza e restaurar a sessão anterior.

Assim:

```text
Neko reinicia
PRoot reinicia
imagem é atualizada
celular reinicia
        ↓
perfil continua
        ↓
logins continuam
```

---

## A decisão em uma frase

> **A melhor solução é executar o Neko Firefox ARM64 dentro do PRoot-Distro no Termux, adaptar a imagem oficial somente onde o PRoot exigir e acessar externamente pelo Tailscale.**

Não escolheria como solução principal:

```text
Puter Browser
iframe/proxy web
desktop completo com VNC
noVNC
Xpra
BrowserBox
```

Xpra permanece como plano B:

```text
Debian PRoot
└── Firefox
    └── Xpra HTML5
```

Ele é mais genérico e talvez seja mais simples de adaptar caso o runtime do Neko não funcione.

Porém, para a experiência final de navegador remoto com vídeo, áudio, login persistente e acesso pelo navegador do João, o Neko é a melhor escolha.

---

## Ordem de execução

```text
1. Testar a imagem oficial Neko Firefox pelo PRoot-Distro.
2. Registrar o primeiro erro real.
3. Corrigir por configuração, bind mount ou variável.
4. Criar uma imagem derivada somente quando necessário.
5. Persistir o perfil do Firefox fora da imagem.
6. Criar o launcher neko-termux.
7. Conectar o POCO e o PC do João pelo Tailscale.
8. Testar páginas comuns.
9. Testar login persistente.
10. Testar áudio e YouTube em 480p e 720p.
```

O objetivo final seria:

No POCO:

```bash
neko-termux start
```

No PC do João:

```text
http://poco-f6:8080
```

---

## Resumo da decisão

```text
Navegador:
Firefox ARM64

Servidor remoto:
Neko

Runtime sem Docker:
PRoot-Distro

Ambiente:
Termux no POCO F6

Acesso externo:
Tailscale

Persistência:
perfil do Firefox fora da rootfs

Plano B:
Xpra HTML5
```
