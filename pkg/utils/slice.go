package utils

// convert slice of strings to map[string]int
func SliceStringToMap(slice []string) (m map[string]int) {

	m = make(map[string]int)

	for i, v := range slice {
		m[v] = i
	}
	return
}
