package httpfunc

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/markoczy/crawler/cli"
)

var (
	matchIllegalPath      = regexp.MustCompile(`\?|\%|\*|\:|\||\"|\<|\>|\,|\;|\=`)
	matchIllegalPathOrSep = regexp.MustCompile(`\?|\%|\*|\:|\||\"|\<|\>|\,|\;|\=|\\|/`)
)

func DownloadFile(cfg cli.CrawlerConfig, url string) error {
	if !cfg.NamingCapture().MatchString(url) {
		return fmt.Errorf("Cannot download: Naming Capture does not match URL string '%s'", url)
	}

	filename := cfg.NamingPattern()
	match := cfg.NamingCapture().FindStringSubmatch(url)
	for i, name := range cfg.NamingCapture().SubexpNames() {
		if i != 0 && name != "" {
			repl := sanitizePath(match[i], !cfg.NamingCaptureFolders())
			filename = strings.ReplaceAll(filename, "<"+name+">", repl)
		}
	}

	return downloadFile(url, filename, cfg.Headers())
}

func downloadFile(url, filename string, headers map[string]interface{}) error {
	var err error
	var req *http.Request
	var resp *http.Response
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return err
	}

	for key, val := range headers {
		req.Header.Set(key, mapToString(val))
	}

	client := &http.Client{}
	if resp, err = client.Do(req); err != nil {
		return err
	}
	defer resp.Body.Close()

	createFolder(filename)
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func mapToString(val interface{}) string {
	switch val.(type) {
	case string:
		return val.(string)
	case []byte:
		return string(val.([]byte))
	default:
		return fmt.Sprintf("%v", val)
	}
}

func createFolder(filename string) error {
	dir := filepath.Dir(filename)
	return os.MkdirAll(dir, os.ModeDir)
}

func sanitizePath(input string, replaceSep bool) string {
	if replaceSep {
		return matchIllegalPathOrSep.ReplaceAllString(input, "_")
	}
	return matchIllegalPath.ReplaceAllString(input, "_")
}
