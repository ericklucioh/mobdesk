# Security Policy

## Reporting a vulnerability

Do not publish details of vulnerabilities involving SSH, authentication,
command execution, installation scripts, updates, or exposed ports in public
issues.

Email the repository maintainer privately at
[contato@ericklucioh.com](mailto:contato@ericklucioh.com). Do not use a public
issue for a vulnerability report. Include:

- Mobdesk version;
- Android and Termux versions;
- device model;
- reproduction steps;
- observed impact;
- logs with passwords, private keys and tokens removed.

Until a supported-version policy exists, fixes are applied to the development
version and, when technically possible, the latest stable release.

The maintainer will review reports sent to this address and coordinate the
response privately. Do not include passwords, private keys, tokens or unrelated
personal data in the report.

Use Mobdesk SSH only on trusted networks or through a secure tunnel. Never
expose port `8022` directly to the public internet.

See [the Brazilian Portuguese security policy](SECURITY.pt-BR.md) for the
translated public version.
