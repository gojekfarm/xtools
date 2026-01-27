package app

// displayModule returns a display-friendly module name.
// Empty string (root module) is shown as "(root)".
func displayModule(mod string) string {
	if mod == "" {
		return "(root)"
	}
	return mod
}
