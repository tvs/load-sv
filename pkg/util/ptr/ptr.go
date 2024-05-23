package ptr

// To returns a pointer to a given value
func To[T any](v T) *T {
	return &v
}

// Deref dereferences ptr and returns the value it points to. If ptr is
// nil, returns the def as a default value.
func Deref[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	}
	return def
}
