# Security Policy

## Reporting a vulnerability

Do not publish details of vulnerabilities involving SSH, authentication,
command execution, installation scripts, updates, or exposed ports in public
issues.

Use a private communication channel with the repository maintainers. Include:

- Mobdesk version;
- Android and Termux versions;
- device model;
- reproduction steps;
- observed impact;
- logs with passwords, private keys and tokens removed.

Until a supported-version policy exists, fixes are applied to the development
version and, when technically possible, the latest stable release.

Use Mobdesk SSH only on trusted networks or through a secure tunnel. Never
expose port `8022` directly to the public internet.

See [the Brazilian Portuguese security policy](SECURITY.pt-BR.md) for the
translated public version.
