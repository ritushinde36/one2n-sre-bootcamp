#!/usr/bin/env bash
# macOS-only. Installs every tool needed to build, run, and test this app -
# both the hard prerequisites (Go, MySQL, Docker, Git, Make) and the optional
# dev tools some Makefile targets use (staticcheck, newman, hadolint), plus
# newman's own transitive dependency (Node.js/npm). Skips anything already
# on PATH. Installation goes through Homebrew, except make/git, which come
# from Xcode's Command Line Tools on macOS.
#
# Run directly to install everything:
#   ./scripts/install-prerequisites.sh
# Or source it to install just one:
#   source scripts/install-prerequisites.sh && install_docker
set -uo pipefail

ensure_homebrew() {
	if command -v brew >/dev/null 2>&1; then
		echo "homebrew: already installed ($(command -v brew))"
		return 0
	fi
	echo "homebrew: installing (required to install everything else below)..."
	/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
}

# make and git both ship with Xcode's Command Line Tools on macOS - installing
# via Homebrew instead would only get you `gmake`/a second git, not replace
# the ones on PATH, so this is the correct source for these two specifically.
ensure_xcode_clt() {
	if xcode-select -p >/dev/null 2>&1; then
		return 0
	fi
	echo "Xcode Command Line Tools: installing (provides make and git)..."
	echo "A GUI installer window will open - finish it, then re-run this script." >&2
	xcode-select --install
	return 1
}

install_make() {
	if command -v make >/dev/null 2>&1; then
		echo "make: already installed ($(command -v make))"
		return 0
	fi
	ensure_xcode_clt
}

install_git() {
	if command -v git >/dev/null 2>&1; then
		echo "git: already installed ($(command -v git))"
		return 0
	fi
	ensure_xcode_clt
}

install_go() {
	if command -v go >/dev/null 2>&1; then
		echo "go: already installed ($(command -v go))"
		return 0
	fi
	echo "go: installing via brew..."
	brew install go
}

install_node() {
	if command -v npm >/dev/null 2>&1; then
		echo "node/npm: already installed ($(command -v npm))"
		return 0
	fi
	echo "node/npm: installing via brew..."
	brew install node
}

install_mysql() {
	if command -v mysql >/dev/null 2>&1; then
		echo "mysql: already installed ($(command -v mysql))"
		return 0
	fi
	echo "mysql: installing via brew (not started - run 'brew services start mysql' when you need it)..."
	brew install mysql
}

install_docker() {
	if command -v docker >/dev/null 2>&1; then
		echo "docker: already installed ($(command -v docker))"
		return 0
	fi
	echo "docker: installing Docker Desktop via brew cask..."
	echo "You'll need to open Docker.app once by hand afterwards to finish setup." >&2
	brew install --cask docker
}

install_staticcheck() {
	if command -v staticcheck >/dev/null 2>&1; then
		echo "staticcheck: already installed ($(command -v staticcheck))"
		return 0
	fi
	if ! command -v go >/dev/null 2>&1; then
		echo "staticcheck: skipped - Go is required first" >&2
		return 1
	fi
	echo "staticcheck: installing via go install..."
	go install honnef.co/go/tools/cmd/staticcheck@latest
	local install_status=$?
	if [ "$install_status" -ne 0 ]; then
		return "$install_status"
	fi
	if command -v staticcheck >/dev/null 2>&1; then
		return 0
	fi
	# go install succeeded, but the binary landed somewhere not on PATH -
	# usually because $(go env GOPATH)/bin was never added to it.
	local go_bin
	go_bin="$(go env GOBIN)"
	[ -n "$go_bin" ] || go_bin="$(go env GOPATH)/bin"
	echo "staticcheck: installed to $go_bin, but that's not on your PATH yet." >&2
	echo "  Add this to your shell profile (~/.zshrc or ~/.bash_profile), then open a new shell:" >&2
	echo "    export PATH=\"\$PATH:$go_bin\"" >&2
	return 1
}

install_newman() {
	if command -v newman >/dev/null 2>&1; then
		echo "newman: already installed ($(command -v newman))"
		return 0
	fi
	if ! command -v npm >/dev/null 2>&1; then
		echo "newman: skipped - Node.js/npm is required first" >&2
		return 1
	fi
	echo "newman: installing via npm..."
	npm install -g newman
}

install_hadolint() {
	if command -v hadolint >/dev/null 2>&1; then
		echo "hadolint: already installed ($(command -v hadolint))"
		return 0
	fi
	echo "hadolint: installing via brew..."
	brew install hadolint
}

# Only auto-run everything when executed directly, not when sourced - keeps
# each function individually callable (e.g. from a fresh shell after sourcing).
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
	if [[ "$(uname -s)" != "Darwin" ]]; then
		echo "This script is macOS-only (relies on Homebrew casks and Xcode Command Line Tools)." >&2
		echo "On other platforms, install each tool from the README's Prerequisites section manually." >&2
		exit 1
	fi
	status=0
	ensure_homebrew || status=1
	install_git || status=1
	install_make || status=1
	install_go || status=1
	install_node || status=1
	install_mysql || status=1
	install_docker || status=1
	install_staticcheck || status=1
	install_newman || status=1
	install_hadolint || status=1
	exit "$status"
fi
