//go:build ignore

package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"regexp"

	"github.com/Peerapon966/blackbox/scraper/internal/crawler"
	_ "github.com/Peerapon966/blackbox/scraper/internal/crawler/engines"
	"github.com/Peerapon966/blackbox/scraper/internal/utils"
)

//go:embed internal/crawler/config/configs.json
var siteConfigs []byte

func main() {
	if len(os.Args) < 2 {
		slog.Error("missing parameter")
		fmt.Println("Usage: go run test_regex.go <param1>")
		return
	}
	url := os.Args[1]
	slog.Debug("received url", "url", url)
	config, err := getConfig()
	if err != nil {
		slog.Error("failed to get config", "error", err)
		return
	}

	slog.Debug("checking episode URL")
	var epListRe *regexp.Regexp
	if config.EpisodeListRegex != "" {
		epListRe, err = utils.CompileRegex(config.EpisodeListRegex)
		if err != nil {
			slog.Error("failed to compile episode list regex", "error", err)
			return
		}
		slog.Debug("compiled episode list regex")
	}

	slog.Debug("configuring targets")
	targets := []crawler.Target{
		{
			URL: url,
		},
	}
	if epListRe != nil && epListRe.MatchString(url) {
		slog.Info("url matches episode list regex, extracting episodes", "url", url)

		c, err := crawler.New(crawler.NewCrawlerInput{
			Config: config,
			Target: crawler.Target{
				URL: url,
			},
		})
		if err != nil {
			slog.Error("failed to create crawler for episode list", "error", err)
			return
		}

		if extractor, ok := c.(crawler.SeriesCrawler); ok {
			targets, err = extractor.ExtractEpisodes()
			if err != nil {
				slog.Error("failed to extract episodes", "error", err)
				return
			}
		} else {
			slog.Warn("crawler does not support episode list extraction", "crawler", config.CrawlerID)
			return
		}
	}
	slog.Info("targets", "targets", targets)

	slog.Debug("starting tasks", "count", len(targets))
	var tasks int
	for _, target := range targets {
		crawler, err := crawler.New(crawler.NewCrawlerInput{
			Config: config,
			Target: target,
		})
		if err != nil {
			slog.Error("failed to create crawler for target", "target", target.URL, "error", err)
			continue
		}

		if err := crawler.Initialize(); err != nil {
			slog.Error("failed to initialize crawler", "target", target.URL, "error", err)
			continue
		}

		if err := crawler.LoadPageLinks(); err != nil {
			slog.Error("failed to load page links", "target", target.URL, "error", err)
			continue
		}

		if err := crawler.LoadTitle(); err != nil {
			slog.Error("failed to load title", "target", target.URL, "error", err)
			continue
		}

		slog.Info("scrape result", "profile_url", crawler.GetURL(), "title", crawler.GetTitle(), "page_count", crawler.GetPageCount(), "image_urls", crawler.GetImageURLs())
		tasks++
	}

	slog.Info("scrape tasks completed", "completed", tasks)
}

func getConfig() (crawler.SiteConfig, error) {
	var configs map[string]crawler.SiteConfig
	err := json.Unmarshal(siteConfigs, &configs)
	if err != nil {
		return crawler.SiteConfig{}, err
	}

	u, err := url.Parse(os.Args[1])
	if err != nil {
		slog.Error("failed to parse url", "error", err)
		return crawler.SiteConfig{}, err
	}

	cid := u.Host
	slog.Info("determined crawler id", "crawler_id", cid)

	if config, exists := configs[cid]; exists {
		slog.Debug("found config for crawler id", "crawler_id", cid)
		return config, nil
	}

	slog.Error("config not found", "crawler_id", cid)
	return crawler.SiteConfig{}, errors.New("Config not found")
}
