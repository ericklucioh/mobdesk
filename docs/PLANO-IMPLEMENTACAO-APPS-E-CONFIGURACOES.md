# Plano de Implementação de Apps e Configurações

**Status:** Fase 11 concluida no fixture automatizado; validacao manual no Termux real pendente

**Documento de decisões:** [`PLANO-REFATORACAO-APPS-E-CONFIGURACOES.md`](PLANO-REFATORACAO-APPS-E-CONFIGURACOES.md)

**Objetivo:** implementar a nova central de apps do Mobdesk com popup de
detalhes, instalação, desinstalação segura, configurações opinativas opcionais,
estimativas de armazenamento e suporte inicial para Neovim/LazyVim.

## 1. Regra deste Plano

Este documento transforma as decisões do produto em tarefas executáveis. As
alternativas já descartadas não fazem parte do escopo.

Decisões obrigatórias:

- O modelo do catálogo será `AppProfile`.
- Neovim será adicionado ao catálogo instalável.
- LazyVim será a única configuração opinativa da primeira entrega.
- Neovim deverá estar instalado antes de `config apply`.
- Configuração existente causará conflito e não será sobrescrita.
- O Mobdesk não criará backup automático no MVP.
- A remoção excluirá somente plugins e arquivos comprovadamente gerenciados.
- Dependências nunca serão removidas automaticamente no MVP.
- O registro de configuração será separado em `state/configurations/<app>.json`.
- O contrato JSON continuará no schema 1 com campos opcionais aditivos.
- A popup será sempre aberta ao tocar em uma linha de app.
- Confirmações destrutivas ocorrerão dentro da popup.
- Apps detectados serão mostrados como instalados, mas não serão considerados gerenciados automaticamente.
- Apps sem proveniência não poderão ser desinstalados pela TUI ou pela CLI.
- A documentação será atualizada junto com cada fase.

## 2. Resultado Final

O usuário deverá conseguir:

1. Abrir a tela `Apps e linguagens`.
2. Tocar ou selecionar um app.
3. Abrir uma popup com nome, descrição, estado e armazenamento estimado.
4. Instalar o app pelo botão da popup.
5. Acompanhar o progresso da instalação.
6. Verificar que o app está instalado.
7. Aplicar a configuração Mobdesk somente quando decidir fazê-lo.
8. Ver a configuração, plugins, caminhos e armazenamento estimado.
9. Remover a configuração sem apagar alterações manuais.
10. Desinstalar o app somente quando o Mobdesk tiver proveniência suficiente.
11. Receber uma explicação quando a operação não for segura ou possível.

O fluxo deverá funcionar no Termux host. Em uma TUI aberta dentro de uma sessão
SSH no Ubuntu, operações de host deverão continuar bloqueadas com mensagem
explicativa.

## 3. Fora do Escopo

Não implementar nesta entrega:

- Segundo perfil de configuração além de Neovim/LazyVim.
- Backup automático de configuração existente.
- Mesclagem automática de dotfiles.
- Remoção automática de dependências compartilhadas.
- Gerenciamento de configuração do usuário fora dos perfis Mobdesk.
- Interface web ou APK.
- Sincronização de configurações com nuvem.
- Marketplace remoto de perfis.
- Atualização automática de todos os apps instalados.
- Reinstalação forçada sem confirmação.

## 4. Estado Atual do Código

### 4.1 Catálogo e instalação

O catálogo efetivo está em `internal/install/install.go`. O tipo atual é
`Language`, embora o catálogo já contenha linguagens, ferramentas, clientes de
IA e gerenciadores de arquivos.

A instalação atual:

- Resolve nomes e aliases.
- Resolve dependências recursivamente.
- Usa `apt`, `npm`, `pipx`, Go e scripts.
- Executa comandos no Ubuntu via `proot-distro`.
- Usa lock global de instalação.
- Persiste registros em `state/installations`.
- Persiste logs em `logs/install`.
- Verifica o executável depois da instalação.

### 4.2 Status

`internal/status` lê registros persistidos e também verifica executáveis no
Ubuntu. Hoje um executável detectado pode aparecer como instalado sem indicar se
foi instalado pelo Mobdesk.

### 4.3 CLI

`internal/cobra/install.go` expõe `mobdesk install`. O registro de comandos raiz
fica em `internal/cobra/config.go`. O contrato comum está em
`internal/cobra/json.go`.

### 4.4 TUI

`internal/tui/screen_tools.go` mostra os apps. `internal/tui/model.go` monta o
catálogo exibido. `internal/tui/tui.go` e `internal/tui/mouse.go` executam a
instalação ao confirmar ou tocar em uma linha.

A TUI já possui:

- Seleção por teclado.
- Mouse e toque.
- Modal de confirmação para sair e parar a workstation.
- Backend real que chama a própria CLI.
- Backend mock para testes visuais.
- Estado de operação e progresso de instalação.

## 5. Arquitetura Alvo

```text
TUI Bubble Tea
  -> Backend real ou mock
  -> CLI Cobra e contrato JSON
  -> serviços internos
      -> catálogo AppProfile
      -> instalação
      -> desinstalação
      -> configuração
      -> estado e proveniência
      -> status
  -> Ubuntu via proot-distro
```

Responsabilidades:

- `internal/install`: perfis, instalação, desinstalação, configuração e estado de apps.
- `internal/status`: fotografia do estado dos apps e configurações.
- `internal/cobra`: flags, runtime, JSON e saída humana.
- `internal/tui`: interação, popup, foco, confirmação e apresentação.
- `internal/paths`: caminhos persistentes canônicos.
- `scripts`: smoke tests reais do catálogo e do fluxo Termux/PRoot.

A TUI não deverá chamar `apt`, `pipx`, `npm`, `proot-distro` ou scripts de
configuração diretamente.

## 6. Ordem e Dependências das Fases

