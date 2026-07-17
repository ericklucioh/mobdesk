# Decisões do Mobdesk

Este documento registra decisões atuais, alternativas futuras e hipóteses ainda não validadas.

## Decisões atuais

### Ubuntu via PRoot é o ambiente principal

O projeto prioriza compatibilidade com ferramentas Linux tradicionais. O Ubuntu persistente fornece `apt`, `glibc`, caminhos convencionais e maior compatibilidade de dependências.

### Termux é o host

O Termux controla a integração com Android, rede, inicialização, wake-lock, SSH, PRoot e Termux:API. Ele não é o ambiente principal de desenvolvimento do usuário.

### Go é a linguagem da central

Go é adequado para o binário único, execução de processos, concorrência, logs, diagnóstico, distribuição ARM64 e manutenção de uma CLI/TUI.

### A TUI vem antes da interface web

O primeiro produto deve funcionar por terminal. Uma interface web ou APK só deve ser criada depois que a instalação e o fluxo de trabalho estiverem comprovados.

### A instalação deve ser guiada

O usuário não deve precisar dominar `pkg`, `proot-distro`, `apt`, mounts ou scripts. O fluxo desejado é instalar o Mobdesk, executar `mobdesk start` e selecionar ferramentas numa TUI.

### O ambiente deve ser persistente

Ubuntu, ferramentas, projetos e configurações não devem ser recriados em cada execução. O instalador deve ser idempotente, ter estado e preservar dados.

## Alternativas adiadas

### Termux nativo como runtime principal

É mais rápido e integrado ao Android, mas pode exigir adaptações para binários glibc, wheels manylinux, ferramentas com caminhos Linux tradicionais e dependências nativas. Pode continuar disponível para ferramentas específicas.

### Nix-on-Droid

É interessante para configuração declarativa, ambientes reproduzíveis, rollback e sincronização de dotfiles. Não entra no núcleo porque aumenta complexidade, armazenamento e esforço de diagnóstico.

### Desktop gráfico

X11, VNC e desktops completos podem oferecer uma experiência visual, mas aumentam consumo, latência e pontos de falha. A alternativa preferida é TUI e, depois, aplicações web individuais.

### Neko

Neko pode ser útil para navegador remoto com Firefox, áudio e WebRTC, mas é uma linha experimental separada. Depende de validar Firefox, Xorg, PulseAudio, GStreamer, WebRTC e PRoot no aparelho real.

### Pesquisa de navegador remoto

Selkies, Neko, noVNC e `jlesage/docker-firefox` foram comparados em
[`PESQUISA_NAVEGADOR_REMOTO.md`](PESQUISA_NAVEGADOR_REMOTO.md). Para o PRoot
atual, o baseline será Firefox + Xvfb + VNC + noVNC; Selkies standalone é o
primeiro candidato para streaming mais eficiente. Neko permanece como opção de
melhor experiência quando houver um runtime real de containers ou quando sua
execução fora de Docker for validada no aparelho.

### Docker e VM

PRoot não é Docker. Docker real e recursos de kernel exigem outro suporte de kernel, root, VM ou dispositivo apropriado. Isso está fora do objetivo educacional do MVP.

## Hipóteses a validar

- Ubuntu ARM64 via PRoot terá desempenho aceitável para aulas e projetos pequenos;
- a instalação completa caberá confortavelmente no armazenamento do POCO;
- HyperOS permitirá manter Termux e Ubuntu ativos durante o uso;
- a TUI será confortável no teclado virtual e no SSH;
- as ferramentas escolhidas estarão disponíveis ou poderão ser instaladas no Ubuntu;
- a montagem entre Termux e Ubuntu preservará o fluxo de projetos;
- o Mobdesk conseguirá acompanhar processos e sessões sem perder saída;
- atalhos Termux:Widget serão suficientes antes de existir um APK próprio.

## Regras de segurança

- não expor SSH diretamente na internet;
- preferir chaves SSH;
- usar Tailscale para acesso externo;
- manter servidores de desenvolvimento em localhost quando possível;
- preservar autenticação nas aplicações futuras;
- não gravar senhas no repositório;
- pedir confirmação antes de apagar Ubuntu, perfil ou projetos;
- fazer backup fora do celular.
