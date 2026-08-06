# Goal: Suporte JVM para Java 21, Kotlin e Spring Boot 4

## Status

- Status: in_progress
- Project root: /home/erick/code/projs/mobdesk
- Base branch: main
- Branch policy: feat/jvm-spring-boot (uma branch para o Goal inteiro)
- Worktree policy: usar o worktree atual
- PR policy: um PR ao final do Goal
- Commit policy: um commit por Stage

## Objetivo

Adicionar ao Mobdesk um ambiente JVM funcional dentro do Ubuntu via PRoot-Distro,
com Java 21 como JDK gerenciado, Kotlin/JVM, Maven e Gradle como apps opcionais
independentes, e compatibilidade validada com Spring Boot 4.x.

O Mobdesk fornecerá o ambiente e os toolchains. Não criará, gerenciará ou
hospedará projetos Spring Boot como uma funcionalidade própria neste Goal.

## Escopo

- Adicionar Java 21 ao catálogo de apps.
- Configurar `java`, `javac`, `jar` e `JAVA_HOME` dentro do Ubuntu.
- Adicionar Kotlin/JVM com compiler oficial em versão fixada e checksum validado.
- Adicionar Gradle como app opcional dependente de Java.
- Adicionar Maven como app opcional dependente de Java.
- Manter Gradle e Maven como instalações separadas.
- Preferir `./gradlew` e `./mvnw` quando um projeto fornecer wrapper.
- Usar os binários globais como fallback.
- Evoluir o catálogo para múltiplos pacotes e executáveis obrigatórios.
- Permitir dependências compartilhadas entre apps.
- Impedir a remoção de Java enquanto Kotlin, Gradle ou Maven dependerem dele.
- Detectar estados instalado, parcial, ausente e conflitante corretamente.
- Aplicar aviso de armazenamento abaixo de 20 GB livres.
- Bloquear instalações abaixo de 10 GB livres.
- Manter caches Gradle e Maven persistentes.
- Validar Java com Spring Boot 4.x usando Gradle.
- Validar Java com Spring Boot 4.x usando Maven.
- Validar Kotlin com Spring Boot 4.x usando Gradle.
- Validar compilação, testes, empacotamento e execução HTTP de aplicações Spring Boot.
- Preservar o limite Termux como host de controle e Ubuntu como ambiente de desenvolvimento.
- Atualizar contratos JSON de forma aditiva e compatível.
- Atualizar documentação de arquitetura, decisões, roadmap e uso.
- Manter funcionamento por CLI, JSON, TUI, teclado e mouse.

## Fora do escopo

- Gerenciamento automático de Java 17.
- Instalação automática de SDKMAN.
- Kotlin/Native e Kotlin/JS.
- Android SDK, Android Studio ou Gradle Android Plugin.
- Criação de projetos Spring Boot pelo Mobdesk.
- Gerenciamento de projetos, sessões ou serviços Spring Boot.
- Tela específica de projetos.
- Spring Framework 4.x sem Spring Boot.
- Spring Boot CLI.
- Docker, Testcontainers ou execução dependente de Docker.
- systemd, cgroups, namespaces completos ou execução como serviço do sistema.
- Native Image, GraalVM e compilação nativa.
- Remoção automática dos caches `~/.gradle` e `~/.m2`.
- Alterações em `site/` ou `docs/LAUNCH-KIT.md`.
- Garantia de compatibilidade com versões arbitrárias de Java, Kotlin, Gradle, Maven ou Spring.

## Regras de execução

