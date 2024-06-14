package iter

import (
	"context"
	"errors"
)

// Done is returned if the iterator does not contain any more elements.
//
//nolint:revive,stylecheck // No leading "Err" as "Done" is like [io.EOF].
var Done = errors.New("no more elements in iterator")

// Element is a type that can be iterated over.
type Element any

// Iter is a function that returns the next element in the iterator. It returns
// the Done error if the iterator does not contain any more elements.
type Iter[T Element] func(context.Context) (T, error)

// Range creates an iterator that executes the given function for each index in
// the specified range.
func Range[T Element](start, end int, f func(context.Context, int) (T, error)) Iter[T] {
	var idx = start
	return func(ctx context.Context) (t T, err error) {
		if ctx.Err() != nil {
			return t, ctx.Err()
		}
		if idx > end {
			return t, Done
		}
		t, err = f(ctx, idx)
		idx++
		return
	}
}

// Slice creates an iterator that executes the given function for each element
// in the slice.
func Slice[T Element](slice []T, f func(context.Context, T) (T, error)) Iter[T] {
	var (
		idx = 0
		end = len(slice) - 1
	)
	return func(ctx context.Context) (t T, err error) {
		if ctx.Err() != nil {
			return t, ctx.Err()
		}
		if idx > end {
			return t, Done
		}
		t, err = f(ctx, slice[idx])
		idx++
		return
	}
}

// Next returns the next [Element] in the iterator.
func (itr Iter[T]) Next(ctx context.Context) (T, error) {
	return itr(ctx)
}

// Take returns up to n elements from the iterator. The iterator is only
// guaranteed to return a slice of length n if the error is [nil].
func (itr Iter[T]) Take(ctx context.Context, n int) ([]T, error) {
	res := make([]T, n)
	for i := 0; i < n; i++ {
		if ctx.Err() != nil {
			return res[:i], ctx.Err()
		}
		var err error
		if res[i], err = itr.Next(ctx); err != nil {
			return res[:i], err
		}
	}
	return res, nil
}

// Range executes the given function for each [Element] in the iterator until it
// is exhausted in which case it returns [nil] instead of [Done].
func (itr Iter[T]) Range(ctx context.Context, f func(context.Context, T) error) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		t, err := itr.Next(ctx)
		if err != nil {
			if err == Done {
				return nil
			}
			return err
		} else if err := f(ctx, t); err != nil {
			return err
		}
	}
}
