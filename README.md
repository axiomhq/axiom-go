# axiom-go [![Go Reference][gopkg_badge]][gopkg] [![Workflow][workflow_badge]][workflow] [![Latest Release][release_badge]][release] [![License][license_badge]][license]

If you use the [Axiom CLI](https://github.com/axiomhq/cli), run
`eval $(axiom config export -f)` to configure your environment variables.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/axiomhq/axiom-go/axiom"
    "github.com/axiomhq/axiom-go/axiom/ingest"
)

func main() {
    ctx := context.Background()

    client, err := axiom.NewClient(
        // If you don't want to configure your client using the environment,
        // pass credentials explicitly.
        // Personal token:
        // axiom.SetPersonalTokenConfig("xapt-xyz", "acme-1234"),
        // API token:
        // axiom.SetToken("xaat-xyz"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    if _, err = client.IngestEvents(ctx, "my-dataset", []axiom.Event{
        {ingest.TimestampField: time.Now(), "foo": "bar"},
        {ingest.TimestampField: time.Now(), "bar": "foo"},
    }); err != nil {
        log.Fatal(err)
    }

    res, err := client.Query(ctx, "['my-dataset'] | where foo == 'bar' | limit 100")
    if err != nil {
        log.Fatal(err)
    } else if res.Status.RowsMatched == 0 {
        log.Fatal("No matches found")
    }

    rows := res.Tables[0].Rows()
    if err := rows.Range(ctx, func(_ context.Context, row query.Row) error {
        _, err := fmt.Println(row)
        return err
    }); err != nil {
        log.Fatal(err)
    }
}
```

For further examples, head over to the [examples](examples) directory.

If you want to use a logging package, check if there is already an adapter in
the [adapters](adapters) directory. We happily accept contributions for new
adapters.


## Install

```shell
go get github.com/axiomhq/axiom-go
```

## Documentation

Read documentation on [axiom.co/docs/guides/go](https://axiom.co/docs/guides/go).

## License

[MIT](LICENSE)

<!-- Badges -->

[gopkg]: https://pkg.go.dev/github.com/axiomhq/axiom-go
[gopkg_badge]: https://img.shields.io/badge/doc-reference-007d9c?logo=go&logoColor=white
[workflow]: https://github.com/axiomhq/axiom-go/actions/workflows/push.yaml
[workflow_badge]: https://img.shields.io/github/actions/workflow/status/axiomhq/axiom-go/push.yaml?branch=main&ghcache=unused
[release]: https://github.com/axiomhq/axiom-go/releases/latest
[release_badge]: https://img.shields.io/github/release/axiomhq/axiom-go.svg?ghcache=unused
[license]: https://opensource.org/licenses/MIT
[license_badge]: https://img.shields.io/github/license/axiomhq/axiom-go.svg?color=blue&ghcache=unused
