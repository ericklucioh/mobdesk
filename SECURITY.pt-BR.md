# Segurança

## Relatar uma vulnerabilidade

Não publique detalhes de vulnerabilidades envolvendo SSH, autenticação, execução de comandos, scripts de instalação ou exposição de portas em issues públicas.

Envie o relato por e-mail privado para
[contato@ericklucioh.com](mailto:contato@ericklucioh.com). Não use uma issue
pública para relatar uma vulnerabilidade. Inclua:

- versão do Mobdesk;
- versão do Android e do Termux;
- modelo do dispositivo;
- passos para reproduzir;
- impacto observado;
- logs sem senhas, chaves privadas ou tokens.

Até existir uma política de versões suportadas, as correções são aplicadas à versão em desenvolvimento e à release estável mais recente quando tecnicamente possível.

O mantenedor analisará os relatos enviados para esse endereço e coordenará a
resposta de forma privada. Não inclua senhas, chaves privadas, tokens ou dados
pessoais que não sejam necessários.

Use o SSH do Mobdesk apenas em redes confiáveis ou por um túnel seguro. Nunca
exponha a porta `8022` diretamente à internet.

Veja também a [política de segurança em inglês](SECURITY.md).