| Fase | Entrega | Depende de |
|---|---|---|
| 0 | Preparação e contrato congelado | Documento de decisões |
| 1 | `AppProfile` e catálogo | Fase 0 |
| 2 | Neovim instalável | Fase 1 |
| 3 | Estimativas de armazenamento | Fases 1 e 2 |
| 4 | Proveniência e estado | Fases 1 e 3 |
| 5 | Desinstalação segura | Fase 4 |
| 6 | Motor de configuração | Fases 2 e 4 |
| 7 | Perfil LazyVim | Fase 6 |
| 8 | CLI e contratos JSON | Fases 5, 6 e 7 |
| 9 | Status e reconciliação | Fases 4 e 8 |
| 10 | Popup da TUI | Fases 8 e 9 |
| 11 | Integração, documentação e validação | Fase 10 |

Uma fase só será considerada concluída depois que seus testes e critérios de
aceite passarem.

## 7. Fase 0: Preparação e Contrato Congelado

### Objetivo

Preparar a implementação sem alterar comportamento de produção e garantir que
os nomes, estados e caminhos estejam definidos de forma única.

### Arquivos

- `docs/PLANO-REFATORACAO-APPS-E-CONFIGURACOES.md`
- `docs/PLANO-IMPLEMENTACAO-APPS-E-CONFIGURACOES.md`
- `internal/install/model.go`
- `internal/status/model.go`
- `internal/cobra/json.go`

### Passos

1. Confirmar que o documento de decisões contém as escolhas consolidadas.
2. Usar `AppProfile` como nome do novo modelo.
3. Reservar `storage_estimate` como campo opcional do perfil e do resultado JSON.
4. Definir estados canônicos de app e configuração.
5. Definir que `source=detected` não implica proveniência de desinstalação.
6. Não modificar ainda a forma de instalação existente.
7. Criar testes de compilação para os novos tipos sem alterar o catálogo final.

### Critérios de aceite

- Não existem alternativas de produto abertas no plano.
- Os nomes de estado são usados de forma consistente.
- O schema 1 continua compatível com os campos atuais.
- `go test ./...` continua passando antes das mudanças funcionais.

### Resultado da Fase 0

- `AppProfile`, `StorageEstimate`, estados canônicos de app e configuração
  foram adicionados sem migrar o catálogo existente.
- `storage_estimate` foi reservado como campo JSON opcional e aditivo.
- `Language` e o fluxo de instalação atual permanecem compatíveis.
- Testes cobrem a compilação dos novos contratos e a compatibilidade do campo
  opcional no resultado JSON.

## 8. Fase 1: Modelo `AppProfile` e Catálogo

### Objetivo

Substituir o modelo conceitual de linguagem por um perfil de app sem perder os
perfis existentes.

### Arquivos

- `internal/install/model.go`
- `internal/install/install.go`
- `internal/install/catalog.go` se a separação reduzir o tamanho de `install.go`
- `internal/tui/model.go`
- `internal/install/install_test.go`
- `internal/tui/tui_test.go`

### Modelo esperado

O perfil deverá conter, no mínimo:

```text
AppProfile
  Name
  Aliases
  Description
  Kind
  Package
  Executable
  VersionArg
  InstallKind
  Requires
  UserBin
  InstallProfile
  UninstallProfile
  ConfigProfile
  ProfileVersion
  StorageEstimate
```

O tipo `StorageEstimate` deverá conter:

```text
StorageEstimate
  AppMinMB
  AppMaxMB
  DependenciesMinMB
  DependenciesMaxMB
  ConfigMinMB
  ConfigMaxMB
  Source
  Version
  Architecture
  MeasuredAt
```

### Passos

1. Criar `AppProfile` com os campos necessários.
2. Migrar o catálogo atual para `[]AppProfile`.
3. Atualizar `Resolve`, `Languages`, `Tools` e consumidores internos.
4. Manter aliases atuais funcionando.
5. Mover descrições que hoje vivem em `internal/tui/model.go` para o catálogo.
6. Manter `Tools` temporariamente como função de leitura, sem duplicar catálogo.
7. Adicionar descrições para todos os apps existentes.
8. Adicionar o campo de capacidade de configuração sem criar configuração para apps que não possuem perfil.
9. Adicionar estimativas iniciais para todos os perfis existentes.
10. Atualizar testes de resolução, aliases, tipos e descrições.

### Regras do catálogo

- Nomes canônicos devem ser minúsculos e estáveis.
- Aliases devem ser normalizados antes da resolução.
- O catálogo não deve aceitar comandos vindos do usuário.
- Caminhos de instalação devem ser declarados no perfil.
- Um app sem estratégia segura de remoção deve declarar essa limitação.
- Um app sem configuração deve informar `ConfigProfile` vazio.

### Critérios de aceite

- Todos os apps atuais continuam resolvíveis.
- Yazi e TUIFI continuam resolvíveis.
- A TUI não mantém frases duplicadas fora do catálogo.
- Os testes confirmam perfil, estratégia e estimativa de cada app.
- Nenhum comportamento de instalação existente é alterado sem teste correspondente.

### Resultado da Fase 1

- O catálogo agora usa `AppProfile` e `Languages`, `Tools` e `Resolve` retornam
  perfis de app.
- As descrições exibidas na TUI foram movidas para os perfis do catálogo.
- Todos os perfis atuais possuem estimativa inicial de armazenamento para Ubuntu
  ARM64, com origem de planejamento e versão do catálogo.
- Aliases existentes, incluindo Yazi e TUIFI Manager, continuam resolvíveis.
- O instalador continua usando as mesmas estratégias e comandos de instalação.

## 9. Fase 2: Neovim Instalável

### Objetivo

Adicionar Neovim como app instalável antes de implementar a configuração LazyVim.

### Arquivos

- `internal/install/install.go` ou `internal/install/catalog.go`
- `internal/install/install_test.go`
- `scripts/test-catalog.sh`
- `docs/PLANO-REFATORACAO-APPS-E-CONFIGURACOES.md`

### Perfil inicial

O primeiro perfil de Neovim deverá:

