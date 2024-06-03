package array

func Contains[T comparable](slice []T, item T) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func Map[T any, R any](souce []T, f func(T) R) []R {
	result := make([]R, 0, len(souce))
	for _, v := range souce {
		result = append(result, f(v))
	}
	return result
}
