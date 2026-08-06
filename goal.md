# Goal: Otimizar o catalog-test

## Status

- Status: pending
- Project root: /home/erick/code/projs/mobdesk
- Base branch: main
- Branch policy: feat/optimize-catalog-test (uma branch para o Goal inteiro)
- Worktree policy: worktree atual, após criação da branch
- PR policy: um PR ao final do Goal
- Commit policy: um commit por Stage

## Objetivo

Reduzir significativamente o tempo de execução do `catalog-test` sem remover
a cobertura completa de instalação, verificação, idempotência e execução dos
aplicativos do catálogo.

O projeto terá uma validação rápida para o ciclo normal de desenvolvimento e
uma validação completa para execução manual ou periódica. A validação completa
continuará exercitando todos os perfis reais e downloads externos.

## Resultado esperado

- `catalog-test-fast` executa em tempo compatível com o ciclo normal de PR.
- `catalog-test-full` preserva a cobertura completa existente.
- O teste identifica claramente qual ferramenta consumiu tempo.
- Verificações redundantes de PRoot são reduzidas.
- Caches de downloads relevantes são reutilizados entre execuções.
- O comportamento de instalação do produto não muda.
- A redução de tempo não depende de paralelizar operações incompatíveis com
  APT, DPKG ou o lock global do Mobdesk.

## Escopo

- Medir o tempo individual de cada instalação e etapa do catálogo.
- Criar modo rápido para validação representativa por estratégia de instalação.
- Manter modo completo para todos os aplicativos do catálogo.
- Reduzir chamadas redundantes a `proot-distro login`.
- Agrupar verificações finais em poucas sessões Ubuntu.
- Reutilizar caches seguros de APT, Go, npm e pip quando aplicável.
- Manter validação de dependências compartilhadas.
- Manter validação de idempotência.
- Documentar quando usar cada modo.
- Atualizar `Makefile`, scripts de teste, Docker Compose e documentação
  relacionada.

## Fora do escopo

- Alterar o comportamento de instalação em produção.
- Criar um novo comando público de instalação em lote.
- Remover aplicativos do catálogo.
- Remover a validação completa de qualquer aplicativo no modo full.
- Paralelizar instalações que compartilham APT, DPKG ou o lock do Mobdesk.
- Substituir instalações reais por mocks no `catalog-test-full`.
- Ignorar checksums, versões fixadas ou validações de executáveis.
- Alterar o contrato JSON ou a TUI.
- Executar `catalog-test` durante a definição deste Goal.
- Adicionar dependência de secrets.
- Alterar o ambiente real do dispositivo Termux fora dos fixtures Docker.

## Regras de execução

- O modo fast deve validar pelo menos uma ferramenta de cada estratégia:
  APT, script/download, Go, npm e pipx.
- O modo fast deve validar pelo menos uma dependência compartilhada.
- O modo fast deve validar pelo menos uma ferramenta com configuração.
- O modo fast deve validar uma instalação repetida.
- O modo full deve continuar instalando todos os perfis atualmente cobertos.
- O modo full deve continuar verificando os executáveis instalados.
- O modo full deve continuar validando downloads, checksums e versões reais.
- Instalações devem permanecer seriais quando houver lock compartilhado.
- Caches persistentes não podem mascarar a ausência de instalação no Ubuntu.
- A fixture deve continuar reproduzível quando os caches forem removidos.
- O teste deve continuar funcionando em `TERMUX_ARCH=latest` e ARM64 quando
  o perfil exigir arquitetura específica.
- Falhas devem identificar a ferramenta e a etapa responsável.
- O tempo medido deve ser escrito de forma legível sem poluir contratos JSON.
- Nenhuma alteração de produção será feita apenas para acelerar o teste.

## Tasks

### Task 1 - Medir o custo real do catalog-test

- Status: pending
- Depends on: none
- Branch: feat/optimize-catalog-test
- Worktree: worktree atual
- PR: incluída no PR final

#### Stage 1.1 - Instrumentar etapas e ferramentas

- Status: pending

##### Objetivo

