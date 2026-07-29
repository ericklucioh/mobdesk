# Plano de Refatoração de Apps e Configurações

**Status:** aguardando decisões de produto antes da implementação

**Objetivo:** transformar a tela atual de ferramentas em uma central de apps com
detalhes, instalação, desinstalação e aplicação opcional das configurações
opinativas do Mobdesk.

**Escopo deste documento:** registrar o que já foi decidido, separar as
decisões ainda abertas e descrever um plano executável sem deixar escolhas
implícitas.

## 1. Decisão Principal

Cada app será apresentado como um perfil do Mobdesk. Um perfil poderá declarar:

- Como o app é instalado.
- Como o app é desinstalado.
- Quais dependências possui.
- Se possui configuração opinativa do Mobdesk.
- Quais arquivos e diretórios a configuração altera.
- Quais plugins ou componentes adicionais são instalados.
- Como a configuração é aplicada.
- Como a configuração é removida com segurança.

Instalar o app e aplicar a configuração serão operações independentes.

O usuário poderá utilizar o app sem aceitar o padrão do Mobdesk. A configuração
opinativa será sempre opcional e identificada explicitamente como configuração
do Mobdesk.

## 2. Decisões Já Fechadas

Estas decisões vêm da missão, arquitetura, decisões existentes do projeto e da
solicitação desta refatoração. Não devem ser reabertas durante a implementação,
salvo mudança explícita de escopo.

### 2.1 Produto

- A TUI é a interface principal do MVP.
- A tela de apps terá uma popup de detalhes.
- A popup exibirá o nome do app.
- A popup exibirá uma frase explicativa.
- A popup terá ação de instalar.
- A popup terá ação de desinstalar.
- A popup terá ação opcional de adicionar configuração, quando o app oferecer um perfil.
- A popup terá ação de remover configuração, quando a configuração estiver aplicada.
- O usuário poderá ignorar completamente a configuração sugerida.
- O padrão de configuração será baseado nas preferências do autor do Mobdesk.
- O padrão será apresentado como sugestão, não como requisito.
- A configuração aplicada deverá ser identificada como pertencente ao Mobdesk.
- A instalação do app não poderá aplicar configuração silenciosamente.

### 2.2 Arquitetura

- Termux continuará sendo o host de controle.
- Ubuntu via PRoot continuará sendo o ambiente principal do usuário.
- Apps e configurações serão aplicados dentro do Ubuntu.
- Ações de host continuarão bloqueadas quando a TUI estiver em uma sessão SSH dentro do Ubuntu.
- A TUI continuará consumindo o contrato JSON da CLI.
- A TUI não terá regras próprias de instalação, desinstalação ou configuração.
- Cobra continuará sendo um adaptador de flags, entrada e saída.
- Serviços internos continuarão independentes da renderização da TUI.
- Operações longas continuarão usando contexto e cancelamento.
- A TUI continuará executando no máximo uma operação de host por vez.

### 2.3 Segurança e persistência

- Operações destrutivas exigirão confirmação explícita.
- Arquivos existentes do usuário não serão sobrescritos silenciosamente.
- Configurações existentes deverão ser detectadas antes da aplicação.
- Arquivos modificados pelo usuário não serão apagados automaticamente durante a remoção.
- Pacotes compartilhados não serão removidos de forma cega.
- O estado das operações será persistido em diretórios privados.
- Logs não deverão conter segredos.
- Operações repetidas deverão preservar dados e estado.
- Falhas parciais não deverão apagar projetos ou configurações sem relação com o app.
- A validação final deverá incluir `make check` e teste em Termux real.

## 3. Situação Atual

### 3.1 Fluxo atual

Hoje o fluxo de apps é:

```text
Selecionar ou tocar em uma linha
  -> instalar diretamente
  -> TUI chama mobdesk install <app>
  -> CLI chama internal/install
  -> proot-distro entra no Ubuntu
  -> instalação é executada
  -> estado e log são persistidos
  -> TUI atualiza o status
```

