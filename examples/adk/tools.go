package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"github.com/axiomhq/axiom-go/axiom"
)

func createTools(datasets *axiom.DatasetsService, root *os.Root) ([]tool.Tool, error) {
	readFile, err := functiontool.New(functiontool.Config{
		Name: "read_file", Description: "Read the contents of a file",
	}, newReadFileFn(root))
	if err != nil {
		return nil, err
	}

	writeFile, err := functiontool.New(functiontool.Config{
		Name: "write_file", Description: "Write content to a file, creating it if it doesn't exist",
	}, newWriteFileFn(root))
	if err != nil {
		return nil, err
	}

	listDir, err := functiontool.New(functiontool.Config{
		Name: "list_directory", Description: "List the contents of a directory",
	}, newListDirectoryFn(root))
	if err != nil {
		return nil, err
	}

	makeDir, err := functiontool.New(functiontool.Config{
		Name: "make_directory", Description: "Create a directory and any necessary parents",
	}, newMakeDirectoryFn(root))
	if err != nil {
		return nil, err
	}

	deleteFile, err := functiontool.New(functiontool.Config{
		Name: "delete_file", Description: "Delete a file",
	}, newDeleteFileFn(root))
	if err != nil {
		return nil, err
	}

	listDatasets, err := functiontool.New(functiontool.Config{
		Name:        "list_datasets",
		Description: "List all Axiom datasets with their name, description, and creation date",
	}, newListDatasetsFn(datasets))
	if err != nil {
		return nil, err
	}

	getDataset, err := functiontool.New(functiontool.Config{
		Name:        "get_dataset",
		Description: "Get details about a specific Axiom dataset",
	}, newGetDatasetFn(datasets))
	if err != nil {
		return nil, err
	}

	queryDataset, err := functiontool.New(functiontool.Config{
		Name:        "query_dataset",
		Description: queryDatasetDescription,
	}, newQueryDatasetFn(datasets))
	if err != nil {
		return nil, err
	}

	fetchURL, err := functiontool.New(functiontool.Config{
		Name:        "fetch_url",
		Description: "Fetch content from a URL via HTTP GET",
	}, newFetchURLFn())
	if err != nil {
		return nil, err
	}

	return []tool.Tool{
		readFile, writeFile, listDir, makeDir, deleteFile,
		listDatasets, getDataset, queryDataset,
		fetchURL,
	}, nil
}

// Filesystem tools backed by os.Root for sandboxed access.

type readFileInput struct {
	Path string `json:"path" jsonschema:"Path to the file to read"`
}

type readFileOutput struct {
	Content string `json:"content"`
}

func newReadFileFn(root *os.Root) func(tool.Context, readFileInput) (readFileOutput, error) {
	return func(_ tool.Context, in readFileInput) (readFileOutput, error) {
		data, err := root.ReadFile(in.Path)
		if err != nil {
			return readFileOutput{}, fmt.Errorf("reading file: %w", err)
		}
		return readFileOutput{Content: string(data)}, nil
	}
}

type writeFileInput struct {
	Path    string `json:"path" jsonschema:"Path to the file to write"`
	Content string `json:"content" jsonschema:"Content to write to the file"`
}

type writeFileOutput struct {
	BytesWritten int `json:"bytesWritten"`
}

func newWriteFileFn(root *os.Root) func(tool.Context, writeFileInput) (writeFileOutput, error) {
	return func(_ tool.Context, in writeFileInput) (writeFileOutput, error) {
		if err := root.WriteFile(in.Path, []byte(in.Content), 0o644); err != nil {
			return writeFileOutput{}, fmt.Errorf("writing file: %w", err)
		}
		return writeFileOutput{BytesWritten: len(in.Content)}, nil
	}
}

type listDirectoryInput struct {
	Path string `json:"path,omitempty" jsonschema:"Path to the directory to list (defaults to root)"`
}

type entryInfo struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	IsDir bool   `json:"isDir"`
}

type listDirectoryOutput struct {
	Entries []entryInfo `json:"entries"`
}

func newListDirectoryFn(root *os.Root) func(tool.Context, listDirectoryInput) (listDirectoryOutput, error) {
	return func(_ tool.Context, in listDirectoryInput) (listDirectoryOutput, error) {
		path := in.Path
		if path == "" {
			path = "."
		}
		dir, err := root.Open(path)
		if err != nil {
			return listDirectoryOutput{}, fmt.Errorf("opening directory: %w", err)
		}
		defer dir.Close()

		entries, err := dir.ReadDir(-1)
		if err != nil {
			return listDirectoryOutput{}, fmt.Errorf("reading directory: %w", err)
		}

		infos := make([]entryInfo, len(entries))
		for i, e := range entries {
			info, err := e.Info()
			if err != nil {
				return listDirectoryOutput{}, fmt.Errorf("getting file info: %w", err)
			}
			infos[i] = entryInfo{
				Name:  e.Name(),
				Size:  info.Size(),
				IsDir: e.IsDir(),
			}
		}

		return listDirectoryOutput{Entries: infos}, nil
	}
}

type makeDirectoryInput struct {
	Path string `json:"path" jsonschema:"Path of the directory to create"`
}

type makeDirectoryOutput struct{}

func newMakeDirectoryFn(root *os.Root) func(tool.Context, makeDirectoryInput) (makeDirectoryOutput, error) {
	return func(_ tool.Context, in makeDirectoryInput) (makeDirectoryOutput, error) {
		if err := root.MkdirAll(in.Path, 0o755); err != nil {
			return makeDirectoryOutput{}, fmt.Errorf("creating directory: %w", err)
		}
		return makeDirectoryOutput{}, nil
	}
}

type deleteFileInput struct {
	Path string `json:"path" jsonschema:"Path of the file to delete"`
}

