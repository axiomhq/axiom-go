# axiom-go

<a href="https://axiom.co">
<picture>
  <source media="(prefers-color-scheme: dark) and (min-width: 600px)" srcset="https://axiom.co/assets/github/axiom-github-banner-light-vertical.svg">
  <source media="(prefers-color-scheme: light) and (min-width: 600px)" srcset="https://axiom.co/assets/github/axiom-github-banner-dark-vertical.svg">
  <source media="(prefers-color-scheme: dark) and (max-width: 599px)" srcset="https://axiom.co/assets/github/axiom-github-banner-light-horizontal.svg">
  <img alt="Axiom.co banner" src="https://axiom.co/assets/github/axiom-github-banner-dark-horizontal.svg" align="right">
</picture>
</a>
&nbsp;

[![Go Reference][gopkg_badge]][gopkg]
[![Workflow][workflow_badge]][workflow]
[![Latest Release][release_badge]][release]
[![License][license_badge]][license]

[Axiom](https://axiom.co) unlocks observability at any scale.

- **Ingest with ease, store without limits:** Axiom's next-generation datastore
  enables ingesting petabytes of data with ultimate efficiency. Ship logs from
  Kubernetes, AWS, Azure, Google Cloud, DigitalOcean, Nomad, and others.
- **Query everything, all the time:** Whether DevOps, SecOps, or EverythingOps,
  query all your data no matter its age. No provisioning, no moving data from
  cold/archive to "hot", and no worrying about slow queries. All your data, all.
  the. time.
- **Powerful dashboards, for continuous observability:** Build dashboards to
  collect related queries and present information that's quick and easy to
  digest for you and your team. Dashboards can be kept private or shared with
  others, and are the perfect way to bring together data from different sources.

For more information check out the
[official documentation](https://axiom.co/docs) and our
[community Discord](https://axiom.co/discord).

## Quickstart

Install using `go get`:

```shell
go get github.com/axiomhq/axiom-go/axiom
```

Import the package:

```go
import "github.com/axiomhq/axiom-go/axiom"
```

If you use the [Axiom CLI](https://github.com/axiomhq/cli), run
`eval $(axiom config export -f)` to configure your environment variables.

Otherwise create a personal token in [the Axiom settings](https://app.axiom.co/profile)
and export it as `AXIOM_TOKEN`. Set `AXIOM_ORG_ID` to the organization ID from
the settings page of the organization you want to access.

> [!NOTE]
> The organization ID is the slug below your organizations full name in [the Axiom settings](https://app.axiom.co/profile) which resembles the full name. It has a copy button next to it. Alternatively you can just extract it from your browsers address bar when browsing the Axiom App: `https://app.dev.axiomtestlabs.co/<your-org-id>/datasets`.

You can also configure the client using [options](https://pkg.go.dev/github.com/axiomhq/axiom-go/axiom#Option)
passed to the `axiom.NewClient` function:

```go
client, err := axiom.NewClient(
    axiom.SetPersonalTokenConfig("AXIOM_TOKEN", "AXIOM_ORG_ID"),
)
```

> [!NOTE]
> When only performing **ingest** or **query** operations, we recommend using
an API token with minimal privileges, only! Create an API token with the
appropriate scopes in
[the Axiom API tokens settings](https://app.axiom.co/settings/api-tokens) and
export it as `AXIOM_TOKEN`.

Create and use a client like this:

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

    client, err := axiom.NewClient()
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

## License

Distributed under the [MIT License](LICENSE).

<!-- Badges -->

[gopkg]: https://pkg.go.dev/github.com/axiomhq/axiom-go
[gopkg_badge]: https://img.shields.io/badge/doc-reference-007d9c?logo=go&logoColor=white
[workflow]: https://github.com/axiomhq/axiom-go/actions/workflows/push.yaml
[workflow_badge]: https://img.shields.io/github/actions/workflow/status/axiomhq/axiom-go/push.yaml?branch=main&ghcache=unused
[release]: https://github.com/axiomhq/axiom-go/releases/latest
[release_badge]: https://img.shields.io/github/release/axiomhq/axiom-go.svg?ghcache=unused
[license]: https://opensource.org/licenses/MIT
[license_badge]: https://img.shields.io/github/license/axiomhq/axiom-go.svg?color=blue&ghcache=unused