O clique em uma linha ainda instala diretamente. Não há popup de detalhes,
desinstalação ou configuração de app.

### 3.2 Componentes atuais

| Área | Arquivos principais | Responsabilidade atual |
|---|---|---|
| Catálogo | `internal/install/model.go` e `internal/install/install.go` | Define e instala linguagens e ferramentas |
| Estado | `internal/install/InstallationRecord` | Persiste estado da instalação |
| Status | `internal/status/model.go` e `internal/status/collect.go` | Lê instalações e detecta executáveis |
| CLI | `internal/cobra/install.go` e `internal/cobra/json.go` | Expõe instalação e JSON |
| TUI | `internal/tui/model.go` | Mantém catálogo exibido e seleção |
| TUI | `internal/tui/screen_tools.go` | Renderiza a tela de apps |
| TUI | `internal/tui/tui.go` e `internal/tui/mouse.go` | Trata teclado, mouse e operações |
| Backend | `internal/tui/backend.go` e `internal/tui/commands.go` | Chama a CLI real ou mock |
| Paths | `internal/paths/paths.go` | Define estado e diretórios persistentes |

### 3.3 Problemas que a refatoração resolve

- O catálogo mistura conceito de linguagem, ferramenta e app.
- Descrições dos apps ficam na camada da TUI.
- A instalação não possui operação inversa.
- O estado não registra suficientemente a origem dos arquivos instalados.
- O sistema não distingue app instalado de configuração aplicada.
- Não há proteção específica para remover configuração modificada pelo usuário.
- A TUI realiza instalação imediatamente ao tocar em uma linha.
- O contrato JSON não representa ações de configuração.

## 4. Experiência Desejada

### 4.1 Tela de apps

Cada card continuará mostrando informações resumidas:

- Nome ou executável principal.
- Frase curta.
- Estado do app.
- Estado da configuração Mobdesk, quando existir.

Exemplo visual de estado:

```text
neovim                         instalado
Editor modal                   configuração não aplicada
```

Tocar ou pressionar `Enter` abrirá a popup. A linha não executará uma operação
destrutiva ou demorada diretamente.

### 4.2 Popup de detalhes

A popup deverá conter:

- Nome do app.
- Descrição curta.
- Estado atual do app.
- Versão conhecida, quando disponível.
- Dependências relevantes.
- Estado da configuração Mobdesk.
- Caminhos que serão afetados pela configuração.
- Plugins ou componentes adicionais, quando houver.
- Ação `Instalar`.
- Ação `Desinstalar`.
- Ação `Adicionar configuração Mobdesk`, quando disponível.
- Ação `Remover configuração Mobdesk`, quando aplicada.
- Ação `Fechar`.

### 4.3 Estados da popup

| Estado | Comportamento |
|---|---|
| App não instalado | Permite instalar; configuração fica bloqueada ou explica o pré-requisito |
| App instalado sem configuração | Permite usar o app ou aplicar a configuração opcional |
| App instalado com configuração aplicada | Permite remover a configuração com confirmação |
| App em operação | Bloqueia outras ações até o resultado final |
| App com conflito | Bloqueia aplicação automática e explica o conflito |
| App sem remoção segura | Desinstalação fica indisponível com explicação objetiva |
| Sessão Ubuntu remota | Mostra que a ação deve ser executada no Termux |

### 4.4 Confirmações

Deverão exigir confirmação:

- Desinstalar o app.
- Remover a configuração.
- Restaurar backup.
- Substituir uma configuração existente, caso essa operação venha a ser permitida.

As confirmações deverão funcionar por teclado e mouse, inclusive em terminal
estreito.

## 5. Modelo de Domínio

### 5.1 Perfil do app

O catálogo deverá evoluir para um perfil que contenha, conceitualmente:

