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

func checkRedirect(cfg cli.CrawlerConfig, url string) (string, error) {
	var err error
	var req *http.Request
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return "", err
	}

	for key, val := range cfg.Headers() {
		req.Header.Set(key, val)
	}

	var via []*http.Request

	client := &http.Client{}
	err = client.CheckRedirect(req, via)
	if err != nil {
		return "", err
	}
	if via == nil || len(via) == 0 {
		return url, nil
	}
	return via[len(via)-1].URL.String(), nil
}

func DownloadFile(cfg cli.CrawlerConfig, log logger.Logger, url string) error {
	var err error
	url, err = checkRedirect(cfg, url)
	if err != nil {
		return err
	}
	log.Info("URL after redirect:", url)
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
