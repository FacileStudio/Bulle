package settings

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Server struct {
	URL string `yaml:"url"`
}

const tokenEnv = "BULLE_TOKEN"

func NormalizeServerURL(raw string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("server url %q must be an absolute http or https address", raw)
	}
	return trimmed, nil
}

func tokenPath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "token"), nil
}

func serverPath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "server"), nil
}

func readTrimmed(path string, err error) string {
	if err != nil {
		return ""
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}

func StoredServerURL() string {
	path, err := serverPath()
	return readTrimmed(path, err)
}

func Token() string {
	if token := strings.TrimSpace(os.Getenv(tokenEnv)); token != "" {
		return token
	}
	path, err := tokenPath()
	return readTrimmed(path, err)
}

func TokenFromEnv() bool {
	return strings.TrimSpace(os.Getenv(tokenEnv)) != ""
}

func SaveLogin(serverURL, token string) error {
	base, err := NormalizeServerURL(serverURL)
	if err != nil {
		return err
	}
	tokenFile, err := tokenPath()
	if err != nil {
		return err
	}
	serverFile, err := serverPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(tokenFile), 0o700); err != nil {
		return err
	}
	if err := writePrivate(tokenFile, token); err != nil {
		return err
	}
	return writePrivate(serverFile, base)
}

func writePrivate(path, content string) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content+"\n"), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func ClearLogin() error {
	for _, locate := range []func() (string, error){tokenPath, serverPath} {
		path, err := locate()
		if err != nil {
			return err
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
