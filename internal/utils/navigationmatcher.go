package utils

func MatchNavigationPath(a, b []string) bool {

	if len(a) != len(b) {
		return false
	}
	//Check if the navigation path matches with single edit
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func GetSingleEditExecutionPath(key string) []string {
	paths := make(map[string][]string)
	paths["select_gender"] = []string{"root", "main", "my_account", "edit_profile", "select_gender"}
	paths["save_location"] = []string{"root", "main", "my_account", "edit_profile", "enter_location"}
	paths["save_yob"] = []string{"root", "main", "my_account", "edit_profile", "enter_yob"}
	return paths[key]
}