- Usar o pacote Ubuntu `neovim` como instalação inicial.
- Verificar o executável `nvim`.
- Usar `nvim --version` como verificação.
- Declarar a versão mínima suportada pelo perfil.
- Declarar que a configuração LazyVim é opcional.
- Declarar que a configuração é aplicada somente depois da instalação.
- Declarar o destino `/root/.config/nvim` dentro do Ubuntu.

### Passos

1. Confirmar a versão de Neovim disponível no Ubuntu fixture.
2. Definir `Name: "neovim"` e alias `nvim`.
3. Definir `Package: "neovim"` e `Executable: "nvim"`.
4. Definir `InstallKind: "apt"`.
5. Definir o perfil de configuração LazyVim.
6. Definir a estimativa de Neovim sem configuração.
7. Adicionar teste de resolução e comando de instalação esperado.
8. Adicionar Neovim ao smoke test do catálogo.
9. Executar instalação no Ubuntu via PRoot.
10. Verificar `nvim --version` dentro do Ubuntu.
11. Executar uma segunda instalação para provar idempotência.

### Critérios de aceite

- `mobdesk install neovim` instala no Ubuntu.
- `mobdesk install nvim` resolve o mesmo perfil.
- `proot-distro login ubuntu -- nvim --version` funciona.
- A segunda instalação não reinstala o pacote sem necessidade.
- A instalação não cria `~/.config/nvim` automaticamente.
- A popup informa que LazyVim ainda não foi aplicado.

### Resultado da Fase 2

- Neovim foi adicionado com alias `nvim`, pacote Ubuntu `neovim` e verificação
  `nvim --version`.
- O perfil declara a versão mínima, o perfil opcional `lazyvim` e o destino
  `/root/.config/nvim`, sem aplicar configuração durante a instalação.
- Testes cobrem resolução por nome e alias e o comando `apt` esperado.
- `scripts/test-catalog.sh` instala Neovim, verifica a versão, confirma que a
  configuração não foi criada e repete a instalação pelo alias.
- O fixture PRoot Docker validou `NVIM v0.11.6` em `x86_64`; a validação ARM64
  ainda depende do Termux/POCO F6 real.

## 10. Fase 3: Estimativas de Armazenamento

### Objetivo

Exibir uma estimativa útil antes de instalar um app ou aplicar configuração,
sem prometer precisão que o sistema não consegue garantir.

### Arquivos

- `internal/install/model.go`
- `internal/install/catalog.go`
- `internal/status/model.go`
- `internal/cobra/json.go`
- `internal/tui/model.go`
- `internal/tui/tool_list.go`
- `internal/tui/tui_test.go`
- `scripts/test-catalog.sh`

### Definição dos valores

- `app_mb`: tamanho do pacote, binário ou runtime principal.
- `dependencies_mb`: dependências que não estejam presentes no Ubuntu.
- `config_mb`: arquivos, plugins, caches e dados criados pela configuração.
- `total`: soma do intervalo mínimo e máximo do perfil isolado.

Dependências compartilhadas não deverão ser contadas novamente quando já
estiverem instaladas. O catálogo deverá manter a observação de que os valores
são estimativas e não somas independentes de todos os apps.

### Passos

1. Copiar as estimativas iniciais do documento de decisões para os perfis.
2. Medir o tamanho dos pacotes em Ubuntu ARM64 limpo.
3. Medir o tamanho dos binários após instalação.
4. Medir dependências adicionais de Yazi, TUIFI e Neovim.
5. Medir o tamanho de uma configuração LazyVim inicial.
6. Registrar versão, arquitetura, origem e data da medição.
7. Atualizar os intervalos quando a diferença for relevante.
8. Expor a estimativa no status e no resultado JSON quando necessário.
9. Renderizar os quatro valores na popup.
10. Mostrar aviso de armazenamento baixo sem iniciar a instalação silenciosamente.

### Fórmula apresentada na popup

```text
Total mínimo = app_min + dependencies_min + config_min
Total máximo = app_max + dependencies_max + config_max
```

### Critérios de aceite

- Todo `AppProfile` possui estimativa ou declara que ela ainda não foi medida.
- A popup exibe app, dependências, configuração e total.
- A estimativa não altera o estado de instalação.
- A estimativa não é confundida com espaço livre do dispositivo.
- Testes cobrem terminal estreito e valores longos.
- O smoke test registra a medição dos perfis principais.

### Resultado da Fase 3

- Totais mínimo e máximo são calculados a partir de app, dependências e
  configuração sem alterar o estado da instalação.
- Resultados de instalação e status incluem `storage_estimate` de forma
  opcional e compatível com o schema 1.
- Registros antigos e apps detectados recebem a estimativa do perfil quando
  existe correspondência no catálogo.
- O payload consumido pela TUI preserva a estimativa para a popup detalhada da
  Fase 10.
- O smoke test valida a estimativa em `install --json` e `status --json`.

## 11. Fase 4: Proveniência e Estado Persistente

### Objetivo

Registrar o que o Mobdesk instalou e separar o ciclo de vida do app do ciclo de
vida da configuração.

### Arquivos

- `internal/paths/paths.go`
- `internal/install/model.go`
- `internal/install/state.go`
- `internal/install/install.go`
- `internal/status/model.go`
- `internal/status/collect.go`
- `internal/install/install_test.go`
- `internal/status/collect_test.go`
- `internal/paths/paths_test.go`

### Caminhos

Adicionar à fonte canônica de paths:

```text
ConfigurationsDir() -> $HOME/.local/share/mobdesk/state/configurations
ConfigurationState(app) -> .../configurations/<app>.json
```

Não criar diretório de backup automático para configuração de app. O estado de
instalação atual e os logs existentes devem permanecer compatíveis.

### Registro de instalação

Ampliar o registro para guardar:

- Nome canônico.
- Pacote principal.
- Executável principal.
- Estratégia usada.
- Dependências resolvidas.
- Pacotes instalados pelo Mobdesk.
- Arquivos e diretórios criados pelo Mobdesk.
- Estado.
- Versão.
- Data da instalação.
- Última tentativa.
- Último erro.
- Log.
- Fonte `mobdesk` ou `detected`.

