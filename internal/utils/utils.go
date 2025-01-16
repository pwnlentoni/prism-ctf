package utils

import "maps"

func MapMerge[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](dst M1, src M2) M1 {
	maps.Copy(dst, src)
	return dst
}
