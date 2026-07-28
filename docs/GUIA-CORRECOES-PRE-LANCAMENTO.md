# Guia De Correcoes Pre-Lancamento

## Objetivo

Este guia descreve como eliminar quatro riscos antes da divulgacao publica do Mobdesk.

Os temas sao autenticidade de update, recuperacao de update, configuracao SSH e fronteira Termux/Ubuntu.

Ele complementa `AUDITORIA-SEGURANCA-2026-07-25.md` e `AUTENTICACAO-SSH.md`.

O foco e uma implementacao pequena, testavel e adequada a Termux, Android e PRoot.

## Ordem Recomendada

1. Corrigir a deteccao de ambiente Termux/Ubuntu.
2. Tornar a troca do binario recuperavel.
3. Assinar os artefatos de release.
4. Definir e aplicar uma politica SSH segura por padrao.

A primeira correcao evita acoes no ambiente errado sem alterar o fluxo do usuario.

A segunda evita que uma falha deixe o comando indisponivel.

A terceira protege a cadeia de distribuicao do binario.

A quarta exige uma decisao de produto, pois altera a experiencia do primeiro acesso remoto.

## 1. Autenticidade Das Atualizacoes

### Problema

O updater baixa o binario e `SHA256SUMS` da mesma release GitHub.

O SHA-256 detecta dano acidental no download, mas nao prova quem publicou o arquivo.

Uma conta, token ou release comprometida pode conter um binario malicioso e seu hash correspondente.

O Mobdesk entao instala esse binario com as permissoes do usuario Termux.

### Decisao Proposta

Usar Minisign para assinar o manifesto de hashes das releases.

Minisign e pequeno, tem um modelo simples de chave publica e atende um binario unico.

A chave publica deve ser compilada no Mobdesk, nunca baixada junto da release.

A chave privada deve ficar fora do repositorio e fora dos logs de CI.

### Assets Da Release

Cada release deve publicar estes tres arquivos:

```text
mobdesk-linux-arm64
SHA256SUMS
SHA256SUMS.minisig
```

O manifesto deve listar somente hashes SHA-256 e nomes de arquivos da release.

A assinatura deve ser criada sobre o conteudo exato de `SHA256SUMS`.

Nao assine cada binario separadamente se um unico manifesto assinado atender todos os targets.

### Custodia Da Chave

Gere a chave em maquina confiavel, de preferencia offline para a criacao inicial.

Armazene a chave privada em cofre de segredos do provedor de CI ou use assinatura manual controlada.

Proteja a chave com senha e mantenha backup seguro fora do GitHub.

Registre o fingerprint da chave publica em `docs/DECISOES.md`.

Nao permita que pull requests tenham permissao de acessar a chave de assinatura.

Revogue a chave e publique nova chave somente por um processo documentado de emergencia.

### Mudancas No Codigo

Crie um valor constante com a chave publica Minisign no pacote `internal/update`.

Adicione `SignatureName` a `update.Options`, com `SHA256SUMS.minisig` como default.

Baixe manifesto e assinatura usando o mesmo cliente HTTP com limites de tamanho.

Verifique a assinatura antes de extrair qualquer hash do manifesto.

Recuse assinatura ausente, invalida, de chave diferente ou com formato inesperado.

Somente apos essa verificacao extraia o hash do asset solicitado.

Mantenha a verificacao SHA-256 do binario depois do download.

### Dependencia Ou Implementacao

Avalie uma biblioteca Go pequena e mantida que verifica o formato Minisign.

Prefira verificacao em Go ao binario externo `minisign`, pois ele nao existe por padrao no Termux.

Antes de adicionar dependencia, valide licenca, manutencao e compatibilidade ARM64.

Se nenhuma biblioteca atender, implemente somente a verificacao minima do formato documentado.

Nao invente criptografia nem aceite chaves ou algoritmos fornecidos pela rede.

### Workflow De Release

O workflow deve compilar o binario reproduzivel para `linux/arm64`.

Ele deve calcular `SHA256SUMS` a partir dos arquivos gerados naquela execucao.

Ele deve assinar o manifesto depois de gerar todos os hashes.

Ele deve anexar binario, manifesto e assinatura a uma release nao draft.

O workflow deve falhar se qualquer asset ou assinatura estiver ausente.

Fixe as GitHub Actions por SHA para reduzir risco na automacao de release.

### Testes

Teste uma assinatura valida criada com uma chave de fixture nao produtiva.

Teste manifesto sem assinatura, assinatura invalida e assinatura de outra chave.

Teste hash valido em manifesto nao assinado e confirme que ele e recusado.