Obter uma decomposição objetiva do tempo gasto por build, provisionamento,
instalação, verificação e segunda passada.

##### Ações

- Medir a duração do build da imagem.
- Medir a duração do provisionamento da fixture Ubuntu.
- Medir cada chamada `mobdesk install`.
- Medir a segunda instalação de cada ferramenta.
- Medir as verificações de versão e fixtures de compilação.
- Identificar falhas e timeouts por etapa.
- Manter a saída resumida e legível no final do teste.

##### Critérios de aceite

- [ ] Cada instalação do catálogo aparece com duração individual.
- [ ] A primeira e a segunda passadas são distinguíveis.
- [ ] O resumo identifica as cinco etapas mais lentas.
- [ ] A instrumentação não altera o resultado de sucesso ou falha.
- [ ] A saída não contém secrets.
- [ ] O timeout continua encerrando o teste corretamente.

##### Validação

```bash
make catalog-test
make integration-test
git diff --check
```

##### Commit

```text
test: measure catalog test stages
```

### Task 2 - Criar o modo catalog-test-fast

- Status: pending
- Depends on: Task 1
- Branch: feat/optimize-catalog-test
- Worktree: worktree atual
- PR: incluída no PR final

#### Stage 2.1 - Definir matriz rápida por estratégia

- Status: pending

##### Objetivo

Validar as principais estratégias de instalação sem reinstalar todos os
aplicativos a cada ciclo de desenvolvimento.

##### Ações

- Definir um conjunto mínimo de perfis representativos:
  - APT;
  - script com download e checksum;
  - Go;
  - npm;
  - pipx;
  - ferramenta com configuração;
  - dependência compartilhada.
- Incluir Java ou outra dependência compartilhada necessária aos perfis.
- Preservar pelo menos uma ferramenta com múltiplos executáveis.
- Preservar pelo menos uma ferramenta com binário em `~/.local/bin`.
- Preservar pelo menos uma ferramenta com validação de configuração.
- Criar alvo explícito `catalog-test-fast`.
- Permitir que o script selecione o modo sem duplicar toda a lógica.

##### Critérios de aceite

- [ ] O modo fast executa todos os grupos de estratégia definidos.
- [ ] O modo fast valida instalação inicial e instalação repetida.
- [ ] O modo fast valida versões e executáveis.
- [ ] O modo fast valida pelo menos uma dependência compartilhada.
- [ ] O modo fast não instala todos os perfis do catálogo.
- [ ] O alvo fast é separado e explícito no `Makefile`.
- [ ] A documentação explica que fast não substitui full.

##### Validação

```bash
make catalog-test-fast
make check
git diff --check
```

##### Commit

```text
test: add fast catalog validation mode
```

#### Stage 2.2 - Agrupar sessões de verificação Ubuntu

- Status: pending

##### Objetivo

Reduzir o overhead de inicializar PRoot repetidamente durante as verificações
que não modificam o ambiente.

##### Ações

- Consolidar verificações de versão compatíveis em uma sessão
  `proot-distro login ubuntu`.
- Consolidar verificações de executáveis relacionados.
- Evitar chamadas repetidas de `uname -m`.
- Manter comandos de instalação separados quando necessário para preservar
  locks, logs e diagnósticos.
- Preservar mensagens de erro específicas por ferramenta.

##### Critérios de aceite

- [ ] O número de sessões PRoot de verificação diminui.
- [ ] Cada executável obrigatório continua sendo validado.
- [ ] Uma falha identifica o executável ou ferramenta responsável.
- [ ] Instalações não são paralelizadas nem agrupadas de forma insegura.
- [ ] O modo fast continua compilando e executando suas fixtures.
- [ ] O modo full continua verificando todos os aplicativos.

##### Validação

```bash
make catalog-test-fast
make catalog-test-full
git diff --check
```

##### Commit

```text
test: consolidate Ubuntu catalog checks
```

### Task 3 - Reutilizar caches sem perder reprodutibilidade

- Status: pending
- Depends on: Task 1
- Branch: feat/optimize-catalog-test
- Worktree: worktree atual
- PR: incluída no PR final

