package utils

// StringInSlice 判断元素是否在数组中
func StringInSlice(a string, list []string) (int, bool) {
	if len(list) == 0 {
		return -1, false
	}
	for i, b := range list {
		if b == a {
			return i, true
		}
	}
	return -1, false
}

// SecondSliceRemoveCommonElement 第二个元素
func SecondSliceRemoveCommonElement(first, second []string) []string {
	var result []string
	secondMap := make(map[string]bool)
	for _, s := range second {
		secondMap[s] = true
	}
	for _, repeat := range first {
		if ok := secondMap[repeat]; ok {
			delete(secondMap, repeat)
		}
	}
	for key := range secondMap {
		result = append(result, key)
	}
	return result
}
