![axiom-go: The official Go bindings for the Axiom API](.github/images/banner-dark.svg#gh-dark-mode-only)
![axiom-go: The official Go bindings for the Axiom API](.github/images/banner-light.svg#gh-light-mode-only)

<div align="center">

[![Go Reference][gopkg_badge]][gopkg] 
[![Workflow][workflow_badge]][workflow]
[![Latest Release][release_badge]][release]
[![License][license_badge]][license]

</div>

[Axiom](https://axiom.co) unlocks observability at any scale.

- **Ingest with ease, store without limits:** Axiom’s next-generation datastore enables ingesting petabytes of data with ultimate efficiency. Ship logs from Kubernetes, AWS, Azure, Google Cloud, DigitalOcean, Nomad, and others.
- **Query everything, all the time:** Whether DevOps, SecOps, or EverythingOps, query all your data no matter its age. No provisioning, no moving data from cold/archive to “hot”, and no worrying about slow queries. All your data, all. the. time.
- **Powerful dashboards, for continuous observability:** Build dashboards to collect related queries and present information that’s quick and easy to digest for you and your team. Dashboards can be kept private or shared with others, and are the perfect way to bring together data from different sources

For more information check out the [official documentation](https://axiom.co/docs).

## Usage

Install using `go get`:

```shell
go get github.com/axiomhq/axiom-go/axiom
```

Create a personal token in [the Axiom settings](https://cloud.axiom.co/settings/profile) and export it as `AXIOM_TOKEN`. Set `AXIOM_ORG_ID` to the organization ID from the settings page of the organization you want to access.

Create and use a client like this:

```go
import (
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/apl"
)

client, err := axiom.NewClient()
if err != nil {
	return err
}

_, err := client.Datasets.IngestEvents(ctx, "my-dataset", axiom.IngestOptions{}, axiom.Event{
	"foo": "bar",
})
if err != nil {
	return err
}

res, err := client.Datasets.APLQuery(context.Background(), apl.Query(`['my-dataset'] | where foo == 'bar' | limit 100`), apl.Options{})
if err != nil {
	return err
}
for _, match := range res.Matches {
	fmt.Println(match.Data)
}
```

For further examples, head over to the [examples](examples) directory.

## License

Distributed under the [MIT License](LICENSE).

<!-- Badges -->

[gopkg]: https://pkg.go.dev/github.com/axiomhq/axiom-go
[gopkg_badge]: https://img.shields.io/badge/doc-reference-007d9c?logo=go&logoColor=white
[workflow]: https://github.com/axiomhq/axiom-go/actions/workflows/push.yml
[workflow_badge]: https://img.shields.io/github/workflow/status/axiomhq/axiom-go/Push?ghcache=unused
[release]: https://github.com/axiomhq/axiom-go/releases/latest
[release_badge]: https://img.shields.io/github/release/axiomhq/axiom-go.svg?ghcache=unused
[license]: https://opensource.org/licenses/MIT
[license_badge]: https://img.shields.io/github/license/axiomhq/axiom-go.svg?color=blue&ghcache=unused
