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
	"github.com/markoczy/crawler/logger"
)

var (
	matchIllegalPath      = regexp.MustCompile(`\?|\%|\*|\:|\||\"|\<|\>|\,|\;|\=`)
	matchIllegalPathOrSep = regexp.MustCompile(`\?|\%|\*|\:|\||\"|\<|\>|\,|\;|\=|\\|/`)
)

func DownloadFile(url string, cfg cli.CrawlerConfig, log logger.Logger) error {
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

	if cfg.SkipExisting() && fileExists(filename) {
		log.Info("Skipping download from url '%s' as local file '%s' already exists", url, filename)
		return nil
	}
	return downloadFile(url, filename, cfg.Headers())
}

func downloadFile(url, filename string, headers map[string]string) error {
	var err error
	var req *http.Request
	var resp *http.Response
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return err
	}

	for key, val := range headers {
		req.Header.Set(key, val)
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

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}
