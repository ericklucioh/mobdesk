# Goal: Padronizar a apresentação visual dos apps

## Status

- Status: in_progress
- Project root: /home/erick/code/projs/mobdesk
- Base branch: main
- Branch policy: uma branch para o Goal inteiro
- Worktree policy: usar o worktree atual
- PR policy: um PR ao final do Goal
- Commit policy: um commit por Stage

## Objetivo

Padronizar a apresentação dos apps na tela normal da TUI, exibindo metadados úteis e consistentes sem renderizar a saída completa de `--help`.

A instalação automática das dependências será preservada.

## Escopo

- Adicionar metadados editoriais padronizados, incluindo uso.
- Exibir nome, descrição, estado, versão, uso, dependências, configuração e armazenamento quando aplicável.
- Corrigir o tratamento de versão para impedir que `--help` apareça como versão.
- Padronizar TTT, Neovim, Yazi e os demais apps do catálogo.
- Manter instalação automática das dependências existentes.
- Garantir o padrão por structs, contratos, comentários, documentação e testes.
- Preservar ações, confirmação, mouse, teclado e terminais estreitos.

## Fora do escopo

- Alterar a instalação automática de dependências.
- Criar tela de “Mais detalhes”.
- Exibir package, caminho do executável ou comandos internos na tela normal.
- Criar descoberta dinâmica de metadados via rede.
- Criar novos apps.
- Alterar `site/` ou `docs/LAUNCH-KIT.md`.

## Regras de execução

- O catálogo é a fonte declarativa dos metadados dos apps.
- A TUI não executa comandos Ubuntu diretamente.
- `VersionArg` não pode resultar em help completo renderizado como versão.
- Metadados editoriais e dados técnicos devem possuir campos distintos.
- A instalação automática atual de `Requires` deve permanecer funcionando.
- Ações destrutivas exigem confirmação.
- Toda ação importante funciona por mouse/toque e teclado.
- A tela deve caber em terminais estreitos.
- Novos campos JSON devem ser aditivos e compatíveis.
- Executar `make check` antes da conclusão.

## Tasks

### Task 1 - Padronizar o contrato de metadados

- Status: completed
- Depends on: none
- Branch: branch única do Goal
- Worktree: worktree atual
- PR: incluída no PR final

#### Stage 1.1 - Adicionar metadados editoriais

- Status: completed

##### Objetivo

Permitir que cada app declare informações próprias para apresentação.

##### Ações

- Adicionar `Usage` ao `AppProfile`.
- Definir a fonte da versão curta quando o app não possui um comando confiável.
- Documentar a diferença entre descrição, uso, versão, comando de verificação e dados técnicos de instalação.
- Preencher os metadados de TTT, Neovim e Yazi.
- Preservar `Requires`, `ConfigProfile` e `StorageEstimate`.

##### Critérios de aceite

- [x] Todos os apps possuem descrição e uso válidos.
- [x] TTT possui uso `ttt [arquivos/pastas/URLs...]`.
- [x] Neovim possui uso `nvim [arquivo ou diretório]`.
- [x] Yazi possui uso `yazi [diretório]`.
- [x] Nenhum app usa help completo como versão visual.
- [x] A instalação automática das dependências continua inalterada.

##### Validação

```bash
go test ./internal/install ./internal/i18n
go vet ./internal/install ./internal/i18n
```

Resultado: aprovado; os dois comandos passaram e `git diff --check` não encontrou erros.

Commit: 8ec67d6

##### Commit

```text
feat: standardize app metadata
```

### Task 2 - Implementar o popup normal compacto

- Status: in_progress
- Depends on: Task 1
- Branch: branch única do Goal
- Worktree: worktree atual
- PR: incluída no PR final

#### Stage 2.1 - Padronizar rendering e ações

- Status: completed

##### Objetivo

Renderizar todos os apps com o padrão visual aprovado.

##### Ações

- Atualizar `renderAppPopup`.
- Exibir nome, descrição, estado, versão, uso e dependências quando aplicável.
- Exibir configuração resumida quando aplicável.
- Exibir armazenamento resumido quando relevante.
- Remover origem, paths, plugins e saída bruta de comandos.
- Mostrar `Reinstalar` para apps instalados.
- Manter `Desinstalar`, `Remover config`, `Instalar` e `Fechar` conforme o estado.
- Preservar confirmação, `Esc`, mouse e teclado.