```text
name
aliases
description
kind
package
executable
version_arg
install_kind
requires
user_bin
install_profile
uninstall_profile
config_profile
profile_version
```

O nome exato dos tipos Go será definido na decisão D1.

O perfil não deve aceitar comandos vindos da entrada do usuário. Comandos e
caminhos serão declarados pelo catálogo versionado do Mobdesk.

### 5.2 Perfil de configuração

Um perfil de configuração deverá declarar:

- Identificador estável.
- Versão do perfil.
- Aplicativo ao qual pertence.
- Descrição legível.
- Caminhos de destino dentro do Ubuntu.
- Arquivos gerados.
- Diretórios gerados.
- Plugins instalados.
- Dependências de configuração.
- Estratégia de backup.
- Estratégia de aplicação.
- Estratégia de remoção.

### 5.3 Registro de instalação

O registro atual em `state/installations/<app>.json` deverá ser ampliado ou
complementado para guardar:

- Nome canônico.
- Pacote principal.
- Executável principal.
- Estado da instalação.
- Versão.
- Dependências resolvidas.
- Pacotes instalados pelo Mobdesk.
- Arquivos instalados pelo Mobdesk.
- Diretórios instalados pelo Mobdesk.
- Data da instalação.
- Última tentativa.
- Último erro.
- Caminho do log.

### 5.4 Registro de configuração

O estado da configuração deverá ser separado do estado de instalação. O
registro deverá guardar:

- App relacionado.
- Identificador do perfil.
- Versão do perfil.
- Estado da aplicação.
- Caminhos gerenciados.
- Arquivos gerados.
- Plugins gerenciados.
- Hashes dos arquivos gerados.
- Caminho do backup.
- Data da aplicação.
- Data da remoção.
- Indicação de alteração manual.
- Indicação de conflito.
- Último erro.

O local recomendado é um diretório específico de configurações dentro do
estado persistente do Mobdesk. O caminho final depende da decisão D8.

## 6. Segurança da Configuração

### 6.1 Antes de aplicar

Antes de modificar qualquer caminho, o serviço deverá:

1. Resolver o app por nome canônico ou alias.
2. Confirmar que o app está instalado.
3. Confirmar que existe um perfil de configuração.
4. Validar que os caminhos estão dentro do HOME do Ubuntu esperado.
5. Inspecionar arquivos e diretórios existentes.
6. Identificar se são vazios, gerenciados pelo Mobdesk ou desconhecidos.
7. Criar backup conforme a política escolhida.
8. Persistir uma tentativa de configuração.
9. Aplicar arquivos e plugins.
10. Validar o resultado.
11. Persistir hashes e manifestos somente após sucesso.

### 6.2 Durante a aplicação

- Usar arquivos temporários no mesmo filesystem do destino.
- Promover arquivos com rename atômico quando possível.
- Registrar cada componente criado.
- Interromper a operação ao primeiro erro não recuperável.
- Tentar rollback dos componentes criados pela operação atual.
- Preservar o log da operação.
- Não executar comandos de shell formados com nomes não validados.

### 6.3 Durante a remoção

O serviço deverá comparar o estado atual com os hashes registrados:

- Hash igual ao gerado: arquivo pode ser removido ou restaurado.
- Hash diferente: arquivo foi alterado; deve ser preservado.
- Arquivo ausente: registrar como já removido.
- Arquivo desconhecido: preservar.
- Backup disponível: oferecer restauração conforme a decisão D5.

### 6.4 Dependências

Dependências como Node, Python, Go e Clang podem ser usadas por mais de um
app. O desinstalador deve preservar dependências quando ainda houver referência
de outro perfil ou quando não houver comprovação de que foram instaladas pelo
Mobdesk para aquele app.

## 7. CLI e Contrato JSON

### 7.1 Comandos

Os comandos planejados são:

```text
mobdesk install <app>
mobdesk uninstall <app>
mobdesk config apply <app>
mobdesk config remove <app>
```

Cada operação deverá aceitar:

