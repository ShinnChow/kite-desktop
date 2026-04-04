#!/usr/bin/env bash

set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "usage: $0 <version-without-v> <homepage-url>" >&2
  exit 1
fi

VERSION="$1"
HOMEPAGE_URL="$2"
export KITE_VERSION="${VERSION}"
export KITE_HOMEPAGE="${HOMEPAGE_URL}"

perl -0pi -e 's/(version:\s*")[^"]+(")/${1}$ENV{KITE_VERSION}$2/g' \
  desktop/build/config.yml \
  desktop/build/linux/nfpm/nfpm.yaml

perl -0pi -e 's!(homepage:\s*")[^"]+(")!${1}$ENV{KITE_HOMEPAGE}$2!g' \
  desktop/build/linux/nfpm/nfpm.yaml

(
  cd desktop
  wails3 task common:update:build-assets
)

echo "desktop release metadata updated: version=${VERSION} homepage=${HOMEPAGE_URL}"
