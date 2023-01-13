package longlegs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
)

// Page is a structure used for serializing/deserializing data.
type Page struct {
	Id            string            `json:"id"`
	Url           *url.URL          `json:"url"`
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

func (site *Site) NewRawPageFromUrl(urlStr string) Page {
	page := Page{Id: urlStr}

	url, err := url.Parse(urlStr)
	if err != nil {
		log.Info().Msgf("Invalid URL: %s", urlStr)
		page.Error = err
		return page
	}
	baseUrl := site.GetUrl()
	url = CanonicalizeUrl(&baseUrl, url)
	page.Id = urlToId(url)
	page.Url = url

	client := &http.Client{}
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		log.Info().Msgf("Failed to Request %s", url.String())
		page.Error = err
		return page
	}

	req.Header.Set("User-Agent", site.UserAgent())

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		log.Info().Msgf("Failed to Request %s", url.String())
		page.Error = err
		return page
	}
	defer resp.Body.Close()
	page.Ms = time.Now().Sub(start).Milliseconds()

	page.StatusCode = resp.StatusCode
	if resp.StatusCode != 200 {
		log.Info().Msgf("Page status %v: %s", resp.StatusCode, url.String())
		page.Error = err
		return page
	}

	page.Headers = resp.Header

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Info().Msgf("Failed to body %s", url.String())
		page.Error = err
		return page
	}
	page.Body = body

	return page
}

func (site *Site) NewPageFromUrl(urlStr string) Page {
	page := site.NewRawPageFromUrl(urlStr)

	contentType := strings.Split(page.Headers.Get("content-type"), ";")[0]
	if contentType != "text/html" {
		log.Info().Msgf("Not HTML: %s", page.String())
		page.Error = errors.New("Not HTML")
		return page
	}

	reader := bytes.NewReader(page.Body)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Info().Msgf("Failed to Parse %s", page.Url.String())
		return Page{Error: err}
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
	url = CanonicalizeUrl(url, url)
	page.Id = urlToId(url)
	page.Url = url

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

func getKeys(hash map[string]bool) []string {
	keys := []string{}
	for k := range hash {
		if k != "" {
			keys = append(keys, k)
		}
	}
	return keys
}

func (page Page) parseLinks() Page {
	doc := page.Document

	links := doc.Find("a").Map(func(i int, s *goquery.Selection) string {
		href, exists := s.Attr("href")
		if exists {
			return href
		}
		return ""
	})

	linksUniq := map[string]bool{}
	externalUniq := map[string]bool{}

	for _, u := range links {
		rurl := ResolveURL(page.Url, u)
		rurl = CanonicalizeUrl(page.Url, rurl)

		if rurl != nil {
			if page.Url.Hostname() == rurl.Hostname() && (rurl.Scheme == "http" || rurl.Scheme == "https") {
				linksUniq[rurl.String()] = true
			} else {
				externalUniq[rurl.String()] = true
			}
		}
	}

	page.Links = getKeys(linksUniq)
	page.ExternalLinks = getKeys(externalUniq)

	return page
}
