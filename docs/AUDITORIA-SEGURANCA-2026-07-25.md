# Resumo Executivo

Mobdesk é uma CLI/TUI Go para Termux que provisiona Ubuntu via PRoot-Distro, instala linguagens, gerencia um `sshd` na porta `8022` e atualiza o próprio binário.

Arquitetura: `cmd/mobdesk` -> Cobra -> serviços `workstation`, `install`, `status`, `update`, `logs`; a TUI Bubble Tea reinvoca a própria CLI e interpreta JSON.

Achados: **0 críticos, 3 altos, 5 médios, 2 baixos, 2 informativos**.

Avaliação: **não está pronto para produção em rede**. O maior bloqueador é expor SSH por senha em todas as interfaces, seguido da atualização sem autenticidade independente e falhas de coordenação/cancelamento na TUI.

# Mapa Da Superfície De Ataque

- Entradas: argumentos e flags CLI, estado JSON local, logs persistidos, `HOME`/`PREFIX`, respostas GitHub, saída de Termux API, rede SSH.
- Operações privilegiadas: `pkg`, `apt-get` no Ubuntu PRoot, criação/início/parada de `sshd`, escrita de configuração e substituição do binário.
- Dados sensíveis: senha da conta Termux, `authorized_keys`, logs de comandos e estado da instalação.
- Integrações: Termux, OpenSSH, PRoot-Distro, repositórios apt/pkg, GitHub Releases.
- Maior risco: SSH publicado em LAN e pipeline de atualização que executa binários baixados.

# Achados Altos

## [H-01] SSH exposto na LAN com senha habilitada e proteção de permissões desabilitada

**Severidade:** Alta  
**Confiança:** Alta  
**Categoria:** Autenticação / configuração insegura  
**Localização:** `internal/workstation/ssh_config.go:24-27`  
**Descrição:** A configuração usa `ListenAddress 0.0.0.0`, `PasswordAuthentication yes`, `KbdInteractiveAuthentication yes` e `StrictModes no`.  
**Causa raiz:** O serviço prioriza acesso remoto por senha sem limitar origem, usuário, tentativas ou exigir chaves.  
**Cenário:** Em Wi-Fi universitário, público ou comprometido, qualquer host acessível tenta autenticação repetidamente na porta 8022. Se um processo conseguir gravar `~/.ssh` do usuário, `StrictModes no` reduz uma defesa do OpenSSH contra arquivos de chaves permissivos.  
**Impacto:** Acesso remoto não autorizado ao shell Ubuntu/PRoot do usuário; possibilidade de comprometimento de dados e projetos acessíveis à sessão.  
**Correção recomendada:** Usar `StrictModes yes`; preferir chaves por padrão; desabilitar senha ou torná-la opt-in; oferecer bind em loopback como padrão e exposição LAN explicitamente confirmada. Definir `MaxAuthTries`, `LoginGraceTime` e, se compatível, `AllowUsers`.  
**Patch sugerido:** Alterar a configuração gerada para `ListenAddress 127.0.0.1`, `PasswordAuthentication no`, `KbdInteractiveAuthentication no`, `StrictModes yes`; criar flag explícita para exposição LAN e senha.  
**Teste de regressão:** Testar `renderSSHConfig` para garantir os defaults seguros e que só flags explícitas habilitam rede/senha.  
**Efeitos colaterais:** Requer chave SSH ou configuração explícita do usuário para acesso pela LAN.

## [H-02] Atualizador aceita binário e checksum comprometidos juntos

