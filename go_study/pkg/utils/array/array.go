package array

import (
	"fmt"
	"slices"
)

func Contains[T comparable](slice []T, item T) bool {
	return slices.Contains(slice, item)
}

func FirstOrNil[T any](slice []T, f func(T) bool) (result T, found bool) {
	found = false
	for _, s := range slice {
		if f(s) {
			return s, true
		}
	}
	return
}

func Map[T any, R any](souce []T, f func(T) R) []R {
	result := make([]R, 0, len(souce))
	for _, v := range souce {
		result = append(result, f(v))
	}
	return result
}

func Join(slice []string, sep string) string {
	if len(slice) == 0 {
		return ""
	}

	var result string
	for _, item := range slice {
		result += fmt.Sprintf("%s%s", item, sep)
	}
	return result[:len(result)-len(sep)]
}
