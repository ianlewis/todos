package testutils

// Must checks the error and panics if not nil.
func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}