**Severidade:** Alta  
**Confiança:** Alta  
**Categoria:** Supply chain / execução de código  
**Localização:** `internal/update/update.go:275-286`, `317-357`  
**Descrição:** O binário e o `SHA256SUMS` são baixados da mesma release e o checksum não possui assinatura ou chave pública fixada.  
**Causa raiz:** SHA-256 detecta corrupção acidental entre downloads, mas não autentica a origem.  
**Cenário:** Comprometimento da conta GitHub, release, token de publicação ou artefatos permite publicar binário malicioso acompanhado de checksum correspondente.  
**Impacto:** `mobdesk update` instala código arbitrário executável como o usuário Termux.  
**Correção recomendada:** Assinar o manifesto de checksums e verificar assinatura contra chave pública embutida no binário. Minisign, cosign ou GPG são opções; a chave não pode ser baixada junto do artefato.  
**Patch sugerido:** Publicar `SHA256SUMS.minisig` no release e validar a assinatura antes de chamar `downloadBinary`.  
**Teste de regressão:** Recusar checksum válido sem assinatura e assinatura feita por chave não confiável.  
**Efeitos colaterais:** Exige gestão segura da chave de release e adaptação do workflow de publicação.

## [H-03] Atualização pode deixar o comando indisponível após interrupção

**Severidade:** Alta  
**Confiança:** Média  
**Categoria:** Integridade / disponibilidade  
**Localização:** `internal/update/update.go:298-308`  
**Descrição:** O binário atual é renomeado para `.bak` antes de instalar o temporário. Encerramento pelo Android, falta de espaço, falha de filesystem ou término do processo entre os dois `Rename` deixa `InstallPath` ausente.  
**Causa raiz:** Não existe recuperação no início da aplicação nem estratégia de troca que preserve um executável válido após interrupção.  
**Cenário:** HyperOS encerra o processo durante `mobdesk update`, comportamento que a própria arquitetura reconhece como possível.  
**Impacto:** O launcher aponta para caminho inexistente e o usuário precisa recuperar manualmente o `.bak`.  
**Correção recomendada:** Antes de atualizar, recuperar automaticamente `.bak` quando o executável principal estiver ausente; manter backup até uma inicialização posterior bem-sucedida.  
**Patch sugerido:** Adicionar rotina de recuperação antes de `Apply` e somente remover `.bak` após validação do binário instalado.  
**Teste de regressão:** Simular falha após o primeiro `Rename`; executar recuperação e verificar que o binário anterior volta ao caminho original.  
**Efeitos colaterais:** Backups podem permanecer temporariamente e exigem política de limpeza.

**Status:** Corrigido em 25/07/2026. A substituição agora renomeia o temporário diretamente sobre o executável no mesmo diretório, eliminando a janela sem binário. O atualizador também recupera `.bak` legado quando o executável principal não existe; os cenários possuem testes de regressão.

# Achados Médios

## [M-01] TUI permite operações concorrentes e aceita mensagens fora de ordem

**Confiança:** Alta  
**Categoria:** Concorrência / bug lógico  
**Localização:** `internal/tui/tui.go:62-80`, `122-128`, `222-223`, `304-315`  
**Descrição:** `busy` é apenas estado visual. Não bloqueia `r`, mouse ou atalhos que iniciam novas operações. Qualquer `statusMessage` encerra a operação visual atual, mesmo que pertença a um refresh anterior.  
**Cenário:** Iniciar `install`, pressionar `r`; o status retorna antes da instalação e limpa `busy`. O usuário pode iniciar outra instalação ou `start`.  
**Impacto:** Estados falsos na UI, execuções duplicadas e disputa pelo lock do `apt`; registros de instalação podem se sobrescrever.  
**Correção:** Bloquear operações host enquanto houver operação ativa; associar cada comando/status a um identificador monotônico e ignorar respostas obsoletas.  
**Teste:** Disparar instalação e refresh com respostas controladas fora de ordem; confirmar que a operação continua marcada e que o snapshot antigo é descartado.

**Status:** Corrigido em 25/07/2026. A TUI bloqueia operações host enquanto uma está ativa, marca operações e snapshots de status com IDs monotônicos e descarta mensagens obsoletas. Testes cobrem tentativa duplicada, snapshot desatualizado e aplicação do snapshot mais recente.

