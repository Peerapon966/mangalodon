package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"strconv"

	"golang.org/x/sync/errgroup"

	"github.com/Peerapon966/blackbox/scraper/internal/apperr"
	"github.com/Peerapon966/blackbox/scraper/internal/crawler"
	"github.com/Peerapon966/blackbox/scraper/internal/utils"
)

type C3 struct {
	target     crawler.Target
	config     crawler.SiteConfig
	id         string
	profileURL string
	viewerURL  string
	cApiData   *cApiData
	gApiData   *gApiData
	imageURLs  []string
	imageExt   string
	pages      int
}

type cApiData struct {
	Image []string `json:"image_servers"`
	Thumb []string `json:"thumb_servers"`
}

type gApiData struct {
	Title struct {
		Pretty string `json:"pretty"`
	} `json:"title"`
	Cover struct {
		Path string `json:"path"`
	} `json:"cover"`
	Pages []struct {
		Path string `json:"path"`
	} `json:"pages"`
}

func init() {
	crawler.RegisterCrawler(crawler.C3, func(target crawler.Target, config crawler.SiteConfig) crawler.Crawler {
		return &C3{target: target, config: config}
	})
}

// Initialize parses the target URL to configure the crawler's internal state.
func (c *C3) Initialize() error {
	if c.config.ImageExt == "" {
		return &apperr.ScraperError{
			Code:    apperr.InvalidSiteConfig,
			Message: fmt.Sprintf("Missing imageExt in SiteConfig for %s crawler.", crawler.C3),
		}
	}
	c.imageExt = c.config.ImageExt

	for format, route := range c.config.Routes {
		if route.Regex.MatchString(c.target.URL) {
			matches := route.Regex.FindStringSubmatch(c.target.URL)
			if matches == nil {
				return &apperr.ScraperError{
					Code:    apperr.UnsupportedURLFormat,
					Message: fmt.Sprintf("failed to parse URL: %s. Invalid URL formats.", c.target.URL),
				}
			}
			c.id = matches[route.Regex.SubexpIndex("id")]

			matchIndices := route.Regex.FindStringSubmatchIndex(c.target.URL)
			if format == crawler.Viewer {
				c.viewerURL = c.target.URL

				var profileURL []byte
				profileURL = route.Regex.ExpandString(profileURL, c.config.Routes[crawler.Profile].Template, c.target.URL, matchIndices)
				c.profileURL = string(profileURL)
			} else {
				c.profileURL = c.target.URL

				var viewerURL []byte
				viewerURL = route.Regex.ExpandString(viewerURL, c.config.Routes[crawler.Viewer].Template, c.target.URL, matchIndices)
				c.viewerURL = string(viewerURL)
			}

			break
		}
	}

	if c.profileURL == "" || c.viewerURL == "" {
		return &apperr.ScraperError{
			Code:    apperr.UnsupportedURLFormat,
			Message: fmt.Sprintf("failed to parse URL: %s. Invalid URL formats.", c.target.URL),
		}
	}

	if c.target.Title == "" {
		return c.LoadTitle()
	}

	return nil
}

// Crawl concurrently fetches all page images and passes them to the provided processor.
func (c *C3) Crawl(ctx context.Context, imgProcessor crawler.ImgProcessor) error {
	strCLimit := os.Getenv("MAX_CONCURRENT_CRAWL")
	if strCLimit == "" {
		slog.Warn("MAX_CONCURRENT_CRAWL environment variable is not set. Using default value.",
			slog.String("provided_value", strCLimit),
			slog.Int("fallback_value", 5),
		)
	}
	cLimit, err := strconv.Atoi(strCLimit)
	if err != nil {
		slog.Warn("Invalid MAX_CONCURRENT_CRAWL env variable (expected integer). Using default value.",
			slog.String("provided_value", strCLimit),
			slog.Int("fallback_value", 5),
		)
	}

	eg, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, cLimit)

	for i, url := range c.imageURLs {
		sem <- struct{}{}

		eg.Go(func() error {
			defer func() { <-sem }()

			body, err := utils.Fetch(utils.FetchInput{
				Ctx: ctx,
				URL: url,
			})
			if err != nil {
				return err
			}

			return imgProcessor(body, i)
		})
	}

	return eg.Wait()
}

