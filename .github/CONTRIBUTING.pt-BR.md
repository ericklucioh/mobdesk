# Contribuindo com o Mobdesk

Obrigado por considerar contribuir com o Mobdesk. O projeto está construindo
uma workstation de desenvolvimento Termux pequena e verificável para Android. O MVP já foi testado
em um aparelho Android real; a validação em uma matriz maior de dispositivos
continua em andamento.

## Antes de começar

Leia:

- o [README](../README.pt-BR.md) para instalação e uso;
- a [Missão](../docs/MISSION.md) para entender o problema e o valor do produto;
- a [Arquitetura](../docs/ARCHITECTURE.md) para entender a fronteira de
  execução somente com Termux;
- o [Roadmap](../docs/ROADMAP.md) para conhecer o escopo futuro;
- as [Decisões](../docs/DECISIONS.md) para respeitar as escolhas atuais;
- o [Código de Conduta](../CODE_OF_CONDUCT.md) para conhecer as expectativas da
  comunidade.

## Ambiente de desenvolvimento

Requisitos recomendados:

- Go `1.26.5` ou mais recente;
- Docker com Docker Compose;
- Git;
- terminal com suporte a TTY;
- Android/Termux para alterações que afetem a integração com o aparelho.

Prepare e execute o projeto com:

```bash
git clone https://github.com/ericklucioh/mobdesk.git
cd mobdesk
make build-image
make dev
```

Use `make shell` para abrir um shell separado no ambiente.

## Verificações obrigatórias

Antes de enviar uma alteração:

```bash
make check
```

Para alterações no Docker, execute também `docker compose config` e
`make build-image`. Para alterações em Termux, integração Android ou SSH, valide também no
Termux real. O Docker não reproduz permissões do Android, rede, comportamento
da bateria ou restrições do kernel. `make integration-test` valida o fluxo
descartável no Docker, mas não substitui o teste em um aparelho.

## Organização do código e regras

- `cmd/mobdesk/`: entrada do executável;
- `internal/cobra/`: comandos e roteamento da CLI;
- `internal/tui/`: telas e componentes Bubble Tea;
- `internal/status/`: coleta do estado do ambiente;
- `internal/install/`: instalação idempotente de ferramentas;
- `internal/update/`: consulta e aplicação de atualizações;
- `docs/`: missão, arquitetura, decisões, roadmap e planos técnicos.

Mantenha operações idempotentes, preserve os dados do usuário, mantenha todas
as operações do host no Termux, use cancelamento em processos longos, valide
entradas antes de formar comandos e nunca grave senhas, tokens ou chaves no
código ou nos logs. Mantenha textos voltados ao usuário nos catálogos de i18n.
Atualize a documentação quando o escopo ou a arquitetura mudar.

## Commits e pull requests

Use commits curtos e descritivos, preferencialmente no formato
`tipo: descrição curta`. Uma pull request deve explicar o problema, a mudança
de comportamento, os testes, os ambientes Termux, Android, SSH ou Docker afetados e
as limitações restantes. Não misture refatorações, mudanças de arquitetura e
correções não relacionadas.

## Escopo atual

O fluxo ativo do MVP é:

```text
Termux -> Mobdesk -> shell local ou SSH
```

O escopo atual inclui a CLI, a TUI, os perfis de ferramentas de desenvolvimento,
os perfis de configuração de aplicativos e o fluxo de atualização. Projetos,
serviços, sessões persistentes, workflows com Tailscale, interfaces web e
outras expansões permanecem para o futuro. Contribuições nessas áreas devem
seguir o roadmap ou atualizar explicitamente a decisão de escopo.

## Relatando problemas

Inclua o modelo do aparelho e a versão do Android, a origem e a versão do
Termux, a versão do Mobdesk, o comando executado, a saída completa do erro e se
o problema ocorreu no Termux, SSH ou Docker. Nunca publique senhas,
chaves privadas, tokens ou dados pessoais nos logs. Use a [política de
segurança](../SECURITY.pt-BR.md) e o e-mail privado indicado nela para
vulnerabilidades.

Veja também o [guia de contribuição em inglês](CONTRIBUTING.md).
