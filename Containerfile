# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.26-trixie AS build

ARG TARGETOS
ARG TARGETARCH
ARG VERSION

WORKDIR /app

ENV GOCACHE=/go-cache
ENV GOMODCACHE=/gomod-cache
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/gomod-cache go mod download

COPY . .
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w -X github.com/nicklasfrahm/kontinuum/pkg/cli.version=${VERSION}" -o /out/kontinuum ./cmd/kontinuum

FROM gcr.io/distroless/static-debian13 AS runtime

WORKDIR /app
COPY --from=build /out/kontinuum /app/kontinuum

EXPOSE 8080
ENTRYPOINT ["/app/kontinuum", "serve"]
