package longlegs

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTest(t *testing.T) {
	assert.Equal(t, 1, 1, "equal")
}

func TestUrlCleanUp(t *testing.T) {
	base, _ := url.Parse("http://www.csh.com")

	assert.Equal(t,
		"http://www.csh.com/",
		ResolvedURL(base, "http://www.csh.com/?thing=true#go").String(),
		"equal")
}

func TestUrlMlDomain(t *testing.T) {
	base, _ := url.Parse("http://toflix.ml/Star-Wars/")

	assert.Equal(t,
		"http://toflix.ml/",
		ResolvedURL(base, "http://toflix.ml/?thing=true#go").String(),
		"equal")

	assert.Equal(t,
		"http://toflix.ml/Star-Wars/Star-Wars/",
		ResolvedURL(base, "Star-Wars/").String(),
		"equal")
}