## [M-02] Subprocessos disparados pela TUI não têm cancelamento nem prazo

**Confiança:** Alta  
**Categoria:** Contexto / leak de processo  
**Localização:** `internal/tui/commands.go:16-34`  
**Descrição:** `runCommand` usa `exec.Command`, não `exec.CommandContext`; setup, instalação e update não recebem timeout.  
**Cenário:** Usuário sai da TUI durante uma instalação longa ou uma conexão GitHub fica pendurada. O filho continua em background sem interface de cancelamento.  
**Impacto:** Processos órfãos, alterações após o usuário acreditar que a operação terminou e indisponibilidade da TUI.  
**Correção:** Criar um contexto de ciclo de vida da TUI, usar `exec.CommandContext`, cancelar no quit e definir limites por operação.  
**Teste:** Iniciar comando bloqueante, encerrar TUI e verificar término do processo filho.

**Status:** Corrigido em 25/07/2026. O backend real passou a manter contexto cancelável do ciclo de vida da TUI; operações, status e shell usam `exec.CommandContext`, e a confirmação de saída cancela os processos filhos.

## [M-03] Update não impõe timeout global ou limites de download

**Confiança:** Alta  
**Categoria:** Disponibilidade / recursos  
**Localização:** `internal/update/update.go:140-177`, `317-326`, `339-357`  
**Descrição:** O cliente HTTP não define `Timeout`; `SHA256SUMS` é lido integralmente e o binário é copiado até EOF sem limite.  
**Cenário:** Resposta lenta que nunca finaliza, manifesto enorme ou asset excessivamente grande.  
**Impacto:** TUI travada, consumo ilimitado de armazenamento e memória.  
**Correção:** Definir timeout de cliente, limites para manifesto e tamanho máximo do binário. Restringir redirects se não forem necessários.  
**Teste:** Servidor `httptest` que bloqueia, envia manifesto maior que o limite e stream acima do máximo.

## [M-04] Logs são lidos integralmente e o estado pode redirecionar leitura de arquivos

**Confiança:** Alta  
**Categoria:** Recursos / leitura arbitrária condicionada  
**Localização:** `internal/logs/logs.go:62-83`, `92-108`  
**Descrição:** `LogPath` vem diretamente do JSON persistido e é passado para `os.ReadFile`; o arquivo inteiro é carregado apesar da opção `--lines`.  
**Cenário:** Um registro de instalação adulterado aponta para arquivo legível pelo usuário; `mobdesk logs` o imprime. Um log grande consome memória integralmente.  
**Impacto:** Exposição local de conteúdo legível pelo mesmo usuário e possível esgotamento de memória.  
**Correção:** Derivar o caminho de log a partir do nome validado do catálogo, rejeitar caminhos fora de `InstallLogsDir`, e implementar tail limitado por tamanho.  
**Teste:** Registro com `log_path` fora da raiz deve ser recusado; log de centenas de MB não deve ser lido integralmente.

**Status:** Parcialmente corrigido em 25/07/2026. O leitor aceita apenas registros canônicos de linguagens do catálogo e deriva o caminho sob `InstallLogsDir`, ignorando `log_path` persistido. O limite de leitura para logs extensos permanece pendente de decisão.

## [M-05] Falhas da CLI perdem a mensagem JSON na TUI

**Confiança:** Alta  
**Categoria:** Tratamento de erros  
**Localização:** `internal/tui/commands.go:23-33`; `internal/cobra/start.go:79-91`; `internal/cobra/install.go:30-37`  
**Descrição:** `cmd.Output()` descarta stderr; quando filho retorna JSON e código não zero, a TUI prioriza `exit status 1` sobre a resposta estruturada. `install --json` nem emite JSON em falha.  
**Cenário:** Tentar iniciar antes do setup ou falhar instalação.  
**Impacto:** Usuário perde instrução recuperável e recebe erro operacional genérico.  
**Correção:** Usar `CombinedOutput` ou pipes separados; decodificar JSON mesmo quando há exit code não zero; garantir contrato JSON de sucesso e erro em todos os subcomandos.  
**Teste:** `start --json` falhando deve mostrar a mensagem de setup pendente na TUI.

