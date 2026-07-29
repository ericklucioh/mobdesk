# Diagnóstico de Consumo de Bateria

## Resumo

O Ubuntu via PRoot não é um serviço que fica rodando sozinho. Ele é um
userland persistente: cada processo `proot-distro login ubuntu ...` existe
enquanto o comando, shell ou sessão que o iniciou estiver ativo.

O risco atual está no ciclo de vida das sessões e processos filhos. O Mobdesk
inicia um servidor SSH no Termux e ativa o `termux-wake-lock`. Uma sessão SSH
pode iniciar um shell Ubuntu via PRoot, e esse shell pode iniciar processos
adicionais.

## Comportamento Atual

- `mobdesk start` verifica o Ubuntu, configura o SSH, ativa o wake-lock e
  inicia o `sshd` na porta 8022.
- `mobdesk stop` envia `SIGTERM` ao processo principal do `sshd` e libera o
  wake-lock.
- `mobdesk shell` inicia uma sessão interativa com `proot-distro`.
- A TUI cancela os comandos que ela iniciou por meio de `context.Context`.
- O filesystem do Ubuntu permanece instalado, mas não consome CPU ou bateria
  quando não há processos ativos.

## Risco Identificado

O `stop` controla diretamente apenas o PID principal do `sshd`. O projeto não
mantém um inventário explícito dos processos filhos das sessões SSH/PRoot.

Podem continuar ativos após o fechamento da TUI ou de uma sessão:

- shells SSH ou PRoot;
- sessões `tmux` ou `zellij`;
- processos iniciados com `nohup` ou `&`;
- servidores locais;
- builds e ferramentas de desenvolvimento de longa duração.

Esses processos podem consumir CPU, memória, rede e impedir a suspensão
profunda do Android. O `wake-lock` ativo aumenta o impacto da bateria.

Não se deve encerrar todos os processos chamados `proot` globalmente, porque
isso poderia matar processos que não pertencem ao Mobdesk.

## Verificação no Dispositivo

Executar no Termux:

```sh
ps -ef | grep -E 'sshd|proot|ubuntu|tmux|zellij|go build|node|python' | grep -v grep
```

Verificar o estado reportado pelo Mobdesk:

```sh
mobdesk status --json
```

Consultar o PID registrado do SSH:

```sh
cat "$HOME/.local/share/mobdesk/ssh/sshd.pid" 2>/dev/null
```

Parar a workstation:

```sh
mobdesk stop
```

Também é necessário fechar sessões SSH e sessões `tmux`/`zellij` abertas.

## Correção Proposta

O comando `stop` deve:

1. validar que o PID registrado pertence ao `sshd` do Mobdesk;
2. identificar apenas os processos descendentes desse servidor e suas sessões;
3. enviar sinais de encerramento de forma ordenada;
4. aguardar o fechamento das sessões e do PRoot;
5. confirmar que a porta 8022 foi liberada;
6. liberar o wake-lock mesmo em caminhos de erro;
7. reportar processos que não puderam ser encerrados.

A implementação deve continuar protegendo processos externos e não pode usar
um `pkill proot` genérico.

## Conclusão

Há uma lacuna de ciclo de vida que pode explicar consumo elevado de bateria:
processos filhos de sessões SSH/PRoot podem sobreviver ao encerramento da TUI
ou do `sshd`. A confirmação deve começar pela inspeção de processos no aparelho
e pelo estado do wake-lock. Depois disso, o `stop` deve ganhar limpeza segura e
verificável das sessões pertencentes ao Mobdesk.
