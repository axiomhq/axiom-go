# Axiom Go

[![GoDoc][godoc_badge]][godoc]
[![Go Workflow][go_workflow_badge]][go_workflow]
[![Coverage Status][coverage_badge]][coverage]
[![Go Report][report_badge]][report]
[![Latest Release][release_badge]][release]
[![License][license_badge]][license]
[![License Status][license_status_badge]][license_status]

> Go language bindings for the [Axiom][1] API.

  [1]: https://axiom.co

---

## Table of Contents

1. [Introduction](#introduction)
1. [Usage](#usage)
1. [Contributing](#contributing)
1. [License](#license)

## Introduction

_Axiom Go_ provides a client library for the Axiom API.

## Usage

### Installation

#### Install using `go get`

With a working Go installation (>=1.15), run:

```shell
$ go get -u github.com/axiomhq/axiom-go/axiom
```

Go 1.11 and higher _should_ be sufficient enough to use `go get` but it is not 
guaranteed that the source code does not use more recent additions to the
standard library which break building.

#### Install from source

This project uses native [go mod][2] support and requires a working Go 1.15
installation.

```shell
$ git clone https://github.com/axiomhq/axiom-go.git
$ cd axiom-go
$ make # Run code generators, linters, sanitizers and test suits
```

  [2]: https://golang.org/cmd/go/#hdr-Module_maintenance

### Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/axiomhq/axiom-go/axiom"
)

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

More examples can be found in the [example folder](example).

## Contributing

Feel free to submit PRs or to fill issues. Every kind of help is appreciated.

Before committing, `make` should run without any issues.

## License

&copy; Axiom, Inc., 2020

Distributed under MIT License (`The MIT License`).

See [LICENSE](LICENSE) for more information.

[![License Status Large][license_status_large_badge]][license_status_large]

<!-- Badges -->

[godoc]: https://github.com/axiomhq/axiom-go/axiom
[godoc_badge]: https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square
[go_workflow]: https://github.com/axiomhq/axiom-go/actions?query=workflow%3Ago
[go_workflow_badge]: https://img.shields.io/github/workflow/status/axiomhq/axiom-go/go?style=flat-square
[coverage]: https://codecov.io/gh/axiomhq/axiom-go
[coverage_badge]: https://img.shields.io/codecov/c/github/axiomhq/axiom-go.svg?style=flat-square
[report]: https://goreportcard.com/report/github.com/axiomhq/axiom-go
[report_badge]: https://goreportcard.com/badge/github.com/axiomhq/axiom-go?style=flat-square
[release]: https://github.com/axiomhq/axiom-go/releases/latest
[release_badge]: https://img.shields.io/github/release/axiomhq/axiom-go.svg?style=flat-square
[license]: https://opensource.org/licenses/MIT
[license_badge]: https://img.shields.io/github/license/axiomhq/axiom-go.svg?color=blue&style=flat-square
[license_status]: https://app.fossa.com/projects/git%2Bgithub.com%2Faxiomhq%2Faxiom-go?ref=badge_shield
[license_status_badge]: https://app.fossa.com/api/projects/git%2Bgithub.com%2Faxiomhq%2Faxiom-go.svg
[license_status_large]: https://app.fossa.com/projects/git%2Bgithub.com%2Faxiomhq%2Faxiom-go?ref=badge_large
[license_status_large_badge]: https://app.fossa.com/api/projects/git%2Bgithub.com%2Faxiomhq%2Faxiom-go.svg?type=large
