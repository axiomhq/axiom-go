# Axiom Go

[![Go Reference][gopkg_badge]][gopkg]
[![Workflow][workflow_badge]][workflow]
[![Coverage Status][coverage_badge]][coverage]
[![Go Report][report_badge]][report]
[![Latest Release][release_badge]][release]
[![License][license_badge]][license]

---

## Introduction

Axiom Go is a Go client library for accessing the [Axiom](https://www.axiom.co/)
API.

Currently, **Axiom Go requires Go 1.16 or greater**.

## Installation

### Install using `go get`

```shell
go get github.com/axiomhq/axiom-go/axiom
```

### Install from source

```shell
git clone https://github.com/axiomhq/axiom-go.git
cd axiom-go
make # Run code generators, linters, sanitizers and test suits
```

## Authentication

The client is initialized with an access token and the users organization ID
when using Axiom Cloud or with the URL of the deployment and an access token
when using Axiom Selfhost. The organization ID can be omitted in case an API
token is used.

The access token can an API token retrieved from the settings page of the Axiom
deployment or a personal token retrieved from the users profile page.

The API token just allows ingestion or querying into or from the datasets the
token is configured for.

The personal access token grants access to all resources available to the user
on his behalf.

## Usage

Simple file ingestion with on the fly compression:

```go
// Export `AXIOM_TOKEN` and `AXIOM_ORG_ID` (when using a personal token) for
// Axiom Cloud.
// Export `AXIOM_URL` and `AXIOM_TOKEN` for Axiom Selfhost.

// 1. Open the file to ingest.
f, err := os.Open("logs.json")
if err != nil {
	log.Fatal(err)
}
defer f.Close()

// 2. Wrap it in a gzip enabled reader.
r, err := axiom.GzipEncoder(f)
if err != nil {
	log.Fatal(err)
}

// 3. Initialize the Axiom API client.
client, err := axiom.NewClient()
if err != nil {
	log.Fatal(err)
}

// 4. Ingest ⚡
// Note the JSON content type and Gzip content encoding being set because the
// client does not auto sense them.
res, err := client.Datasets.Ingest(context.Background(), dataset, r, axiom.JSON, axiom.Gzip, axiom.IngestOptions{})
if err != nil {
	log.Fatal(err)
}

// 5. Make sure everything went smoothly.
for _, fail := range res.Failures {
	log.Print(fail.Error)
}
```

For more sample code snippets, head over to the [examples](examples) directory.

## Documentation

- Visit the Go bindings documentation for the Axiom API on [go.dev](https://pkg.go.dev/github.com/axiomhq/axiom-go/axiom). 

- Also get a detailed overview on [Constants](https://pkg.go.dev/github.com/axiomhq/axiom-go/axiom#pkg-constants), [variables](https://pkg.go.dev/github.com/axiomhq/axiom-go/axiom#pkg-variables), [Functions](https://pkg.go.dev/github.com/axiomhq/axiom-go/axiom#pkg-functions), and [Types](https://pkg.go.dev/github.com/axiomhq/axiom-go/axiom#pkg-types), and [Source Files](https://pkg.go.dev/github.com/axiomhq/axiom-go/axiom#section-sourcefiles) in the [go.dev documentation](https://pkg.go.dev/github.com/axiomhq/axiom-go/axiom#pkg-index). 



You can find the Axiom and Axiom Go documentation
[on the docs website](https://docs.axiom.co/).

The documentation is divided into several sections:

- [Axiom Cloud](https://www.axiom.co/docs/install/cloud)
- [Getting Started](https://docs.axiom.co/usage/getting-started/)
- [Ingesting Data](https://docs.axiom.co/usage/ingest/)
- [Analyzing Data](https://docs.axiom.co/usage/analyze/)
- [Streaming Data](https://docs.axiom.co/usage/stream/)
- [Runing Axiom on Kubernetes](https://docs.axiom.co/install/kubernetes/)
- [Run Axiom on your Desktop](https://docs.axiom.co/install/demo/)
- [Manage deployments with Axiom CLI](https://www.axiom.co/docs/reference/cli)
- [Ingest using Elastic Beats](https://docs.axiom.co/data-shippers/elastic-beats/)
- [Ingesting via Elasticsearch API](https://docs.axiom.co/data-shippers/api/)
- [Where to Get Support](https://axiom.co/support)

## Contributing

The main aim of this repository is to continue developing and advancing
Axiom Go, making it faster and more simplified to use. Kindly check our
[contributing guide](https://github.com/axiomhq/axiom-go/blob/main/Contributing.md)
on how to propose bugfixes and improvements, and submitting pull requests to the
project.

## License

&copy; Axiom, Inc., 2022

Distributed under MIT License (`The MIT License`).

See [LICENSE](LICENSE) for more information.

[![License Status][license_status_badge]][license_status]

<!-- Badges -->

[gopkg]: https://pkg.go.dev/github.com/axiomhq/axiom-go
[gopkg_badge]: https://img.shields.io/badge/doc-reference-007d9c?logo=go&logoColor=white&style=flat-square
[workflow]: https://github.com/axiomhq/axiom-go/actions/workflows/push.yml
[workflow_badge]: https://img.shields.io/github/workflow/status/axiomhq/axiom-go/Push?style=flat-square&ghcache=unused
[coverage]: https://codecov.io/gh/axiomhq/axiom-go
[coverage_badge]: https://img.shields.io/codecov/c/github/axiomhq/axiom-go.svg?style=flat-square&ghcache=unused
[report]: https://goreportcard.com/report/github.com/axiomhq/axiom-go
[report_badge]: https://goreportcard.com/badge/github.com/axiomhq/axiom-go?style=flat-square&ghcache=unused
[release]: https://github.com/axiomhq/axiom-go/releases/latest
[release_badge]: https://img.shields.io/github/release/axiomhq/axiom-go.svg?style=flat-square&ghcache=unused
[license]: https://opensource.org/licenses/MIT
[license_badge]: https://img.shields.io/github/license/axiomhq/axiom-go.svg?color=blue&style=flat-square&ghcache=unused
[license_status]: https://app.fossa.com/projects/git%2Bgithub.com%2Faxiomhq%2Faxiom-go
[license_status_badge]: https://app.fossa.com/api/projects/git%2Bgithub.com%2Faxiomhq%2Faxiom-go.svg?type=large&ghcache=unused
