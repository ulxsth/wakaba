package util

import (
	"regexp"
)

var urlRegex = regexp.MustCompile(`https?://[\w!?/+\-_~=;.,*&@#$%()'[\]]+`)

// 文字列から URL を抽出する
func ExtractURLs(text string) []string {
	return urlRegex.FindAllString(text, -1)
}