### Registro de configuração

Criar registro separado por app com:

- App relacionado.
- Perfil e versão do perfil.
- Estado da configuração.
- Caminhos gerenciados.
- Arquivos gerados.
- Plugins gerenciados.
- Hash esperado de cada arquivo.
- Data de aplicação.
- Data de remoção.
- Indicação de modificação manual.
- Conflitos encontrados.
- Último erro.

### Passos

1. Adicionar paths canônicos.
2. Criar structs persistíveis com JSON estável.
3. Usar escrita temporária e rename atômico.
4. Criar diretórios com permissão `0700`.
5. Criar arquivos de estado com permissão `0600`.
6. Registrar tentativa antes de iniciar uma operação mutável.
7. Atualizar o registro somente após cada etapa concluída.
8. Registrar estado `partial` em falha recuperável.
9. Ler registros antigos sem exigir campos novos.
10. Rejeitar nomes de app que possam escapar do diretório de estado.

### Critérios de aceite

- Estado de instalação e configuração fica separado.
- Repetição não perde registros válidos.
- Falha no rename não deixa estado final falso.
- Arquivos de estado têm permissões privadas.
- Testes usam diretório temporário.
- Registros antigos continuam legíveis.

### Resultado da Fase 4

- Registros de instalação persistem estratégia, dependências, pacotes, arquivos
  declarados e fonte `mobdesk`.
- Status normaliza registros antigos como gerenciados e mantém detecções futuras
  separadas como `source=detected` e não gerenciadas.
- Foi criado o registro separado `state/configurations/<app>.json` com schema
  estável para perfil, estado, caminhos, hashes, conflitos e erros.
- Estado de configuração usa diretório `0700`, arquivos `0600` e rename
  atômico, com rejeição de nomes que escapem do diretório privado.

## 12. Fase 5: Desinstalação Segura

### Objetivo

Implementar `Uninstall` sem remover dependências compartilhadas ou arquivos sem
proveniência comprovada.

### Arquivos

- `internal/install/uninstall.go`
- `internal/install/install.go`
- `internal/install/state.go`
- `internal/install/model.go`
- `internal/install/install_test.go`
- `internal/install/uninstall_test.go`

### Regras de proveniência

- Registro `mobdesk`: pode usar estratégia declarada.
- Registro `detected`: pode informar instalação encontrada, mas não pode remover.
- Sem registro: não remover automaticamente.
- Arquivo divergente do hash: preservar e marcar `modified`.
- Dependência compartilhada: preservar no MVP.

### Estratégias

Implementar estratégias explícitas por tipo:

- `apt`: remover somente o pacote principal declarado.
- `npm`: remover somente o pacote global declarado.
- `pipx`: remover somente o ambiente do app.
- `go`: remover somente o binário registrado.
- `script`: executar remoção somente se o perfil declarar arquivos próprios.
- `neovim`: remover o pacote somente quando a instalação tiver proveniência Mobdesk.
- `yazi`: remover `yazi` e `ya` somente quando os arquivos estiverem registrados.
- `tuifi`: remover o ambiente pipx somente quando criado pelo Mobdesk.

Dependências como Go, Python, Node, Clang, pipx e bibliotecas não serão removidas
automaticamente.

### Passos

1. Resolver o app e carregar seu registro.
2. Recusar desinstalação para estado `detected` sem proveniência.
3. Adquirir o mesmo lock usado pela instalação.
4. Marcar o registro como `uninstalling`.
5. Verificar dependências e arquivos registrados.
6. Verificar hashes quando a estratégia remover arquivos.
7. Preservar arquivos modificados.
8. Executar somente a estratégia do perfil.
9. Persistir itens removidos e preservados.
10. Marcar `uninstalled`, `modified`, `partial` ou `failed`.
11. Atualizar o log.
12. Nunca executar `autoremove` no MVP.

### Critérios de aceite

- App instalado pelo Mobdesk pode ser removido pela CLI.
- App apenas detectado não pode ser removido automaticamente.
- Dependências continuam instaladas.
- Arquivos modificados são preservados.
- Falha parcial é observável no estado e no JSON.
- Repetir a remoção é seguro.
- Não existe remoção genérica baseada apenas no nome do executável.

### Resultado da Fase 5

- `Uninstall` usa o mesmo lock da instalação e exige proveniência `mobdesk`.
- Registros apenas detectados, pacotes compartilhados e caminhos sem
  proveniência comprovada são recusados com segurança.
- Estratégias `apt`, `node`, `npm`, `pipx`, `script`, `go`, `ttt`, `cargo` e
  `gh-extension` foram separadas sem remover dependências automaticamente.
- Arquivos gerenciados persistem hashes; arquivos alterados são preservados e
  marcados como `modified`.
- Estado, arquivos removidos/preservados e falhas permanecem observáveis no
  registro persistente.

## 13. Fase 6: Motor de Configuração

### Objetivo

Criar operações declarativas e idempotentes de aplicar e remover configuração.

### Arquivos

- `internal/install/config.go`
- `internal/install/config_state.go` se a separação for necessária
- `internal/install/model.go`
- `internal/install/state.go`
- `internal/paths/paths.go`
- `internal/install/config_test.go`

### Perfil de configuração

O perfil deverá declarar:

- ID estável.
- Versão.
- App associado.
- Descrição.
- Caminhos de destino no Ubuntu.
- Arquivos gerados.
- Diretórios gerados.
- Plugins gerenciados.
- Comandos de validação declarados internamente.
- Política de conflito.
- Estratégia de aplicação.
- Estratégia de remoção.
- Estimativa de configuração.

### Política de conflito

Considerar conflito quando o destino existir e não estiver registrado como
gerenciado pelo próprio perfil. Isso inclui diretório existente, mesmo vazio,
quando não houver registro anterior da configuração.

Reaplicação é permitida quando:

- O registro pertence ao mesmo perfil.
- O estado anterior é conhecido.
- O arquivo atual corresponde ao hash gerado ou está marcado como gerenciado.

