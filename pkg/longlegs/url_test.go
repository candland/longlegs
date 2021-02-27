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
		{"http://www.csh.com/", "http://www.csh.com/?thing=true#go", "http://www.csh.com/?thing=true#go"},
		{"http://toflix.ml/Star-Wars/", "http://toflix.ml/?thing=true#go", "http://toflix.ml/?thing=true#go"},
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

func TestCanonicalizeUrl(t *testing.T) {
	t.Parallel() // marks TLog as capable of running in parallel with other tests
	tests := []struct {
		base     string
		url      string
		expected string
	}{
		{"https://www.csh.com/", "http://www.csh.com/?thing=true#go", "https://www.csh.com?thing=true"},
		{"https://candland.net", "https://candland.net/assets/css/main.css", "https://candland.net/assets/css/main.css"},
		{"https://candland.net", "https://candland.net/", "https://candland.net"},
		{"https://candland.net", "https://candland.net/sub/", "https://candland.net/sub/"},
		{"https://candland.net", "https://candland.net/SUB/", "https://candland.net/SUB/"},
		{"https://candland.net", "https://candland.net/sub/?query=2&another=1", "https://candland.net/sub/?another=1&query=2"},
		{"https://candland.net", "https://candland.net/sub/page.html#frag", "https://candland.net/sub/page.html"},
		{"https://candland.net", "https://CANDLAND.net/sub/page.html#frag", "https://candland.net/sub/page.html"},
		{"https://candland.net", "https://candland.net:3000/sub/page.html#frag", "https://candland.net:3000/sub/page.html"},
		{"https://candland.net", "http://toflix.ml/?thing=true#go", "http://toflix.ml/?thing=true#go"},
		{"https://candland.net", "http://TOFLIX.ml/?thing=true#go", "http://TOFLIX.ml/?thing=true#go"},
	}
	for _, tt := range tests {
		tt := tt // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tt.url, func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			t.Log(tt.url)
			baseUrl, err := url.Parse(tt.base)
			assert.Nil(t, err)
			url, err := url.Parse(tt.url)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, CanonicalizeUrl(baseUrl, url).String())
		})
	}
}

func Test_urlToId(t *testing.T) {
	t.Parallel() // marks TLog as capable of running in parallel with other tests
	tests := []struct {
		url      string
		expected string
	}{
		{"http://www.csh.com/?thing=true#go", "www.csh.com?thing=true"},
		{"https://candland.net/assets/css/main.css", "candland.net/assets/css/main.css"},
		{"https://candland.net/", "candland.net"},
		{"https://candland.net/sub/", "candland.net/sub/"},
		{"https://candland.net/sub/?query=2&another=1", "candland.net/sub/?another=1&query=2"},
		{"https://candland.net/sub/page.html#frag", "candland.net/sub/page.html"},
		{"https://CANDLAND.net/sub/page.html#frag", "candland.net/sub/page.html"},
		{"https://candland.net:3000/sub/page.html#frag", "candland.net:3000/sub/page.html"},
	}
	for _, tt := range tests {
		tt := tt // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tt.url, func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			t.Log(tt.url)
			url, err := url.Parse(tt.url)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, urlToId(url))
		})
	}
}
