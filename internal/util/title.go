package util

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

var titleRegex = regexp.MustCompile(`(?i)<title>(.*?)</title>`)

// FetchPageTitle fetches the title of the web page at the given URL.
// It returns the title or an error if fetch fails or title is missing.
func FetchPageTitle(url string) (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code: %d", resp.StatusCode)
	}

	// Limit reader to avoid downloading huge files
	limitReader := io.LimitReader(resp.Body, 500*1024) // 500KB

	// Create a reader that converts to UTF-8
	utf8Reader, err := charset.NewReader(limitReader, resp.Header.Get("Content-Type"))
	if err != nil {
		// Fallback to raw reader if detection fails
		utf8Reader = limitReader
	}

	bodyBytes, err := io.ReadAll(utf8Reader)
	if err != nil {
		return "", err
	}

	content := string(bodyBytes)
	// Replace newlines to handle multi-line titles in regex
	content = strings.ReplaceAll(content, "\n", " ")

	match := titleRegex.FindStringSubmatch(content)
	if len(match) > 1 {
		title := strings.TrimSpace(match[1])
		// Basic HTML entity unescape could be added here if needed,
		// but for now returning as is or basic clean.
		return title, nil
	}

	return "", fmt.Errorf("title not found")
}
