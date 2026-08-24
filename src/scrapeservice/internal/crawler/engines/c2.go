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

type C2 struct {
	target     crawler.Target
	config     crawler.SiteConfig
	id         string
	profileURL string
	viewerURL  string
	imageURLs  []string
	imageExt   string
	pages      int
}

func init() {
	crawler.RegisterCrawler(crawler.C2, func(target crawler.Target, config crawler.SiteConfig) crawler.Crawler {
		return &C2{target: target, config: config}
	})
}

// Initialize parses the target URL to configure the crawler's internal state.
func (c *C2) Initialize() error {
	if c.config.ImageExt == "" {
		return &apperr.ScraperError{
			Code:    apperr.InvalidSiteConfig,
			Message: fmt.Sprintf("Missing imageExt in SiteConfig for %s crawler.", crawler.C2),
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
func (c *C2) Crawl(ctx context.Context, imgProcessor crawler.ImgProcessor) error {
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
func (c *C2) GetEpisode() int {
	return c.target.Episode
}

// GetImageExt returns the file extension used for the scraped images.
func (c *C2) GetImageExt() string {
	return c.imageExt
}

// GetImageURLs returns the list of extracted image URLs.
func (c *C2) GetImageURLs() []string {
	return c.imageURLs
}

// GetPageCount returns the total number of image pages found.
func (c *C2) GetPageCount() int {
	return c.pages
}

// GetTarget returns the crawler's target configuration.
func (c *C2) GetTarget() crawler.Target {
	return c.target
}

// GetTitle returns the extracted series title.
func (c *C2) GetTitle() string {
	return c.target.Title
}

// GetURL returns the profile URL for the series.
func (c *C2) GetURL() string {
	return c.profileURL
}

// LoadPageLinks fetches the viewer HTML and extracts all image URLs and the page count.
func (c *C2) LoadPageLinks() error {
	html, err := c.fetchViewerHTML(1)
	if err != nil {
		return err
	}

	metadataRe, err := utils.CompileRegex(`id="pages" value="(?P<pgCount>\d+)"[\s\S]+id="image_dir" value="(?P<imgDir>\d+)"[\s\S]+id="gallery_id" value="(?P<gid>.+)"[\s\S]+id="server_id" value="(?P<sid>\d+)"`)
	if err != nil {
		return err
	}

	pageRe, err := utils.CompileRegex(c.config.PageRegex)
	if err != nil {
		return err
	}

	metadataSubmatch := metadataRe.FindStringSubmatch(html)
	if len(metadataSubmatch) < 1 {
		return &apperr.ScraperError{
			Code:    apperr.ImageExtractFailed,
			Message: "Error extracting metadata: The website HTML structure may have changed, or no match was found.",
		}
	}
	if metadataRe.SubexpIndex("pgCount") < 0 ||
		metadataRe.SubexpIndex("imgDir") < 0 ||
		metadataRe.SubexpIndex("gid") < 0 ||
		metadataRe.SubexpIndex("sid") < 0 {
		return &apperr.ScraperError{
			Code: apperr.ImageExtractFailed,
			Message: fmt.Sprintf(
				"Couldn't extract all required metadata. SubexpIndexes are; pgCount: %d, imgDir: %d, gid: %d, sid: %d",
				metadataRe.SubexpIndex("pgCount"),
				metadataRe.SubexpIndex("imgDir"),
				metadataRe.SubexpIndex("gid"),
				metadataRe.SubexpIndex("sid"),
			),
		}
	}

	pgCount := metadataSubmatch[metadataRe.SubexpIndex("pgCount")]
	imgDir := metadataSubmatch[metadataRe.SubexpIndex("imgDir")]
	gid := metadataSubmatch[metadataRe.SubexpIndex("gid")]
	sid := metadataSubmatch[metadataRe.SubexpIndex("sid")]
	c.pages, err = strconv.Atoi(pgCount)
	if err != nil {
		return &apperr.ScraperError{
			Code:    apperr.TypeConversionFailed,
			Message: fmt.Sprintf("Couldn't convert page count '%v' to integer; not an integer string.", pgCount),
			Err:     err.Error(),
		}
	}

	c.imageURLs = make([]string, 0, c.pages+1)
	c.imageURLs = append(c.imageURLs, fmt.Sprintf(c.config.CoverTemplate, sid, imgDir, gid))
	for i := range c.pages {
		html, err := c.fetchViewerHTML(i + 1)
		if err != nil {
			return err
		}

		submatch := pageRe.FindStringSubmatch(html)
		if len(submatch) < 1 {
			return &apperr.ScraperError{
				Code:    apperr.ImageExtractFailed,
				Message: "Error extracting image URLs: The website HTML structure may have changed, or no match was found.",
			}
		}

		c.imageURLs = append(c.imageURLs, submatch[1])
	}

	slog.Info(fmt.Sprintf("Found %d page links.", c.pages))

	return nil
}

// LoadTitle fetches the viewer HTML and extracts the series title.
func (c *C2) LoadTitle() error {
	html, err := c.fetchViewerHTML(1)
	if err != nil {
		return err
	}

	re, err := utils.CompileRegex(c.config.TitleRegex)
	if err != nil {
		return err
	}
	match := re.FindStringSubmatch(html)
	if len(match) < 1 {
		return &apperr.ScraperError{
			Code:    apperr.TitleExtractFailed,
			Message: "Error searching title, the website HTML structure may have changed, or no match was found.",
		}
	}

	c.target.Title = match[re.SubexpIndex("title")]

	return nil
}

// fetchViewerHTML requests the target viewer page and return HTML string.
func (c *C2) fetchViewerHTML(page int) (string, error) {
	body, err := utils.Fetch(utils.FetchInput{
		URL: fmt.Sprintf(c.viewerURL, page),
	})
	if err != nil {
		return "", err
	}

	viewerHTML := string(body)

	return viewerHTML, nil
}
