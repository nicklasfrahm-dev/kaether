# kontinuum

A Kubernetes-style API server built on [kommodity](https://github.com/kommodity-io/kommodity)'s `libkapi` package. It embeds a generic apiserver + apiextensions (CRD) server + aggregation layer, backed by pluggable storage (SQLite, PostgreSQL, etcd, ...).

> **Warning:** The server ships with no TLS and no authentication by default. Put a TLS-terminating, authenticating proxy in front before exposing it outside a trusted network.

## Quick start

### Prerequisites

- [Go 1.26+](https://go.dev/dl/) — we recommend using [gvm](https://github.com/moovweb/gvm) to manage Go versions:
  ```sh
  gvm install go1.26.4 -B
  gvm use go1.26.4 --default
  ```
- [Docker](https://docs.docker.com/get-docker/) (for the dev environment)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

### Build & run

```sh
make build
make run
```

### Development environment

Starts PostgreSQL and [air](https://github.com/air-verse/air) hot-reload via Docker Compose:

```sh
# start dev environment (air + postgres)
make dev
# stop dev environment
make dev-down
# stop and remove volumes
make dev-clean
```

### Connect with kubectl

```sh
export KUBECONFIG=kontinuum.yaml
kubectl get namespaces
kubectl create namespace demo
```

## Configuration

Configuration is loaded from `KONTINUUM_`-prefixed environment variables. Env-var names are auto-derived from the config struct path (e.g. `Server.Addr` → `KONTINUUM_SERVER_ADDR`).

| Env var                    | Description                                                                        | Default                 |
| -------------------------- | ----------------------------------------------------------------------------------- | ----------------------- |
| `KONTINUUM_SERVER_ADDR`    | Listener address                                                                     | `:8080`                 |
| `KONTINUUM_SERVER_STORAGE` | Storage connection string (`sqlite://`, `postgres://`, `mysql://`, `etcd://`, ...)   | `sqlite://kontinuum.db` |
| `KONTINUUM_LOG_LEVEL`      | Log level (`debug`, `info`, `warn`, `error`)                                         | `warn`                  |
| `KONTINUUM_LOG_FORMAT`     | Log format (`console`, `text`, `json`)                                              | `json`                  |

Flags override environment variables when explicitly set:

```sh
kontinuum serve --addr :9090 --storage postgres://user:pass@host/db
```

## Make targets

```
Usage:
  make <target>

General
  help           Display this help

Development
  build          Build the binary
  run            Run the server locally with dev-friendly logging (info, console)
  dev            Start development environment with hot reload (air + postgres)
  dev-down       Stop development environment
  dev-clean      Stop development environment and remove volumes
  image          Build the container image

Quality
  test           Run tests
  vet            Run go vet
  lint           Run golangci-lint
  lint-fix       Run golangci-lint and fix issues
  tidy           Download and tidy dependencies

Cleanup
  clean          Remove build artifacts
```

## Container

```sh
# builds kontinuum:<version> via distroless/static
make image
docker run -p 8080:8080 -e KONTINUUM_SERVER_STORAGE=postgres://... kontinuum:latest
```

The container image is built on `distroless/static` with `CGO_ENABLED=0`, so SQLite storage is not available — use PostgreSQL or etcd.

## License

Apache License 2.0.
