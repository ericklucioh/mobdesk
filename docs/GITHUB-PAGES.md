# GitHub Pages

The Mobdesk landing page is a static site in `site/`. It is available in
Portuguese at the root and in English at `site/en/`. It does not require a
server, a custom domain, Jekyll or an external hosting provider.

## Deployment

The workflow [`.github/workflows/pages.yml`](../.github/workflows/pages.yml)
publishes the site through the official GitHub Pages deployment actions whenever
the `main` branch changes.

To enable it in the repository:

1. Open **Settings > Pages** on GitHub.
2. Under **Build and deployment**, choose **GitHub Actions** as the source.
3. Push to `main` or run **Deploy GitHub Pages** from the Actions tab.

The default project URL is:

`https://ericklucioh.github.io/mobdesk/`

The page uses only local HTML and CSS. There is no `CNAME` file because the
project does not use a custom domain.
