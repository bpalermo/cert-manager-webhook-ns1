# Changelog

## [2.0.0](https://github.com/bpalermo/cert-manager-webhook-ns1/compare/v1.0.0...v2.0.0) (2026-06-20)


### ⚠ BREAKING CHANGES

* images and the Helm chart are now published to ghcr.io/bpalermo instead of Quay and Docker Hub; the chart's default image.repository changed accordingly.

### Features

* migrate build to Bazel and publish OCI artifacts to GHCR ([#11](https://github.com/bpalermo/cert-manager-webhook-ns1/issues/11)) ([cb317d3](https://github.com/bpalermo/cert-manager-webhook-ns1/commit/cb317d330ecb2c1f85095dc9e5da4470b69bc276))
