# Mobdesk

[Português](README.pt-BR.md) | [English](README.md)

> **MVP / experimental:** o Mobdesk funciona e já foi testado em um aparelho
> Android real. A validação em uma matriz maior de dispositivos ainda está em
> andamento. Use-o para estudo, desenvolvimento e serviços locais leves, não
> para cargas de produção.

## Sua workstation Termux no bolso

O Mobdesk transforma um celular Android em uma workstation Termux pessoal:

```text
Android
  Termux -> Mobdesk -> shell local ou SSH na porta 8022
```

O Termux é tanto a camada de integração com o Android quanto o único ambiente
de desenvolvimento. O workspace nativo permanece no celular. Você pode
trabalhar localmente ou conectar outro computador por uma rede confiável. O
Mobdesk não exige root, máquina virtual, Docker no celular, systemd ou desktop
gráfico.

## O que está disponível

- setup repetível do Termux, SSH, rede e workspace;
- servidor SSH dedicado do Mobdesk na porta `8022`;
- acesso nativo ao shell local do Termux com `mobdesk shell`;
- TUI com suporte a toque, mouse e teclado;
- status e saída JSON para automação;
- perfis de instalação idempotente pelo `pkg` do Termux para Git, Neovim,
  tmux, Go, Python, Node.js, C/C++, Lua, GitHub CLI, tree, htop, ncdu e Micro;
- atualizações do binário com rollback;
- apresentação em inglês (`en-US`) e português do Brasil (`pt-BR`).

Projetos, sessões persistentes, serviços e uma interface web permanecem nos
próximos estágios do roadmap.

A configuração de aplicativos, incluindo Neovim/LazyVim, está adiada para o
primeiro sprint. Ferramentas JVM gerenciadas e Spring Boot também não fazem
parte do escopo atual do sprint.

## Requisitos

- celular Android ARM64;
- Termux pelo [F-Droid](https://f-droid.org/packages/com.termux/) ou pelos
  [releases oficiais](https://github.com/termux/termux-app/releases);
- espaço para Termux, projetos e ferramentas instaladas; o Mobdesk avisa abaixo
  de 20 GB livres e bloqueia novas instalações abaixo de 10 GB;
- rede local confiável para acesso SSH remoto.

## Instalação

O método recomendado usa o binário ARM64 publicado e não exige Go. O último
binário estável está disponível na
[página de releases](https://github.com/ericklucioh/mobdesk/releases).

### Binário ARM64 publicado

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

O checksum verifica a integridade. Os releases ainda não possuem assinatura
independente, portanto o checksum não autentica sua origem.

### Instalações PRoot existentes

O primeiro sprint somente com Termux não migra instalações do PRoot-Distro ou
Ubuntu. Faça backup do que precisar, execute um reset completo do Termux e
instale o Mobdesk novamente. Não tente atualização ou migração no local.

### Compilar com Go

O projeto exige Go `1.26.5` ou mais recente:

```bash
pkg update
pkg install -y golang git
go version
go install github.com/ericklucioh/mobdesk/cmd/mobdesk@latest
~/go/bin/mobdesk setup
```

`@latest` significa a última tag estável com versionamento semântico. Não
significa um commit sem tag nem um prerelease `test-v*`. Use `@v0.6.0`, ou outra
tag explícita, quando a instalação precisar ser reproduzível.

## Primeira execução

O setup instala os pacotes necessários no Termux, cria o workspace nativo
persistente, configura o SSH e instala o launcher `mobdesk`. Depois da primeira
execução:

```bash
mobdesk status
mobdesk start
mobdesk shell
mobdesk stop
```

`mobdesk start` inicia o SSH e exibe os dados de conexão. Use `mobdesk shell`
para acesso local nativo ao Termux ou o comando SSH exibido a partir de outro
computador. As sessões SSH usam a mesma workstation e o mesmo workspace do
Termux.

## TUI e idioma

Execute `mobdesk tui` no Termux. `Tab` muda o foco, `Enter` ativa uma ação,
`Esc` volta e `q` inicia a confirmação de saída. A mesma TUI pode ser executada
por SSH e usa a mesma workstation Termux, não um ambiente separado de Ubuntu
ou PRoot.

Inglês é o idioma padrão. Selecione português do Brasil com:

```bash
mobdesk tui --locale pt-BR
MOBDESK_LOCALE=pt-BR mobdesk tui
```

A TUI ainda não possui um botão interno de idioma; reinicie-a com o locale
desejado.

## Segurança e limitações

Use SSH apenas em redes confiáveis ou por um túnel privado. Nunca exponha a
porta `8022` diretamente à internet. O MVP atual usa autenticação por senha e
escuta na rede local.

O Android pode suspender o Termux, e o projeto não foi criado para cargas
pesadas de produção, Docker real, systemd, módulos do kernel ou acesso
privilegiado a dispositivos.

## Mais informações

O [README da raiz](../README.md) contém a documentação técnica completa.

- [Missão](../docs/MISSION.md)
- [Arquitetura](../docs/ARCHITECTURE.md)
- [Roadmap](../docs/ROADMAP.md)
- [Changelog](../CHANGELOG.md)
- [Contribuição](CONTRIBUTING.pt-BR.md)
- [Código de Conduta](../CODE_OF_CONDUCT.md)
- [Suporte](SUPPORT.pt-BR.md)
- [Política de segurança](../SECURITY.pt-BR.md)

## Licença

O Mobdesk é distribuído sob a [licença MIT](../LICENSE).
