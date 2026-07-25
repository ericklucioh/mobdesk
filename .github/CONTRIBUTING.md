# Contribuindo com o Mobdesk

Obrigado por considerar contribuir com o Mobdesk. O projeto está construindo uma workstation Ubuntu para Android, começando por um MVP pequeno e verificável.

## Antes de começar

Leia estes documentos:

- [README](../README.md) para instalar e executar o projeto;
- [Missão](../docs/MISSAO.md) para entender o problema que estamos resolvendo;
- [Arquitetura](../docs/ARQUITETURA.md) para entender a separação entre Termux e Ubuntu;
- [Roadmap](../docs/ROADMAP.md) para saber o que pertence ao futuro;
- [Decisões](../docs/DECISOES.md) para respeitar as escolhas atuais do projeto.

## Ambiente de desenvolvimento

Requisitos recomendados:

- Go `1.26.5`;
- Docker com Docker Compose;
- Git;
- terminal com suporte a TTY;
- Android/Termux para validação definitiva de integrações.

Prepare o ambiente:

```bash
git clone https://github.com/ericklucioh/mobdesk.git
cd mobdesk
make build-image
```

Durante o desenvolvimento:

```bash
make dev
```

O Air recompila o programa quando arquivos Go observados são alterados. Para abrir um shell separado:

```bash
make shell
```

## Verificações obrigatórias

Antes de enviar uma alteração:

```bash
make check
```

Quando a alteração envolver o ambiente Docker:

```bash
docker compose config
make build-image
```

Quando a alteração envolver Termux, SSH ou PRoot, valide também no Termux real. O Docker não reproduz completamente Android, permissões, rede, bateria ou restrições do kernel.

Para validar automaticamente o fluxo principal em um userland Termux descartável:

```bash
make integration-test
```

Esse smoke test não substitui a validação em um aparelho Android.

## Organização do código

- `cmd/mobdesk/`: entrada do executável;
- `internal/cobra/`: comandos e roteamento da CLI;
- `internal/tui/`: telas e componentes Bubble Tea;
- `internal/status/`: coleta e modelo do estado do ambiente;
- `internal/install/`: instalação idempotente de ferramentas;
- `internal/update/`: consulta e aplicação de atualizações;
- `docs/`: missão, decisões, arquitetura, roadmap e planos técnicos.

Introduza novos pacotes apenas quando houver comportamento real para organizar. Prefira a biblioteca padrão antes de adicionar dependências.

## Regras de implementação

- mantenha operações idempotentes;
- não remova Ubuntu, volumes ou projetos sem confirmação explícita;
- não misture comandos do Termux com comandos do Ubuntu;
- use contexto e cancelamento em processos longos;
- não bloqueie a TUI com instalação ou diagnóstico;
- valide entradas antes de montar comandos;
- não grave senhas, tokens ou chaves em código e logs;
- mantenha mensagens de erro acionáveis;
- atualize a documentação quando mudar escopo ou arquitetura.

## Commits e pull requests

Use commits curtos e descritivos, preferencialmente no formato:

```text
tipo: descrição curta
```

Exemplos:

```text
feat: adiciona diagnóstico do ambiente
fix: corrige detecção do ip no Termux
docs: atualiza instruções de instalação
test: cobre parser de endereços
```

Uma pull request deve explicar:

- qual problema resolve;
- qual comportamento foi alterado;
- como foi testada;
- se altera o ambiente Termux, Docker ou Ubuntu;
- quais limitações permanecem.

Não misture refatorações grandes, mudanças de arquitetura e correções não relacionadas na mesma alteração.

## Escopo atual

O MVP-1 concentra-se em:

```text
Termux → Mobdesk → SSH → Ubuntu via PRoot
```

TUI, ferramentas de desenvolvimento, tmux, Tailscale, projetos, serviços e interface web pertencem aos próximos estágios. Uma contribuição nessas áreas deve respeitar o roadmap ou atualizar explicitamente a decisão de escopo.

## Relatando problemas

Ao abrir uma issue, inclua:

- modelo do celular e versão do Android;
- origem e versão do Termux;
- versão do Mobdesk;
- comando executado;
- saída completa do erro;
- se o problema ocorreu no Termux, Ubuntu, SSH ou Docker.

Nunca publique senhas, chaves privadas, tokens ou dados pessoais nos logs.