#### Stage 3.1 - Persistir caches de downloads

- Status: pending

##### Objetivo

Evitar downloads repetidos entre execuções sem persistir o estado instalado
dos aplicativos.

##### Ações

- Avaliar e adicionar volumes para caches seguros de APT.
- Avaliar cache npm utilizado pelos perfis npm.
- Avaliar cache pip/pipx utilizado pelo perfil pipx.
- Preservar os volumes Go existentes.
- Garantir que caches sejam montados apenas nos caminhos de cache.
- Não montar o rootfs Ubuntu inteiro como volume persistente.
- Não transformar uma execução anterior em pré-condição para uma execução
  limpa.

##### Critérios de aceite

- [ ] Segunda execução com caches reutiliza downloads disponíveis.
- [ ] Execução sem caches continua funcionando.
- [ ] Remover caches não deixa ferramentas falsamente instaladas.
- [ ] Os caches não incluem estado privado ou secrets.
- [ ] O modo full continua executando instalações reais.
- [ ] A documentação identifica os caches persistentes.

##### Validação

```bash
make catalog-test-fast
make catalog-test-full
docker compose config
git diff --check
```

##### Commit

```text
test: persist catalog download caches
```

#### Stage 3.2 - Validar execução limpa e cacheada

- Status: pending

##### Objetivo

Comprovar que a otimização de cache melhora a segunda execução sem quebrar
a reprodutibilidade da primeira.

##### Ações

- Validar execução sem volumes de cache.
- Validar execução com volumes de cache.
- Comparar tempos das duas execuções.
- Confirmar que a fixture Ubuntu continua correta nos dois casos.
- Registrar o ganho observado e qualquer custo residual.

##### Critérios de aceite

- [ ] Execução limpa passa.
- [ ] Execução cacheada passa.
- [ ] Execução cacheada não depende de estado instalado anterior.
- [ ] O resumo registra a diferença de duração.
- [ ] Falhas de cache produzem fallback ou diagnóstico objetivo.

##### Validação

```bash
make catalog-test-fast
make catalog-test-fast
make catalog-test-full
```

##### Commit

```text
test: verify clean and cached catalog runs
```

### Task 4 - Preservar cobertura completa no modo full

- Status: pending
- Depends on: Task 2, Task 3
- Branch: feat/optimize-catalog-test
- Worktree: worktree atual
- PR: incluída no PR final

#### Stage 4.1 - Separar full da matriz rápida

- Status: pending

##### Objetivo

Garantir que o modo full continue sendo a autoridade para cobertura real de
todo o catálogo.

##### Ações

- Criar alvo explícito `catalog-test-full`.
- Mover a lista completa de perfis para o modo full.
- Manter a segunda instalação dos perfis necessários para idempotência.
- Manter verificações reais de cada executável.
- Manter a validação condicional de `zellij` em ARM64.
- Manter as fixtures Java e Kotlin existentes.
- Evitar duplicar código entre fast e full.

##### Critérios de aceite

- [ ] Todos os perfis atualmente cobertos continuam no modo full.
- [ ] Todos os executáveis atualmente verificados continuam sendo verificados.
- [ ] Checksums e downloads reais continuam sendo executados.
- [ ] Idempotência continua coberta.
- [ ] O resultado do full diferencia instalação, verificação e fixture.
- [ ] O full não depende de o fast ter sido executado antes.

##### Validação

```bash
make catalog-test-full
make check
git diff --check
```

##### Commit

```text
test: preserve complete catalog validation
```

#### Stage 4.2 - Validar cobertura contra o catálogo declarativo

- Status: pending

##### Objetivo

Evitar que novos perfis sejam adicionados ao catálogo sem decisão explícita
sobre sua cobertura live.

##### Ações

- Comparar os perfis declarados em `internal/install/install.go` com as listas
  de fast e full.
- Definir como o teste reage a um novo perfil sem cobertura.
- Preferir uma verificação automática para o modo full.
- Registrar exceções arquiteturais como `zellij`.

##### Critérios de aceite

