package lo

import (
	"cmp"
	"sync"
)

// MinLen 获取[][]T中 最短的[]T的长度
func MinLen[T any](s [][]T) int {
	mm := len(s[0])
	for _, v := range s {
		if len(v) < mm {
			mm = len(v)
		}
	}
	return mm
}

// Min 获取数组中的最小值
func Min[T cmp.Ordered](s ...T) T {
	mm := s[0]
	for _, v := range s {
		if v < mm {
			mm = v
		}
	}
	return mm
}

func MapConcurrent[T any, R any](collection []T, iteratee func(item T, index int) R) []R {
	result := make([]R, len(collection))

	var wg sync.WaitGroup
	wg.Add(len(collection))

	for i, item := range collection {
		go func(_item T, _i int) {
			res := iteratee(_item, _i)

			result[_i] = res

			wg.Done()
		}(item, i)
	}

	wg.Wait()

	return result
}

type Either[L any, R any] struct {
	l L
	r R
}

func NewEither[L any, R any](l L, r R) Either[L, R] {
	return Either[L, R]{l: l, r: r}
}

func (e Either[L, R]) Left() L {
	return e.l
}

func (e Either[L, R]) Right() R {
	return e.r
}