```text
--json
--progress
```

### 7.2 Resultado estruturado

O resultado deverá representar, no mínimo:

```text
schema_version
command
target
action
success
state
changed
message
version
config_state
log_path
conflicts
paths
```

O resultado estruturado deverá ser emitido tanto em sucesso quanto em falha,
sem misturar mensagens humanas no stdout JSON.

### 7.3 Estados sugeridos

Estados de app:

```text
available
installing
installed
uninstalling
uninstalled
partial
failed
```

Estados de configuração:

```text
unavailable
not_applied
applying
applied
removing
removed
modified
conflict
failed
```

## 8. Perfil Piloto: Neovim e LazyVim

O primeiro caso de configuração será Neovim/LazyVim, por ser o exemplo
concreto que motivou a funcionalidade.

O perfil deverá especificar:

- Como instalar o Neovim.
- Qual versão é suportada.
- Qual diretório de configuração será usado.
- Qual bootstrap do LazyVim será usado.
- Quais plugins serão instalados.
- Se o gerenciador de plugins será instalado ou reutilizado.
- Quais arquivos pertencem ao Mobdesk.
- Como verificar que a configuração está funcional.
- Como criar backup.
- Como remover ou restaurar a configuração.

Nenhum outro app deverá receber uma configuração genérica apenas para preencher
a interface. Apps sem perfil deverão informar que não possuem configuração
Mobdesk disponível.

## 9. Plano de Implementação

### Fase 1: contrato e catálogo

Arquivos prováveis:

- `internal/install/model.go`
- `internal/install/install.go`
- `internal/tui/model.go`
- `internal/install/install_test.go`

Entregas:

- Definir o perfil de app.
- Mover descrições para o catálogo.
- Adicionar perfil de configuração opcional.
- Definir estratégias suportadas de instalação e remoção.
- Adicionar o perfil inicial do Neovim/LazyVim, conforme D2 e D3.
- Testar resolução por nome e alias.

### Fase 2: proveniência e estado

Arquivos prováveis:

- `internal/install/model.go`
- `internal/install/install.go`
- `internal/paths/paths.go`
- `internal/status/model.go`
- `internal/status/collect.go`
- `internal/install/install_test.go`
- `internal/status/collect_test.go`

Entregas:

- Registrar arquivos e pacotes pertencentes ao Mobdesk.
- Criar registros separados para configuração.
- Definir permissões privadas.
- Persistir estados parciais e falhas.
- Preservar registros existentes durante a evolução do formato.
- Testar leitura de estado antigo e novo conforme D8.

### Fase 3: desinstalação segura

Arquivos prováveis:

- `internal/install/uninstall.go`
- `internal/install/install.go`
- `internal/install/install_test.go`

Entregas:

- Implementar remoção por estratégia.
- Implementar proteção para dependências compartilhadas.
- Implementar verificação de proveniência.
- Registrar arquivos removidos e preservados.
- Retornar estado parcial quando a remoção não puder ser completa.
- Bloquear remoção sem estratégia segura.

### Fase 4: motor de configuração

Arquivos prováveis:

- `internal/install/config.go` ou arquivo equivalente dentro do serviço existente
- `internal/paths/paths.go`
- `internal/install/config_test.go`

Entregas:

- Aplicar configuração declarada pelo perfil.
- Criar manifestos e hashes.
- Criar backup conforme D5.
- Detectar conflito conforme D4.
- Instalar plugins declarados.
- Remover somente arquivos gerenciados ou restaurar backup.
- Preservar arquivos modificados manualmente.
- Implementar rollback de falha parcial.

### Fase 5: comandos CLI

Arquivos prováveis:

- `internal/cobra/install.go`
- `internal/cobra/uninstall.go`
- `internal/cobra/config.go`
- `internal/cobra/json.go`
- `internal/cobra/install_test.go`
- novos testes de CLI

Entregas:

