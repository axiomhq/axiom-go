# Examples

This directory contains examples that showcase the usage of Axiom Go. Each
example is a self-contained Go package that can be run with `go run`:

```shell
go run ./{example}
```

## Before you start

Axiom Go and the adapters automatically pick up their configuration from the
environment, if not otherwise specified. To learn more about configuration,
check the
[documentation](https://pkg.go.dev/github.com/axiomhq/axiom-go/adapters).

To quickstart, export the environment variables below.

> [!NOTE]
> If you have the [Axiom CLI](github.com/axiomhq/cli) installed and are logged
> in, you can easily export most of the required environment variables:
>
>```shell
>eval $(axiom config export -f)
>```

### Required environment variables

- `AXIOM_TOKEN`: **API** or **Personal** token. Can be created under
  `Settings > API Tokens` or `Profile`. For security reasons it is advised to
  use an API token with minimal privileges only.
- `AXIOM_ORG_ID`: Organization identifier of the organization to (when using a
  personal token).
- `AXIOM_DATASET`: Dataset to use. Must exist prior to using it. You can use
  [Axiom CLI](github.com/axiomhq/cli) to create a dataset:
  `axiom dataset create`.

## Package usage

- [ingestevent](ingestevent/main.go): How to ingest events into Axiom.
- [ingestfile](ingestfile/main.go): How to ingest the contents of a file into
  Axiom and compress them on the fly.
- [ingesthackernews](ingesthackernews/main.go): How to ingest the contents of
  Hacker News into Axiom.
- [query](query/main.go): How to query a dataset using the Kusto-like Axiom
  Processing Language (APL).
- [querylegacy](querylegacy/main.go): How to query a dataset using the legacy
  query datatypes.

## Adapter usage

- [apex](apex/main.go): How to ship logs to Axiom using the popular
  [Apex](https://github.com/apex/log) logging package.
- [logrus](logrus/main.go): How to ship logs to Axiom using the popular
  [Logrus](https://github.com/sirupsen/logrus) logging package.
- [slog](slog/main.go): How to ship logs to Axiom using the standard libraries
  [Slog](https://pkg.go.dev/log/slog) structured logging package.
- [slogx](slogx/main.go): How to ship logs to Axiom using the
  [golang.org/x/exp/slog](https://pkg.go.dev/golang.org/x/exp/slog) structured
  logging package (pre Go 1.21).
- [zap](zap/main.go): How to ship logs to Axiom using the popular
  [Zap](https://github.com/uber-go/zap) logging package.

## OpenTelemetry usage

- [otelinstrument](otelinstrument/main.go): How to instrument the Axiom Go
  client using OpenTelemetry.
- [oteltraces](oteltraces/main.go): How to ship traces to Axiom using the
  OpenTelemetry Go SDK and the Axiom SDKs `otel` helper package.
