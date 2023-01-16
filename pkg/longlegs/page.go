package longlegs

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
)

// Page is a structure used for serializing/deserializing data.
type Page struct {
	Id            string            `json:"id"`
	Url           url.URL           `json:"url"`
	StatusCode    int               `json:"status_code"`
	Headers       http.Header       `json:"headers"`
	Body          []byte            `json:"-"`
	Document      *goquery.Document `json:"-"`
	Links         []string          `json:"links"`
	ExternalLinks []string          `json:"external_links"`
	Error         error             `json:"-"`
	Ms            int64             `json:"ms"`
}

func (page Page) String() string {
	return fmt.Sprintf("Page: %s (%s)", page.Id, page.Url.String())
}

func (page *Page) setPageUrl(baseUrl url.URL, url url.URL) {
	// url = CanonicalizeUrl(&baseUrl, url)
	page.Id = urlToId(url)
	page.Url = url
}

func (site *Spider) NewRawPageFromUrl(url url.URL) Page {
	page := Page{}

	page.setPageUrl(site.url, url)

	resp, ms, err := site.Get(url)
	if err != nil {
		page.Error = err
		return page
	}
	defer resp.Body.Close()
	page.Ms = ms

	if resp.Request.URL.String() != page.Url.String() {
		// Wonder if we need a history entry for this.
		log.Warn().Msgf("Page was redirected to %s from %s", resp.Request.URL, page.Url)
		page.setPageUrl(site.url, *resp.Request.URL)
	}

	page.StatusCode = resp.StatusCode
	if resp.StatusCode != 200 {
		log.Info().Msgf("Page status %d: %s", resp.StatusCode, url.String())
		page.Error = ErrNotOk
		return page
	}

	page.Headers = resp.Header

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Info().Msgf("Failed to read body %s", url.String())
		page.Error = ErrCantReadBody
		return page
	}
	page.Body = body

	return page
}

func (site *Spider) NewPageFromUrl(url url.URL) Page {
	page := site.NewRawPageFromUrl(url)

	if page.Error != nil {
		return page
	}

	contentType := strings.Split(page.Headers.Get("content-type"), ";")[0]
	if contentType != "text/html" {
		log.Info().Msgf("Not HTML: %s", page.String())
		page.Error = ErrNotHTML
		return page
	}

	reader := bytes.NewReader(page.Body)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Info().Msgf("Failed to Parse %s", page.Url.String())
		page.Error = ErrCantParseHTML
		return page
	}

	page.Document = doc
	page = page.parseLinks()

	return page
}

func NewPageFromFile(urlStr string, path string) Page {
	page := Page{}

	url, err := url.Parse(urlStr)
	if err != nil {
		log.Info().Msgf("Invalid URL: %s", urlStr)
		page.Error = err
		return page
	}
	url = CanonicalizeUrl(*url, url)
	page.Id = urlToId(*url)
	page.Url = *url

	reader, err := os.Open(path)
	if err != nil {
		page.Error = err
		return page
	}
	defer reader.Close()

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Info().Msgf("Failed to Parse %s", url.String())
		page.Error = err
		return page
	}

	page.Document = doc
	page = page.parseLinks()

	return page
}
