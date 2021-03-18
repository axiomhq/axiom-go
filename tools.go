// +build tools

package axiom

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "golang.org/x/tools/cmd/stringer"
	_ "gotest.tools/gotestsum"
)
