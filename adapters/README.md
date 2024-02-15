# Adapters

Adapters integrate Axiom Go into well known Go logging libraries.

We currently support a bunch of adapters right out of the box.

## Standard Library

* [Slog](https://pkg.go.dev/log/slog): `import adapter "github.com/axiomhq/axiom-go/adapters/slog"`

## Third Party Packages

* [Apex](https://github.com/apex/log): `import adapter "github.com/axiomhq/axiom-go/adapters/apex"`
* [Logrus](https://github.com/sirupsen/logrus): `import adapter "github.com/axiomhq/axiom-go/adapters/logrus"`
* [Zap](https://github.com/uber-go/zap): `import adapter "github.com/axiomhq/axiom-go/adapters/zap"`
