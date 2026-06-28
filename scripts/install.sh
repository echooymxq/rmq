#!/usr/bin/env bash
set -euo pipefail

RMQ_GITHUB_REPO="echooymxq/rmq"
RMQ_BINARY="rmq"
RMQ_INSTALL_DIR="${BINDIR:-/usr/local/bin}"
RMQ_REQUESTED_VERSION="${VERSION:-latest}"

fail() {
  echo "rmq installer: $*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

detect_rmq_os() {
  case "$(uname -s)" in
    Darwin)
      echo "darwin"
      ;;
    Linux)
      echo "linux"
      ;;
    *)
      fail "unsupported OS: $(uname -s)"
      ;;
  esac
}

detect_rmq_arch() {
  case "$(uname -m)" in
    x86_64 | amd64)
      echo "amd64"
      ;;
    arm64 | aarch64)
      echo "arm64"
      ;;
    *)
      fail "unsupported architecture: $(uname -m)"
      ;;
  esac
}

resolve_rmq_release_tag() {
  if [ -n "$RMQ_REQUESTED_VERSION" ] && [ "$RMQ_REQUESTED_VERSION" != "latest" ]; then
    case "$RMQ_REQUESTED_VERSION" in
      v*)
        echo "$RMQ_REQUESTED_VERSION"
        ;;
      *)
        echo "v${RMQ_REQUESTED_VERSION}"
        ;;
    esac
    return
  fi

  local latest_url
  latest_url="$(curl -fsSLI -o /dev/null -w "%{url_effective}" "https://github.com/${RMQ_GITHUB_REPO}/releases/latest")" ||
    fail "failed to resolve latest release for ${RMQ_GITHUB_REPO}"

  local version
  version="${latest_url##*/}"
  if [ -z "$version" ] || [ "$version" = "latest" ]; then
    fail "failed to parse latest release version from ${latest_url}"
  fi
  echo "$version"
}

find_rmq_binary() {
  local root="$1"
  local path
  path=""
  while IFS= read -r candidate; do
    path="$candidate"
    break
  done < <(find "$root" -type f -name "$RMQ_BINARY")
  if [ -z "$path" ]; then
    fail "archive does not contain ${RMQ_BINARY}"
  fi
  echo "$path"
}

require_command curl
require_command tar
require_command uname
require_command find
require_command mktemp

rmq_os="$(detect_rmq_os)"
rmq_arch="$(detect_rmq_arch)"
rmq_release_tag="$(resolve_rmq_release_tag)"
rmq_version="${rmq_release_tag#v}"
rmq_archive="${RMQ_BINARY}_${rmq_version}_${rmq_os}_${rmq_arch}.tar.gz"
rmq_download_url="https://github.com/${RMQ_GITHUB_REPO}/releases/download/${rmq_release_tag}/${rmq_archive}"

tmpdir="$(mktemp -d "${TMPDIR:-/tmp}/rmq.install.XXXXXX")"
trap 'rm -rf "$tmpdir"' EXIT

echo "rmq installer: downloading ${rmq_download_url}"
curl -fsSL "$rmq_download_url" -o "${tmpdir}/${rmq_archive}" ||
  fail "failed to download ${rmq_archive}"

tar -xzf "${tmpdir}/${rmq_archive}" -C "$tmpdir" ||
  fail "failed to extract ${rmq_archive}"

binary_path="$(find_rmq_binary "$tmpdir")"

mkdir -p "$RMQ_INSTALL_DIR" || fail "failed to create install directory: ${RMQ_INSTALL_DIR}"
if [ ! -w "$RMQ_INSTALL_DIR" ]; then
  fail "install directory is not writable: ${RMQ_INSTALL_DIR}; set BINDIR to a writable directory"
fi

install_path="${RMQ_INSTALL_DIR%/}/${RMQ_BINARY}"
cp "$binary_path" "$install_path" || fail "failed to install ${RMQ_BINARY} to ${install_path}"
chmod 0755 "$install_path" || fail "failed to make ${install_path} executable"

echo "rmq installer: installed rmq ${rmq_version} to ${install_path}"
case ":${PATH}:" in
  *":${RMQ_INSTALL_DIR}:"*) ;;
  *) echo "rmq installer: warning: ${RMQ_INSTALL_DIR} is not in PATH" ;;
esac