Teste assinatura valida cujo manifesto nao contem o asset requisitado.

Teste que o binario nao e aberto nem instalado quando a assinatura falha.

### Criterio De Aceite

`mobdesk update` instala somente binarios cobertos por manifesto assinado pela chave embutida.

Uma release adulterada com checksum correspondente falha antes da troca do executavel.

A mensagem de erro informa que a autenticidade da release nao foi validada.

## 2. Atualizacao Recuperavel

### Problema

Uma atualizacao pode ser interrompida por falta de armazenamento, rede, encerramento pelo Android ou falha de filesystem.

O Mobdesk nunca deve remover seu unico binario funcional antes de ter uma alternativa validada.

O atual `recoverInterruptedUpdate` reconhece um `.bak`, mas o fluxo deve criar e administrar esse backup de forma consistente.

Uma troca atomica reduz a janela de falha, mas nao substitui validacao do binario recebido.

### Estado Proposto

Use tres caminhos no mesmo diretorio do binario instalado.

```text
mobdesk          binario em uso
mobdesk.bak      ultima versao funcional
.mobdesk-new-*   download temporario
```

Arquivos no mesmo diretorio permitem `rename` atomico no mesmo filesystem.

Nunca escreva o download diretamente no caminho do executavel ativo.

### Fluxo De Aplicacao

1. Recuperar `.bak` se o binario principal estiver ausente.
2. Baixar o novo binario em arquivo temporario com permissoes `0600`.
3. Verificar assinatura, hash, tamanho e fechamento do arquivo temporario.
4. Alterar o temporario para `0755` somente depois de validado.
5. Executar o temporario com `version --json` e prazo curto.
6. Renomear o binario atual para `.bak`.
7. Renomear o temporario para o caminho final.
8. Executar o novo binario com `version --json` e validar versao esperada.
9. Manter `.bak` ate uma proxima inicializacao ou politica de limpeza segura.

Se o passo 7 ou 8 falhar, restaure `.bak` imediatamente.

Se o processo morrer entre os passos, a proxima execucao deve restaurar o backup automaticamente.

### Validacao Do Binario

O autoteste deve executar o caminho absoluto do temporario, nao `mobdesk` pelo `PATH`.

Use um contexto com timeout pequeno, por exemplo dez segundos.

Exija codigo zero e JSON valido contendo a versao compilada.

Compare a versao retornada com a tag da release, respeitando o formato usado pelo pacote `version`.

Nao considere apenas permissao executavel ou tamanho de arquivo como validacao suficiente.

### Recuperacao No Inicio

Chame `recoverInterruptedUpdate` antes de atualizar e antes de executar caminhos que dependam do binario instalado.

Se o executavel existe, nunca o substitua automaticamente pelo `.bak` antigo.

Se ambos existem, preserve ambos e limpe backup apenas por politica conhecida.

Se o principal nao existe e o backup existe, renomeie backup de volta e informe recuperacao na proxima saida humana.

Se nenhum existe, retorne erro claro com instrucoes de reinstalacao via `go install`.

### Espaco Em Disco

O processo precisa de espaco para temporario, binario atual e backup simultaneamente.

Antes do download, consulte espaco livre quando a plataforma permitir.

Mesmo sem pre-check confiavel, trate `ENOSPC` e preserve o binario atual.

Nao apague logs, workspace ou dados Ubuntu para abrir espaco automaticamente.

### Testes

Teste download valido seguido de troca e preservacao do `.bak`.

Teste binario com checksum valido que falha no autoteste e confirme rollback.

Teste falha entre os dois renames com filesystem ou funcoes injetaveis.

Teste inicio com somente `.bak` e confirme restauracao do executavel.

Teste falta de espaco e permissao negada sem perder o binario anterior.

Teste cancelamento de contexto durante download e durante autoteste.

### Criterio De Aceite

Uma atualizacao falha sem remover o comando funcional anterior.

Uma interrupcao deixa executavel ou backup recuperavel no mesmo diretorio.

O binario novo so se torna ativo apos passar autoteste e autenticidade.

## 3. SSH Seguro Por Padrao

### Problema

O SSH atual escuta em `0.0.0.0`, aceita senha e usa `StrictModes no`.

Isso expoe uma superficie de ataque em qualquer Wi-Fi compartilhado.

O produto precisa equilibrar acesso facil para estudantes e protecao de projetos e contas pessoais.

### Politica Recomendada

Use chave SSH como autenticacao padrao.

Use `StrictModes yes` sempre.

