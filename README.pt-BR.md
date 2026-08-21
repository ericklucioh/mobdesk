# Mobdesk

Transforme seu celular Android em uma workstation Termux pessoal.

**[Abrir landing page](https://ericklucioh.github.io/mobdesk/)**

> **MVP / experimental:** o Mobdesk funciona e já foi testado em um aparelho
> Android real. A validação em uma matriz maior de dispositivos ainda está em
> andamento. Use-o para estudo, desenvolvimento e serviços locais leves, não
> para cargas de produção.

O Mobdesk usa o Termux tanto como camada de integração com o Android quanto como
o único ambiente de desenvolvimento:

```text
Android
  Termux -> Mobdesk -> shell local ou SSH na porta 8022
```

O projeto não exige root, máquina virtual, Docker no celular, systemd, desktop
gráfico ou módulos do kernel.

## O que está disponível

- setup repetível do Termux, SSH, rede e workspace;
- servidor SSH dedicado do Mobdesk na porta `8022`;
- acesso local ao Termux com `mobdesk shell`;
- saída humana e JSON para automação e para a TUI;
- instalação idempotente por `pkg` de Git, Neovim, tmux, Go, Python, Node.js,
  Clang/C++, Lua, GitHub CLI, tree, htop, ncdu e Micro;
- perfis privados selecionados para TUIFI, Bitwarden CLI e Resterm;
- fluxos offline de Git, Go, Python, Node/npm, C, C++, Lua, Neovim e tmux
  validados em Docker no workspace do Termux;
- telas de status, setup, ferramentas, shell, sistema e atualização na TUI;
- atualização verificável do binário com rollback e recuperação;
- apresentação em inglês (`en-US`) e português do Brasil (`pt-BR`).

Projetos, sessões persistentes, serviços e uma interface web permanecem fora do
MVP atual e pertencem aos próximos estágios do roadmap.

A configuração de aplicativos, incluindo Neovim/LazyVim, está adiada para o
primeiro sprint. Ferramentas JVM gerenciadas e Spring Boot também não fazem parte
do escopo atual do sprint.

## Requisitos

- celular Android com arquitetura ARM64;
- Termux de uma fonte confiável, preferencialmente o
  [F-Droid](https://f-droid.org/packages/com.termux/) ou os
  [releases oficiais](https://github.com/termux/termux-app/releases);
- espaço para Termux, projetos e ferramentas instaladas; o Mobdesk avisa
  abaixo de 20 GB livres e bloqueia novas instalações abaixo de 10 GB;
- rede local confiável se você for conectar outro computador.

O Mobdesk não exige root. O desempenho e a estabilidade dependem da memória,
temperatura e bateria do aparelho e das políticas de processos em segundo
plano do Android/HyperOS.

## Instalação

Instale o Termux por uma fonte confiável antes de escolher um dos métodos
abaixo. O método recomendado usa o binário ARM64 publicado e não exige Go.

### Opção 1: binário ARM64 publicado

O último binário estável está disponível pela
[página de releases](https://github.com/ericklucioh/mobdesk/releases). No
Termux, execute:

```bash
pkg update
pkg install -y curl coreutils

BASE_URL="https://github.com/ericklucioh/mobdesk/releases/latest/download"
mkdir -p "$HOME/.local/bin"
cd "$HOME/.local/bin"
curl -fL -o mobdesk-linux-arm64 "${BASE_URL}/mobdesk-linux-arm64"
curl -fL -o SHA256SUMS "${BASE_URL}/SHA256SUMS"
sha256sum -c SHA256SUMS
mv mobdesk-linux-arm64 mobdesk
chmod 0755 mobdesk
"$HOME/.local/bin/mobdesk" setup
```

O checksum verifica a integridade do arquivo. Os releases ainda não possuem
assinatura, portanto essa verificação não autentica independentemente a origem
do release. O primeiro setup instala os pacotes necessários no Termux, cria o
workspace persistente, configura o SSH e instala o launcher `mobdesk`.

### Instalações PRoot existentes

O primeiro sprint somente com Termux não migra instalações do PRoot-Distro ou
Ubuntu. Faça backup do que precisar, execute um reset completo do Termux e
instale o Mobdesk novamente. Não tente atualização ou migração no local.

### Opção 2: compilar com Go

Use este método quando quiser compilar a partir da última tag estável do
módulo Go. O projeto exige Go `1.26.5` ou mais recente:

```bash
pkg update
pkg install -y golang git
go version
go install github.com/ericklucioh/mobdesk/cmd/mobdesk@latest
~/go/bin/mobdesk setup
```

`@latest` significa a última tag estável com versionamento semântico; não
significa o último commit sem tag nem um prerelease `test-v*`. Para uma
instalação reproduzível, fixe a versão explicitamente, por exemplo:

```bash
go install github.com/ericklucioh/mobdesk/cmd/mobdesk@v0.6.0
```

Executar o setup por `~/go/bin/mobdesk` na primeira vez é intencional. O setup
cria o launcher do Termux em `$PREFIX/bin/mobdesk`, permitindo usar `mobdesk`
normalmente nos comandos seguintes. O setup pode ser executado novamente após
uma interrupção sem apagar o workspace.

## Uso básico

Inicie e consulte a workstation:

```bash
mobdesk status
mobdesk start
```

`mobdesk start` inicia o servidor SSH e exibe os dados de conexão. Para abrir o
Termux localmente sem SSH, use:

```bash
mobdesk shell
```

Para conectar de outro computador na mesma rede confiável, use o comando SSH
mostrado pelo Mobdesk, por exemplo:

```bash
ssh -p 8022 android@192.168.1.50
```

Pare o servidor SSH quando terminar:

```bash
mobdesk stop
```

Use `mobdesk status --json` para automação. Use `mobdesk logs --name <name>`
para ler um log limitado de instalação quando uma ferramenta falhar.

## Segurança e uso em rede

O MVP atual usa autenticação por senha e escuta na rede local para permitir SSH
a partir de outro computador. Use-o apenas em uma rede confiável ou por meio de
um túnel privado. Nunca exponha a porta `8022` diretamente à internet. Mantenha
backups dos projetos importantes fora do celular.

As atualizações verificam a integridade SHA-256, mas a autenticidade dos
releases ainda não é validada por assinatura independente. Trate os binários
publicados como experimentais até a assinatura de releases ser implementada.

Veja a [política de segurança](SECURITY.pt-BR.md) para relatos privados de
vulnerabilidades.

## TUI

Execute `mobdesk tui` no Termux. A TUI oferece ações visíveis para status, setup,
ferramentas, shell e atualizações do sistema. As ações importantes funcionam
com toque/mouse e teclado. `Tab` muda o foco, `Enter` ativa uma ação, `Esc`
volta e `q` inicia a confirmação de saída.

A TUI também pode ser aberta por uma sessão SSH. Essa sessão usa a mesma
workstation Termux, não um ambiente separado de Ubuntu ou PRoot.

### Idioma

O Mobdesk usa inglês (`en-US`) por padrão e oferece suporte a português do
Brasil (`pt-BR`). Escolha o idioma ao iniciar a CLI ou a TUI:

```bash
mobdesk tui --locale pt-BR
mobdesk --locale en-US status
```

Também é possível definir o idioma com `MOBDESK_LOCALE`:

```bash
MOBDESK_LOCALE=pt-BR mobdesk tui
```

O idioma selecionado é repassado para as operações da TUI e para os comandos
filhos. A TUI ainda não possui um botão interno para trocar o idioma; reinicie
a TUI com o idioma desejado. Identificadores técnicos, como comandos, flags,
chaves JSON e estados, permanecem em inglês.

## Limitações

- Docker, systemd, módulos do kernel e acesso privilegiado a dispositivos não
  estão disponíveis pelo Mobdesk;
- o Android pode suspender ou encerrar o Termux; quando apropriado, isente o
  Termux da otimização de bateria;
- a validação em Docker não reproduz permissões do Android, comportamento da
  bateria, wake-lock, suspensão de processos pelo HyperOS ou todos os aparelhos
  ARM64;
- o projeto não foi criado para cargas pesadas de produção nem exposição
  pública do SSH.

## Desenvolvimento

```bash
git clone https://github.com/ericklucioh/mobdesk.git
cd mobdesk
make build-image
make check
```

Antes de contribuir, leia o [guia de contribuição](.github/CONTRIBUTING.pt-BR.md).
O repositório usa Docker para verificações reproduzíveis, mas alterações em
Termux/SSH também exigem validação no Termux real quando houver um aparelho
disponível.

## Documentação e comunidade

- [Landing page e publicação no GitHub Pages](docs/GITHUB-PAGES.md)
- [Missão](docs/MISSION.md)
- [Arquitetura](docs/ARCHITECTURE.md)
- [Decisões](docs/DECISIONS.md)
- [Roadmap](docs/ROADMAP.md)
- [Changelog](CHANGELOG.md)
- [Contribuição](.github/CONTRIBUTING.pt-BR.md)
- [Código de Conduta](CODE_OF_CONDUCT.md)
- [Suporte](.github/SUPPORT.pt-BR.md)
- [Política de segurança](SECURITY.pt-BR.md)
- [README em inglês](README.md)

## Licença

O Mobdesk é distribuído sob a [licença MIT](LICENSE).
