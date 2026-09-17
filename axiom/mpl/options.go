package mpl

import "maps"

// Options specifies the optional parameters for an MPL query.
type Options struct {
	// Params are the values of the parameters that the query declares.
	Params map[string]string
}

// An Option applies an optional parameter to an MPL query.
type Option func(*Options)

// SetParam sets the value of a parameter that the query declares with
// "param $name: type;". Pass the name without the "$". The value is an MPL
// literal, so a string value includes its quotes, for example `"frontend"`.
func SetParam(name, value string) Option {
	return func(o *Options) {
		if o.Params == nil {
			o.Params = make(map[string]string, 1)
		}
		o.Params[name] = value
	}
}

// SetParams sets the values of the parameters that the query declares. It
// copies the map and overwrites any existing parameters. Refer to [SetParam]
// for the format of names and values.
func SetParams(params map[string]string) Option {
	return func(o *Options) { o.Params = maps.Clone(params) }
}