Não criar backup automático. O usuário deverá mover ou remover manualmente uma
configuração existente antes de tentar aplicar o perfil Mobdesk.

### Aplicação

1. Resolver app e perfil.
2. Confirmar instalação do app.
3. Confirmar que o perfil existe.
4. Validar caminhos contra o HOME do Ubuntu.
5. Inspecionar o destino.
6. Recusar conflito.
7. Persistir estado `applying`.
8. Criar somente diretórios declarados.
9. Escrever arquivos temporários.
10. Promover arquivos com rename atômico.
11. Instalar plugins declarados.
12. Registrar manifestos e hashes.
13. Executar validação do app.
14. Persistir estado `applied` somente após sucesso.

### Falha parcial

Quando uma aplicação falhar:

- Remover arquivos criados pela tentativa atual que ainda estejam intactos.
- Preservar qualquer arquivo que tenha sido alterado depois da criação.
- Registrar componentes removidos e preservados.
- Marcar estado `failed` ou `modified`.
- Não apagar conteúdo anterior do usuário.
- Não criar backup automático.

### Remoção

1. Carregar o registro de configuração.
2. Marcar estado `removing`.
3. Comparar hashes atuais com hashes registrados.
4. Remover arquivos intactos e gerenciados.
5. Remover plugins comprovadamente gerenciados.
6. Preservar arquivos alterados.
7. Preservar arquivos desconhecidos.
8. Marcar `removed` se tudo gerenciado foi removido.
9. Marcar `modified` se algo alterado foi preservado.
10. Persistir detalhes no registro e no JSON.

### Critérios de aceite

- Aplicar configuração inexistente funciona.
- Aplicar novamente é idempotente.
- Aplicar sobre configuração existente retorna conflito.
- Nenhum backup automático é criado.
- Remover preserva alterações manuais.
- Plugins gerenciados são removidos quando seus hashes permitem.
- Falha parcial não corrompe arquivos externos.
- O registro separado é atualizado corretamente.

### Resultado da Fase 6

- Criado motor declarativo `ApplyConfig`/`RemoveConfig` com perfis estáticos,
  caminhos gerenciados, arquivos, validações e estimativas.
- Aplicação exige app instalado, rejeita conflitos e é idempotente para o mesmo
  perfil e hashes conhecidos.
- Arquivos são escritos por dados base64 e rename no Ubuntu, sem concatenar
  conteúdo do usuário em shell.
- Estado `applying`, `applied`, `removing`, `removed`, `modified` e `failed` é
  persistido no registro separado.
- Remoção compara hashes e preserva arquivos alterados; falhas tentam remover
  somente componentes criados pela tentativa atual.

## 14. Fase 7: Perfil Neovim/LazyVim

### Objetivo

Entregar a configuração opinativa inicial do Mobdesk de forma versionada,
explícita, reproduzível e reversível dentro das regras escolhidas.

### Arquivos prováveis

- `internal/install/profiles/neovim/` para arquivos estáticos do perfil
- `internal/install/config.go`
- `internal/install/model.go`
- `internal/install/config_test.go`
- `internal/install/neovim_profile_test.go`
- `internal/tui/model.go`
- `docs/PLANO-REFATORACAO-APPS-E-CONFIGURACOES.md`

### Conteúdo do perfil

O perfil deverá declarar:

- Destino `/root/.config/nvim`.
- Arquivos Mobdesk gerenciados.
- Versão do bootstrap LazyVim.
- Versão do gerenciador de plugins.
- Lista de plugins adicionais do padrão Mobdesk.
- Arquivo de lock quando aplicável.
- Comando de validação `nvim --headless`.
- Estimativa de configuração.

Os arquivos de configuração deverão ficar versionados no repositório ou ser
embutidos no binário. O perfil não deverá baixar código sem versão fixa nem
executar conteúdo arbitrário retornado por uma rede.

### Passos

1. Adicionar o perfil de Neovim ao catálogo.
2. Confirmar que Neovim está instalado.
3. Criar os arquivos estáticos do padrão Mobdesk.
4. Declarar o bootstrap LazyVim e suas versões.
5. Definir plugins obrigatórios e opcionais do padrão.
6. Definir os diretórios de plugins gerenciados.
7. Detectar `/root/.config/nvim` antes da aplicação.
8. Recusar se o diretório não pertencer ao Mobdesk.
9. Criar a configuração declarada.
10. Instalar plugins durante `config apply`.
11. Validar com `nvim --headless`.
12. Registrar hashes e componentes.
13. Testar remoção dos arquivos e plugins gerenciados.
14. Testar preservação de arquivo modificado.

### Critérios de aceite

- Neovim instalado sem configuração continua utilizável.
- LazyVim só é aplicado após ação explícita.
- Configuração existente gera conflito claro.
- Aplicação gera o padrão definido pelo autor.
- Plugins são instalados durante `config apply`.
- Remoção não apaga arquivos modificados.
- Estado `modified` é exibido quando necessário.
- O tamanho de LazyVim aparece na popup.

### Resultado da Fase 7

- O perfil `lazyvim` é embutido no binário e contém `init.lua`, módulos Lua e
  `lazy-lock.json` versionados.
- O perfil declara `lazy.nvim`, LazyVim e nvim-treesitter por repositório HTTPS,
  commit de 40 caracteres e diretório gerenciado no Ubuntu.
- `config apply` clona, busca e faz checkout apenas das revisões declaradas;
  `config remove` remove somente plugins limpos pertencentes ao manifesto.
- A configuração continua opcional, exige Neovim instalado pelo Mobdesk e
  preserva conflitos existentes antes de qualquer escrita ou clone.
- Testes cobrem o manifesto, revisões fixas, aplicação, persistência e remoção
  dos plugins gerenciados. A validação headless do Neovim permanece pendente no
  dispositivo Termux/POCO F6 real.

## 15. Fase 8: CLI e Contratos JSON

### Objetivo

Expor as operações para a TUI e para usuários avançados sem duplicar regras.

### Arquivos

