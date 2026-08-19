package chrome

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// DetectExecutable finds a Chrome-compatible executable without starting it.
// CHROMEPROBE_CHROME takes precedence over platform defaults.
func DetectExecutable() (string, error) {
	if override := os.Getenv("CHROMEPROBE_CHROME"); override != "" {
		if isExecutable(override) {
			return override, nil
		}
		return "", fmt.Errorf("CHROMEPROBE_CHROME is not executable: %s", override)
	}

	var candidates []string
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		candidates = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			filepath.Join(home, "Applications/Google Chrome.app/Contents/MacOS/Google Chrome"),
			"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
		}
	case "linux":
		candidates = []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser"}
	case "windows":
		candidates = []string{
			os.Getenv("PROGRAMFILES") + `\Google\Chrome\Application\chrome.exe`,
			os.Getenv("PROGRAMFILES(X86)") + `\Google\Chrome\Application\chrome.exe`,
		}
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if filepath.IsAbs(candidate) {
			if isExecutable(candidate) {
				return candidate, nil
			}
			continue
		}
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("chrome executable not found; set CHROMEPROBE_CHROME")
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Mode()&0o111 != 0
}
