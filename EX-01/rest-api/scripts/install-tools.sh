#!/usr/bin/env bash
# Installs the optional developer tools some Makefile targets require
# (staticcheck, newman), skipping any that are already on PATH.
#
# Run directly to install both:
#   ./scripts/install-tools.sh
# Or source it to install just one:
#   source scripts/install-tools.sh && install_newman
set -uo pipefail

install_staticcheck() {
	if command -v staticcheck >/dev/null 2>&1; then
		echo "staticcheck: already installed ($(command -v staticcheck))"
		return 0
	fi
	if ! command -v go >/dev/null 2>&1; then
		echo "staticcheck: skipped - Go is required first (https://go.dev/dl/)" >&2
		return 1
	fi
	echo "staticcheck: installing via go install..."
	go install honnef.co/go/tools/cmd/staticcheck@latest
}

install_newman() {
	if command -v newman >/dev/null 2>&1; then
		echo "newman: already installed ($(command -v newman))"
		return 0
	fi
	if ! command -v npm >/dev/null 2>&1; then
		echo "newman: skipped - Node.js/npm is required first (https://nodejs.org)" >&2
		return 1
	fi
	echo "newman: installing via npm..."
	npm install -g newman
}

# Only auto-run both when executed directly, not when sourced - keeps each
# function individually callable (e.g. from a fresh shell after sourcing).
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
	status=0
	install_staticcheck || status=1
	install_newman || status=1
	exit "$status"
fi
