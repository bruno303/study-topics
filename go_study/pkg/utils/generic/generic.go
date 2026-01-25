package generic

func CastValueIfNoError[T any](value any, err error) (T, error) {
	var empty T
	if err != nil {
		return empty, err
	}

	return value.(T), nil
}
