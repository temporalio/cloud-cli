package temporalcloudcli

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	defaultConfigDir string
)

func init() {
	defaultConfigDir = filepath.Join(os.Getenv("HOME"), ".config", "temporal-cloud-cli")
}

func parseURL(s string) (*url.URL, error) {
	// Without a scheme, url.Parse would interpret the path as a relative file path.
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = fmt.Sprintf("%s%s", "https://", s)
	}

	u, err := url.ParseRequestURI(s)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		u.Scheme = "https"
	}

	return u, err
}

func openBrowser(cctx *CommandContext, message string, url string) error {
	// Print to stderr so other tooling can parse the command output.
	fmt.Fprintf(os.Stderr, "%s: %s\n", message, url)

	if cctx.RootCommand.DisablePopUp {
		return nil
	}

	switch runtime.GOOS {
	case "linux":
		if err := exec.Command("xdg-open", url).Start(); err != nil {
			return err
		}
	case "windows":
		if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start(); err != nil {
			return err
		}
	case "darwin":
		if err := exec.Command("open", url).Start(); err != nil {
			return err
		}
	default:
	}
	return nil
}
