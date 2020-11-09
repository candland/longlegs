package longlegs

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveURL(t *testing.T) {
	t.Parallel() // marks TLog as capable of running in parallel with other tests
	tests := []struct {
		base     string
		url      string
		expected string
	}{
		{"http://www.csh.com/", "http://www.csh.com/?thing=true#go", "http://www.csh.com/"},
		{"http://toflix.ml/Star-Wars/", "http://toflix.ml/?thing=true#go", "http://toflix.ml/"},
		{"http://toflix.ml/Star-Wars/", "Star-Wars/", "http://toflix.ml/Star-Wars/Star-Wars/"},
		{"https://candland.net", "https://maxcdn.bootstrapcdn.com/font-awesome/4.6.3/css/font-awesome.min.css", "https://maxcdn.bootstrapcdn.com/font-awesome/4.6.3/css/font-awesome.min.css"},
		{"https://candland.net", "/assets/css/main.css", "https://candland.net/assets/css/main.css"},
		{"https://candland.net", "/assets/images/apple-touch-icon.png", "https://candland.net/assets/images/apple-touch-icon.png"},
		{"https://candland.net", "http://|!|vcvUploadUrl|!|/assets/images/apple-touch-icon.png", ""},
		// {"https://candland.net", "", ""},
	}
	for _, tt := range tests {
		tt := tt // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tt.url, func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			t.Log(tt.url)
			baseUrl, err := url.Parse(tt.base)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, ResolveURL(baseUrl, tt.url).String())
		})
	}
}
