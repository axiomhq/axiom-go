# Examples

This directory contains examples that showcase the usage of Axiom Go. Each
example is a self-contained Go package that can be run with `go run`:

```shell
go run ./{example}
```

## Before you start

Axiom Go and the Adapters automatically pick up their configuration from the
environment, if not otherwise specified. To learn more about configuration,
check the [Documentation](https://pkg.go.dev/github.com/axiomhq/axiom-go).

To quickstart, export the environment variables below.

💡 _If you have the [Axiom CLI](https://direnv.net/) installed and are
logged in, you can easily export most of the required environment variables:_

```shell
eval $(axiom config export -f)
```

### When using Axiom Cloud

- `AXIOM_TOKEN`: **API** or **Personal Access** token. Can be created under
  `Settings > API Tokens` or `Profile`. For security reasons it is advised to
  use an API token with minimal privileges only.
- `AXIOM_ORG_ID`: Organization identifier of the organization to use on Axiom
  Cloud (when using a personal access token).
- `AXIOM_DATASET`: Dataset to use. Must exist prior to using it.

### When using Axiom Selfhost

- `AXIOM_URL`: URL of the Axiom deployment to use.
- `AXIOM_TOKEN`: **API** or **Personal Access** token. Can be created under
  `Settings > API Tokens` or `Profile`. For security reasons it is advised to
  use an API token with minimal privileges only.
- `AXIOM_DATASET`: Dataset to use. Must exist prior to using it.

## Package usage

- [apl](apl/main.go): How to query a dataset using the Kusto-like Axiom
  Processing Language (APL).
- [ingestevent](ingestevent/main.go): How to ingest events into Axiom.
- [ingestfile](ingestfile/main.go): How to ingest the contents of a file into
  Axiom and compress them on the fly.
- [ingesthackernews](ingesthackernews/main.go): How to ingest the contents of
  Hacker News into Axiom.
- [query](query/main.go): How to query a dataset using the datatypes provided by
  Axiom Go.

## Adapter usage

- [apex](apex/main.go): How to ship logs to Axiom using the popular
  [Apex](https://github.com/apex/log) logging package.
- [logrus](logrus/main.go): How to ship logs to Axiom using the popular
  [Logrus](https://github.com/sirupsen/logrus) logging package.
- [zap](zap/main.go): How to ship logs to Axiom using the popular
  [Zap](https://github.com/uber-go/zap) logging package.

## OpenTelemetry usage

- [oteltraces](oteltraces/main.go): How to ship traces to Axiom using the
  OpenTelemetry Go SDK and the Axiom SDKs `otel` helper package.
