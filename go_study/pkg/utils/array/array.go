package array

func Contains[T comparable](slice []T, item T) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
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
