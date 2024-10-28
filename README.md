# axiom-go [![Go Reference][gopkg_badge]][gopkg] [![Workflow][workflow_badge]][workflow] [![Latest Release][release_badge]][release] [![License][license_badge]][license]

```go
package main

import (
    "ctx"
    "log"

    "github.com/axiomhq/axiom-go/axiom"
)

func main() {
    ctx := context.Background()

    client, err := axiom.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    if _, err = client.IngestEvents(ctx, "my-dataset", []axiom.Event{
        {"foo": "bar"},
        {"bar": "foo"},
    }); err != nil {
        log.Fatal(err)
    }

    res, err := client.Query(ctx, "['my-dataset'] | where foo == 'bar' | limit 100")
    if err != nil {
        log.Fatal(err)
    }
}
```

## Install

```sh
go get github.com/axiomhq/axiom-go
```

## Documentation

Read documentation on [axiom.co/docs/guides/go](https://axiom.co/docs/guides/go).

## License

[MIT](LICENSE).

<!-- Badges -->

[gopkg]: https://pkg.go.dev/github.com/axiomhq/axiom-go
[gopkg_badge]: https://img.shields.io/badge/doc-reference-007d9c?logo=go&logoColor=white
[workflow]: https://github.com/axiomhq/axiom-go/actions/workflows/push.yaml
[workflow_badge]: https://img.shields.io/github/actions/workflow/status/axiomhq/axiom-go/push.yaml?branch=main&ghcache=unused
[release]: https://github.com/axiomhq/axiom-go/releases/latest
[release_badge]: https://img.shields.io/github/release/axiomhq/axiom-go.svg?ghcache=unused
[license]: https://opensource.org/licenses/MIT
[license_badge]: https://img.shields.io/github/license/axiomhq/axiom-go.svg?color=blue&ghcache=unused