- Adicionar `uninstall`.
- Adicionar `config apply`.
- Adicionar `config remove`.
- Adicionar progresso JSON.
- Adicionar resultado estruturado de sucesso e falha.
- Validar runtime Termux antes de qualquer operação.
- Manter a saída humana existente de `install`.

### Fase 6: popup da TUI

Arquivos prováveis:

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

Entregas:

- Abrir popup ao selecionar app.
- Remover instalação direta pelo toque na linha.
- Adicionar estado de popup e foco de ação.
- Adicionar ações de instalar e desinstalar.
- Adicionar ações de aplicar e remover configuração.
- Adicionar confirmações destrutivas.
- Atualizar mock backend.
- Bloquear ações durante operações.
- Atualizar status após cada operação.
- Garantir layout em terminal estreito.
- Manter explicação de restrição no Ubuntu remoto.

### Fase 7: validação e documentação

Arquivos a atualizar:

- `docs/ARQUITETURA.md`
- `docs/DECISOES.md`
- `docs/ROADMAP.md`
- este documento, registrando as decisões escolhidas

Validações:

- `gofmt` nos arquivos Go alterados.
- Testes unitários dos serviços.
- Testes de contrato JSON.
- Testes da TUI.
- `make check`.
- Smoke test no Ubuntu via PRoot.
- Teste de instalação e remoção no Termux real.
- Teste de configuração existente e modificada manualmente.
- Teste de interrupção durante aplicação e remoção.

## 10. Critérios de Aceite

### Produto

- Tocar em um app abre detalhes em vez de instalar imediatamente.
- A popup mostra nome e explicação.
- A popup permite instalar quando o app está disponível.
- A popup permite desinstalar somente quando existe estratégia segura.
- A popup mostra claramente se existe configuração Mobdesk.
- O usuário pode aplicar ou ignorar a configuração.
- A aplicação da configuração exige ação explícita.
- A remoção da configuração exige confirmação.

### Arquitetura

- A TUI não contém comandos de instalação ou configuração próprios.
- A CLI e a TUI usam o mesmo serviço interno.
- A operação não é executada a partir de Ubuntu remoto.
- Uma operação host por vez continua garantida.
- Operações podem ser canceladas.

### Segurança

- Arquivos existentes são detectados antes da aplicação.
- Backups são criados conforme a política escolhida.
- Arquivos alterados pelo usuário não são removidos automaticamente.
- Dependências compartilhadas não são quebradas.
- Logs não expõem segredos.
- Falhas parciais deixam estado recuperável.

### Qualidade

- Testes cobrem sucesso, falha, repetição e cancelamento.
- Testes cobrem terminal estreito.
- Testes cobrem teclado e mouse.
- `make check` passa.
- O fluxo é validado no Termux real antes de ser considerado concluído.

## 11. Decisões Pendentes

Cada decisão abaixo possui exatamente duas opções. A implementação não deverá
assumir uma escolha sem que ela seja registrada na seção 12.

### D1. Nome do modelo de catálogo

**Pergunta:** o tipo atual `Language` deve ser substituído por um conceito de
app?

**Opção A, recomendada: introduzir `AppProfile`**

- Representa linguagens, ferramentas, editores e clientes de forma correta.
- Permite evoluir para configuração sem distorcer o modelo.
- Exige atualizar usos e testes do catálogo.

**Opção B: manter `Language` e apenas adicionar campos**

- Menor alteração imediata.
- Preserva o nome atual mesmo para apps que não são linguagens.
- Pode aumentar a dívida conceitual do pacote.

### D2. Primeiro escopo do catálogo configurável

**Pergunta:** quais apps terão configuração na primeira entrega?

**Opção A, recomendada: somente Neovim/LazyVim**

- Permite validar segurança, backup, plugins e remoção em um caso real.
- Reduz o risco de criar várias configurações frágeis.
- Outros apps continuam instaláveis sem configuração Mobdesk.

