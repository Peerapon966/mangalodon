package engines

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/Peerapon966/blackbox/scraper/internal/apperr"
	"github.com/Peerapon966/blackbox/scraper/internal/crawler"
	"github.com/Peerapon966/blackbox/scraper/internal/utils"
)

type C4 struct {
	target      crawler.Target
	config      crawler.SiteConfig
	id          string
	profileURL  string
	viewerURL   string
	gg          *gg
	galleryInfo *galleryInfo
	imageURLs   []string
	imageExt    string
	pages       int
}

type gg struct {
	o int
	s [2]int
	b int
	g []int
}

type galleryInfo struct {
	title  string
	hashes []string
}

func init() {
	crawler.RegisterCrawler(crawler.C4, func(target crawler.Target, config crawler.SiteConfig) crawler.Crawler {
		return &C4{target: target, config: config}
	})
}

// Initialize parses the target URL to configure the crawler's internal state.
func (c *C4) Initialize() error {
	if c.config.ImageExt == "" {
		return &apperr.ScraperError{
			Code:    apperr.InvalidSiteConfig,
			Message: fmt.Sprintf("Missing imageExt in SiteConfig for %s crawler.", crawler.C4),
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
func (c *C4) Crawl(ctx context.Context, imgProcessor crawler.ImgProcessor) error {
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
				Headers: map[string]string{
					"referer": c.profileURL,
				},
			})
			if err != nil {
				return err
			}

			return imgProcessor(body, i)
		})

		// target has a strict rate limit, so this is required
		time.Sleep(1 * time.Second)
	}

	return eg.Wait()
}

// GetEpisode returns the target episode number.
func (c *C4) GetEpisode() int {
	return c.target.Episode
}

// GetImageExt returns the file extension used for the scraped images.
func (c *C4) GetImageExt() string {
	return c.imageExt
}

// GetImageURLs returns the list of extracted image URLs.
func (c *C4) GetImageURLs() []string {
	return c.imageURLs
}

// GetPageCount returns the total number of image pages found.
func (c *C4) GetPageCount() int {
	return c.pages
}

// GetTarget returns the crawler's target configuration.
func (c *C4) GetTarget() crawler.Target {
	return c.target
}

// GetTitle returns the extracted series title.
func (c *C4) GetTitle() string {
	return c.target.Title
}

// Returns a profile URL
// GetURL returns the profile URL for the series.
func (c *C4) GetURL() string {
	return c.profileURL
}

// Sets imageURLs, pages
// LoadPageLinks fetches the viewer HTML and extracts all image URLs and the page count.
func (c *C4) LoadPageLinks() error {
	if c.gg == nil {
		err := c.fetchGG()
		if err != nil {
			return err
		}
	}

	if c.galleryInfo == nil {
		err := c.fetchGalleryInfo()
		if err != nil {
			return err
		}
	}

	c.imageURLs = make([]string, 0, len(c.galleryInfo.hashes)+1)
	for _, hash := range c.galleryInfo.hashes {
		hexRe, err := utils.CompileRegex(`[0-9a-f]{61}([0-9a-f]{2})([0-9a-f]{1})`)
		if err != nil {
			return err
		}
		hexSubmatches := hexRe.FindStringSubmatch(hash)
		if len(hexSubmatches) < 3 {
			return &apperr.ScraperError{
				Code:    apperr.ImageExtractFailed,
				Message: "Error dissecting hash: Hash length is shorter than 64",
			}
		}

		hex := hexSubmatches[c.gg.s[0]] + hexSubmatches[c.gg.s[1]]
		g64, err := strconv.ParseInt(hex, 16, 0)
		if err != nil {
			return &apperr.ScraperError{
				Code:    apperr.TypeConversionFailed,
				Message: fmt.Sprintf("Couldn't convert 'g' value '%v' to integer; not a valid hex string.", g64),
				Err:     err.Error(),
			}
		}

		g := int(g64)
		o := utils.If(slices.Contains(c.gg.g, g),
			utils.If(c.gg.o-1 < 0, 1, c.gg.o-1),
			c.gg.o,
		)

		// add cover image url as a first item
		if len(c.imageURLs) < 1 {
			c.imageURLs = append(c.imageURLs, strings.Replace(
				fmt.Sprintf("%s/webpsmalltn/%s/%s/%s.%s",
					c.viewerURL,
					hexSubmatches[c.gg.s[0]],
					hexSubmatches[c.gg.s[1]],
					hash,
					c.imageExt),
				"<subdomain>",
				fmt.Sprintf("%stn", [2]string{"a", "b"}[rand.IntN(2)]),
				1,
			))

		}

		c.imageURLs = append(c.imageURLs, strings.Replace(
			fmt.Sprintf("%s/%d/%d/%s.%s", c.viewerURL, c.gg.b, g, hash, c.imageExt),
			"<subdomain>",
			fmt.Sprintf("w%d", 1+o),
			1,
		))
	}
	c.pages = len(c.imageURLs) - 1
	slog.Info(fmt.Sprintf("Found %d page links.", c.pages))

	return nil
}

// LoadTitle fetches the viewer HTML and extracts the series title.
func (c *C4) LoadTitle() error {
	if c.galleryInfo == nil {
		err := c.fetchGalleryInfo()
		if err != nil {
			return err
		}
	}

	c.target.Title = c.galleryInfo.title

	return nil
}

