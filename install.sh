#!/bin/sh
# Install sate: curl -fsSL https://raw.githubusercontent.com/sawirricardo/sate/main/install.sh | sh
set -e

REPO="sawirricardo/sate"
BIN_DIR="${SATE_INSTALL_DIR:-$HOME/.local/bin}"

case "$(uname -s)" in
	Darwin) os=darwin ;;
	Linux) os=linux ;;
	*) echo "unsupported OS: $(uname -s) — grab a binary from https://github.com/$REPO/releases" >&2; exit 1 ;;
esac
case "$(uname -m)" in
	x86_64 | amd64) arch=amd64 ;;
	arm64 | aarch64) arch=arm64 ;;
	*) echo "unsupported arch: $(uname -m) — grab a binary from https://github.com/$REPO/releases" >&2; exit 1 ;;
esac

url="https://github.com/$REPO/releases/latest/download/sate-$os-$arch"
echo "downloading $url"
mkdir -p "$BIN_DIR"
curl -fsSL "$url" -o "$BIN_DIR/sate"
chmod +x "$BIN_DIR/sate"
echo "installed $("$BIN_DIR/sate" --version) -> $BIN_DIR/sate"

case ":$PATH:" in
	*":$BIN_DIR:"*) ;;
	*) echo "note: $BIN_DIR is not on your PATH" ;;
esac