// GetEpisode returns the target episode number.
func (c *C3) GetEpisode() int {
	return c.target.Episode
}

// GetImageExt returns the file extension used for the scraped images.
func (c *C3) GetImageExt() string {
	return c.imageExt
}

// GetImageURLs returns the list of extracted image URLs.
func (c *C3) GetImageURLs() []string {
	return c.imageURLs
}

// GetPageCount returns the total number of image pages found.
func (c *C3) GetPageCount() int {
	return c.pages
}

// GetTarget returns the crawler's target configuration.
func (c *C3) GetTarget() crawler.Target {
	return c.target
}

// GetTitle returns the extracted series title.
func (c *C3) GetTitle() string {
	return c.target.Title
}

// Returns a profile URL
// GetURL returns the profile URL for the series.
func (c *C3) GetURL() string {
	return c.profileURL
}

// Sets imageURLs, pages
// LoadPageLinks fetches the viewer HTML and extracts all image URLs and the page count.
func (c *C3) LoadPageLinks() error {
	if c.cApiData == nil {
		err := c.fetchCApiData()
		if err != nil {
			return err
		}
	}

	if c.gApiData == nil {
		err := c.fetchGApiData()
		if err != nil {
			return err
		}
	}

	iServer := c.cApiData.Image[rand.IntN(len(c.cApiData.Image))]
	tServer := c.cApiData.Thumb[rand.IntN(len(c.cApiData.Thumb))]
	c.imageURLs = make([]string, 0, len(c.gApiData.Pages)+1)
	c.imageURLs = append(c.imageURLs, fmt.Sprintf("%s/%s", tServer, c.gApiData.Cover.Path))
	for _, page := range c.gApiData.Pages {
		c.imageURLs = append(c.imageURLs, fmt.Sprintf("%s/%s", iServer, page.Path))
	}
	c.pages = len(c.imageURLs) - 1
	slog.Info(fmt.Sprintf("Found %d page links.", c.pages))

	return nil
}

// LoadTitle fetches the viewer HTML and extracts the series title.
func (c *C3) LoadTitle() error {
	if c.gApiData == nil {
		err := c.fetchGApiData()
		if err != nil {
			return err
		}
	}

	c.target.Title = c.gApiData.Title.Pretty

	return nil
}

// fetchCApiData requests CDN server data from the API and caches it.
func (c *C3) fetchCApiData() error {
	data, err := utils.Fetch(utils.FetchInput{
		URL: fmt.Sprintf("%s/cdn", c.viewerURL),
	})
	if err != nil {
		return err
	}

	var cApiData cApiData
	err = json.Unmarshal(data, &cApiData)
	if err != nil {
		return &apperr.ScraperError{
			Code:    apperr.UnmarshalFailed,
			Message: "Couldn't unmarshal site configs.",
			Err:     err.Error(),
		}
	}

	c.cApiData = &cApiData
	return nil
}

// fetchGApiData requests gallery data from the API and caches it.
func (c *C3) fetchGApiData() error {
	data, err := utils.Fetch(utils.FetchInput{
		URL: fmt.Sprintf("%s/galleries/%s", c.viewerURL, c.id),
	})
	if err != nil {
		return err
	}

	var gApiData gApiData
	err = json.Unmarshal(data, &gApiData)
	if err != nil {
		return &apperr.ScraperError{
			Code:    apperr.UnmarshalFailed,
			Message: "Couldn't unmarshal site configs.",
			Err:     err.Error(),
		}
	}

	c.gApiData = &gApiData
	return nil
}