// fetchGG requests CDN server data from the API and caches it.
func (c *C4) fetchGG() error {
	data, err := utils.Fetch(utils.FetchInput{
		URL: fmt.Sprintf("%s/gg.js",
			strings.Replace(c.viewerURL, "<subdomain>", "ltn", 1),
		),
		Headers: map[string]string{
			"referer": c.profileURL,
		},
	})
	if err != nil {
		return err
	}

	var gg gg
	rawGG := string(data)
	oRe, err := utils.CompileRegex(`var o = (?P<o>\d+);`)
	if err != nil {
		return err
	}

	sRe, err := utils.CompileRegex(`m\[(?P<s1>\d{1})\]?\+m\[(?P<s2>\d{1})\]?`)
	if err != nil {
		return err
	}

	bRe, err := utils.CompileRegex(`b:\s*'(?P<b>\d+)/'`)
	if err != nil {
		return err
	}

	gRe, err := utils.CompileRegex(`case\s+(?P<g>\d+):`)
	if err != nil {
		return err
	}

	oSubmatches := oRe.FindStringSubmatch(rawGG)
	if len(oSubmatches) < 1 {
		return &apperr.ScraperError{
			Code:    apperr.InfoExtractFailed,
			Message: "Error searching required info 'o', the website HTML structure may have changed, or no match was found.",
		}
	}
	o, err := strconv.Atoi(oSubmatches[oRe.SubexpIndex("o")])
	if err != nil {
		return &apperr.ScraperError{
			Code:    apperr.TypeConversionFailed,
			Message: fmt.Sprintf("Couldn't convert 'o' value '%v' to integer; not an integer string.", o),
			Err:     err.Error(),
		}
	}
	gg.o = o

	sSubmatches := sRe.FindStringSubmatch(rawGG)
	if len(sSubmatches) < 1 {
		return &apperr.ScraperError{
			Code:    apperr.InfoExtractFailed,
			Message: "Error searching required info 's', the website HTML structure may have changed, or no match was found.",
		}
	}
	s1, err := strconv.Atoi(sSubmatches[sRe.SubexpIndex("s1")])
	if err != nil {
		return &apperr.ScraperError{
			Code:    apperr.TypeConversionFailed,
			Message: fmt.Sprintf("Couldn't convert 's1' value '%v' to integer; not an integer string.", s1),
			Err:     err.Error(),
		}
	}
	s2, err := strconv.Atoi(sSubmatches[sRe.SubexpIndex("s2")])
	if err != nil {
		return &apperr.ScraperError{
			Code:    apperr.TypeConversionFailed,
			Message: fmt.Sprintf("Couldn't convert 's2' value '%v' to integer; not an integer string.", s2),
			Err:     err.Error(),
		}
	}
	gg.s = [2]int{s1, s2}

	bSubmatches := bRe.FindStringSubmatch(rawGG)
	if len(bSubmatches) < 1 {
		return &apperr.ScraperError{
			Code:    apperr.InfoExtractFailed,
			Message: "Error searching required info 'b', the website HTML structure may have changed, or no match was found.",
		}
	}
	b, err := strconv.Atoi(bSubmatches[bRe.SubexpIndex("b")])
	if err != nil {
		return &apperr.ScraperError{
			Code:    apperr.TypeConversionFailed,
			Message: fmt.Sprintf("Couldn't convert 'b' value '%v' to integer; not an integer string.", b),
			Err:     err.Error(),
		}
	}
	gg.b = b

	gSubmatches := gRe.FindAllStringSubmatch(rawGG, -1)
	if len(gSubmatches) < 1 {
		return &apperr.ScraperError{
			Code:    apperr.InfoExtractFailed,
			Message: "Error extracting required info 'g': The website HTML structure may have changed, or no match was found.",
		}
	}

	for _, submatch := range gSubmatches {
		g, err := strconv.Atoi(submatch[gRe.SubexpIndex("g")])
		if err != nil {
			return &apperr.ScraperError{
				Code:    apperr.TypeConversionFailed,
				Message: fmt.Sprintf("Couldn't convert 'g' value '%v' to integer; not an integer string.", g),
				Err:     err.Error(),
			}
		}
		gg.g = append(gg.g, g)
	}

	c.gg = &gg

	return nil
}

// fetchGalleryInfo requests gallery data from the API and caches it.
func (c *C4) fetchGalleryInfo() error {
	data, err := utils.Fetch(utils.FetchInput{
		URL: fmt.Sprintf("%s/galleries/%s.js",
			strings.Replace(c.viewerURL, "<subdomain>", "ltn", 1),
			c.id,
		),
		Headers: map[string]string{
			"referer": c.profileURL,
		},
	})
	if err != nil {
		return err
	}

	var galleryInfo galleryInfo
	rawGalleryInfo := string(data)
	titleRe, err := utils.CompileRegex(c.config.TitleRegex)
	if err != nil {
		return err
	}

	pageRe, err := utils.CompileRegex(c.config.PageRegex)
	if err != nil {
		return err
	}

	titleSubmatches := titleRe.FindStringSubmatch(rawGalleryInfo)
	if len(titleSubmatches) < 1 {
		return &apperr.ScraperError{
			Code:    apperr.TitleExtractFailed,
			Message: "Error searching title, the website HTML structure may have changed, or no match was found.",
		}
	}
	galleryInfo.title = titleSubmatches[titleRe.SubexpIndex("title")]

	pageSubmatches := pageRe.FindAllStringSubmatch(rawGalleryInfo, -1)
	if len(pageSubmatches) < 1 {
		return &apperr.ScraperError{
			Code:    apperr.ImageExtractFailed,
			Message: "Error extracting hashes: The website HTML structure may have changed, or no match was found.",
		}
	}

	for _, submatch := range pageSubmatches {
		galleryInfo.hashes = append(galleryInfo.hashes, submatch[pageRe.SubexpIndex("hash")])
	}

	c.galleryInfo = &galleryInfo

	return nil
}