- `internal/cobra/config.go` para registrar comandos raiz
- `internal/cobra/uninstall.go`
- `internal/cobra/app_config.go`
- `internal/cobra/install.go`
- `internal/cobra/json.go`
- `internal/cobra/install_test.go`
- `internal/cobra/uninstall_test.go`
- `internal/cobra/config_test.go`

O nome `app_config.go` evita confundir o comando de configuração de apps com o
arquivo atual `internal/cobra/config.go`, que registra `RootCmd`.

### Comandos

Adicionar:

```text
mobdesk uninstall <app>
mobdesk config apply <app>
mobdesk config remove <app>
```

Todos deverão aceitar `--json` e `--progress` quando a operação for longa.

### Contrato

Adicionar campos opcionais ao resultado atual:

```text
target
action
changed
config_state
storage_estimate
conflicts
paths
source
```

Manter campos existentes como `language` durante a transição para não quebrar
consumidores atuais. Os novos comandos usarão `target`; instalação existente
continuará preenchendo `language` quando aplicável.

### Regras de erro

- `--json` sempre produz um JSON final válido.
- Falha de runtime Termux aparece como resultado estruturado.
- Falha de conflito não executa alteração parcial.
- Falha parcial informa estado e caminhos preservados.
- `stderr` fica reservado para mensagens auxiliares.
- Progresso não substitui o resultado final.

### Passos

1. Criar adaptador CLI para `Uninstall`.
2. Criar subcomando `config` com `apply` e `remove`.
3. Adicionar as operações ao `RootCmd`.
4. Implementar flags JSON e progresso.
5. Expandir `operationResult` com campos opcionais.
6. Adicionar mensagens humanas para cada estado.
7. Garantir JSON em sucesso, conflito, falha parcial e erro de runtime.
8. Testar argumentos exatos e bloqueio fora do Termux.

### Critérios de aceite

- Os quatro fluxos têm comando CLI funcional.
- A TUI pode chamar todos pelo mesmo backend.
- O schema continua compatível.
- Erros não corrompem stdout JSON.
- Operações destrutivas continuam exigindo confirmação na TUI; a CLI informa a operação sem criar confirmação interativa.

### Resultado da Fase 8

- A CLI agora expõe `uninstall`, `config apply` e `config remove`, todos com
  `--json` e `--progress`.
- Os resultados preservam o schema 1, mantêm `language` no install e adicionam
  `target`, `action`, `changed`, `config_state`, `source`, `paths`, `conflicts`
  e `storage_estimate` quando aplicáveis.
- Sucesso, conflito, falha parcial, app sem proveniência e runtime Ubuntu
  remoto retornam um resultado JSON final sem diagnóstico humano no stdout.
- O progresso usa eventos JSON separados do resultado final; a confirmação
  destrutiva continua sendo responsabilidade da TUI nas fases seguintes.
- Testes cobrem argumentos exatos, conversão do contrato, falha estruturada e
  bloqueio fora do Termux.

## 16. Fase 9: Status e Reconciliação

### Objetivo

Exibir corretamente instalação, detecção, proveniência e configuração aplicada.

### Arquivos

- `internal/status/model.go`
- `internal/status/collect.go`
- `internal/status/render.go`
- `internal/status/collect_test.go`
- `internal/tui/model.go`
- `internal/tui/messages.go`

### Modelo

Adicionar campos opcionais a `InstallationStatus`:

```text
source
managed
storage_estimate
config_state
```

Adicionar uma coleção separada quando necessário:

```text
ConfigurationStatus
  app
  profile
  state
  managed_paths
  modified_paths
  conflicts
```

### Reconciliação

1. Ler registros persistidos.
2. Detectar executáveis no Ubuntu.
3. Criar estado `installed` com `source=detected` quando o executável existir sem registro.
4. Preservar `source=mobdesk` quando houver registro válido.
5. Nunca transformar detecção em proveniência de remoção.
6. Ler registros de configuração separados.
7. Associar configuração ao app canônico.
8. Marcar configuração ausente, aplicada, modificada ou em conflito.
9. Atualizar alertas sem tratar app detectado como erro.

### Critérios de aceite

- Status mostra app instalado pelo Mobdesk.
- Status mostra app apenas detectado.
- Status mostra configuração separadamente.
- Status não oferece desinstalação de app sem proveniência.
- Estado antigo continua sendo lido.
- TUI recebe os campos sem lógica duplicada.

### Resultado da Fase 9

- `status --json` agora separa `installations` de `configurations` e associa os
  dois por nome canônico.
- Apps configuráveis sem registro aparecem como `not_applied`; caminhos
  existentes sem registro aparecem como `conflict`.
- Registros aplicados são reconciliados por hash quando a inspeção Ubuntu
  retorna um hash válido, produzindo `modified` e `modified_paths` sem apagar
  arquivos ou tratar falha de inspeção como conflito.
- A fonte `detected` continua não gerenciada e não habilita desinstalação; os
  estados de conflito, modificação e falha entram nos alertas do status.
- Testes cobrem associação app/configuração, estado ausente, hash divergente e
  serialização JSON compatível.

## 17. Fase 10: Popup da TUI

### Objetivo

Substituir a instalação direta ao toque por uma experiência segura de detalhes
e ações.

### Arquivos

- `internal/tui/model.go`
- `internal/tui/screen_tools.go`
- `internal/tui/tool_list.go`
- `internal/tui/tui.go`
- `internal/tui/mouse.go`
- `internal/tui/messages.go`
- `internal/tui/backend.go`
- `internal/tui/commands.go`
- `internal/tui/styles.go`
- `internal/tui/tui_test.go`

### Estado da TUI

Adicionar ao `Model`:

```text
appPopupOpen
popupAppIndex
popupFocus
popupAction
popupConfirm
popupMessage
```

O estado da popup não deverá substituir `busy`, `operationID` ou os IDs de
status existentes. A operação continuará passando pelo mecanismo de host atual.

### Fluxo de teclado