- Java 21 é o único JDK instalado e gerenciado automaticamente pelo Mobdesk.
- Java 17 manual é permitido, mas não será substituído, removido ou administrado pelo Mobdesk.
- Java, Kotlin, Gradle e Maven são perfis independentes na lista de apps.
- Kotlin, Gradle e Maven declaram Java 21 como dependência.
- A instalação de um dependente pode instalar Java 21 automaticamente.
- A desinstalação deve proteger todos os pacotes compartilhados.
- A instalação ocorre no Ubuntu via `proot-distro`, nunca no Termux.
- O JDK do Termux não deve ser usado como JVM do Ubuntu.
- `JAVA_HOME` deve ser descoberto dentro do Ubuntu sem caminho fixo de arquitetura.
- O ambiente exportado para shells e comandos deve incluir `JAVA_HOME` e o `PATH` correto.
- Downloads externos usam HTTPS, versão fixada, arquitetura validada quando aplicável e checksum.
- O compiler Kotlin é instalado em caminho privado e gerenciado pelo Mobdesk.
- Gradle Wrapper e Maven Wrapper têm precedência sobre binários globais.
- Caches `~/.gradle` e `~/.m2` sobrevivem a reinstalações e desinstalações.
- O espaço livre é verificado antes de qualquer instalação.
- Abaixo de 20 GB livres, a operação mostra aviso e exige confirmação compatível.
- Abaixo de 10 GB livres, a instalação é bloqueada sem apagar dados.
- Toda operação longa aceita cancelamento e preserva estado, logs e dados após falha.
- A TUI não executa `apt`, `proot-distro`, Gradle, Maven ou scripts diretamente.
- Instalação e atualização continuam bloqueadas dentro de SSH/Ubuntu.
- Novos campos JSON são aditivos e preservam o contrato versionado.
- Estados parciais, bloqueados, ocupados, conflitantes, concluídos e falhos são visíveis.
- A interface funciona em terminais estreitos e mantém equivalência mouse/teclado.
- Nenhum teste depende de secrets.
- A documentação declara as limitações reais do PRoot.

## Tasks

### Task 1 - Evoluir o contrato de catálogo e instalações compartilhadas

- Status: in_progress
- Depends on: none
- Branch: branch única do Goal
- Worktree: worktree atual
- PR: incluída no PR final

#### Stage 1.1 - Modelar múltiplos pacotes e executáveis

- Status: completed

##### Objetivo

Permitir vários pacotes e executáveis obrigatórios sem quebrar registros JSON existentes.

##### Ações

- Evoluir `AppProfile` e registros para múltiplos pacotes e executáveis.
- Manter leitura dos campos antigos `package` e `executable`.
- Definir agregação de versões e erros.
- Atualizar o catálogo existente sem alterar seu comportamento.

##### Critérios de aceite

- [ ] Registros antigos continuam sendo lidos.
- [x] O modelo aceita declarações de múltiplos pacotes e executáveis para Java, Kotlin, Gradle e Maven.
- [x] Perfil só é instalado quando todos os executáveis obrigatórios existem.
- [x] Testes cobrem registros antigos e múltiplos executáveis.

##### Validação

```bash
go test ./internal/install ./internal/status
go vet ./internal/install ./internal/status
```

Resultado: testes de install/status e vet passaram; `git diff --check` passou.

##### Commit

```text
feat: extend toolchain installation contract
```

Commit: c4a0a09

#### Stage 1.2 - Proteger dependências compartilhadas na desinstalação

- Status: completed

##### Objetivo

Impedir que a remoção de um app JVM desinstale Java ou pacote ainda utilizado.

##### Ações

- Comparar todos os pacotes e dependências diretas e transitivas.
- Preservar pacotes compartilhados.
- Registrar pacotes removidos e preservados.
- Representar estados parciais e cobrir operações repetidas.

##### Critérios de aceite

- [x] Pacotes compartilhados e dependências de perfil bloqueiam a remoção.
- [x] A regra protege Java quando usado por um perfil dependente.
- [x] Caches e dados do usuário não são removidos.
- [x] Arquivos modificados continuam protegidos.

##### Validação

```bash
go test ./internal/install -run 'Test.*[Uu]ninstall|Test.*[Ss]hared|Test.*[Dd]ep'
go vet ./internal/install
```

Resultado: testes de desinstalação, compartilhamento e dependências, vet e `git diff --check` passaram.

Commit: fe96f0f

##### Commit

```text
feat: protect shared toolchain dependencies
```

#### Stage 1.3 - Aplicar política global de armazenamento

- Status: completed

##### Objetivo

Impedir que instalações consumam o armazenamento crítico do dispositivo.

##### Ações

- Definir aviso em 20 GB e bloqueio em 10 GB.
- Aplicar a política a toda instalação do Mobdesk.
- Expor motivo em texto e JSON.
- Verificar espaço antes de modificar o sistema.
- Adicionar testes para estados suficiente, aviso e bloqueado.