Mantenha acesso LAN como escolha explicita no setup, com aviso sobre redes nao confiaveis.

Mantenha senha apenas como opt-in confirmado pelo usuario, nunca como default silencioso.

Nao exponha o servidor diretamente na internet; acesso externo deve usar tunel ou VPN apropriada.

### Fluxo Inicial Sugerido

Durante `mobdesk setup`, pergunte se o usuario deseja acesso remoto pela rede local.

Se recusar, gere configuracao apenas para loopback ou nao inicie SSH automaticamente.

Se aceitar, explique que outros dispositivos da mesma rede poderao tentar conectar.

Mostre como adicionar uma chave publica pelo comando ou pela TUI.

Ofereca senha temporaria apenas se o usuario confirmar o risco e definir senha forte.

No futuro, o pareamento por aprovacao no celular pode substituir esse atrito inicial.

### Configuracao Base

Uma configuracao segura deve conter estas diretrizes:

```text
PasswordAuthentication no
KbdInteractiveAuthentication no
PermitEmptyPasswords no
PubkeyAuthentication yes
StrictModes yes
MaxAuthTries 3
LoginGraceTime 30
MaxStartups 10:30:60
```

`ListenAddress 127.0.0.1` e o default mais restritivo.

`ListenAddress 0.0.0.0` so deve ser emitido para modo LAN explicitamente habilitado.

Confirme com `sshd -t -f` antes de instalar a configuracao final.

### Modelo De Configuracao

Troque parametros globais por uma estrutura de opcoes SSH validada.

Ela pode conter `ExposeLAN`, `AllowPassword` e `AuthorizedKeysPath`.

Nao aceite valores livres vindos de flags sem validacao e sem persistencia privada.

Persista configuracao em diretorio `0700` e arquivos em `0600` quando aplicavel.

Recrie `sshd_config` a partir dessas opcoes, em vez de editar texto do usuario.

### Chaves SSH

Aceite somente chaves publicas OpenSSH suportadas pelo `sshd` instalado.

Valide que uma chave ocupa uma linha, sem caracteres de controle e sem opcoes perigosas inesperadas.

Grave `authorized_keys` com diretorio `.ssh` em `0700` e arquivo em `0600`.

Nunca grave chaves privadas, senhas ou codigos temporarios nos logs.

Ofereca listar fingerprints e remover chaves para revogacao futura.

### Senha Como Fallback

Se senha for mantida, exija confirmacao explicita a cada habilitacao.

Mostre no `status` que senha e exposicao LAN estao ativas.

Nao permita senha vazia e mantenha `MaxAuthTries` e `LoginGraceTime` restritos.

Documente que esse modo nao e indicado para Wi-Fi publico ou desconhecido.

Nao tente implementar rate limit proprio antes de verificar recursos do OpenSSH no Termux.

### Compatibilidade PRoot

Valide `StrictModes yes` no aparelho real com o caminho de home usado pelo Termux.

Se PRoot alterar ownership ou permissoes percebidas, corrija os diretorios gerados, nao desative a protecao.

Teste login por chave e por senha opt-in usando o wrapper que entra no Ubuntu.

Confirme que arquivos de chave permanecem no host Termux esperado e nao vazam para logs.

### Testes

Teste `renderSSHConfig` sem opcoes e confirme loopback, chaves e `StrictModes yes`.

Teste modo LAN explicito e confirme somente a alteracao de endereco esperada.

Teste modo senha explicito e confirme que ele nao aparece no default.

Teste que `sshd -t` rejeita configuracao invalida antes da troca final.

No aparelho, teste acesso por chave em LAN, recusa de senha default e parada do servidor.

### Criterio De Aceite

Uma instalacao nova nao aceita senha nem escuta na LAN sem escolha explicita.

Chaves validas funcionam com permissoes estritas no Termux real.

O usuario consegue identificar no status quando escolheu um modo de maior exposicao.

## 4. Fronteira Entre Termux E Ubuntu

### Problema

O Mobdesk deve executar `setup`, `start`, `stop`, `install` e `update` somente no host Termux.

Uma sessao SSH entra no Ubuntu via PRoot e deve bloquear essas acoes.

`TERMUX_VERSION` pode ser herdada pelo processo Ubuntu e causar classificacao incorreta.

Hoje a deteccao consulta essa variavel antes de verificar `/etc/os-release`.

### Correcao Minima

Em `detectTermuxRuntime`, verifique primeiro se `/etc/os-release` identifica Ubuntu.

Se for Ubuntu, retorne `false` mesmo que `TERMUX_VERSION` e `PREFIX` estejam definidos.