**Opção B: Neovim/LazyVim e mais um app escolhido pelo usuário**

- Valida dois formatos de configuração desde o início.
- Aumenta o tempo e a superfície de testes.
- Exige definir agora o segundo perfil completo.

### D3. Instalação do Neovim

**Pergunta:** o perfil do Neovim deve instalar o app caso ele ainda não exista?

**Opção A, recomendada: exigir que o app esteja instalado antes da configuração**

- Mantém instalação e configuração independentes.
- Torna o fluxo mais previsível e fácil de desfazer.
- A popup orienta o usuário a instalar primeiro.

**Opção B: permitir que `config apply` instale o Neovim automaticamente**

- Reduz cliques para iniciantes.
- Cria uma operação composta com mais pontos de falha.
- Exige tratar instalação parcial e rollback conjunto.

### D4. Configuração existente

**Pergunta:** o que fazer quando já existir `~/.config/nvim`?

**Opção A, recomendada: recusar e informar conflito**

- Nunca sobrescreve a configuração do usuário.
- É a opção mais segura para o MVP.
- O usuário pode remover ou mover o diretório manualmente e tentar novamente.

**Opção B: fazer backup e substituir pela configuração Mobdesk**

- Entrega uma experiência mais automática.
- Aumenta o risco de surpresa e perda percebida.
- Exige restauração confiável e confirmação específica.

### D5. Política de backup

**Pergunta:** onde e como guardar o backup?

**Opção A, recomendada: backup local versionado pelo Mobdesk**

- Guardar em `~/.local/share/mobdesk/backups/<app>/<timestamp>`.
- Permite restauração sem depender de outro dispositivo.
- Consome armazenamento e exige limpeza futura.

**Opção B: não criar backup automático; apenas recusar conflito**

- Simplifica o motor de configuração.
- Evita o Mobdesk assumir responsabilidade sobre cópias.
- Torna a automação menos conveniente para o usuário.

### D6. Remoção de plugins

**Pergunta:** remover a configuração deve remover também os plugins instalados?

**Opção A, recomendada: remover apenas plugins comprovadamente gerenciados**

- Usa manifesto e hashes para preservar alterações manuais.
- Remove componentes pertencentes ao perfil quando for seguro.
- Pode deixar arquivos preservados em conflitos.

**Opção B: remover apenas os arquivos de configuração**

- Minimiza risco de apagar dados de plugins.
- Pode deixar consumo de espaço e componentes sem uso.
- Torna a remoção menos completa.

### D7. Desinstalação de dependências

**Pergunta:** a desinstalação deve remover dependências automáticas?

**Opção A, recomendada: nunca remover dependências automaticamente no MVP**

- Evita quebrar outros apps.
- Mantém a remoção conservadora.
- Pode deixar pacotes não utilizados no Ubuntu.

**Opção B: remover dependências com contagem de referências**

- Reduz pacotes sobrando.
- Exige proveniência precisa e tratamento de pacotes instalados manualmente.
- Aumenta consideravelmente a complexidade do desinstalador.

### D8. Local do registro de configuração

**Pergunta:** como persistir o estado da configuração?

**Opção A, recomendada: arquivo separado por app**

- Usar `state/configurations/<app>.json`.
- Mantém instalação e configuração independentes.
- Facilita leitura, recuperação e testes.

**Opção B: adicionar campos ao registro de instalação**

- Reduz a quantidade de arquivos persistidos.
- Mistura dois ciclos de vida diferentes.
- Torna estados parciais e múltiplas configurações mais difíceis de representar.

### D9. Versão do contrato JSON

**Pergunta:** as novas ações devem alterar a versão do schema?

**Opção A, recomendada: manter schema 1 com campos aditivos opcionais**

- Preserva consumidores atuais.
- Evita migração imediata da TUI e scripts existentes.
- Exige cuidado para não mudar o significado dos campos antigos.

**Opção B: criar schema 2 para todas as operações de apps**

