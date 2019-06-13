package validator

/*
This bool_pointer_helpers.go file are a set of helpers to validate a pointer to
a bool value and avoid issues with invalid memory address calls to dereferenced
pointers. Instead of constantly checking if is nil then if is T/F you can call
these helpers.
*/

// isTrue checks that a value is not nil then if it is true
func isTrue(boolPtr *bool) bool {
	if isNilBool(boolPtr) {
		return false
	}

	if *boolPtr {
		return true
	}

	return false
}

func isFalse(boolPtr *bool) bool {
	if isNilBool(boolPtr) {
		return false
	}

	if *boolPtr == false {
		return true
	}

	return false
}

// isFalseOrNil will return true, if a pointer to a bool is either nil or if it is false
func isFalseOrNil(boolPtr *bool) bool {
	if isNilBool(boolPtr) {
		return true
	}

	if *boolPtr == false {
		return true
	}

	return false
}

// isNilBool is to check if the bool is nil
func isNilBool(boolPtr *bool) bool {
	if boolPtr == (*bool)(nil) {
		return true
	}

	return false
}
