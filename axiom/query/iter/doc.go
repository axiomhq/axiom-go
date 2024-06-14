// Package iter provides a generic iterator implementation and helper functions
// to construct iterators from slices and ranges.
//
// To construct an [Iter], use the [Range] or [Slice] functions:
//
//	// Construct an iterator that returns a, b and c on successive calls.
//	slice := []string{"a", "b", "c"}
//	itr := iter.Slice(slice, func(_ context.Context, item string) (string, error) {
//	   return item, nil
//	})
//
// An [Iter] always returns a [Done] error when it is exhausted.
package iter
