package lo

import (
	"cmp"
	"golang.org/x/exp/rand"
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

// Delete 从列表中删除元素
func Delete[T comparable](s []T, i T) []T {
	for idx, v := range s {
		if v == i {
			return append(s[:idx], s[idx+1:]...)
		}
	}
	return s
}

func DeleteFunc[T any](s []T, f func(T) bool) []T {
	var result []T
	for _, v := range s {
		if !f(v) {
			result = append(result, v)
		}
	}
	return result
}

func Map[T any, R any](s []T, f func(T) R) []R {
	result := make([]R, len(s))
	for i, v := range s {
		result[i] = f(v)
	}
	return result
}

// RandInt 返回 [min, max) 范围内的随机整数
func RandInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func RandOne[T any](slice []T) T {
	return slice[RandInt(0, len(slice))]
}

func IfThen[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}