- [ ] O modo full cobre todos os perfis declarados ou falha com diagnóstico.
- [ ] O modo fast documenta seus perfis representativos.
- [ ] Perfis condicionais por arquitetura são tratados explicitamente.
- [ ] A verificação não exige rede.
- [ ] A manutenção futura do catálogo não depende de memória manual.

##### Validação

```bash
make test
make check
make catalog-test-fast
```

##### Commit

```text
test: verify catalog coverage declarations
```

### Task 5 - Documentar uso e concluir o Goal

- Status: pending
- Depends on: Task 4
- Branch: feat/optimize-catalog-test
- Worktree: worktree atual
- PR: incluída no PR final

#### Stage 5.1 - Documentar os modos e resultados

- Status: pending

##### Objetivo

Tornar claro quando executar cada validação e quais garantias cada uma
oferece.

##### Ações

- Documentar `catalog-test-fast`.
- Documentar `catalog-test-full`.
- Documentar caches persistentes e execução limpa.
- Documentar que full pode ser lento por downloads externos e PRoot.
- Documentar ausência de paralelismo nas instalações.
- Registrar tempos observados após a otimização.
- Atualizar documentação de desenvolvimento e decisões quando necessário.

##### Critérios de aceite

- [ ] README ou documentação de desenvolvimento explica os dois modos.
- [ ] O `Makefile` descreve corretamente os alvos.
- [ ] A diferença de cobertura está explícita.
- [ ] O uso dos caches está documentado.
- [ ] O tempo observado e as limitações externas estão registrados.
- [ ] `i18n-check` passa.

##### Validação

```bash
./scripts/i18n-check.sh
make check
git diff --check
```

##### Commit

```text
docs: document catalog test modes
```

#### Stage 5.2 - Validação final e fechamento

- Status: pending

##### Objetivo

Confirmar o ganho de tempo, a preservação de cobertura e a ausência de
regressões.

##### Ações

- Executar `catalog-test-fast`.
- Executar `catalog-test-full`.
- Executar `make check`.
- Executar `make integration-test`.
- Comparar os tempos instrumentados com a linha de base.
- Revisar o diff completo.
- Atualizar este Goal com resultados reais.
- Validar o branch antes do PR.

##### Critérios de aceite

- [ ] `catalog-test-fast` passa.
- [ ] `catalog-test-full` passa.
- [ ] `make check` passa.
- [ ] `make integration-test` passa.
- [ ] O fast é substancialmente mais rápido que o full.
- [ ] A cobertura full não foi reduzida.
- [ ] Instalações e verificações continuam idempotentes.
- [ ] `git diff --check` passa.
- [ ] O worktree contém apenas alterações relacionadas ao Goal.

##### Validação

```bash
make catalog-test-fast
make catalog-test-full
make check
make integration-test
git status --short
```

##### Commit

```text
test: finalize catalog test optimization
```

## Ordem de execução

1. Task 1 / Stage 1.1
2. Task 2 / Stage 2.1
3. Task 2 / Stage 2.2
4. Task 3 / Stage 3.1
5. Task 3 / Stage 3.2
6. Task 4 / Stage 4.1
7. Task 4 / Stage 4.2
8. Task 5 / Stage 5.1
9. Task 5 / Stage 5.2

## Bloqueios e decisões

- O `catalog-test` completo não será executado durante a definição deste
  Goal.
- O modo full continua sendo obrigatório para preservar a cobertura real.
- O modo fast é uma validação adicional, não uma substituição silenciosa.
- Instalações APT, DPKG e operações protegidas pelo lock do Mobdesk permanecem
  seriais.
- Ainda não está decidido se os caches APT, npm e pip produzirão ganho
  relevante; a medição da Task 1 orienta a implementação.
- O ganho mínimo aceitável do modo fast será definido após a linha de base.
- O ambiente real Termux/ARM64 continua sendo a validação final de integração.

## Conclusão

O Goal termina quando existirem modos fast e full claramente separados,
o modo full preservar a cobertura completa do catálogo, os caches forem
validados sem comprometer reprodutibilidade, os tempos forem medidos e
documentados, e todas as validações finais passarem.
