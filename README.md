# NS1 Webhook for Cert Manager
This is a webhook solver for NS1, for use with cert-manager, to solve ACME DNS01 challenges.

> **NOTE:** This is a fork of [cert-manager-webhook-ns1](https://github.com/ns1/cert-manager-webhook-ns1)
> with arm64 support, a Bazel-native build, and OCI artifacts published to GHCR.

## Install

The image and Helm chart are published as OCI artifacts to GitHub Container Registry:

- Image: `ghcr.io/bpalermo/cert-manager-webhook-ns1`
- Chart: `oci://ghcr.io/bpalermo/charts/cert-manager-webhook-ns1`

```sh
helm install cert-manager-webhook-ns1 \
  oci://ghcr.io/bpalermo/charts/cert-manager-webhook-ns1 \
  --namespace cert-manager
```

The NS1 API key is supplied via a Secret referenced from the Issuer's
`webhook.config.apiKeySecretRef`.

## Development

The build is driven by [Bazel](https://bazel.build) (via Bazelisk; the version
is pinned in `.bazelversion`).

```sh
bazel build //...                 # build everything (binary + image + chart)
bazel test //...                  # run tests
bazel run //:gazelle              # regenerate Go BUILD files
bazel build //image:image         # build the multi-arch container image
bazel run //image:push            # push the image to GHCR (needs auth)
bazel build //chart:chart         # package the Helm chart (.tgz)
bazel run //chart:chart.push_registry  # push the chart to GHCR (needs auth)
```

After adding/removing Go imports, run `bazel run //:gazelle` and `bazel mod tidy`.