Depois disso, use `TERMUX_VERSION` e o prefixo Termux como sinais do host.

Documente a precedencia: raiz visivel do PRoot e mais confiavel que ambiente herdado.

### Defesa Em Profundidade

O wrapper SSH deve limpar variaveis exclusivas do host antes de chamar `proot-distro`.

Exemplos sao `TERMUX_VERSION`, `PREFIX` e qualquer sentinel futuro usado pelo Mobdesk.

Nao limpe `HOME`, `USER`, `TERM` ou `PATH` sem estudar o efeito no shell Ubuntu.

Use `env -u` somente se a shell do Termux suportar o comportamento verificado.

Alternativamente, gere o wrapper com atribuicoes vazias antes do `exec`.

### Fonte De Verdade

Nao use somente `PATH`, pois ele pode ser herdado ou alterado pelo usuario.

Nao use somente `PREFIX`, pois PRoot pode preserva-lo.

Nao use somente `TERMUX_VERSION`, pois e uma variavel de ambiente herdavel.

Use `/etc/os-release` como sinal forte do sistema visivel na sessao.

Mantenha uma mensagem segura quando a deteccao for inconclusiva: bloqueie acoes de host.

### Impacto Na TUI

A TUI ja sabe ocultar acoes de host quando `Host.Termux` e falso.

A correcao deve alimentar esse campo corretamente quando executada por SSH no Ubuntu.

O usuario remoto ainda deve poder abrir shell local e consultar informacoes do workspace.

Nao tente executar `pkg`, `termux-wake-lock` ou `proot-distro` dentro da sessao Ubuntu.

Explique no texto da TUI que o controle da workstation requer abrir o Mobdesk no Termux.

### Testes Unitarios

Extraia ou injete a leitura de ambiente e de `/etc/os-release` se isso simplificar testes deterministas.

Teste Ubuntu com `TERMUX_VERSION` e prefixo Termux herdados; o resultado deve ser falso.

Teste host Termux com `TERMUX_VERSION`; o resultado deve ser verdadeiro.

Teste prefixo Termux sem arquivo Ubuntu; o resultado deve ser verdadeiro.

Teste ambiente desconhecido; o resultado deve ser falso e conservador.

### Teste De Integracao

Inicie SSH pelo fluxo normal do Mobdesk no ambiente de integracao.

Conecte via SSH e execute `mobdesk status --json` dentro do Ubuntu.

Confirme que `host.termux` e `false` na saida estruturada.

Abra `mobdesk tui` remotamente e confirme ausencia das acoes de host.

Tente chamar os comandos host pela TUI e confirme bloqueio com mensagem explicativa.

Repita o teste com `TERMUX_VERSION` e `PREFIX` exportados antes do login SSH.

### Criterio De Aceite

Uma sessao Ubuntu nunca executa acao exclusiva do host, mesmo com ambiente Termux herdado.

Uma sessao Termux real continua reconhecida como host e conserva todas as operacoes existentes.

O comportamento e coberto por teste unitario e de integracao.

## Plano De Entrega

Divida o trabalho em quatro pull requests pequenos ou commits atomicos.

O primeiro altera deteccao e testes de fronteira de ambiente.

O segundo introduz backup, autoteste e recuperacao do updater.

O terceiro adiciona assinatura de release, verificacao e workflow CI.

O quarto define opcoes SSH, migra defaults e atualiza a experiencia de setup.

Nao misture mudanca de UX SSH com a implementacao de assinatura de update.

Cada entrega deve manter `make check` verde e adicionar seu teste de regressao.

## Validacao Final No Aparelho

Execute setup novo em Android ARM64 com Termux atualizado.

Teste reinicio, wake-lock, inicio e parada do SSH.

Teste SSH por chave em uma rede LAN confiavel e confirme recusa de senha default.

Teste TUI remota e confirme bloqueio de operacoes do host.

Teste update assinado em rede normal e simulacao de binario invalido em ambiente seguro.

Teste encerramento do processo durante update e confirme recuperacao no proximo inicio.

Registre modelo do aparelho, versao Android, Termux e resultado no checklist de release.

## Definition Of Done

Os quatro criterios de aceite deste documento devem passar no CI e no aparelho real.

O README deve explicar a politica SSH escolhida e o comportamento de update seguro.

`docs/DECISOES.md` deve registrar Minisign, politica de backup e defaults SSH.

Nenhuma chave privada, senha, token ou assinatura de fixture deve entrar no Git.

Somente depois disso o Mobdesk deve ser divulgado como pronto para uso remoto por terceiros.
