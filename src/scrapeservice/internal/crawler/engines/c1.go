package engines

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"golang.org/x/sync/errgroup"

	"github.com/Peerapon966/blackbox/scraper/internal/apperr"
	"github.com/Peerapon966/blackbox/scraper/internal/crawler"
	"github.com/Peerapon966/blackbox/scraper/internal/utils"
)

type C1 struct {
	target     crawler.Target
	config     crawler.SiteConfig
	id         string
	profileURL string
	viewerURL  string
	viewerHTML *string
	imageURLs  []string
	imageExt   string
	pages      int
}

func init() {
	crawler.RegisterCrawler(crawler.C1, func(target crawler.Target, config crawler.SiteConfig) crawler.Crawler {
		return &C1{target: target, config: config}
	})
}

// Initialize parses the target URL to configure the crawler's internal state.
func (c *C1) Initialize() error {
	if c.config.ImageExt == "" {
		return &apperr.ScraperError{
			Code:    apperr.InvalidSiteConfig,
			Message: fmt.Sprintf("Missing imageExt in SiteConfig for %s crawler.", crawler.C1),
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
func (c *C1) Crawl(ctx context.Context, imgProcessor crawler.ImgProcessor) error {
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
func (c *C1) GetEpisode() int {
	return c.target.Episode
}

// GetImageExt returns the file extension used for the scraped images.
func (c *C1) GetImageExt() string {
	return c.imageExt
}

// GetImageURLs returns the list of extracted image URLs.
func (c *C1) GetImageURLs() []string {
	return c.imageURLs
}

// GetPageCount returns the total number of image pages found.
func (c *C1) GetPageCount() int {
	return c.pages
}

// GetTarget returns the crawler's target configuration.
func (c *C1) GetTarget() crawler.Target {
	return c.target
}

// GetTitle returns the extracted series title.
func (c *C1) GetTitle() string {
	return c.target.Title
}

// GetURL returns the profile URL for the series.
func (c *C1) GetURL() string {
	return c.profileURL
}

// LoadPageLinks fetches the viewer HTML and extracts all image URLs and the page count.
func (c *C1) LoadPageLinks() error {
	if c.viewerHTML == nil {
		err := c.fetchViewerHTML()
		if err != nil {
			return err
		}
	}

	re, err := utils.CompileRegex(c.config.PageRegex)
	if err != nil {
		return err
	}
	submatches := re.FindAllStringSubmatch(*c.viewerHTML, -1)
	if len(submatches) < 1 {
		return &apperr.ScraperError{
			Code:    apperr.ImageExtractFailed,
			Message: "Error extracting image URLs: The website HTML structure may have changed, or no match was found.",
		}
	}

	c.imageURLs = make([]string, 0, len(submatches)+1)
	c.imageURLs = append(c.imageURLs, fmt.Sprintf(c.config.CoverTemplate, c.id))
	for _, match := range submatches {
		if len(match) > 1 {
			c.imageURLs = append(c.imageURLs, match[re.SubexpIndex("url")])
		}
	}
	c.pages = len(c.imageURLs) - 1
	slog.Info(fmt.Sprintf("Found %d page links.", c.pages))

	return nil
}

// LoadTitle fetches the viewer HTML and extracts the series title.
func (c *C1) LoadTitle() error {
	if c.viewerHTML == nil {
		err := c.fetchViewerHTML()
		if err != nil {
			return err
		}
	}

	re, err := utils.CompileRegex(c.config.TitleRegex)
	if err != nil {
		return err
	}
	match := re.FindStringSubmatch(*c.viewerHTML)
	if len(match) < 1 {
		return &apperr.ScraperError{
			Code:    apperr.TitleExtractFailed,
			Message: "Error searching title, the website HTML structure may have changed, or no match was found.",
		}
	}

	c.target.Title = match[re.SubexpIndex("title")]

	return nil
}

// fetchViewerHTML requests the target webpage and caches its HTML body.
func (c *C1) fetchViewerHTML() error {
	body, err := utils.Fetch(utils.FetchInput{
		URL: c.viewerURL,
	})
	if err != nil {
		return err
	}

	viewerHTML := string(body)
	c.viewerHTML = &viewerHTML

	return nil
}
