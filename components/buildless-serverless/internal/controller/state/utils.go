package state

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		if v, ok := b[key]; !ok || v != value {
			return false
		}
	}
	return true
}
