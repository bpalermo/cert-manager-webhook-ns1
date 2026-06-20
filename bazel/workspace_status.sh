#!/usr/bin/env bash
#
# Emits build stamps consumed by rules_img / rules_helm.
#
# STABLE_IMAGE_COMMIT_SHA identifies the app/image version: the full SHA of the
# last commit that touched any image build input. A chart-only change leaves it
# unchanged, so the image is not re-tagged. CI sets $IMAGE_COMMIT_SHA (computed
# once); locally it is derived from git.
set -euo pipefail

# Paths whose changes affect the built image (everything except chart/).
IMAGE_PATHS=(cmd pkg go.mod go.sum MODULE.bazel MODULE.bazel.lock)

image_sha="${IMAGE_COMMIT_SHA:-}"
if [[ -z "${image_sha}" ]]; then
  image_sha="$(git log -1 --format=%H -- "${IMAGE_PATHS[@]}" 2>/dev/null || true)"
  [[ -z "${image_sha}" ]] && image_sha="dev"
fi

echo "STABLE_IMAGE_COMMIT_SHA ${image_sha}"
echo "STABLE_GIT_SHA $(git rev-parse HEAD 2>/dev/null || echo unknown)"
