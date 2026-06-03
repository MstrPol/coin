#!/usr/bin/env bash
# Определяет GOARCH / DOCKER_PLATFORM для текущего хоста.
set -euo pipefail

host_arch() {
  case "$(uname -m)" in
    arm64|aarch64) echo "arm64" ;;
    x86_64|amd64) echo "amd64" ;;
    *)
      echo "unsupported arch: $(uname -m)" >&2
      exit 1
      ;;
  esac
}

coin_platform() {
  echo "linux/$(host_arch)"
}

if [[ "${1:-}" == "--goarch" ]]; then
  host_arch
elif [[ "${1:-}" == "--platform" ]]; then
  coin_platform
else
  echo "usage: $0 --goarch|--platform" >&2
  exit 1
fi
