#!/bin/sh
# install.sh — install the latest googledrive-mcp release with one command:
#
#   curl -fsSL https://raw.githubusercontent.com/3leafcoder/googledrive-ai/main/install.sh | sh
#
# Environment overrides:
#   GOOGLEDRIVE_MCP_VERSION      release tag to install (default: latest)
#   GOOGLEDRIVE_MCP_INSTALL_DIR  destination directory (default: $HOME/.local/bin)
#
# The default install directory is user-owned ($HOME/.local/bin) so the
# installed binary can self-update in place without sudo. A directory under
# $HOME is never installed with sudo — a root-owned binary there would break
# sudo-free self-updates; if the $HOME path cannot be created or written, the
# script errors and points at ownership/permissions instead of escalating. Only
# an GOOGLEDRIVE_MCP_INSTALL_DIR pointing outside $HOME (e.g. /usr/local/bin) falls
# back to sudo when that directory is not writable.
set -eu

REPO="3leafcoder/googledrive-ai"
BINARY="googledrive-mcp"

err() {
	printf 'error: %s\n' "$1" >&2
	exit 1
}

# --- detect platform ---------------------------------------------------------
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
	darwin) os=darwin ;;
	linux) os=linux ;;
	mingw* | msys* | cygwin* | windows*)
		err "Windows is not supported by this script — download a release from https://github.com/$REPO/releases" ;;
	*) err "unsupported OS: $os — download a release from https://github.com/$REPO/releases" ;;
esac

arch=$(uname -m)
case "$arch" in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*) err "unsupported architecture: $arch — download a release from https://github.com/$REPO/releases" ;;
esac

asset="$BINARY-$os-$arch"

# --- resolve download URL ----------------------------------------------------
version="${GOOGLEDRIVE_MCP_VERSION:-}"
if [ -n "$version" ]; then
	url="https://github.com/$REPO/releases/download/$version/$asset"
else
	url="https://github.com/$REPO/releases/latest/download/$asset"
fi

# --- download to a temp file, cleaned up on exit -----------------------------
tmp=$(mktemp "${TMPDIR:-/tmp}/googledrive-mcp-install.XXXXXX") || err "could not create temp file"
trap 'rm -f "$tmp"' EXIT

if command -v curl >/dev/null 2>&1; then
	curl -fsSL "$url" -o "$tmp" || err "download failed: $url"
elif command -v wget >/dev/null 2>&1; then
	wget -qO "$tmp" "$url" || err "download failed: $url"
else
	err "need curl or wget to download $url"
fi
chmod +x "$tmp"

# --- resolve install directory ----------------------------------------------
install_dir="${GOOGLEDRIVE_MCP_INSTALL_DIR:-$HOME/.local/bin}"
dest="$install_dir/$BINARY"

# Is the destination under $HOME? A trailing slash makes "$HOME" itself match.
under_home=0
case "$install_dir/" in
	"$HOME"/*) under_home=1 ;;
esac

if [ "$under_home" -eq 1 ]; then
	# User-owned target: never sudo. Fail fast on permission problems.
	if ! mkdir -p "$install_dir" 2>/dev/null; then
		err "cannot create $install_dir — check ownership/permissions (a previous sudo install may have left it root-owned). Fix its ownership, or set GOOGLEDRIVE_MCP_INSTALL_DIR to a directory you own."
	fi
	if [ ! -w "$install_dir" ]; then
		err "$install_dir is not writable — check ownership/permissions (a previous sudo install may have left it root-owned). Fix its ownership, or set GOOGLEDRIVE_MCP_INSTALL_DIR to a directory you own."
	fi
	mv "$tmp" "$dest"
	chmod 755 "$dest"
elif mkdir -p "$install_dir" 2>/dev/null && [ -w "$install_dir" ]; then
	# Outside $HOME but writable: plain install, still no sudo.
	mv "$tmp" "$dest"
	chmod 755 "$dest"
else
	# Outside $HOME and not writable: fall back to sudo.
	printf 'Installing to %s (requires sudo)\n' "$install_dir"
	sudo mkdir -p "$install_dir" || err "could not create $install_dir"
	sudo mv "$tmp" "$dest" || err "could not install to $dest"
	sudo chmod 755 "$dest" || err "could not set permissions on $dest"
fi
trap - EXIT

# --- report ------------------------------------------------------------------
installed_version=$("$dest" --version 2>/dev/null || echo "unknown")
printf 'Installed %s %s to %s\n' "$BINARY" "$installed_version" "$dest"

# Warn (to stderr) if the install dir is not on $PATH, with a copy-paste fix.
# Exact component match: the surrounding colons prevent a substring like
# "$install_dir"ary from counting as a match. The install itself succeeded, so
# we still exit 0.
case ":${PATH}:" in
	*":$install_dir:"*) ;;
	*)
		# Infer the rc file from $SHELL where reasonable.
		case "$(basename "${SHELL:-}")" in
			zsh) rc="~/.zshrc" ;;
			bash) if [ "$os" = darwin ]; then rc="~/.bash_profile"; else rc="~/.bashrc"; fi ;;
			*) rc="" ;;
		esac
		{
			printf '\nWarning: %s is not on your PATH; the %s command will not be found until you add it.\n' "$install_dir" "$BINARY"
			printf 'Add this line to %s:\n' "${rc:-your shell config file}"
			printf '  export PATH="%s:$PATH"\n' "$install_dir"
			if [ -n "$rc" ]; then
				printf 'Then restart your shell, or run: source %s\n' "$rc"
			else
				printf 'Then restart your shell so the change takes effect.\n'
			fi
		} >&2
		;;
esac

printf '\nGet started: %s --help\n' "$BINARY"
