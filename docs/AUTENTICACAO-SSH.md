# Autenticação e aprovação de conexões SSH

Este documento descreve duas formas de adicionar uma camada de autorização ao
SSH do Mobdesk:

1. aprovação da conexão pelo celular;
2. código temporário de uso único.

As duas alternativas devem proteger o acesso ao ambiente Ubuntu sem afastar o
Mobdesk da sua proposta de uso simples por TUI.

## Contexto

No MVP, o Mobdesk executa um servidor SSH dedicado na porta `8022` e encaminha
o usuário para o Ubuntu persistente via PRoot. A autenticação atual é feita
por senha do Termux.

O objetivo desta evolução é reduzir a dependência de senha fixa e permitir que
o usuário controle, pelo próprio celular, quais computadores podem iniciar uma
sessão.

Esta camada não substitui as proteções básicas:

- não expor o SSH diretamente à internet;
- preferir rede local ou Tailscale para acesso externo;
- usar chaves SSH quando possível;
- proteger arquivos de configuração, estado e credenciais;
- registrar tentativas sem gravar senhas ou códigos.

## Opção 2: aprovação da conexão pelo celular

### Conceito

Quando um computador tenta abrir uma sessão SSH, o Mobdesk apresenta uma
solicitação pendente na TUI do celular. O usuário vê os dados da tentativa e
decide se permite ou recusa o acesso.

Exemplo de solicitação:

```text
Nova conexão SSH

Dispositivo: notebook-erick
IP:          192.168.3.40
Horário:     19:42
Fingerprint: SHA256:...

[A] Aprovar   [R] Recusar   [X] Bloquear dispositivo
```

### Fluxo esperado

```text
Computador                Mobdesk no celular             SSH/Ubuntu
     |                             |                         |
     |--- inicia conexão ---------->|                         |
     |                             |                         |
     |                             |--- mostra solicitação   |
     |                             |                         |
     |                             |<-- usuário aprova ------|
     |                             |                         |
     |<---------- autorização -----|                         |
     |------------------- sessão SSH autorizada ------------->|
```

### Vantagens

- experiência simples: aprovar ou recusar;
- mostra contexto antes de liberar o acesso;
- permite revogar dispositivos pelo Mobdesk;
- reduz o risco de uma senha vazada ser suficiente para entrar;
- pode funcionar junto com uma chave SSH como segunda camada.

### Cuidados

- a aprovação precisa estar vinculada à tentativa específica, não apenas ao IP;
- deve expirar rapidamente se o usuário não responder;
- o usuário precisa identificar o fingerprint do cliente quando parear um novo
  dispositivo;
- o Mobdesk deve negar por padrão quando não houver resposta;
- uma aprovação não deve ser reutilizada para outra sessão ou dispositivo.

### Observação de implementação

Um `Banner` do SSH pode exibir informações, mas não autentica o usuário. A
aprovação precisa participar de uma etapa real de autenticação ou autorização,
por exemplo por autenticação interativa, um componente auxiliar integrado ao
`sshd` ou uma chave temporária liberada somente para aquela conexão.

## Opção 3: código temporário de uso único

### Conceito

O Mobdesk gera um código curto, aleatório e de validade limitada. O usuário
digita o código no cliente SSH para concluir a autorização.

O código deve ser mostrado exclusivamente no celular. Se ele for exibido no
terminal remoto e depois digitado no Mobdesk, um invasor que iniciou a conexão
também poderá capturá-lo.

### Requisitos mínimos

- validade curta, por exemplo de 30 a 60 segundos;
- uso único;
- geração criptograficamente segura;
- limite de tentativas;
- invalidação após sucesso, expiração ou cancelamento;
- vínculo com usuário, dispositivo e tentativa de conexão;
- nenhum registro do valor do código nos logs.

### Vantagens

- fallback útil quando a aprovação visual não estiver disponível;
- pode ser usado para o primeiro pareamento;
- não exige que o usuário configure manualmente uma chave SSH;
- é familiar para usuários que já conhecem códigos de autenticação.

### Limitações

- é mais trabalhoso que apenas tocar em “Aprovar”;
- códigos curtos podem sofrer tentativas automáticas se não houver rate limit;
- um código digitado em um computador comprometido pode ser capturado;
- se o segredo e o servidor estiverem no mesmo celular, o código não protege
  contra comprometimento completo do aparelho;
- sincronização de relógio pode causar problemas se o modelo usar TOTP baseado
  apenas em tempo.

Por isso, o Mobdesk deve preferir um desafio aleatório associado à tentativa,
em vez de um código TOTP independente da conexão.

## Comparação

| Critério | Aprovação no celular | Código temporário |
|---|---|---|
| Facilidade de uso | Muito alta | Média |
| Clareza para o usuário | Alta, mostra o dispositivo | Média |
| Resistência a senha vazada | Alta | Alta |
| Risco de captura durante a digitação | Baixo | Médio |
| Implementação | Mais complexa | Moderada |
| Melhor uso | Conexões normais | Pareamento e fallback |
| Necessita expiração | Sim | Sim |

## Recomendação

O Mobdesk deve priorizar a **aprovação da conexão pelo celular** como a
experiência principal.

O código temporário deve existir como mecanismo de pareamento e fallback:

1. o usuário inicia o Mobdesk e habilita o SSH;
2. um computador solicita o primeiro acesso;
3. o Mobdesk exibe um código de pareamento ou uma solicitação de aprovação;
4. após a aprovação, o computador instala uma chave SSH Ed25519;
5. nas próximas conexões, a chave identifica o dispositivo;
6. a aprovação pelo celular pode continuar obrigatória ou ser ativada como
   segunda camada;
7. o código temporário permanece disponível para recuperação ou novo
   pareamento.

### Política sugerida

- negar conexões novas por padrão;
- exigir aprovação explícita no primeiro acesso;
- identificar dispositivos por chave e fingerprint;
- permitir revogar uma chave individual ou todas as chaves;
- expirar solicitações pendentes;
- bloquear temporariamente muitas tentativas falhas;
- manter o SSH restrito à rede local ou Tailscale;
- não permitir que a aprovação seja reutilizada em outra conexão.

## Escopo por estágio

### MVP-1

- manter o SSH dedicado na porta `8022`;
- manter a autenticação atual por senha;
- não expor o SSH diretamente na internet;
- documentar a preparação para chaves e aprovação futura.

### Pós-MVP-1

- criar pareamento de dispositivos;
- gerar e instalar chaves Ed25519 automaticamente;
- exibir solicitações pendentes na TUI;
- implementar aprovação e recusa vinculadas à conexão;
- adicionar código temporário de uso único como fallback;
- adicionar revogação, expiração, rate limit e auditoria.

## Referência técnica

O OpenSSH possui métodos como `publickey`, `password` e
`keyboard-interactive`, além de permitir exigir mais de um método com
`AuthenticationMethods`. A documentação oficial está disponível em
[`sshd_config(5)`](https://man.openbsd.org/sshd_config).

