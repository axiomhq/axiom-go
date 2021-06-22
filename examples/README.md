# Examples

## Before you start

Axiom Go and the Adapters automatically pick up their configuration from the
environment, if not otherwise specified. To learn more about cunfiguration,
check the [Documentation](https://pkg.go.dev/github.com/axiomhq/axiom-go).

To quickstart, export the environment variables below.

### When using Axiom Cloud

* `AXIOM_TOKEN`: **Personal Access** or **Ingest** token. Can be
  created under `Profile` or `Settings > Ingest Tokens`. For security reasons it
  is advised to use an Ingest Token with minimal privileges only.
* `AXIOM_ORG_ID`: Organization identifier of the organization to use on Axiom
   Cloud.
* `AXIOM_DATASET`: Dataset to use. Must exist prior to using it.

### When using Axiom Selfhost

* `AXIOM_URL`: URL of the Axiom deployment to use.
* `AXIOM_TOKEN`: **Personal Access** or **Ingest** token. Can be
  created under `Profile` or `Settings > Ingest Tokens`. For security reasons it
  is advised to use an Ingest Token with minimal privileges only.
* `AXIOM_DATASET`: Dataset to use. Must exist prior to using it.

## Package usage

* [ingestfile](ingestfile/main.go): How to ingest the contents of a file into
  Axiom and compress them on the fly.
* [query](query/main.go): How to query a dataset using the Kusto-like Axiom
  Processing Language (APL).

## Adapter usage

* [apex](apex/main.go): How to ship logs to Axiom using the popular
  [Apex](https://github.com/apex/log) logging package.
* [logrus](logrus/main.go): How to ship logs to Axiom using the popular
  [Logrus](https://github.com/sirupsen/logrus) logging package.
* [zap](zap/main.go): How to ship logs to Axiom using the popular
  [Zap](https://github.com/uber-go/zap) logging package.
