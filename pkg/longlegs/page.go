package longlegs

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Page is a structure used for serializing/deserializing data.
type Page struct {
	Id            string              `json:"id"`
	Url           *url.URL            `json:"url"`
	StatusCode    int                 `json:"status_code"`
	Headers       map[string][]string `json:"headers"`
	HTML          string              `json:"-"`
	Document      *goquery.Document   `json:"-"`
	Title         string              `json:"title"`
	Description   string              `json:"description"`
	Image         string              `json:"image,omitempty"`
	Links         []string            `json:"links"`
	ExternalLinks []string            `json:"external_links"`
	Error         error               `json:"-"`
}

func (page Page) String() string {
	return fmt.Sprintf("Page: %s (%s)", page.Title, page.Url.String())
}

func NewPageFromUrl(urlStr string) Page {

	page := Page{Id: urlStr}

	url, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("Invalid URL: %s\n", urlStr)
		page.Error = err
		return page
	}
	page.Id = CanonicalizeUrl(url)
	page.Url = url

	resp, err := http.Get(url.String())
	if err != nil {
		log.Printf("Failed to Request %s\n", url.String())
		page.Error = err
		return page
	}
	defer resp.Body.Close()

	page.StatusCode = resp.StatusCode
	if resp.StatusCode != 200 {
		log.Printf("Page status %v: %s\n", resp.StatusCode, url.String())
		page.Error = err
		return page
	}

	page.Headers = resp.Header
	contentType := strings.Split(resp.Header.Get("content-type"), ";")[0]
	if contentType != "text/html" {
		log.Printf("Not HTML: %s\n", url.String())
		page.Error = errors.New("Not HTML")
		return page
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		log.Printf("Failed to Parse %s\n", url.String())
		return Page{Error: err}
	}

	page.Document = doc
	page = page.parseLinks().parseTitle().parseDescription().parseImage()

	return page
}

func NewPageFromFile(urlStr string, path string) Page {
	page := Page{}

	url, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("Invalid URL: %s\n", urlStr)
		page.Error = err
		return page
	}
	page.Id = CanonicalizeUrl(url)
	page.Url = url

	reader, err := os.Open(path)
	if err != nil {
		page.Error = err
		return page
	}
	defer reader.Close()

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Printf("Failed to Parse %s\n", url.String())
		page.Error = err
		return page
	}

	page.Document = doc
	page = page.parseLinks().parseTitle().parseDescription().parseImage()

	return page
}

func (page Page) parseTitle() Page {
	title := page.Document.Find("title").First().Text()
	page.Title = strings.TrimSpace(title)
	return page
}

func (page Page) parseDescription() Page {
	doc := page.Document

	description := ""
	metas := doc.Find("meta")
	for i := range metas.Nodes {
		s := metas.Eq(i)
		if name, _ := s.Attr("name"); strings.EqualFold(name, "description") {
			description, _ = s.Attr("content")
			break
		}
	}
	page.Description = description
	return page
}

func (page Page) parseImage() Page {
	imageURLStr := ""
	exists := false
	doc := page.Document

	if imageURLStr, exists = doc.Find("meta[property='og:image'],meta[name='twitter:image:src'],meta[name='twitter:image']").Attr("content"); !exists {
		// log.Println("No Image")
		// } else {
		// log.Printf("IMAGE STR: %v\n", imageURLStr)
	}
	page.Image = imageURLStr
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
	doc.Find("script").Remove()
	doc.Find("template").Remove()

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
		rurl := ResolvedURL(page.Url, u)
		if rurl != nil {
			if page.Url.Hostname() == rurl.Hostname() {
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
