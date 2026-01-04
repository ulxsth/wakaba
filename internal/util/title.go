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

// 受け取ったurl先のタイトルを取得して返す
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

	limitReader := io.LimitReader(resp.Body, 500*1024)
	utf8Reader, err := charset.NewReader(limitReader, resp.Header.Get("Content-Type"))
	if err != nil {
		utf8Reader = limitReader
	}

	bodyBytes, err := io.ReadAll(utf8Reader)
	if err != nil {
		return "", err
	}

	content := string(bodyBytes)
	content = strings.ReplaceAll(content, "\n", " ")

	match := titleRegex.FindStringSubmatch(content)
	if len(match) > 1 {
		title := strings.TrimSpace(match[1])
		return title, nil
	}

	return "", fmt.Errorf("title not found")
}
