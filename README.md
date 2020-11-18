# Axiom-Go

[![GoDoc][godoc_badge]][godoc]
[![Go Workflow][go_workflow_badge]][go_workflow]
[![Coverage Status][coverage_badge]][coverage]
[![Go Report][report_badge]][report]
[![Latest Release][release_badge]][release]
[![License][license_badge]][license]

--------

## Table of Contents

1. [Introduction](#introduction)
1. [Installation](#Installation)
1. [Usage](#usage)
1. [Authentication](#authentication)
1. [Documentation](#documentation)
1. [Contributing](#contributing)
1. [License](#license)

## Introduction

Axiom-Go is a Go client library for accessing the [Axiom](https://www.axiom.co/)
API.

Currently, **Axiom-Go requires Go 1.12 or greater**.

## Installation

1. #### Install using `go get`

With a working Go installation, run:

```shell
$ go get github.com/axiomhq/axiom-go/axiom
```

2.  #### Install from source

This project uses native
[go mod](https://golang.org/cmd/go/#hdr-Module_maintenance) support and requires
a working Go 1.15 installation.

```shell
$ git clone https://github.com/axiomhq/axiom-go.git
$ cd axiom-go
$ make # Run code generators, linters, sanitizers and test suits
```

## Usage

The purpose of this how to use the Axiom-Go client library to access the
[Axiom](https://www.axiom.co/) API. This example shows how to stream the
contents of a JSON using the Axiom-Go Library.

We have several examples [on the website](https://docs.axiom.co/).

`import "github.com/axiomhq/axiom-go/axiom"` // import path

------

Set the `AXIOM_DEPLOYMENT_URL` & `AXIOM_ACCESS_TOKEN` environment variables.

```go

func main() {
	var (
		deploymentURL = os.Getenv("AXM_DEPLOYMENT_URL")
		accessToken   = os.Getenv("AXM_ACCESS_TOKEN")
	)
```

----

```go
	// Open the file to ingest.
	f, err := os.Open("logs.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	//  Wrap in a gzip enabled reader.
	r, err := axiom.GZIPStreamer(f, gzip.BestSpeed)
	if err != nil {
		log.Fatal(err)
    }
```

------

Construct a new Axiom Client, then use the various services on the client to
access different parts of the Axiom API. For example:

```go
    // Initialize the Axiom API client.
	client, err := axiom.NewClient(deploymentURL, accessToken)
	if err != nil {
		log.Fatal(err)
    }
```

----

 Ingest the Data ⚡

 ```go
	// Note the JSON content type and GZIP content encoding being set because
	// the client does not auto sense them.

	res, err := client.Datasets.Ingest(context.Background(), "test", r, axiom.JSON, axiom.GZIP, axiom.IngestOptions{})
	if err != nil {
		log.Fatal(err)
    }
```

Make sure everything runs smoothly.

```go
	for _, fail := range res.Failures {
		log.Print(fail.Error)
	}
}
```

For more sample code snippets, head over to the[examples](examples) directory.

## Authentication

```go
func main() {
	client, err := axiom.NewClient("https://my-axiom.example.com", "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX")
	if err != nil {
		log.Fatal(err)
	}

	datasets, err := client.Datasets.List(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(datasets)
}
```

## Documentation

You can find the Axiom and Axiom-Go documentation
[on the website](https://docs.axiom.co/)

Check out the [Getting Started](https://docs.axiom.co/) page for a quick
overview.

The documentation is divided into several sections:

- [Tutorial](https://docs.axiom.co/getting-started/)
- [Ingesting](https://docs.axiom.co/usage/ingest/)
- [Analyzing](https://docs.axiom.co/usage/analyze/)
- [Streaming](https://docs.axiom.co/usage/stream/)
- [Alerting](https://docs.axiom.co/usage/alerts/)
- [Integrations](https://docs.axiom.co/usage/integrations/)
- [Where to Get Support](axiom.co/community)
- [Contributing Guide](https://docs.axiom.co/how-to-contribute/)

## Contributing

The main aim of this repository is to continue developing and advancing
Axiom-Go, making it faster and more simplified to use. Kindly check our
[contributing guide](https://github.com/axiomhq/axiom-go/blob/main/Contributing.md)
on how to propose bugfixes and improvements, and submitting pull requests to the
project.

## License

&copy; Axiom, Inc., 2020

Distributed under MIT License (`The MIT License`).

See [LICENSE](LICENSE) for more information.

[![License Status Large][license_status_badge]][license_status]

<!-- Badges -->

[godoc]: https://github.com/axiomhq/axiom-go/axiom
[godoc_badge]: https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square&dummy=unused
[go_workflow]: https://github.com/axiomhq/axiom-go/actions?query=workflow%3Ago
[go_workflow_badge]: https://img.shields.io/github/workflow/status/axiomhq/axiom-go/go?style=flat-square&dummy=unused
[coverage]: https://codecov.io/gh/axiomhq/axiom-go
[coverage_badge]: https://img.shields.io/codecov/c/github/axiomhq/axiom-go.svg?style=flat-square&dummy=unused
[report]: https://goreportcard.com/report/github.com/axiomhq/axiom-go
[report_badge]: https://goreportcard.com/badge/github.com/axiomhq/axiom-go?style=flat-square&dummy=unused
[release]: https://github.com/axiomhq/axiom-go/releases/latest
[release_badge]: https://img.shields.io/github/release/axiomhq/axiom-go.svg?style=flat-square&dummy=unused
[license]: https://opensource.org/licenses/MIT
[license_badge]: https://img.shields.io/github/license/axiomhq/axiom-go.svg?color=blue&style=flat-square&dummy=unused
[license_status]: https://app.fossa.com/projects/git%2Bgithub.com%2Faxiomhq%2Faxiom-go
[license_status_badge]: https://app.fossa.com/api/projects/git%2Bgithub.com%2Faxiomhq%2Faxiom-go.svg?type=large&dummy=unused