- Torna o novo contrato explícito.
- Exige atualizar todos os consumidores ao mesmo tempo.
- Pode simplificar a modelagem futura.

### D10. Ação de desinstalação indisponível

**Pergunta:** como tratar apps sem estratégia de remoção segura?

**Opção A, recomendada: desabilitar o botão e explicar o motivo**

- Impede uma remoção genérica perigosa.
- Mantém honestidade na interface.
- Permite adicionar a estratégia posteriormente.

**Opção B: oferecer remoção manual orientada**

- Ajuda o usuário a concluir a limpeza.
- Exige exibir comandos e caminhos específicos.
- Não entrega uma desinstalação automatizada pelo Mobdesk.

### D11. Fluxo de confirmação

**Pergunta:** onde confirmar operações destrutivas?

**Opção A, recomendada: confirmação dentro da popup**

- Mantém o usuário no contexto do app.
- Funciona bem com mouse e teclado.
- Exige adicionar mais um estado de foco à TUI.

**Opção B: fechar a popup e usar um modal global de confirmação**

- Reutiliza parte da confirmação existente.
- Pode reduzir o contexto visual da ação.
- Exige transportar a operação pendente entre telas.

### D12. Estado após remoção parcial

**Pergunta:** como exibir uma remoção que preservou arquivos modificados?

**Opção A, recomendada: estado `modified` com detalhes preservados**

- Mostra que a remoção foi parcial por segurança.
- Permite nova ação depois de intervenção do usuário.
- Exige incluir detalhes na popup e no JSON.

**Opção B: estado `removed` e aviso apenas no log**

- Simplifica o modelo visual.
- Esconde uma informação importante da tela principal.
- Aumenta o risco de o usuário achar que tudo foi removido.

### D13. Aplicação de plugins

**Pergunta:** quando instalar plugins declarados pelo perfil?

**Opção A, recomendada: instalar durante `config apply`**

- A ação representa exatamente a intenção do usuário.
- O manifesto registra plugins como parte da configuração.
- A operação pode ser longa e exige progresso confiável.

**Opção B: criar configuração primeiro e instalar plugins na primeira abertura do app**

- Reduz o tempo da ação inicial.
- Torna o primeiro uso menos previsível.
- Exige integração com o app ou script de inicialização.

### D14. Ação de toque na lista

**Pergunta:** o toque em uma linha deve fazer o quê depois da refatoração?

**Opção A, recomendada: sempre abrir a popup**

- Evita ações acidentais.
- Mantém teclado e mouse com o mesmo comportamento.
- Exige um toque ou Enter adicional para instalar.

**Opção B: tocar no nome abre popup e tocar no estado executa a ação principal**

- Pode reduzir cliques para usuários experientes.
- Cria alvos diferentes na mesma linha.
- Aumenta o risco de instalação ou remoção acidental em telas móveis.

### D15. Descoberta de apps instalados

**Pergunta:** como reconciliar estado persistido e executáveis encontrados?

**Opção A, recomendada: mostrar estado `detectado` até haver registro completo**

- Diferencia instalação feita fora do Mobdesk.
- Evita alegar proveniência que não foi comprovada.
- Exige um estado adicional na interface.

**Opção B: tratar executável encontrado como `instalado`**

- Mantém o comportamento atual mais simples.
- Pode permitir desinstalação sem saber quem instalou o arquivo.
- Aumenta o risco de remover instalação externa.

### D16. Escopo de atualização de documentação

**Pergunta:** quando registrar as decisões escolhidas nos documentos oficiais?

**Opção A, recomendada: atualizar documentação junto com cada fase**

- Mantém arquitetura e implementação sincronizadas.
- Reduz divergência durante a refatoração.
- Exige disciplina em cada entrega.

**Opção B: atualizar toda documentação ao final da refatoração**

- Reduz alterações intermediárias.
- Mantém documentação temporariamente desatualizada.
- Aumenta o risco de esquecer decisões importantes.

