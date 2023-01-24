package iter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom/query/iter"
)

func TestRange(t *testing.T) {
	itr := iter.Range(1, 2, func(_ context.Context, idx int) (int, error) {
		return idx, nil
	})

	ctx := context.Background()

	res, err := itr.Next(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, res)

	res, err = itr.Next(ctx)
	require.NoError(t, err)

	assert.Equal(t, 2, res)

	res, err = itr.Next(ctx)
	require.Error(t, err)

	assert.Equal(t, iter.Done, err)
	assert.Zero(t, res)
}

func TestSlice(t *testing.T) {
	slice := []int{1, 2}
	itr := iter.Slice(slice, func(_ context.Context, item int) (int, error) {
		return item, nil
	})

	ctx := context.Background()

	res, err := itr.Next(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, res)

	res, err = itr.Next(ctx)
	require.NoError(t, err)

	assert.Equal(t, 2, res)

	res, err = itr.Next(ctx)
	require.Error(t, err)

	assert.Equal(t, iter.Done, err)
	assert.Zero(t, res)
}

func TestIter_Next(t *testing.T) {
	itr := iter.Iter[int](func(context.Context) (int, error) {
		return 1, nil
	})

	ctx := context.Background()

	res1, _ := itr(ctx)
	res2, _ := itr.Next(ctx)

	assert.Equal(t, res1, res2)
}

func TestIter_Take(t *testing.T) {
	itr := iter.Iter[int](func(context.Context) (int, error) {
		return 1, nil
	})

	ctx := context.Background()

	res, _ := itr.Take(ctx, 3)
	if assert.Len(t, res, 3) {
		assert.Equal(t, []int{1, 1, 1}, res)
	}
}

func TestIter_Take_Error(t *testing.T) {
	var count int
	itr := iter.Iter[int](func(context.Context) (int, error) {
		if count > 1 {
			return 0, errors.New("an error")
		}
		count++
		return 1, nil
	})

	ctx := context.Background()

	res, err := itr.Take(ctx, 3)

	if assert.Error(t, err) {
		assert.EqualError(t, err, "an error")
	}
	if assert.Len(t, res, 2) {
		assert.Equal(t, []int{1, 1}, res)
	}
}

func TestIter_Range(t *testing.T) {
	itr := iter.Range(1, 5, func(_ context.Context, idx int) (int, error) {
		return idx, nil
	})

	var res int
	err := itr.Range(context.Background(), func(ctx context.Context, i int) error {
		res += i
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, 15, res)
}
