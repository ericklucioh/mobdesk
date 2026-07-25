# Pesquisa: navegador remoto no Mobdesk

**Data da pesquisa:** 2026-07-17  
**Status:** investigação; nenhuma alternativa foi aprovada para o MVP.

## Problema

O Mobdesk pode, em uma fase posterior, executar um navegador gráfico completo
dentro do Ubuntu via PRoot e disponibilizar sua interface para outro navegador
por uma porta local ou pela rede privada.

O objetivo não é expor o HTML das páginas visitadas. O objetivo é transmitir a
área de trabalho do Firefox, incluindo abas, barra de endereço, menus, teclado,
mouse e, quando possível, áudio.

Esse modelo precisa ser entendido como:

```text
HTTP       carrega a interface web
WebSocket  ou WebRTC transporta tela, áudio e eventos
X server   fornece a tela virtual para o Firefox
Firefox    executa o navegador no Ubuntu/PRoot
```

“Uma porta HTTP” normalmente significa uma URL inicial. O transporte
interativo continua usando WebSocket ou WebRTC. A exceção prática é configurar
um proxy ou multiplexação para concentrar os canais em uma única porta.

## Resumo das alternativas

| Alternativa | O que é | Encaixe no PRoot | Avaliação inicial |
| --- | --- | --- | --- |
| [Selkies](https://github.com/selkies-project/selkies) | Plataforma de desktop remoto HTML5 com WebSocket ou WebRTC | Promissor | Melhor candidato para o primeiro spike fora de Docker |
| [Neko](https://github.com/m1k1o/neko) | Navegador virtual pronto, focado em WebRTC | Experimental | Melhor experiência específica de navegador, mas Docker-first |
| [jlesage/docker-firefox](https://github.com/jlesage/docker-firefox) | Firefox empacotado com GUI web baseada em VNC | Baixo/médio | A alternativa simples que provavelmente motivou esta pesquisa |
| [noVNC](https://github.com/novnc/noVNC) | Cliente VNC HTML5, não um navegador completo | Alto como componente | Base simples; precisa ser combinado com Xvfb, Firefox e VNC |
| [Apache Guacamole](https://guacamole.apache.org/) | Gateway HTML5 para VNC, RDP e SSH | Médio | Útil para gateway geral, não ideal para vídeo/áudio do navegador |
| [Kasm Workspaces](https://www.kasmweb.com/) | Plataforma VDI e sessões Docker | Baixo | Grande demais para o telefone e incompatível com o modelo PRoot atual |

## Selkies

Selkies é uma plataforma de streaming de desktops Linux para navegador. O
cliente HTML5 funciona em navegadores modernos, inclusive Chromium, Firefox e
Safari. O projeto oferece transporte WebSocket por padrão e WebRTC como opção.

O ponto mais importante para o Mobdesk é que a documentação descreve uma
instalação como pacote Python e também distribui um tarball portátil ARM64. Isso
torna Selkies mais compatível com um processo executado diretamente dentro do
Ubuntu/PRoot do que uma solução cuja instalação pressupõe Docker.

O projeto também possui um container mínimo de referência com Xfce4 e Firefox,
mas esse container é apenas uma forma conveniente de começar. A execução fora
de container ainda precisa ser validada com Xvfb, Firefox, GStreamer e um
servidor X compatível.

### Vantagens

- documenta execução standalone;
- suporta ARM64/aarch64;
- WebSocket usa uma única porta TCP por padrão;
- WebRTC pode ser habilitado quando a latência justificar;
- é adequado para transmitir um desktop inteiro, não apenas uma aplicação;
- permite evoluir depois para outras aplicações gráficas.

### Riscos

- ainda exige Xorg/Xvfb, GStreamer e um pipeline de captura/encode;
- a configuração é mais geral e complexa que a de um navegador pronto;
- aceleração gráfica não deve ser presumida no Android;
- a compatibilidade específica com PRoot e HyperOS não está comprovada;
- o consumo de CPU pode ser alto mesmo sem GPU.

### Papel proposto

Selkies deve ser o primeiro candidato do spike `browser-remote`, usando
resolução baixa, Firefox e Xvfb. O teste inicial deve preferir WebSocket para
evitar a complexidade de STUN/TURN e das faixas UDP do WebRTC.

## Neko

Neko é um navegador virtual self-hosted com interface web e WebRTC. Suas
imagens oficiais incluem Firefox e possuem variantes multi-arquitetura no GHCR.
O Neko usa X Server e Openbox para a área de trabalho, PulseAudio para áudio e
Pion para a comunicação WebRTC.

### Vantagens

- experiência pronta e específica para navegador;
- Firefox já empacotado;
- áudio e vídeo via WebRTC;
- controle remoto, autenticação e reconexão;
- suporte a perfil persistente do navegador;
- Firefox não exige as capacidades adicionais que algumas imagens Chromium
  exigem.

### Riscos

- a instalação oficialmente recomendada é Docker;
- a documentação do projeto considera outros métodos fora do escopo;
- o exemplo padrão usa HTTP e uma faixa de portas UDP para WebRTC;
- rodar a imagem extraída manualmente no PRoot seria uma adaptação não
  suportada, não uma instalação oficial;
- requer validar Firefox, X, PulseAudio, GStreamer, WebRTC e consumo de RAM no
  aparelho real.

### Papel proposto

Neko permanece como segunda implementação do spike. Ele deve voltar a ser
priorizado se o Mobdesk ganhar um runtime com Docker/Podman real, uma VM ou um
host Linux completo. No PRoot atual, não deve ser tratado como dependência do
núcleo.

## jlesage/docker-firefox

Este é provavelmente o projeto “VNC simples só para o browser” mencionado na
discussão. Ele fornece um container Firefox com a interface gráfica completa
acessível por navegador moderno, sem instalar um cliente no dispositivo. O
exemplo expõe a GUI web na porta `5800` e persiste estado, configuração e logs
em `/config`.

Ele também oferece áudio web, clipboard, gerenciador de arquivos, terminal web,
autenticação e conexão por cliente VNC. Conceitualmente, é mais simples de
entender que Neko:

```text
Firefox + desktop virtual + VNC + websockify/noVNC + HTTP
```

### Vantagens

- uma porta HTTP para a GUI (`5800` no exemplo);
- Firefox, perfil persistente e desktop já integrados;
- pode ser acessado por navegador móvel;
- oferece cliente VNC direto em uma porta separada;
- modelo fácil de incorporar como serviço externo no Mobdesk.

### Riscos

- é Docker-first;
- não é uma solução standalone documentada para PRoot;
- VNC/WebSocket tende a ser menos eficiente que WebRTC para vídeo contínuo;
- áudio, autenticação e proxy reverso exigem configuração adicional;
- o projeto é um empacotamento comunitário, não uma distribuição oficial do
  Firefox.

### Papel proposto

Usar como referência de simplicidade e como possível baseline de comparação.
Não assumir que a imagem Docker possa ser executada dentro do PRoot apenas por
ser baseada em Linux. Para o Mobdesk, a variante equivalente seria montar
manualmente os componentes: Firefox, Xvfb, um servidor VNC e noVNC/websockify.

## noVNC

noVNC não executa o navegador e não é um servidor VNC. Ele é o cliente HTML5
que renderiza uma sessão VNC no navegador do usuário. Para o Mobdesk, precisaria
ser combinado com:

```text
Xvfb       tela virtual
Firefox    navegador
x11vnc     ou outro servidor VNC
websockify ponte WebSocket/TCP
noVNC      interface web
```

O projeto continua ativo e suporta navegadores móveis, gestos de toque,
clipboard, escala e várias codificações VNC. O script `novnc_proxy` pode
servir a interface web e criar a ponte WebSocket para um VNC existente.

### Vantagens

- componentes conhecidos e relativamente independentes;
- pode funcionar sem Docker;
- normalmente concentra a interface em uma porta TCP;
- mais fácil de diagnosticar que uma cadeia WebRTC completa;
- bom candidato para provar primeiro que Firefox + Xvfb funcionam no PRoot.

### Riscos

- não foi desenhado como streaming de vídeo de baixa latência;
- áudio não faz parte do núcleo VNC/noVNC;
- qualidade e consumo dependem muito do servidor VNC e da codificação;
- exige montar e supervisionar vários processos;
- continua dependendo de Firefox e servidor X funcionando no PRoot.

### Papel proposto

noVNC deve ser o baseline técnico do primeiro teste. Se Firefox + Xvfb + VNC
forem estáveis no POCO, o Mobdesk poderá comparar noVNC com Selkies usando a
mesma resolução e o mesmo cenário de navegação.

## Apache Guacamole

Guacamole é um gateway HTML5 que acessa servidores VNC, RDP e SSH através de
uma aplicação web e do daemon `guacd`. Ele não substitui o Firefox nem cria
uma área de trabalho sozinho.

É interessante para uma futura central de acesso do Mobdesk, mas adiciona uma
camada de gateway Java/`guacd` que não é necessária para o primeiro navegador
remoto. A própria arquitetura do projeto separa cliente web, aplicação,
`guacd` e o servidor remoto.

## Kasm Workspaces

Kasm entrega desktops e aplicações no navegador, mas é uma plataforma VDI
orientada a Docker. A documentação atual indica, como mínimo, 2 CPUs, 4 GB de
RAM e 50 GB de armazenamento, além de exigir Docker para os serviços e sessões.

Isso o torna inadequado para o Mobdesk no POCO/PRoot, embora seja uma referência
útil para funcionalidades futuras como autenticação, provisionamento e gestão
de sessões.

## Decisão de pesquisa

Para o ambiente atual do Mobdesk:

1. **Baseline:** Firefox + Xvfb + servidor VNC + noVNC.
2. **Candidato principal após o baseline:** Selkies standalone.
3. **Candidato de melhor UX, se houver runtime real de containers:** Neko.
4. **Não adotar agora:** Kasm Workspaces e Apache Guacamole como núcleo.

O teste precisa medir no POCO F6:

- tempo para iniciar e reconectar;
- memória do Firefox, Xvfb e encoder;
- CPU durante navegação e vídeo;
- latência de toque, teclado e rolagem;
- estabilidade do perfil persistente;
- áudio, clipboard e upload/download;
- comportamento com Termux em segundo plano;
- necessidade de manter `termux-wake-lock`;
- exposição apenas em localhost, LAN ou Tailscale.

## Fontes consultadas

- [Selkies — repositório oficial](https://github.com/selkies-project/selkies)
- [Selkies — início e instalação standalone](https://selkies-project.github.io/selkies/start/)
- [Neko — repositório oficial](https://github.com/m1k1o/neko)
- [Neko — instalação](https://neko.m1k1o.net/docs/v3/installation)
- [Neko — quick start e requisitos](https://neko.m1k1o.net/docs/v3/quick-start)
- [Neko — imagens Docker](https://neko.m1k1o.net/docs/v3/installation/docker-images)
- [noVNC — repositório oficial](https://github.com/novnc/noVNC)
- [jlesage/docker-firefox — repositório](https://github.com/jlesage/docker-firefox)
- [Apache Guacamole — arquitetura](https://guacamole.apache.org/doc/1.5.2/gug/guacamole-architecture.html)
- [Kasm — requisitos do sistema](https://www.kasmweb.com/docs/latest/install/system_requirements.html)
- [Termux PRoot-Distro — limitações](https://github.com/termux/proot-distro)
