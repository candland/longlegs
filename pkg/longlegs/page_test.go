package longlegs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadPageFromFile(t *testing.T) {
	assert.Equal(t, 1, 1, "equal")
	page := NewPageFromFile("https://candland.net", "../../test/data/candland.net.html")
	assert.Nil(t, page.Error)

	assert.Equal(t, "candland.net", page.Id)
	// assert.Equal(t, "Notes about things | candland.net", page.Title)
	// assert.Equal(t, "Notes about things Recent posts Jupyter Labs and Ruby Dusty Candland | Thu Apr 09 2020 | asdf, bundler, jupyter, ruby Wanted to setup Jupyter for Ruby to test...", page.Description)
	// assert.Equal(t, "https://candland.net/assets/images/candland-icon.png", page.Image)
	assert.Equal(t, 20, len(page.Links))
	assert.Equal(t, 21, len(page.ExternalLinks))
}

func NoTestLoadPageFromUrl(t *testing.T) {
	assert.Equal(t, 1, 1, "equal")
	page := NewPageFromUrl(&Site{}, "https://candland.net")
	assert.Nil(t, page.Error)

	assert.Equal(t, "candland.net", page.Id)
	// assert.Equal(t, "Notes about things | candland.net", page.Title)
	// assert.Equal(t, "Notes about things Recent posts Jupyter Labs and Ruby Dusty Candland | Thu Apr 09 2020 | asdf, bundler, jupyter, ruby Wanted to setup Jupyter for Ruby to test...", page.Description)
	// assert.Equal(t, "https://candland.net/assets/images/candland-icon.png", page.Image)
	assert.Equal(t, 20, len(page.Links))
	assert.Equal(t, 21, len(page.ExternalLinks))
}
