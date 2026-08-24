package crawler

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/Peerapon966/blackbox/scraper/internal/apperr"
	"github.com/Peerapon966/blackbox/scraper/internal/crypto"
)

type Crawler interface {
	Initialize() error
	Crawl(ctx context.Context, imgProcessor ImgProcessor) error
	GetEpisode() int
	GetImageExt() string
	GetImageURLs() []string
	GetPageCount() int
	GetTarget() Target
	GetTitle() string
	GetURL() string
	LoadTitle() error
	LoadPageLinks() error
}

type SeriesCrawler interface {
	ExtractEpisodes() ([]Target, error)
}

type CrawlerFactory func(target Target, config SiteConfig) Crawler
type ImgProcessor func(img []byte, page int) error

type NewCrawlerInput struct {
	Config SiteConfig
	Target Target
}

type SiteConfig struct {
	CrawlerID        CrawlerID        `json:"crawlerID"`
	EpisodeListRegex string           `json:"episodeListRegex"`
	PageRegex        string           `json:"pageRegex"`
	TitleRegex       string           `json:"titleRegex"`
	CoverTemplate    string           `json:"coverTemplate"`
	ImageExt         string           `json:"imageExt"`
	Routes           map[Format]Route `json:"routes"`
}

type Target struct {
	URL     string
	Title   string
	Episode int
}

type Route struct {
	Template string
	Regex    *regexp.Regexp
}

type Format string

const (
	Viewer  Format = "viewer"
	Profile Format = "profile"
)

type CrawlerID string

const (
	C1 CrawlerID = "C1"
	C2 CrawlerID = "C2"
	C3 CrawlerID = "C3"
	C4 CrawlerID = "C4"
)

var crawlerRegistry = make(map[CrawlerID]CrawlerFactory)

//go:embed config/configs.json.enc
var encSiteConfigs []byte

func New(params NewCrawlerInput) (Crawler, error) {
	constructor, exists := crawlerRegistry[params.Config.CrawlerID]
	if !exists {
		return nil, &apperr.ScraperError{
			Code:    apperr.NoSuchCrawler,
			Message: fmt.Sprintf("Couldn't get crawler %s. No such crawler exists in the factory.", params.Config.CrawlerID),
		}
	}

	crawler := constructor(params.Target, params.Config)

	return crawler, nil
}

func DownloadSiteConfigs(ctx context.Context, secrets crypto.Secrets) (map[string]SiteConfig, error) {
	decContent, err := crypto.DecryptBlob(encSiteConfigs, secrets.DEK)
	if err != nil {
		return nil, err
	}

	var siteConfigs map[string]SiteConfig
	err = json.Unmarshal(decContent, &siteConfigs)
	if err != nil {
		return nil, &apperr.ScraperError{
			Code:    apperr.UnmarshalFailed,
			Message: "Couldn't unmarshal site configs.",
			Err:     err.Error(),
		}
	}

	return siteConfigs, nil
}

func RegisterCrawler(id CrawlerID, factory CrawlerFactory) {
	crawlerRegistry[id] = factory
}
