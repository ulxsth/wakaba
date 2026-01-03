package util

import (
	"reflect"
	"testing"
)

func TestExtractURLs(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{
			name: "single url",
			text: "Check this out https://google.com",
			want: []string{"https://google.com"},
		},
		{
			name: "multiple urls",
			text: "Start http://example.com and end https://test.org/path?q=1",
			want: []string{"http://example.com", "https://test.org/path?q=1"},
		},
		{
			name: "no url",
			text: "Just some text 1234",
			want: nil,
		},
		{
			name: "url with special chars",
			text: "Complex: https://site.com/foo_bar-baz?a=b&c=d#anchor",
			want: []string{"https://site.com/foo_bar-baz?a=b&c=d#anchor"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractURLs(tt.text); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractURLs() = %v, want %v", got, tt.want)
			}
		})
	}
}