# Achados Baixos

## [L-01] Argumentos posicionais inesperados são ignorados

**Confiança:** Alta  
**Categoria:** Validação de entrada  
**Localização:** `internal/cobra/setup.go:14-22`; `internal/cobra/start.go:14-33`; `internal/cobra/shell.go:17-22`; `internal/cobra/status.go:21-26`; `internal/cobra/tui.go:9-20`  
**Correção:** Adicionar `Args: cobra.NoArgs` aos comandos sem argumentos.

## [L-02] Saída de logs não neutraliza sequências de controle de terminal

**Confiança:** Média  
**Categoria:** Terminal injection  
**Localização:** `internal/install/install.go:149-164`; `internal/cobra/logs.go:61-75`  
**Descrição:** Saídas de processos são gravadas e reimpressas sem filtragem.  
**Cenário:** Ferramenta ou pacote controlado escreve ANSI/OSC em stderr/stdout.  
**Correção:** Remover sequências de escape ao renderizar logs humanos ou exibir conteúdo codificado/seguro.

# Problemas De Concorrência

- Não foi identificada data race Go confirmada por inspeção.
- Há condição de corrida funcional na TUI: operações e snapshots podem concluir fora de ordem.
- Não há serialização de instalações em processos distintos; comandos duplicados podem concorrer pelo `apt` e sobrescrever o mesmo registro JSON.
- `waitForSSH` e `waitForPortClosed` respeitam `ctx.Done()`, mas o contexto raiz de vários comandos CLI não define deadline.
- Não há goroutine leak confirmado no código de status: as duas goroutines chamam `Done` e o `Wait` sincroniza o retorno.

# Dependências E Supply Chain

- `go.mod` possui dependências diretas restritas e sem `replace` suspeito.
- Há `govulncheck` agendado no CI e Dependabot semanal.
- Não foi possível executar `govulncheck`, `go mod verify` ou `go list -m -u all` no modo somente leitura.
- Workflows usam tags mutáveis de actions, inclusive no workflow de release com `contents: write`: `.github/workflows/release.yml:8-9`, `21-30`. Fixar actions por SHA é defesa em profundidade recomendada.
- `Dockerfile.termux:4,17` usa imagem e Air em `latest`; é ambiente de desenvolvimento, mas reduz reprodutibilidade e aumenta risco de supply chain.

# Lacunas De Testes

- Operações TUI simultâneas, refresh fora de ordem e saída durante operação.
- Cancelamento e timeout de subprocessos da TUI.
- Update interrompido entre os renames.
- Downloads lentos, infinitos e grandes.
- Rejeição de `LogPath` fora da raiz prevista.
- Limite de memória para logs extensos.
- Configuração SSH segura por padrão.
- Execução regular de `go test -race ./...` no CI.

# Plano Priorizado

1. **Corrigir imediatamente**
   - Tornar SSH seguro por padrão: loopback/chaves/`StrictModes yes`.
   - Assinar releases com chave pública fixada.
   - Adicionar recuperação atômica do binário após update interrompido.

2. **Antes do próximo release**
   - Serializar operações TUI e propagar cancelamento/timeouts.
   - Limitar downloads e leitura de logs.
   - Corrigir o contrato JSON de erros da CLI/TUI.

3. **Próximo ciclo**
   - Validar origem de `LogPath`.
   - Adicionar testes de concorrência, cancelamento e falhas de filesystem.
   - Fixar GitHub Actions por SHA.

4. **Preventivo**
   - Rodar `go test -race ./...`, `govulncheck ./...` e linter no CI.
   - Pin de imagem Docker e ferramentas de desenvolvimento.

# Comandos Executados

