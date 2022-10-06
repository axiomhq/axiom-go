package version

import "runtime/debug"

var version string

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/axiomhq/axiom-go" {
				version = dep.Version
				break
			}
		}
	}
}

// Get returns the Go module version of the axiom-go module.
func Get() string {
	return version
}
