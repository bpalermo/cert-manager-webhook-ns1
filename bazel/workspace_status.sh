#!/usr/bin/env bash
#
# Emits stable/volatile build stamps consumed by rules_img / rules_helm.
# VERSION is taken from $RELEASE_VERSION when set (CI), otherwise derived
# from git, falling back to a dev marker.
set -euo pipefail

version="${RELEASE_VERSION:-}"
if [[ -z "${version}" ]]; then
  version="$(git describe --tags --always --dirty 2>/dev/null || echo "0.0.0-dev")"
fi

git_sha="$(git rev-parse HEAD 2>/dev/null || echo "unknown")"

echo "STABLE_VERSION ${version#v}"
echo "STABLE_GIT_SHA ${git_sha}"