##### Critérios de aceite

- [x] 25 GB livres prossegue sem aviso.
- [x] 19 GB livres informa aviso.
- [x] 10 GB livres ou mais não é bloqueado somente pela política.
- [x] Menos de 10 GB bloqueia sem executar APT, downloads ou remoções.
- [x] Bloqueio aparece no resultado da instalação e no JSON.
- [x] A política vale para apps existentes e novos.

##### Validação

```bash
go test ./internal/install ./internal/status
go vet ./internal/install ./internal/status
```

Resultado: testes de limites, install/status/cobra/i18n, vet, `i18n-check` e `git diff --check` passaram.

##### Commit

```text
feat: enforce global storage thresholds
```

Commit: e89d03b

### Task 2 - Implementar o ambiente Java 21

- Status: in_progress
- Depends on: Task 1
- Branch: branch única do Goal
- Worktree: worktree atual
- PR: incluída no PR final

#### Stage 2.1 - Adicionar o perfil Java 21

- Status: completed

##### Objetivo

Instalar e validar o JDK 21 oficial do Ubuntu ARM64 dentro do PRoot.

##### Ações

- Adicionar perfil `java` usando `openjdk-21-jdk` via APT no Ubuntu.
- Declarar pacotes e executáveis obrigatórios.
- Validar `java --version`, `javac --version` e `jar --version`.
- Registrar pacotes, versão e logs sem secrets.
- Manter instalação idempotente.

##### Critérios de aceite

- [x] `mobdesk install java` funciona em Termux.
- [x] APT ocorre somente dentro do Ubuntu.
- [x] Java 21 é encontrado e validado.
- [x] `javac` compila e `jar` empacota uma classe simples.
- [x] Segunda instalação não reinstala desnecessariamente.
- [x] Perfil aparece na CLI, JSON e TUI.

##### Validação

```bash
go test ./internal/install
go vet ./internal/install
```

Resultado: testes de install, vet, localização, smoke test de catálogo e `git diff --check` passaram. O fixture confirmou Java 21, `javac`, `jar`, compilação e execução do JAR.

##### Commit

```text
feat: add managed Java 21 profile
```

#### Stage 2.2 - Configurar JAVA_HOME e ambiente de shell

- Status: completed

##### Objetivo

Garantir que shells, Mobdesk e build tools usem a mesma JVM dentro do Ubuntu.

##### Ações

- Descobrir o diretório real do JDK.
- Gerar configuração privada e idempotente do shell Ubuntu.
- Exportar `JAVA_HOME` e ajustar `PATH`.
- Não transportar ambiente do Termux.
- Preservar configuração existente do usuário.

##### Critérios de aceite

- [x] `JAVA_HOME` aponta para o JDK 21 no Ubuntu.
- [x] `java` e `javac` apontam para o ambiente Ubuntu.
- [x] Gradle e Maven recebem o mesmo `JAVA_HOME`.
- [x] Configuração é repetível e não apaga configuração existente.
- [x] Caminho não depende de arquitetura codificada.

##### Validação

```bash
go test ./internal/install ./internal/workstation
go vet ./internal/install ./internal/workstation
```

Resultado: configuração de shell, preservação do `.bashrc`, descoberta dinâmica do JDK e ausência de ambiente Termux foram cobertas por testes; testes e vet passaram.

##### Commit

```text
feat: configure Ubuntu Java environment
```

### Task 3 - Implementar Kotlin/JVM fixado

- Status: pending
- Depends on: Task 2
- Branch: branch única do Goal
- Worktree: worktree atual
- PR: incluída no PR final

#### Stage 3.1 - Instalar compiler oficial Kotlin

- Status: pending

##### Objetivo

Disponibilizar Kotlin/JVM atual sem usar o pacote Ubuntu obsoleto.

##### Ações

- Definir versão compatível com Spring Boot 4.x.
- Fixar URL, versão e checksum.
- Validar HTTPS, arquitetura e integridade.
- Instalar em diretório privado e expor `kotlinc` e `kotlin`.
- Declarar Java 21 como dependência.
- Registrar arquivos e hashes para remoção segura.

##### Critérios de aceite

