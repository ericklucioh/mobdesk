# Plano: refresh global da TUI

## Objetivo

Permitir que a tecla `R` atualize o estado exibido em qualquer tela da TUI, mantendo o Cobra como única fonte de verdade e sem criar streaming ou eventos de progresso.

O refresh será uma nova leitura estática do estado atual. Ele não executará operações como setup, instalação, start, stop ou atualização.

## Contrato de comunicação

O backend real deve consultar somente comandos existentes no Cobra:

```text
mobdesk status --json
mobdesk version --json
```

`status --json` fornece o estado do ambiente, setup, SSH, Ubuntu, rede, bateria e instalações.

`version --json` fornece os dados exibidos na tela Sistema.

O refresh de estado não deve executar automaticamente:

```text
mobdesk update --check --json
```

A verificação de atualização continua sendo uma ação explícita do botão `Verificar`.

## Comportamento do refresh global

As teclas `r` e `R` devem executar a mesma ação em todas as telas:

| Tela | Dados atualizados |
|---|---|
| Home | estado da workstation, SSH e ambiente |
| Apps | instalações e estados das ferramentas |
| Setup | fases concluídas e estado do ambiente |
| Status | tabela e resumo completo do sistema |
| Sistema | estado do ambiente e versão instalada |
| Shell | estado atual da workstation |

O refresh deve ser ignorado enquanto:

- uma operação estiver em andamento;
- o modal de confirmação estiver aberto;
- a aplicação estiver encerrando.

O refresh deve preservar a tela atual, posição de rolagem e foco sempre que possível.

## Operações sem tempo real no fluxo real

O fluxo real da TUI não receberá eventos de progresso nem fará polling periódico.

Ao iniciar uma operação, a TUI exibirá uma mensagem fixa:

```text
Operação em andamento...
Aguarde a conclusão.
```

Não serão exibidos percentuais, etapas ou estados intermediários inventados.

Quando o comando Cobra terminar:

1. a TUI interpreta o JSON final;
2. mostra sucesso ou erro;
3. executa o refresh global automaticamente;
4. atualiza todas as telas que dependem do estado compartilhado.

O mock continuará existindo para testar manualmente as telas e os fluxos. Ele poderá:

- atrasar artificialmente o resultado para manter a tela de operação visível;
- simular sucesso, erro e estados degradados;
- alterar o snapshot retornado depois de uma operação;
- simular ferramentas instaladas e não instaladas;
- permitir testar start, stop, setup, instalação, update e shell sem tocar no sistema real.

Esse atraso é uma ferramenta de teste, não uma representação de progresso real. Tanto o mock quanto o backend real devem exibir uma mensagem fixa durante a operação, sem etapas ou percentuais inventados.

## Implementação prevista

- Centralizar o refresh em uma única função do modelo.
- Fazer `r` e `R` chamarem essa função antes dos atalhos específicos.
- Reutilizar o mesmo comando de refresh no botão `[R] Atualizar status`.
- Atualizar status, versão, tabela e lista de ferramentas a partir das mensagens recebidas.
- Remover o uso de spinner, percentual e etapas falsas da tela de operação.
- Manter as operações reais delegadas aos comandos Cobra válidos.
- Fazer o mock rejeitar qualquer operação que não exista no contrato do CLI.

## Testes

Adicionar testes para:

- refresh por `r` e `R` na Home;
- refresh por `r` e `R` em Apps, Setup, Status, Sistema e Shell;
- atualização da lista de ferramentas após uma instalação simulada;
- atualização das fases de setup após uma operação simulada;
- atualização do status da workstation após start e stop;
- carregamento da versão pelo contrato do CLI;
- refresh ignorado durante operação;
- refresh ignorado durante confirmação;
- preservação da tela atual após refresh;
- ausência de polling automático;
- ausência de etapas e percentuais falsos na tela de operação real;
- permanência dos cenários mock para testar operação pendente, sucesso e erro;
- sucesso e erro retornados pelo Cobra/mock.

## Critérios de aceitação

- `R` atualiza os dados de qualquer tela sem navegar para Status.
- Nenhuma tecla ou botão de refresh executa operações destrutivas ou de alteração.
- A TUI usa somente comandos existentes no Cobra.
- A tela de operação não afirma progresso que o Cobra não fornece.
- O refresh manual e o refresh automático após uma operação produzem o mesmo estado.
- O comportamento real e o comportamento mock seguem o mesmo contrato.
- O fluxo continua funcionando em terminal estreito e largo.

## Fora do escopo

- streaming de eventos;
- polling periódico;
- percentual real de progresso;
- WebSocket ou servidor de eventos;
- alteração do formato final dos JSONs do Cobra;
- execução automática de update, setup ou instalação durante o refresh.
