package longlegs

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/temoto/robotstxt"
)

type ISite interface {
	Process(*Spider, Page)
	ProcessError(*Spider, Page)
	Blocked(*Spider, url.URL)
}

type Spider struct {
	hostname       string
	url            url.URL
	userAgent      string
	history        History
	robots         Robots
	robotsData     *robotstxt.RobotsData
	crawlerGroup   *robotstxt.Group
	delayMin       time.Duration
	delayRange     time.Duration
	cookieJar      http.CookieJar
	captchaSolver  CaptchaSolver
	requestTimeout time.Duration
}

func (spider *Spider) Robots() Robots {
	return spider.robots
}

func (spider *Spider) Hostname() string {
	return spider.hostname
}

func (spider *Spider) Url() url.URL {
	return spider.url
}

func (spider *Spider) History() History {
	return spider.history
}

func (spider *Spider) UserAgent() string {
	return spider.userAgent
}

// SetCaptchaSolver registers a captcha solver that will be invoked when a
// meta-refresh redirect is detected in a page body. The solver is responsible
// for solving the captcha (e.g. a proof-of-work challenge) and submitting the
// solution, after which the cookie jar captures any cookies set by the server.
func (spider *Spider) SetCaptchaSolver(solver CaptchaSolver) {
	spider.captchaSolver = solver
}

// SetRequestTimeout sets the timeout for individual HTTP requests made by the
// spider. This covers the entire request lifecycle (DNS, connection, TLS
// handshake, headers, and body reading). The default is 30 seconds. Set to 0
// to disable the timeout (not recommended for production use).
func (spider *Spider) SetRequestTimeout(d time.Duration) {
	spider.requestTimeout = d
}

// WarmUp performs an initial request to the spider's seed URL. This triggers
// meta-refresh following and captcha solving (if a solver is registered) so
// that the cookie jar is populated before any checks or crawling begins.
func (spider *Spider) WarmUp() error {
	page := spider.NewRawPageFromUrl(spider.url)
	if page.Error != nil {
		return page.Error
	}
	return nil
}

func NewSpider(userAgent string, urlStr string, delayMin time.Duration, delayRange time.Duration) (*Spider, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	spider := &Spider{
		userAgent:      userAgent,
		delayMin:       delayMin,
		delayRange:     delayRange,
		cookieJar:      jar,
		requestTimeout: 30 * time.Second,
	}
	spider.setUrl(urlStr)

	spider.robotsData = spider.getRobots()
	spider.crawlerGroup = spider.robotsData.FindGroup(spider.userAgent)

	if !spider.CanCrawl(spider.url) {
		log.Fatal().Msg("Site blocked by robots.txt")
	}

	_, err = spider.HeadNoRedirects(spider.url)
	if err != nil {
		if err, success := errors.Unwrap(err).(*ErrRedirect); success {
			spider.setUrl(err.Location)
		} else {
			return nil, err
		}
	}
	spider.WaitCrawlDelay()

	history := make(History)
	history[spider.url] = &HistoryEntry{Crawled: false}

	spider.history = history

	log.Info().Msgf("Found Crawler Group %s with robots delay of %s, using min delay %s with range %s",
		spider.crawlerGroup.Agent,
		spider.crawlerGroup.CrawlDelay,
		spider.delayMin,
		spider.delayRange)
	spider.robots = NewRobots(spider.robotsData)

	return spider, nil
}

func (spider *Spider) setUrl(urlStr string) {
	url, err := url.Parse(urlStr)
	if err != nil {
		log.Fatal().Err(err).Msgf("Invalid Redirected URL: %s", urlStr)
	}
	spider.url = *CanonicalizeUrl(*url, nil)
	spider.hostname = strings.ToLower(url.Hostname())
}