- [ ] `mobdesk install kotlin` instala Java 21 se necessário.
- [ ] Não usa `apt install kotlin`.
- [ ] Checksum inválido interrompe sem ativar arquivos.
- [ ] `kotlinc --version` funciona.
- [ ] Kotlin compila JAR executável executado por `java -jar`.
- [ ] Segunda instalação é idempotente.
- [ ] Arquivos modificados não são removidos silenciosamente.

##### Validação

```bash
go test ./internal/install
go vet ./internal/install
```

##### Commit

```text
feat: add pinned Kotlin JVM compiler
```

#### Stage 3.2 - Validar Kotlin com Java 21

- Status: pending

##### Objetivo

Confirmar interoperabilidade básica entre Kotlin/JVM, Java 21 e PRoot.

##### Ações

- Adicionar fixture Kotlin console.
- Compilar e executar com `kotlinc` e `java -jar`.
- Validar caminhos, temporários e saída em ARM64.
- Cobrir cancelamento e falha de download.

##### Critérios de aceite

- [ ] Fixture compila e executa no Ubuntu ARM64 via PRoot.
- [ ] Fixture usa Java 21 do Ubuntu.
- [ ] Falha de JDK ausente produz diagnóstico objetivo.
- [ ] Teste não exige IDE, Android SDK ou Kotlin/Native.

##### Validação

```bash
make catalog-test
```

##### Commit

```text
test: validate Kotlin JVM toolchain
```

### Task 4 - Adicionar Maven e Gradle como apps opcionais

- Status: pending
- Depends on: Task 2
- Branch: branch única do Goal
- Worktree: worktree atual
- PR: incluída no PR final

#### Stage 4.1 - Adicionar perfis independentes de Gradle e Maven

- Status: pending

##### Objetivo

Disponibilizar os dois build tools sem criar dependência entre eles.

##### Ações

- Adicionar `gradle` dependente de Java.
- Adicionar `maven` dependente de Java.
- Definir estratégias e versões verificáveis.
- Validar `gradle --version` e `mvn --version`.
- Manter caches persistentes e perfis opcionais.

##### Critérios de aceite

- [ ] Gradle e Maven aparecem como apps opcionais.
- [ ] Instalar Gradle não instala Maven e vice-versa.
- [ ] Ambos instalam Java 21 somente quando necessário.
- [ ] Ambos mostram a versão do Java usado.
- [ ] Desinstalação preserva Java enquanto houver dependentes.
- [ ] TUI mostra dependências sem subconfiguração dentro do Java.

##### Validação

```bash
go test ./internal/install ./internal/status ./internal/tui
go vet ./internal/install ./internal/status ./internal/tui
```

##### Commit

```text
feat: add optional Maven and Gradle profiles
```

#### Stage 4.2 - Respeitar wrappers de projetos

- Status: pending

##### Objetivo

Garantir que projetos existentes controlem suas próprias versões.

##### Ações

- Dar precedência a `./gradlew` sobre `gradle`.
- Dar precedência a `./mvnw` sobre `mvn`.
- Validar wrappers com `JAVA_HOME` configurado.
- Não alterar arquivos de projeto automaticamente.
- Definir diagnóstico para wrapper ausente, inválido ou sem permissão.

##### Critérios de aceite

- [ ] Wrappers têm precedência sobre binários globais.
- [ ] Wrapper inválido produz diagnóstico objetivo.
- [ ] Cancelamento não deixa estado inconsistente.
- [ ] Documentação explica o fallback global.

##### Validação

```bash
go test ./internal/install ./internal/executil
go vet ./internal/install ./internal/executil
```

##### Commit

```text
feat: document project build wrapper precedence
```

### Task 5 - Integrar status, JSON e TUI

- Status: pending
- Depends on: Task 1, Task 2, Task 3, Task 4
- Branch: branch única do Goal
- Worktree: worktree atual
- PR: incluída no PR final

#### Stage 5.1 - Representar múltiplos executáveis e estados parciais

- Status: pending

##### Objetivo

Evitar que perfis incompletos apareçam como instalados.

##### Ações

- Verificar todos os executáveis obrigatórios.
- Diferenciar ferramenta gerenciada de ferramenta externa.
- Expor executáveis ausentes e dependências bloqueadas.
- Atualizar estados parciais e alertas.
- Reconciliar registros antigos.

