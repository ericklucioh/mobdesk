# Todo Da Auditoria De Segurança

Baseado em `docs/AUDITORIA-SEGURANCA-2026-07-25.md`.

## Decisões Pendentes

### H-01: Política de acesso SSH

**Decisão:** definir exposição de rede e método de autenticação padrão.

| Opção | Vantagens | Custos e riscos |
| --- | --- | --- |
| Loopback + chaves SSH por padrão (recomendada) | Menor superfície de ataque; elimina brute force remoto; mantém acesso por túnel. | Exige criação e gestão inicial de chave; não atende acesso LAN imediato. |
| LAN + chaves SSH por padrão | Permite conectar outro dispositivo na mesma rede sem senha. | Porta continua exposta na LAN; requer instrução clara sobre redes não confiáveis. |
| LAN + senha opt-in | Menor atrito para iniciantes e preserva o fluxo atual quando explicitamente escolhido. | Senhas podem sofrer brute force; requer aviso, limites de tentativas e confirmação explícita. |

**Por que decidir:** esta escolha muda o fluxo de primeiro acesso e o nível de risco aceito em Wi-Fi de faculdade, público ou compartilhado.

### H-02: Autenticidade das releases

**Decisão:** escolher o mecanismo de assinatura e a custódia da chave privada de publicação.

| Opção | Vantagens | Custos e riscos |
| --- | --- | --- |
| Minisign (recomendada) | Pequeno, simples para binário único e fácil de verificar com chave pública embutida. | Exige proteger e fazer backup da chave privada; adiciona etapa de assinatura ao release. |
| Cosign/Sigstore | Integra bem com proveniência e identidades OIDC do CI. | Ecossistema mais complexo para usuários e binário local; pode ser excesso para o MVP. |
| GPG | Padrão conhecido e amplamente suportado. | UX e gestão de chaves mais complexas; maior chance de erros operacionais. |

**Por que decidir:** checksum sem assinatura não protege contra release ou conta GitHub comprometida. A solução escolhida define o formato dos assets e o workflow de publicação.

### M-03: Limites do atualizador

**Decisão:** definir timeout, tamanho máximo de download e política de redirects.

| Opção | Vantagens | Custos e riscos |
| --- | --- | --- |
| 2 min total, 64 MiB de binário, 1 MiB de checksums, até 3 redirects (recomendada) | Protege armazenamento e evita TUI bloqueada; é confortável para binário ARM64 em rede móvel. | Pode rejeitar release futura acima do limite; exige ajustar valor quando o binário crescer. |
| Timeout apenas no contexto chamador, sem limite de tamanho | Implementação mínima e flexível. | Mantém risco de conexão infinita e esgotamento de armazenamento. |
| Limites muito restritos, como 30 s e 16 MiB | Defesa forte contra abuso e consumo de dados. | Falha com redes móveis lentas ou releases normais maiores. |

**Por que decidir:** esses valores são política de produto e precisam equilibrar conectividade móvel, tamanho da release e proteção de recursos do celular.

### M-04: Limite de leitura de logs

**Decisão:** definir quantas linhas e bytes `mobdesk logs` pode ler por arquivo.

| Opção | Vantagens | Custos e riscos |
| --- | --- | --- |
| 200 linhas e 1 MiB por log (recomendada) | Mantém saída útil para diagnóstico sem carregar logs enormes em memória. | Pode truncar contexto de falhas antigas. |
| Apenas limite de linhas | UX simples. | Ainda requer ler arquivo inteiro com a implementação atual; não resolve consumo de memória. |
| 1.000 linhas e 8 MiB | Retém mais contexto para diagnóstico. | Aumenta consumo de memória e saída no terminal. |

**Por que decidir:** o limite define a experiência de diagnóstico e a proteção contra logs produzidos por instaladores ou ferramentas com saída excessiva.

### CI: Política de validação e supply chain

**Decisão:** definir gates obrigatórios e rigor de proveniência na automação.

| Opção | Vantagens | Custos e riscos |
| --- | --- | --- |
| `go test -race`, `govulncheck`, linter, Actions fixadas por SHA e cobertura informativa (recomendada) | Adiciona defesa relevante sem bloquear PRs por uma meta arbitrária de cobertura. | CI fica mais lenta; exige atualizar SHAs de Actions periodicamente. |
| Mesmo conjunto + mínimo de 70% de cobertura | Impede queda mensurável de cobertura. | A cobertura atual é inferior; pode incentivar testes superficiais só para atingir número. |
| Manter CI atual | Menor custo e tempo de execução. | Não detecta races em CI e deixa risco de supply chain das Actions sem mitigação. |

**Por que decidir:** define o custo de manutenção aceito em troca de detectar regressões e vulnerabilidades antes da release.
