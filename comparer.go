package mapper

func compareStructFields(src map[string]StructField, tgt map[string]StructField) bool {
	if len(src) != len(tgt) {
		return false
	}
	for key := range src {
		if _, exist := tgt[key]; !exist {
			return false
		}
	}
	return true
}