1. Em `toolsScreen`, `Enter` abre a popup do app selecionado.
2. `Tab` e `Shift+Tab` percorrem ações.
3. `Enter` executa a ação focada.
4. `Esc` fecha a popup sem executar operação.
5. `Y` ou `S` confirma ação destrutiva.
6. `N` ou `Esc` cancela confirmação.
7. Durante operação, ações ficam bloqueadas.
8. Após resultado, a TUI solicita novo status.

### Fluxo de mouse

1. Toque em qualquer ponto da linha abre a popup.
2. Toque em uma ação focada executa a ação.
3. Toque em desinstalar ou remover configuração abre confirmação na própria popup.
4. Toque fora da área de ação não inicia operação.
5. Em terminal remoto, o toque mostra a restrição do host.

### Conteúdo da popup

Renderizar:

- Nome.
- Descrição.
- Estado do app.
- Fonte `Mobdesk` ou `detectado`.
- Versão.
- Dependências.
- Configuração disponível ou não.
- Estado da configuração.
- Caminhos afetados.
- Plugins.
- App em MB.
- Dependências em MB.
- Configuração em MB.
- Total em MB.
- Ações permitidas.
- Mensagem de conflito ou indisponibilidade.

### Regras de ação

- App não instalado: habilitar somente `Instalar`.
- App instalado gerenciado: habilitar `Desinstalar`.
- App detectado: exibir instalado, bloquear `Desinstalar` e explicar.
- Configuração disponível e app instalado: habilitar `Adicionar configuração`.
- Configuração aplicada: habilitar `Remover configuração` com confirmação.
- Configuração em conflito: bloquear aplicação e mostrar motivo.
- App sem estratégia: desabilitar remoção e explicar.
- Sessão Ubuntu remota: bloquear todas as operações de host.

### Backend e mensagens

Atualizar:

- `realBackend.OperationCmd` para reconhecer novos comandos.
- `runCommand` para comandos JSON simples.
- `runInstallCommand` somente quando progresso de instalação for necessário.
- `operationMessage` com ação, alvo, estado e configuração.
- `operationMessageText` para mensagens de instalação, remoção e configuração.
- Mock backend para simular estados de sucesso, conflito, falha e remoção parcial.

### Layout

- Usar `modalStyle` como base.
- Limitar largura à área de conteúdo.
- Quebrar descrição, caminhos e mensagens longas.
- Manter todos os botões dentro da largura do terminal.
- Mostrar uma ação por linha em terminal estreito.
- Não depender de mouse para nenhuma ação.
- Evitar esconder o estado de armazenamento.

### Critérios de aceite

- Tocar em uma linha nunca instala diretamente.
- `Enter` abre a popup.
- Ações funcionam por teclado e mouse.
- Confirmações ocorrem dentro da popup.
- Popup cabe em largura mínima suportada.
- Apps detectados não podem ser removidos.
- Configuração não disponível aparece claramente.
- Estado é atualizado após cada operação.
- Sessão remota continua bloqueando operações de host.

### Resultado da Fase 10

- Toque ou `Enter` em uma linha abre a popup de detalhes sem iniciar uma
  instalação diretamente.
- A popup mostra estado, origem, versão, dependências, configuração, caminhos,
  plugins, estimativa de armazenamento e motivos de indisponibilidade.
- Instalação, desinstalação, aplicação e remoção de configuração usam o
  backend CLI; desinstalação e remoção de configuração exigem confirmação
  interna por teclado ou mouse.
- Ações bloqueadas no Ubuntu remoto, apps detectados e conflitos ficam visíveis
  sem disparar operações; operações concorrentes continuam bloqueadas pelo
  estado `busy` existente.
- Testes cobrem abertura/fechamento, hit-test da linha e ações, foco por
  teclado, confirmações, app detectado, runtime remoto e terminal estreito.

## 18. Fase 11: Testes e Integração

### Objetivo

Validar o comportamento isolado, os contratos e o fluxo completo no ambiente
Termux/PRoot.

### Testes unitários de catálogo

Arquivos:

- `internal/install/install_test.go`
- `internal/install/catalog_test.go` se necessário

Cobertura:

- Resolução de todos os apps.
- Aliases.
- Perfil Neovim.
- Perfil Yazi.
- Perfil TUIFI.
- Estratégias declaradas.
- Estimativas de armazenamento.
- Dependências declaradas.
- Rejeição de nomes inválidos.

### Testes de instalação

Cobertura:

- Instalação ausente.
- Instalação já existente.
- Dependência existente.
- Dependência compartilhada.
- Falha de comando.
- Timeout.
- Cancelamento.
- Persistência de sucesso.
- Persistência de falha.
- Yazi em ARM64 e x86_64 quando o fixture suportar ambos.
- TUIFI via pipx com comando `tuifi --version`.
- Neovim via apt com comando `nvim --version`.

### Testes de desinstalação

Cobertura:

- Remoção de app gerenciado.
- App apenas detectado.
- Pacote compartilhado.
- Arquivo com hash igual.
- Arquivo modificado.
- Arquivo ausente.
- Falha parcial.
- Repetição da remoção.
- Ausência de estratégia segura.

### Testes de configuração

Cobertura:

- Aplicação em diretório inexistente.
- Conflito com diretório existente.
- Reaplicação do próprio perfil.
- Registro separado.
- Manifesto e hashes.
- Instalação de plugin durante `config apply`.
- Remoção de plugin gerenciado.
- Preservação de plugin modificado.
- Estado `modified`.
- Falha parcial.
- Ausência de backup automático.
- LazyVim validado com Neovim headless.

### Testes CLI

Cobertura:

- Argumentos válidos.
- Argumentos ausentes.
- `--json` em sucesso.
- `--json` em falha.
- `--progress` em operação longa.
- Conflito estruturado.
- Remoção parcial estruturada.
- Runtime Ubuntu remoto bloqueado.
- Compatibilidade do schema 1.

### Testes TUI

Cobertura:

- Popup abre por `Enter`.
- Popup abre por mouse.
- Popup fecha por `Esc`.
- Foco percorre ações.
- Instalação dispara os argumentos corretos.
- Desinstalação exige confirmação.
- Configuração exige instalação prévia.
- Remoção de configuração exige confirmação.
- App detectado bloqueia remoção.
- Conflito aparece na popup.
- Estado de armazenamento aparece.
- Terminal estreito não quebra linhas.
- Mock simula sucesso e falha.
- Modo remoto bloqueia host.

### Smoke test do catálogo

Atualizar `scripts/test-catalog.sh` para:

1. Instalar Neovim.
2. Instalar Yazi.
3. Verificar `yazi`, `ya` e `nvim`.
4. Instalar TUIFI Manager.
5. Verificar `tuifi --version`.
6. Repetir instalação dos três perfis.
7. Confirmar que a segunda passagem é idempotente.

Adicionar teste específico de configuração quando o motor LazyVim estiver
implementado. O teste deverá usar um diretório de estado limpo e não deve
alterar o HOME real do host.

### Resultado da Fase 11

- `make check` passou com `go fmt`, `go vet`, `go test ./...` e build dentro do
  fixture Termux.
- Os testes unitários cobrem catálogo, instalação, desinstalação, configuração,
  contratos JSON, status, popup, confirmações, conflitos, runtime remoto e
  terminal estreito.
- O smoke test do catálogo já passou após as mudanças de catálogo e instalação,
  incluindo Neovim, Yazi e TUIFI; não foi repetido nesta fase porque as mudanças
  finais não alteraram catálogo, instalação ou PRoot.
- A validação manual descrita na seção 19 continua pendente no Termux/POCO F6
  real e deve ser executada antes de declarar o fluxo operacional validado no
  dispositivo.

## 19. Validação Manual no Termux Real

### Preparação

1. Atualizar o binário Mobdesk no Termux.
2. Confirmar que Ubuntu está acessível.
3. Confirmar espaço livre suficiente conforme estimativas.
4. Fazer backup externo do ambiente conforme a política geral do projeto.
5. Garantir que não exista configuração Neovim que o usuário queira preservar sem mover manualmente.

### Cenário de instalação

1. Abrir `mobdesk tui` no Termux.
2. Entrar em `Apps e linguagens`.
3. Abrir Neovim.
4. Confirmar a estimativa.
5. Instalar Neovim.
6. Verificar versão no shell Ubuntu.
7. Reabrir a popup.
8. Aplicar configuração Mobdesk.
9. Confirmar plugins e inicialização headless.

### Cenário de conflito

1. Criar manualmente uma configuração em `/root/.config/nvim` no Ubuntu.
2. Abrir a popup do Neovim.
3. Tentar adicionar configuração.
4. Confirmar que a operação é recusada.
5. Confirmar que nenhum arquivo existente é alterado.
6. Confirmar estado `conflict` no status e no JSON.

### Cenário de remoção

1. Aplicar a configuração Mobdesk em diretório limpo.
2. Alterar um arquivo gerenciado manualmente.
3. Remover a configuração.
4. Confirmar que o arquivo alterado é preservado.
5. Confirmar estado `modified`.
6. Confirmar que plugins intactos e gerenciados são removidos.

### Cenário remoto

1. Abrir a TUI através do SSH Mobdesk dentro do Ubuntu.
2. Entrar na tela de apps.
3. Selecionar um app.
4. Confirmar que instalação, remoção e configuração estão bloqueadas.
5. Confirmar que a mensagem orienta o usuário a voltar ao Termux.

## 20. Atualização da Documentação

Após cada fase concluída:

- Atualizar este plano com o status da fase.
- Atualizar `docs/ARQUITETURA.md` quando uma fronteira técnica mudar.
- Atualizar `docs/DECISOES.md` somente para decisões de arquitetura confirmadas.
- Atualizar `docs/ROADMAP.md` com o avanço do gerenciamento de apps.
- Atualizar estimativas de armazenamento quando houver medição real.
- Registrar limitações do Termux, Ubuntu e PRoot.

Não alterar o documento de decisões para reabrir alternativas já descartadas.

## 21. Critérios Globais de Conclusão

A implementação estará concluída quando todos os itens forem verdadeiros:

- `AppProfile` substitui o conceito de `Language` no catálogo.
- Neovim está disponível para instalação.
- Yazi e TUIFI continuam instaláveis e verificáveis.
- Todo perfil possui estratégia de instalação explícita.
- Todo perfil possui estimativa de armazenamento ou limitação documentada.
- A popup mostra nome, descrição, estado, ações e armazenamento.
- Toque na linha sempre abre a popup.
- Instalação, desinstalação e configuração usam serviços internos.
- A configuração LazyVim é opcional.
- Configuração existente gera conflito.
- Nenhum backup automático é criado.
- Arquivos modificados manualmente são preservados.
- Dependências compartilhadas não são removidas.
- Apps detectados não são desinstalados sem proveniência.
- O estado da configuração fica separado do estado do app.
- O contrato JSON continua válido no schema 1.
- A TUI funciona com teclado e mouse.
- A TUI bloqueia operações no Ubuntu remoto.
- `make check` passa.
- O smoke test do catálogo passa.
- O fluxo completo passa no Termux/PRoot real.

## 22. Ordem de Commits por Fase

Cada fase deverá terminar com exatamente um commit isolado e verificável. O
commit só pode ser criado depois dos testes e critérios de aceite da fase
passarem. Arquivos modificados antes da fase devem permanecer fora do commit.

0. `chore: freeze app management contracts`
1. `refactor: introduce app profiles`
2. `feat: add neovim app profile`
3. `feat: add storage estimates`
4. `feat: persist app provenance`
5. `feat: add safe app uninstall`
6. `feat: add app configuration engine`
7. `feat: add lazyvim configuration profile`
8. `feat: expose app operations in cli`
9. `feat: show app configuration status`
10. `feat: add app details popup`
11. `test: validate app lifecycle and document rollout`

Cada commit deverá passar os testes da fase correspondente. Nenhum commit deve
misturar mudança visual não relacionada, alteração de arquitetura fora do plano
ou ajuste de outro subsistema sem justificativa registrada.