##### Critérios de aceite

- [ ] Java sem `javac` não é completamente instalado.
- [ ] Kotlin sem `kotlinc` aparece como parcial ou ausente.
- [ ] Gradle sem Java mostra dependência faltante.
- [ ] Ferramentas externas não são removíveis pelo Mobdesk.
- [ ] JSON mantém schema compatível.
- [ ] Status respeita a fronteira Termux/SSH.

##### Validação

```bash
go test ./internal/status ./internal/install
go vet ./internal/status ./internal/install
```

##### Commit

```text
feat: report complete toolchain states
```

#### Stage 5.2 - Atualizar catálogo visual e localizações

- Status: pending

##### Objetivo

Apresentar Java, Kotlin, Gradle e Maven na lista e nos popups existentes.

##### Ações

- Adicionar descrições e usos em inglês e português.
- Mostrar dependências, bloqueios e estados parciais.
- Preservar ações, confirmação, mouse, teclado e terminais estreitos.

##### Critérios de aceite

- [ ] Quatro apps aparecem na lista.
- [ ] Gradle e Maven são instaláveis separadamente.
- [ ] Falta de espaço mostra bloqueio na TUI.
- [ ] App parcial não oferece ação incompatível sem explicação.
- [ ] Nenhum popup mostra help ou saída bruta.
- [ ] `i18n-check` passa.

##### Validação

```bash
go test ./internal/tui ./internal/i18n
go vet ./internal/tui ./internal/i18n
./scripts/i18n-check.sh
```

##### Commit

```text
feat: expose JVM tools in the TUI
```

### Task 6 - Validar Spring Boot 4.x no Ubuntu ARM64

- Status: pending
- Depends on: Task 3, Task 4, Task 5
- Branch: branch única do Goal
- Worktree: worktree atual
- PR: incluída no PR final

#### Stage 6.1 - Criar fixtures de Spring Boot 4.x

- Status: pending

##### Objetivo

Validar o ambiente em aplicações reais sem criar um gerenciador de projetos.

##### Ações

- Criar fixture Java com Gradle.
- Criar fixture Java com Maven.
- Criar fixture Kotlin com Gradle.
- Fixar versão exata de Spring Boot 4.x.
- Criar endpoint HTTP mínimo em porta não privilegiada.
- Não adicionar Docker, Testcontainers ou banco externo.

##### Critérios de aceite

- [ ] As três fixtures compilam.
- [ ] Os testes passam.
- [ ] JARs executáveis são gerados.
- [ ] Versões não dependem de `latest`.
- [ ] Fixtures não dependem de IDE.

##### Validação

```bash
make catalog-test
```

##### Commit

```text
test: add Spring Boot JVM fixtures
```

#### Stage 6.2 - Executar e validar aplicações Spring

- Status: pending

##### Objetivo

Confirmar aplicações Spring Boot como processos de usuário dentro do PRoot.

##### Ações

- Executar Java/Gradle com `bootRun`.
- Executar Java/Maven com `spring-boot:run`.
- Executar JARs Java e Kotlin.
- Verificar HTTP local e configuração de endereço/porta.
- Testar cancelamento e registrar limitações de memória, file watching e suspensão.

##### Critérios de aceite

- [ ] Java/Gradle inicia e responde HTTP.
- [ ] Java/Maven inicia e responde HTTP.
- [ ] Kotlin/Gradle inicia e responde HTTP.
- [ ] Processos usam porta não privilegiada.
- [ ] Processos podem ser encerrados sem órfãos.
- [ ] Validação não assume systemd, Docker ou cgroups.
- [ ] Falhas de rede, memória e cancelamento têm diagnóstico utilizável.

##### Validação

```bash
make catalog-test
make integration-test
```

##### Commit

```text
test: validate Spring Boot runtime in PRoot
```

### Task 7 - Documentar, testar e fechar o Goal

- Status: pending
- Depends on: Task 1, Task 2, Task 3, Task 4, Task 5, Task 6
- Branch: branch única do Goal
- Worktree: worktree atual
- PR: incluída no PR final

#### Stage 7.1 - Atualizar documentação e decisões

- Status: pending

##### Objetivo

Documentar toolchains, decisões e limitações de compatibilidade.