- `go version`: `go1.26.5 linux/amd64`.
- `go env`: ambiente Go local inspecionado.
- `golangci-lint --version`: disponível, versão `2.12.2`.
- Inspeção estática de código, testes, Docker, scripts e workflows.
- Não executei `go test`, `go vet`, `go mod verify`, `govulncheck` ou linters porque o modo solicitado estava estritamente somente leitura e essas ferramentas podem alterar caches ou baixar módulos.

# Limitações Da Auditoria

- Não houve validação dinâmica em Termux/Android/PRoot real.
- Não houve execução de detector de races, scanner de vulnerabilidades ou fuzzing.
- Não há banco de dados, HTTP server, JWT, criptografia de aplicação ou autorização por papéis no escopo atual.
- Não foram encontrados segredos versionados; nenhum valor sensível foi exposto neste relatório.

## Respostas Diretas

1. **Está seguro para produção?** Não, especialmente se o SSH ficar acessível além de uma rede local confiável.
2. **Três maiores riscos:** SSH por senha exposto na LAN; updater sem assinatura independente; update interrompido quebrando o executável.
3. **Bloqueador de release?** Sim: H-01 e H-02 para qualquer release que ofereça SSH remoto e autoatualização.
4. **Exploração realista?** Brute force/ataque em LAN contra SSH e comprometimento de release/update são realistas. Leitura arbitrária de logs exige adulteração local prévia.
5. **Correções primeiro:** endurecer SSH, autenticar releases, tornar update recuperável, depois corrigir cancelamento/concorrência TUI.
6. **Testes prioritários:** configuração SSH segura, assinatura inválida, interrupção durante update, download sem fim/grande, operações TUI sobrepostas e cancelamento de filho.

# Complemento De Validação Dinâmica

Esta seção complementa a auditoria original. Ela substitui a limitação anterior sobre ferramentas que ainda não haviam sido executadas. Nenhum dos patches abaixo foi aplicado ao código de produção.

## Resultados Das Ferramentas

| Comando | Resultado | Evidência / observação |
| --- | --- | --- |
| `go mod verify` | Passou | Todos os módulos no cache foram verificados. |
| `go mod tidy -diff` | Passou | Não sugeriu alteração em `go.mod` ou `go.sum`. |
| `go vet ./...` | Passou | Sem diagnósticos. |
| `go test -count=1 ./...` | Passou | Todos os pacotes testáveis passaram. |
| `go test -race ./...` | Passou | Nenhuma race foi detectada pelos cenários de teste existentes. Isto não invalida a corrida funcional M-01. |
| `go test -shuffle=on -count=1 ./...` | Passou | Não houve dependência de ordem detectada nesta execução. |
| `go test -coverprofile=coverage.out ./...` | Passou | Cobertura total de statements: 57,4%. |
| `golangci-lint run` | Falhou | 59 diagnósticos: 27 `errcheck`, 25 `staticcheck` e 7 `unused`. |
| `make integration-test` | Passou | Smoke test isolado no Docker completou setup, instalação, SSH por chave, stop e limpeza de PID obsoleto. |

`govulncheck`, `staticcheck` e `gosec` não estão instalados como executáveis independentes no ambiente. O `golangci-lint` disponível executou regras de `staticcheck`; não foi instalado software adicional apenas para esta auditoria.

## Cobertura E Lacunas Confirmadas

- Cobertura total: **57,4%**.
- `internal/workstation`: **35,3%**. Caminhos reais de processo, socket, lock, PTY e filesystem continuam pouco cobertos.
- `internal/cobra`: **12,8%**. Os fluxos públicos CLI e seu contrato JSON quase não possuem teste direto.
- `internal/tui/commands.go`: `runCommand` tem **8,3%** e `runStatusCommand` **6,7%**; os fluxos que originam M-01, M-02 e M-05 não têm proteção suficiente.
- `internal/update`: **66,0%**, porém faltam testes para timeout, corpo excessivo, interrupção entre renames e autenticidade de release.
- A integração Docker valida o userland simulado, mas não testa Termux:API, suspensão/encerramento pelo HyperOS, permissões Android ou PRoot real em um aparelho ARM64.

