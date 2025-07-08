package browser

func openBrowser(url string, runCmd func(program string, args ...string) error) error {
	return runCmd("open", url)
}
