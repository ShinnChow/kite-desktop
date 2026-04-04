#!/usr/bin/env bash

set -euo pipefail

if [[ $# -lt 4 ]]; then
  echo "usage: $0 <tag> <repository> <assets-dir> <output-file>" >&2
  exit 1
fi

TAG_NAME="$1"
REPOSITORY="$2"
ASSETS_DIR="$3"
OUTPUT_FILE="$4"
VERSION="${TAG_NAME#v}"
RELEASE_BASE_URL="https://github.com/${REPOSITORY}/releases/download/${TAG_NAME}"

release_url() {
  local filename="$1"
  printf '%s/%s' "${RELEASE_BASE_URL}" "${filename}"
}

asset_exists() {
  local filename="$1"
  [[ -f "${ASSETS_DIR}/${filename}" ]]
}

emit_download_row() {
  local platform="$1"
  local architecture="$2"
  local package_type="$3"
  local filename="$4"

  if asset_exists "${filename}"; then
    printf '| %s | %s | %s | [`%s`](%s) |\n' \
      "${platform}" "${architecture}" "${package_type}" "${filename}" "$(release_url "${filename}")"
  fi
}

{
  echo "# ${TAG_NAME}"
  echo
  echo "## Downloads"
  echo
  echo "| Platform | Architecture | Package | Download |"
  echo "| --- | --- | --- | --- |"
  emit_download_row "Windows" "x64" "Installer" "Kite-${TAG_NAME}-windows-amd64-installer.exe"
  emit_download_row "Windows" "ARM64" "Installer" "Kite-${TAG_NAME}-windows-arm64-installer.exe"
  emit_download_row "macOS" "Intel" "DMG" "Kite-${TAG_NAME}-macos-intel.dmg"
  emit_download_row "macOS" "Apple Silicon" "DMG" "Kite-${TAG_NAME}-macos-apple-silicon.dmg"
  emit_download_row "Linux" "x64" "AppImage" "Kite-${TAG_NAME}-linux-amd64.AppImage"
  emit_download_row "Linux" "x64" "DEB" "Kite-${TAG_NAME}-linux-amd64.deb"
  emit_download_row "Linux" "x64" "RPM" "Kite-${TAG_NAME}-linux-amd64.rpm"
  echo

  if asset_exists "SHA256SUMS"; then
    echo "## Checksums"
    echo
    echo "- [SHA256SUMS]($(release_url "SHA256SUMS"))"
    echo
  fi

  local_server_assets=()
  for candidate in \
    "kite-amd64-${TAG_NAME}.tar.gz" \
    "kite-arm64-${TAG_NAME}.tar.gz" \
    "kite-${VERSION}.tgz"
  do
    if asset_exists "${candidate}"; then
      local_server_assets+=("${candidate}")
    fi
  done

  if [[ ${#local_server_assets[@]} -gt 0 ]]; then
    echo "## Server Assets"
    echo
    for filename in "${local_server_assets[@]}"; do
      printf -- '- [`%s`](%s)\n' "${filename}" "$(release_url "${filename}")"
    done
    echo
  fi

  echo "## Changes"
  echo
  CHANGELOG_FILE="$(mktemp)"

  PREVIOUS_TAG="$(git tag --sort=-creatordate | grep -E '^v' | grep -v "^${TAG_NAME}$" | head -n 1 || true)"
  if [[ -n "${PREVIOUS_TAG}" ]] && git rev-parse -q --verify "${PREVIOUS_TAG}^{tag}" >/dev/null 2>&1; then
    git log "${PREVIOUS_TAG}..${TAG_NAME}" --no-merges --pretty=format:'- %s (%h)' > "${CHANGELOG_FILE}" || true
    if [[ -s "${CHANGELOG_FILE}" ]]; then
      cat "${CHANGELOG_FILE}"
    else
      echo "- No changes recorded."
    fi
    echo
    echo
    echo "---"
    echo
    echo "Full Changelog: https://github.com/${REPOSITORY}/compare/${PREVIOUS_TAG}...${TAG_NAME}"
  elif git rev-parse -q --verify "${TAG_NAME}^{tag}" >/dev/null 2>&1; then
    git log "${TAG_NAME}" --no-merges --pretty=format:'- %s (%h)' > "${CHANGELOG_FILE}" || true
    if [[ -s "${CHANGELOG_FILE}" ]]; then
      cat "${CHANGELOG_FILE}"
    else
      echo "- Initial release."
    fi
  else
    echo "- Initial release."
  fi
} > "${OUTPUT_FILE}"