type deleteFileOutput struct{}

func newDeleteFileFn(root *os.Root) func(tool.Context, deleteFileInput) (deleteFileOutput, error) {
	return func(_ tool.Context, in deleteFileInput) (deleteFileOutput, error) {
		if err := root.Remove(in.Path); err != nil {
			return deleteFileOutput{}, fmt.Errorf("deleting file: %w", err)
		}
		return deleteFileOutput{}, nil
	}
}

// Axiom tools backed by *axiom.DatasetsService.

type datasetInfo struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type listDatasetsInput struct{}

type listDatasetsOutput struct {
	Datasets []datasetInfo `json:"datasets"`
	Count    int           `json:"count"`
}

func newListDatasetsFn(datasets *axiom.DatasetsService) func(tool.Context, listDatasetsInput) (listDatasetsOutput, error) {
	return func(ctx tool.Context, _ listDatasetsInput) (listDatasetsOutput, error) {
		ds, err := datasets.List(ctx)
		if err != nil {
			return listDatasetsOutput{}, fmt.Errorf("listing datasets: %w", err)
		}

		infos := make([]datasetInfo, len(ds))
		for i, d := range ds {
			infos[i] = datasetInfo{
				Name:        d.Name,
				Description: d.Description,
				CreatedAt:   d.CreatedAt,
			}
		}

		return listDatasetsOutput{
			Datasets: infos,
			Count:    len(infos),
		}, nil
	}
}

type getDatasetInput struct {
	Name string `json:"name" jsonschema:"Name of the dataset"`
}

func newGetDatasetFn(datasets *axiom.DatasetsService) func(tool.Context, getDatasetInput) (datasetInfo, error) {
	return func(ctx tool.Context, in getDatasetInput) (datasetInfo, error) {
		ds, err := datasets.Get(ctx, in.Name)
		if err != nil {
			return datasetInfo{}, fmt.Errorf("getting dataset: %w", err)
		}

		return datasetInfo{
			Name:        ds.Name,
			Description: ds.Description,
			CreatedAt:   ds.CreatedAt,
		}, nil
	}
}

type queryDatasetInput struct {
	APL string `json:"apl" jsonschema:"APL query to execute"`
}

type queryDatasetOutput struct {
	Rows      []map[string]any `json:"rows"`
	Count     int              `json:"count"`
	Truncated bool             `json:"truncated,omitempty"`
}

const (
	maxQueryRows = 1000

	queryDatasetDescription = `Run an APL (Axiom Processing Language) query and return results.

APL Quick Reference:
  Query structure: ['dataset-name'] | operator1 | operator2
  Always filter by time first: | where _time >= ago(1h)

  Filtering:    | where field == "value"
                | where field contains "text"
                | where field >= 500
  Projection:   | project field1, field2
                | extend new_field = expression
  Aggregation:  | summarize count() by field
                | summarize avg(duration), max(duration) by bin(_time, 5m)
  Sorting:      | sort by field desc
                | top 10 by count_
  Limiting:     | take 50

  Bracket notation for dotted fields: ['service.name'], ['status.code']
  Time functions: ago(1h), ago(24h), ago(7d), now()
  Common aggregations: count(), countif(cond), dcount(field), avg(), sum(),
    min(), max(), percentile(field, N)
  String predicates: contains, has, startswith, endswith, matches regex

  For complete APL documentation, fetch https://axiom.co/docs/apl.md

  Example:
    ['my-logs'] | where _time >= ago(1h) | summarize count() by status | sort by count_ desc`
)

func newQueryDatasetFn(datasets *axiom.DatasetsService) func(tool.Context, queryDatasetInput) (queryDatasetOutput, error) {
	return func(ctx tool.Context, in queryDatasetInput) (queryDatasetOutput, error) {
		result, err := datasets.Query(ctx, in.APL)
		if err != nil {
			return queryDatasetOutput{}, fmt.Errorf("querying dataset: %w", err)
		}

		var rows []map[string]any
		truncated := false

		if len(result.Tables) > 0 {
			table := result.Tables[0]
			for row := range table.Rows() {
				if len(rows) >= maxQueryRows {
					truncated = true
					break
				}
				m := make(map[string]any, len(table.Fields))
				for i, field := range table.Fields {
					if i < len(row) {
						m[field.Name] = row[i]
					}
				}
				rows = append(rows, m)
			}
		}

		return queryDatasetOutput{
			Rows:      rows,
			Count:     len(rows),
			Truncated: truncated,
		}, nil
	}
}

// Web tool for fetching documentation or external resources.

type fetchURLInput struct {
	URL string `json:"url" jsonschema:"URL to fetch via HTTP GET"`
}

type fetchURLOutput struct {
	Content    string `json:"content"`
	StatusCode int    `json:"statusCode"`
	Truncated  bool   `json:"truncated,omitempty"`
}

const maxFetchBytes = 1024 * 64

func newFetchURLFn() func(tool.Context, fetchURLInput) (fetchURLOutput, error) {
	return func(ctx tool.Context, in fetchURLInput) (fetchURLOutput, error) {
		httpClient := &http.Client{Timeout: time.Second * 30}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, in.URL, nil)
		if err != nil {
			return fetchURLOutput{}, fmt.Errorf("creating request: %w", err)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return fetchURLOutput{}, fmt.Errorf("fetching url: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(io.LimitReader(resp.Body, maxFetchBytes+1))
		if err != nil {
			return fetchURLOutput{}, fmt.Errorf("reading response: %w", err)
		}

		truncated := len(body) > maxFetchBytes
		if truncated {
			body = body[:maxFetchBytes]
		}

		return fetchURLOutput{
			Content:    string(body),
			StatusCode: resp.StatusCode,
			Truncated:  truncated,
		}, nil
	}
}
