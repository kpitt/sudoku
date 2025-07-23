package solver

func mapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func filterMap[K comparable, V any](
	m map[K]V, filter func(K, V) bool,
) map[K]V {
	filtered := make(map[K]V)
	for k, v := range m {
		if filter(k, v) {
			filtered[k] = v
		}
	}
	return filtered
}