##### Ações

- Atualizar `README.md` e `README.pt-BR.md`.
- Atualizar `docs/ARCHITECTURE.md`, `docs/DECISIONS.md` e `docs/ROADMAP.md`.
- Documentar fronteira Termux/Ubuntu, Java 17 manual, wrappers, caches e limites de armazenamento.
- Documentar limitações de Spring Boot no PRoot.

##### Critérios de aceite

- [ ] Documentação não afirma que JDK Termux é usado pelo Ubuntu.
- [ ] Ambiente JVM é diferenciado de gerenciamento de projetos.
- [ ] Maven e Gradle aparecem como apps opcionais.
- [ ] Spring Boot 4.x e as exclusões estão explícitos.
- [ ] Limites de 20 GB e 10 GB estão documentados.

##### Validação

```bash
./scripts/i18n-check.sh
```

##### Commit

```text
docs: document JVM and Spring Boot support
```

#### Stage 7.2 - Validação final e regressão

- Status: pending

##### Objetivo

Garantir que o suporte JVM não quebre o catálogo nem os fluxos existentes.

##### Ações

- Executar testes unitários, vet, localização, build, catálogo e integração.
- Validar instalações repetidas, desinstalações compartilhadas, JSON e estados bloqueados.
- Validar no Termux ARM64 real do POCO F6.
- Registrar limitações dependentes do dispositivo.

##### Critérios de aceite

- [ ] `make check` passa.
- [ ] `make catalog-test` passa.
- [ ] `make integration-test` passa.
- [ ] Java, Kotlin, Gradle e Maven passam as validações previstas.
- [ ] Matriz Spring reduzida passa.
- [ ] Catálogo existente continua funcionando.
- [ ] Instalações repetidas são idempotentes.
- [ ] Java não é removido prematuramente.
- [ ] Aviso e bloqueio de armazenamento funcionam.
- [ ] `git diff --check` passa.

##### Validação

```bash
make check
make catalog-test
make integration-test
```

##### Commit

```text
test: verify JVM toolchain integration
```

## Ordem de execução

1. Task 1 / Stage 1.1
2. Task 1 / Stage 1.2
3. Task 1 / Stage 1.3
4. Task 2 / Stage 2.1
5. Task 2 / Stage 2.2
6. Task 3 / Stage 3.1
7. Task 3 / Stage 3.2
8. Task 4 / Stage 4.1
9. Task 4 / Stage 4.2
10. Task 5 / Stage 5.1
11. Task 5 / Stage 5.2
12. Task 6 / Stage 6.1
13. Task 6 / Stage 6.2
14. Task 7 / Stage 7.1
15. Task 7 / Stage 7.2

## Bloqueios e decisões

- O alvo é Spring Boot 4.x, não Spring Framework 4.x isolado.
- Java 21 é o único JDK gerenciado automaticamente.
- Java 17 manual não será administrado pelo Mobdesk.
- Kotlin é Kotlin/JVM com compiler oficial fixado.
- Kotlin/Native fica fora deste Goal.
- Maven e Gradle são apps opcionais independentes.
- Ambos dependem de Java 21, mas não dependem um do outro.
- Wrappers de projeto têm precedência sobre instalações globais.
- O Mobdesk fornece o ambiente, mas não cria nem gerencia projetos.
- A matriz Spring é Java/Gradle, Java/Maven e Kotlin/Gradle.
- Docker, Testcontainers, systemd, Native Image e serviços persistentes não são critérios de aceite.
- Aviso global ocorre abaixo de 20 GB livres.
- Instalações são bloqueadas abaixo de 10 GB livres.
- Caches Maven e Gradle são preservados.
- Validação real depende do POCO F6 e da rede disponível.
- Nenhuma implementação de produto ocorre durante a definição deste Goal.

## Conclusão

O Goal termina quando o Mobdesk instala e valida Java 21, Kotlin/JVM, Maven e
Gradle dentro do Ubuntu via PRoot, representa corretamente dependências e estados
compartilhados, aplica os limites de armazenamento, expõe os apps pela CLI/JSON/TUI,
executa a matriz reduzida de aplicações Spring Boot 4.x, documenta as limitações e
passa toda a validação automatizada e real disponível.
