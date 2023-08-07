# Adapters

Adapters integrate Axiom Go into well known Go logging libraries.

💡 _Go **1.21** will feature the `slog` package, a structured logging package in
the Go standard library. You can try it out now by importing
`golang.org/x/exp/slog` and we already provide [an adapter](slog)._

We currently support a bunch of adapters right out of the box.

## Standard Library

* [Slog](https://pkg.go.dev/golang.org/x/exp/slog): `import "github.com/axiomhq/axiom-go/adapters/slog"`

## Third Party Packages

* [Apex](https://github.com/apex/log): `import "github.com/axiomhq/axiom-go/adapters/apex"`
* [Logrus](https://github.com/sirupsen/logrus): `import "github.com/axiomhq/axiom-go/adapters/logrus"`
* [Zap](https://github.com/uber-go/zap): `import "github.com/axiomhq/axiom-go/adapters/zap"`
