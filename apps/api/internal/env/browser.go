package env

// loadBrowser is the path of the bulle-browser bundle the API copies into
// each VM, or empty while BULLE_BROWSER is off. The bundle needs an image
// baked with Node and Playwright, so browsing is opt-in until it is.
func loadBrowser() (string, error) {
	enabled, err := boolOrDefault("BULLE_BROWSER", false)
	if err != nil || !enabled {
		return "", err
	}
	return valueOrDefault("BULLE_BROWSER_BUNDLE", "/browser/bulle-browser.mjs"), nil
}
