package mapx

import "sort"

func Sort[T comparable, Y any](m map[T]Y, less func(i, j T) bool) []Y {
	var keys []T
	for key := range m {
		keys = append(keys, key)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return less(keys[i], keys[j])
	})

	var result []Y
	for _, key := range keys {
		result = append(result, m[key])
	}

	return result
}