## 12. Registro das Escolhas

Preencher uma opção por decisão antes da implementação.

| Decisão | Escolha | Observação |
|---|---|---|
| D1. Nome do modelo | [ ] A  [ ] B | |
| D2. Escopo configurável | [ ] A  [ ] B | |
| D3. Instalação do Neovim | [ ] A  [ ] B | |
| D4. Configuração existente | [ ] A  [ ] B | |
| D5. Backup | [ ] A  [ ] B | |
| D6. Remoção de plugins | [ ] A  [ ] B | |
| D7. Dependências | [ ] A  [ ] B | |
| D8. Registro persistente | [ ] A  [ ] B | |
| D9. Schema JSON | [ ] A  [ ] B | |
| D10. Remoção indisponível | [ ] A  [ ] B | |
| D11. Confirmação | [ ] A  [ ] B | |
| D12. Remoção parcial | [ ] A  [ ] B | |
| D13. Plugins | [ ] A  [ ] B | |
| D14. Toque na lista | [ ] A  [ ] B | |
| D15. App detectado | [ ] A  [ ] B | |
| D16. Documentação | [ ] A  [ ] B | |

## 13. Dependências e Riscos

### Riscos técnicos

- A instalação atual usa várias estratégias e não possui remoção simétrica.
- `c` e `cpp` compartilham o pacote `clang`.
- Node, Python e Go podem ser dependências de vários apps.
- Executáveis detectados pelo status podem ter sido instalados fora do Mobdesk.
- Scripts de release podem instalar arquivos fora de um diretório simples.
- A configuração do usuário pode existir antes da aplicação.
- Plugins podem alterar arquivos depois da instalação.
- A TUI precisa funcionar em telas móveis estreitas.
- Operações de configuração podem durar mais que uma interação normal.
- Uma interrupção do Android pode acontecer durante a operação.

### Mitigações obrigatórias

- Catálogo com estratégias explícitas.
- Manifesto de proveniência.
- Hashes de arquivos gerenciados.
- Backup ou recusa de conflito conforme D5.
- Proteção contra dependências compartilhadas.
- Confirmação para ações destrutivas.
- Lock compartilhado entre instalação, remoção e configuração.
- Rollback da operação atual quando possível.
- Estado parcial observável.
- Teste em Termux real.

## 14. Ordem Final de Execução

1. Registrar as escolhas D1 a D16.
2. Atualizar este documento com as decisões selecionadas.
3. Implementar o modelo de perfil de app.
4. Implementar estado e proveniência.
5. Implementar desinstalação segura.
6. Implementar o perfil Neovim/LazyVim.
7. Implementar comandos CLI e contratos JSON.
8. Implementar a popup da TUI.
9. Implementar testes unitários, de contrato e de interação.
10. Atualizar `docs/ARQUITETURA.md`, `docs/DECISOES.md` e `docs/ROADMAP.md`.
11. Executar `make check`.
12. Validar instalação, configuração, remoção e recuperação no Termux real.

## 15. Definição de Concluído

A refatoração será considerada concluída quando:

- Todas as decisões pendentes estiverem registradas.
- A popup estiver funcionando por teclado e mouse.
- O toque não instalar diretamente sem abrir detalhes.
- Instalação e desinstalação usarem serviços internos comuns.
- Configuração e instalação possuírem estados separados.
- O perfil Neovim/LazyVim puder ser aplicado de forma opcional.
- A configuração existente não for sobrescrita silenciosamente.
- A remoção preservar alterações manuais.
- Dependências compartilhadas permanecerem intactas.
- A CLI emitir JSON válido em sucesso e falha.
- A TUI funcionar em sessão Termux e bloquear corretamente o modo Ubuntu remoto.
- Os testes cobrirem os fluxos principais e as falhas de segurança.
- `make check` passar.
- O fluxo completo tiver sido testado no Termux/PRoot real.
