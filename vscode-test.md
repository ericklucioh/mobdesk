# VS Code Tunnel Experiment

**Status:** personal historical experiment; non-authoritative and not part of
the Mobdesk MVP. It records a manual test rather than a supported product
workflow.

Run the experiment inside Debian or Ubuntu through PRoot, not in native Termux.
The Microsoft standalone CLI provides a Linux ARM64 download, but PRoot on
Android is not officially certified by Microsoft.

## Installation experiment

```bash
pkg update
pkg install proot-distro
proot-distro install debian
proot-distro login debian
uname -m
apt update
apt install -y curl ca-certificates tar
cd /tmp
curl -L "https://update.code.visualstudio.com/latest/cli-linux-arm64/stable" \
  -o vscode_cli.tar.gz
tar -xzf vscode_cli.tar.gz
install -m 0755 code /usr/local/bin/code
code --version
code tunnel --accept-server-license-terms
```

The expected architecture is `aarch64`. `code tunnel` displays an authentication
flow and a `vscode.dev/tunnel/...` URL for the client browser. The tunnel is
available only while the process runs. It needs outbound GitHub or Microsoft
HTTPS access and does not require an Android inbound port.

## Troubleshooting notes

For an Ubuntu PRoot test, refresh certificates and inspect:

```bash
apt update
apt install --reinstall -y ca-certificates curl openssl
update-ca-certificates
getent hosts github.com
curl -Iv https://github.com/login/device
env | grep -i proxy
unset HTTP_PROXY HTTPS_PROXY ALL_PROXY http_proxy https_proxy all_proxy
code tunnel --accept-server-license-terms --verbose --log trace
```

Compare the same HTTPS request in native Termux. A failure only in Ubuntu
usually points to PRoot DNS, certificates or proxy configuration; a failure in
both points to Android networking, VPN, DNS, firewall or blocking. A successful
installation does not prove product readiness: the experiment ended with
communication working but product validation still pending.

References:

- [VS Code FAQ](https://code.visualstudio.com/docs/supporting/faq)
- [Remote development with Linux](https://code.visualstudio.com/docs/remote/linux)
- [Remote tunnels](https://code.visualstudio.com/docs/remote/tunnels)
- [related authentication issue](https://github.com/microsoft/vscode/issues/235874)
