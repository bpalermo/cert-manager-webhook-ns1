# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.19 AS build
WORKDIR /src
ARG TARGETOS
ARG TARGETARCH
RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    go mod download
RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -o /go/bin/webhook ./cmd/webhook

FROM gcr.io/distroless/static-debian11
WORKDIR /
COPY --from=build /go/bin/webhook /webhook
ENTRYPOINT ["/webhook"]
