package utils

// ServiceArgsFromMap processes a map of string pointers and categorizes them into update and delete lists.
// - If the value pointer is nil, it adds the argument name to the delete list.
// - If the value pointer is not nil, it adds the argument and its value to the update map.
func ServiceArgsFromMap(args map[string]*string) (map[string]string, []string) {
	updateArgs := make(map[string]string)
	deleteArgs := make([]string, 0)

	for arg, val := range args {
		if val == nil {
			deleteArgs = append(deleteArgs, arg)
		} else {
			updateArgs[arg] = *val
		}
	}
	return updateArgs, deleteArgs
}
