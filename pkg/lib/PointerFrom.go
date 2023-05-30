package lib

func PointerFrom[T any](t T) *T {
	return &t
}
