#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

failed=0

report_matches() {
	local label=$1
	local path=$2
	local pattern=$3
	local matches
	if matches=$(rg -n --no-heading --glob '*.go' "${pattern}" "${path}" 2>/dev/null); then
		printf 'i18n-check: %s\n%s\n' "${label}" "${matches}" >&2
		failed=1
	fi
}

# Portuguese characters in Go source indicate untranslated comments, names or
# presentation literals. Catalog JSON and external language fixtures are not Go
# source and are deliberately outside this check.
report_matches "Portuguese text in Go source" "cmd" '[À-ÿ]'
report_matches "Portuguese text in Go source" "internal" '[À-ÿ]'

if rg -n --no-heading '[À-ÿ]' Makefile scripts --glob '*.sh' --glob 'Makefile' --glob '!i18n-check.sh' 2>/dev/null; then
	printf '%s\n' 'i18n-check: Portuguese text in Makefile or shell scripts' >&2
	failed=1
fi

# Presentation packages must print catalog results, not direct user-facing
# literals. This intentionally ignores fmt.Printf/Fprintf because those paths
# emit commands, paths, external output, or other technical values.
if rg -n --no-heading --glob '*.go' --glob '!**/*_test.go' \
	'(fmt\.(Println|Fprintln)|Message[[:space:]]*:)[[:space:]]*"' \
	internal/cobra internal/tui internal/status 2>/dev/null; then
	printf '%s\n' 'i18n-check: direct presentation literal outside a catalog' >&2
	failed=1
fi

# Catalog completeness is validated by the package test, which checks every
# required MessageID in both supported locales without scanning fixtures or docs.
while IFS= read -r message_id; do
	for catalog in internal/i18n/locale/en-US.json internal/i18n/locale/pt-BR.json; do
		if ! rg -q "\"${message_id}\"[[:space:]]*:" "${catalog}"; then
			printf 'i18n-check: missing catalog ID %s in %s\n' "${message_id}" "${catalog}" >&2
			failed=1
		fi
	done
done < <(rg -o 'MessageID = "[^"]+"' internal/i18n/message.go | rg -o '"[^"]+"' | tr -d '"')

if ! go test ./internal/i18n >/dev/null; then
	printf '%s\n' 'i18n-check: catalog validation failed' >&2
	failed=1
fi

check_markdown_links() {
	local source target path
	while IFS=$'\t' read -r source target; do
		case "${target}" in
			http://*|https://*|mailto:*|\#*|//* ) continue ;;
		esac
		target=${target%%\#*}
		target=${target%%\?*}
		[[ -z "${target}" ]] && continue
		path="$(dirname "${source}")/${target}"
		if [[ ! -e "${path}" ]]; then
			printf 'i18n-check: broken Markdown link %s -> %s\n' "${source}" "${target}" >&2
			failed=1
		fi
	done < <(
		while IFS= read -r source; do
			[[ -f "${source}" ]] || continue
			perl -ne 'while (/\[[^]]+\]\(([^)[:space:]]+)/g) { print "$ARGV\t$1\n" }' "${source}"
		done < <(git ls-files '*.md' 'AGENTS.md')
	)
}

check_markdown_links

if (( failed != 0 )); then
	exit 1
fi

printf '%s\n' 'i18n-check: PASS'