# Achados Adicionais

## [L-03] Erros de fechamento de recursos são ignorados em caminhos relevantes

**Severidade:** Baixa  
**Confiança:** Alta  
**Categoria:** Gerenciamento de recursos / tratamento de erro  
**Localização:** `internal/cobra/shell.go:60`; `internal/install/install.go:148,179-185`; `internal/update/update.go:188,232,285,287,322,344,371`; `internal/workstation/network.go:40`; `internal/workstation/ssh_config.go:64`  
**Descrição:** `golangci-lint` identificou 27 retornos ignorados. Alguns são limpeza best-effort aceitável, mas outros envolvem `Close` de arquivo temporário e PTY.  
**Causa raiz:** Uso de `defer Close()` sem decidir se uma falha de flush/fechamento deve ser propagada.  
**Cenário:** Falha de armazenamento no fechamento de arquivo temporário ou log; em particular, a operação pode reportar conclusão sem registrar integralmente a saída.  
**Impacto:** Principalmente observabilidade incompleta e tratamento impreciso de falhas de filesystem; não é vulnerabilidade de acesso remoto.  
**Correção recomendada:** Propagar erro de `Close` quando concluir escrita ou instalação depender dele; documentar explicitamente limpeza best-effort com `_ =`.  
**Patch sugerido:** Em `appendLog`, trocar o `defer file.Close()` por fechamento explícito no caminho de sucesso e incluir seu erro no retorno. Manter `_ = os.Remove(...)` somente para limpeza que não altera o resultado principal.  
**Teste de regressão:** Injetar writer/arquivo que falha no fechamento e verificar que a operação retorna erro quando a persistência é essencial.  
**Efeitos colaterais da correção:** Pode expor falhas antes silenciosas ao usuário, o que é desejável para arquivos de estado e update.

## [I-01] Código morto e APIs depreciadas reduzem sinal de manutenção

**Severidade:** Informativo  
**Confiança:** Alta  
**Categoria:** Qualidade / manutenção  
**Localização:** `internal/cobra/json.go:15`; `internal/install/install.go:129`; `internal/tui/components.go:41`; `internal/tui/styles.go:34,85,92`; `internal/tui/tui.go:553`; `internal/cobra/install.go:43`  
**Descrição:** O linter reportou sete símbolos não usados e uso de `strings.Title`, depreciado desde Go 1.18. Há também chamadas `lipgloss.Style.Copy` depreciadas pela versão atual da biblioteca.  
**Impacto:** Não há exploração confirmada; aumenta ruído do linter e dificulta identificar regressões reais.  
**Correção recomendada:** Remover símbolos mortos; substituir `strings.Title` por capitalização explícita do catálogo, já que os nomes aceitos são conhecidos, ou por alternativa Unicode adequada. Atualizar chamadas Lip Gloss conforme a API atual.  
**Teste de regressão:** Cobrir a renderização de linguagem com entrada do catálogo e com caracteres Unicode se a entrada deixar de ser restrita.  
**Efeitos colaterais da correção:** Nenhum efeito funcional esperado se as remoções forem precedidas por busca de consumidores.

**Status:** Corrigido em 25/07/2026. Foram removidos símbolos mortos, substituídas APIs depreciadas e propagados erros de escrita relevantes. `golangci-lint run` agora conclui sem diagnósticos.

## [I-02] Atualizações de dependências disponíveis exigem triagem, não atualização cega