##### Critérios de aceite

- [x] TTT corresponde ao layout aprovado.
- [x] Neovim mostra `Config LazyVim aplicada` quando aplicável.
- [x] Yazi mostra estado, uso e espaço total.
- [x] Apps sem configuração não mostram configuração vazia.
- [x] Apps sem dependências não mostram linha vazia.
- [x] Apps instalados mostram `Reinstalar`.
- [x] Nenhum popup mostra a saída completa de `--help`.
- [x] Nenhuma linha excede a largura do terminal.
- [x] Não existe tela secundária de detalhes.

##### Validação

```bash
go test ./internal/tui
go vet ./internal/tui
```

Resultado: aprovado; testes de TUI, i18n e install, vet, i18n-check e `git diff --check` passaram.

Commit: pendente

##### Commit

```text
feat: simplify app detail popup
```

#### Stage 2.2 - Atualizar localização

- Status: completed

##### Objetivo

Garantir textos consistentes em inglês e português.

##### Ações

- Adicionar mensagens para uso, reinstalação e armazenamento resumido.
- Atualizar labels e ações necessárias.
- Validar todos os IDs usados pelo popup.
- Manter textos curtos para terminais móveis.

##### Critérios de aceite

- [x] Inglês e português possuem todas as mensagens novas.
- [x] Nenhum texto aparece como tradução ausente.
- [x] Textos localizados cabem em terminais estreitos.
- [x] Hit-tests não dependem de textos não localizados.

##### Validação

```bash
./scripts/i18n-check.sh
go test ./internal/i18n ./internal/tui
```

Resultado: aprovado; `i18n-check`, testes localizados e `git diff --check` passaram.

Commit: pendente

##### Commit

```text
feat: localize standardized app popup
```

### Task 3 - Garantir o padrão contra regressões

- Status: pending
- Depends on: Task 2
- Branch: branch única do Goal
- Worktree: worktree atual
- PR: incluída no PR final

#### Stage 3.1 - Adicionar testes e documentação

##### Objetivo

Impedir que novos apps sejam cadastrados ou renderizados de forma inconsistente.

##### Ações

- Adicionar testes de campos obrigatórios do catálogo.
- Adicionar testes para impedir help completo como versão.
- Adicionar testes de TTT, Neovim e Yazi.
- Adicionar testes de mouse, teclado, confirmação e terminal estreito.
- Atualizar `docs/ARCHITECTURE.md`.
- Atualizar `docs/DECISIONS.md`.
- Atualizar `docs/ROADMAP.md` se necessário.
- Adicionar comentários concisos nos contratos e regras não óbvias.

##### Critérios de aceite

- [ ] Perfil sem uso ou descrição falha nos testes.
- [ ] Help completo não aparece na versão.
- [ ] A instalação automática de dependências continua coberta pelos testes existentes.
- [ ] Popup compacto possui cobertura de rendering.
- [ ] Mouse e teclado continuam equivalentes.
- [ ] `make check` passa.
- [ ] Documentação explica como cadastrar novos metadados.

##### Validação

```bash
make check
```

##### Commit

```text
test: enforce standardized app presentation
```

## Ordem de execução

1. Task 1 / Stage 1.1
2. Task 2 / Stage 2.1
3. Task 2 / Stage 2.2
4. Task 3 / Stage 3.1

## Bloqueios e decisões

- As Tasks são dependentes e serão executadas em uma única branch.
- A instalação automática de dependências permanece como está.
- O comportamento de `Requires` não será refatorado neste Goal.
- Alterações pré-existentes em `site/` e `docs/LAUNCH-KIT.md` ficam fora do Goal.
- A versão de apps sem comando curto confiável deve vir de metadado declarado no catálogo.

## Conclusão

O Goal termina quando TTT, Neovim, Yazi e os demais apps seguem o popup compacto aprovado, nenhum app renderiza help completo como versão, a instalação automática continua funcionando, os contratos e testes estão atualizados e `make check` passa.
