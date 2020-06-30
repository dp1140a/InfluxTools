package util

// posString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func PosString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// containsString returns true iff slice contains element
func ContainsString(slice []string, element string) bool {
	return !(PosString(slice, element) == -1)
}