**Severidade:** Informativo  
**Confiança:** Alta  
**Categoria:** Dependências  
**Descrição:** `go list -m -u all` informou versões mais recentes para dependências transitivas como `github.com/spf13/pflag`, `golang.org/x/sync`, `github.com/mattn/go-runewidth` e `github.com/charmbracelet/ultraviolet`. Isso não confirma vulnerabilidade ou necessidade de atualização imediata.  
**Correção recomendada:** Executar `govulncheck ./...` no CI e avaliar cada atualização por changelog, compatibilidade com Bubble Tea/Lip Gloss e impacto no Android/Termux.  
**Efeitos colaterais da correção:** Atualizações transitivas podem modificar comportamento de terminal; validar no aparelho real antes de release.

# Patches Mínimos Propostos

Os trechos abaixo são direcionais, não patches aplicados. Eles devem ser refinados e testados antes de alterar o comportamento público.

## H-01: Defaults SSH Seguros

```go
fmt.Fprintf(&builder, "ListenAddress 127.0.0.1\nPidFile %s\n", p.SSHPID())
builder.WriteString("PasswordAuthentication no\n")
builder.WriteString("KbdInteractiveAuthentication no\n")
builder.WriteString("PermitEmptyPasswords no\nStrictModes yes\n")
```

Uma opção explícita, confirmada pelo usuário, deve habilitar LAN e senha. O teste precisa verificar que nenhuma configuração padrão contém `0.0.0.0`, `PasswordAuthentication yes` ou `StrictModes no`.

## H-02: Autenticidade De Release

```text
release assets:
  SHA256SUMS
  SHA256SUMS.minisig

updater:
  download SHA256SUMS e SHA256SUMS.minisig com limites
  verificar assinatura contra chave pública compilada no binário
  só então extrair o hash esperado e baixar o executável
```

O workflow de release deve assinar o manifesto com credencial protegida fora do repositório. A chave pública pode ser distribuída no código; a privada nunca.

## H-03: Recuperação Após Troca Interrompida

```go
func recoverInterruptedUpdate(installPath string) error {
	backup := installPath + ".bak"
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		if _, backupErr := os.Stat(backup); backupErr == nil {
			return os.Rename(backup, installPath)
		}
	}
	return nil
}
```

Executar a recuperação antes de substituir o binário e antes de iniciar a operação normal. O backup só deve ser removido quando a nova versão puder ser iniciada e validada por estratégia definida.

## M-01 e M-02: Operação TUI Serializada E Cancelável

```go
if m.busy {
	return m, nil
}
m.operationID++
operationID := m.operationID
m.busy = true
return m, m.backend.OperationCmd(operationID, args...)
```

As mensagens de status e operação devem carregar o identificador correspondente. A TUI deve ignorar mensagens cujo identificador seja anterior ao request corrente. O backend real deve usar `exec.CommandContext` com contexto cancelado no encerramento da TUI e timeout específico por operação.

# Separação De Ambientes

| Área | Docker de desenvolvimento | Termux/Android real |
| --- | --- | --- |
| Setup, PRoot e pacote Ubuntu | Coberto pelo smoke test | Deve ser repetido em aparelho ARM64. |
| SSH por chave e stop | Coberto em loopback no container | Exposição LAN, políticas Wi-Fi e permissões Android não cobertas. |
| Termux:API | Não reproduzido pelo Docker | Requer Termux:API instalado e permissões concedidas. |
| Suspensão e encerramento | Não reproduzido | Requer teste de background, economia de bateria e HyperOS. |
| Atualização GitHub | Não coberta pela integração | Requer teste com release assinada, rede lenta e interrupção real. |

# Plano Atualizado

1. Corrigir H-01, H-02 e H-03 antes de oferecer SSH remoto ou autoatualização como funcionalidade de produção.
2. Corrigir M-01, M-02 e M-03 antes do próximo release da TUI.
3. Adicionar testes de integração para cancelamento, falha de troca de binário, limites HTTP e estado JSON malicioso.
4. Resolver L-03 e remover código morto para tornar `golangci-lint run` verde.
5. Adicionar ao CI: `go test -race ./...`, cobertura com mínimo acordado, `govulncheck ./...` e actions fixadas por SHA.
