package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/Peerapon966/blackbox/scraper/internal/apperr"
	"github.com/Peerapon966/blackbox/scraper/internal/crawler"
	_ "github.com/Peerapon966/blackbox/scraper/internal/crawler/engines"
	"github.com/Peerapon966/blackbox/scraper/internal/crypto"
	"github.com/Peerapon966/blackbox/scraper/internal/s3"
	"github.com/Peerapon966/blackbox/scraper/internal/task"
	"github.com/Peerapon966/blackbox/scraper/internal/utils"
)

type Request struct {
	URL     string `json:"url"`
	Title   string `json:"title,omitempty"`
	Episode string `json:"episode"`
}

var deregInputTmpl task.DeregisterTaskInput
var response = errors.New("Scrape task failed.")

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, request Request) error {
	slog.Info("Starting scrape task(s)...")

	slog.Debug("Initializing an S3 client...")
	s3Client, err := s3.NewClient(ctx)
	if err != nil {
		slog.Error("Init task error. Failed to init S3 client.",
			slog.String("error", err.Error()),
		)
		return response
	}

	slog.Debug("Initializing a secrets object...")
	secrets, err := crypto.GetSecrets(ctx)
	if err != nil {
		slog.Error("Init task error. Failed to get secrets.",
			slog.String("error", err.Error()),
		)
		return response
	}

	slog.Debug("Decrypting request...")
	req, err := decryptRequest(request, secrets.DEK)
	if err != nil {
		slog.Error("Init task error. Failed to decrypt request.",
			slog.String("error", err.Error()),
		)
		return response
	}

	deregInputTmpl = task.DeregisterTaskInput{
		S3Client: s3Client,
		Secrets:  secrets,
		Ctx:      ctx,
		Status:   task.Failed,
	}

	slog.Debug("Validating request...")
	err = validateRequest(req)
	if err != nil {
		failedTaskHandler(crawler.Target{
			URL:     req.URL,
			Title:   req.Title,
			Episode: req.Episode,
		}, err)
		return response
	}

	slog.Debug("Downloading site configs...")
	configs, err := crawler.DownloadSiteConfigs(ctx, secrets)
	if err != nil {
		failedTaskHandler(crawler.Target{
			URL:     req.URL,
			Title:   req.Title,
			Episode: req.Episode,
		}, err)
		return response
	}

	slog.Debug("Finding a matching cralwer...")
	u, err := url.Parse(req.URL)
	if err != nil {
		return &apperr.ScraperError{
			Code:    apperr.UnsupportedURLFormat,
			Message: fmt.Sprintf("failed to parse URL: %s. Invalid URL formats.", req.URL),
			Err:     err.Error(),
		}
	}

	cid := u.Host

	var config crawler.SiteConfig
	if _, exists := configs[cid]; exists {
		config = configs[cid]
	} else {
		failedTaskHandler(crawler.Target{
			URL:     req.URL,
			Title:   req.Title,
			Episode: req.Episode,
		}, &apperr.ScraperError{
			Code:    apperr.NoMatchingCrawlerFounded,
			Message: fmt.Sprintf("Couldn't find a matching crawler for URL: %s, Host: %s, CID: %s", req.URL, u.Host, cid),
		})
		return response
	}
	slog.Debug("Crawler found", slog.String("cralwer_id", string(config.CrawlerID)))

	slog.Debug("Checking episode URL...")
	var re *regexp.Regexp
	if config.EpisodeListRegex != "" {
		re, err = utils.CompileRegex(config.EpisodeListRegex)
		if err != nil {
			failedTaskHandler(crawler.Target{
				URL:     req.URL,
				Title:   req.Title,
				Episode: req.Episode,
			}, err)
			return response
		}
	}

	slog.Debug("Configurating target(s)...")
	targets := []crawler.Target{
		{
			URL:     req.URL,
			Title:   req.Title,
			Episode: req.Episode,
		},
	}
	if re != nil && re.MatchString(req.URL) {
		c, err := crawler.New(crawler.NewCrawlerInput{
			Config: config,
			Target: crawler.Target{
				URL:   req.URL,
				Title: req.Title,
			},
		})
		if err != nil {
			failedTaskHandler(crawler.Target{
				URL:     req.URL,
				Title:   req.Title,
				Episode: req.Episode,
			}, err)
			return response
		}

		if extractor, ok := c.(crawler.SeriesCrawler); ok {
			targets, err = extractor.ExtractEpisodes()
			if err != nil {
				failedTaskHandler(crawler.Target{
					URL:     req.URL,
					Title:   req.Title,
					Episode: req.Episode,
				}, err)
				return response
			}
		} else {
			failedTaskHandler(crawler.Target{
				URL:     req.URL,
				Title:   req.Title,
				Episode: req.Episode,
			}, &apperr.ScraperError{
				Code:    apperr.MissingExtractEpisodesMethod,
				Message: fmt.Sprintf("An episode list URL was given, but the crawler (%s) doesn't support it.", config.CrawlerID),
			})
			return response
		}
	}

	slog.Debug("Starting task(s)...")
	var tasks []*task.Task
	for _, target := range targets {
		crawler, err := crawler.New(crawler.NewCrawlerInput{
			Config: config,
			Target: target,
		})
		if err != nil {
			failedTaskHandler(target, err)
			continue
		}

		task, err := task.New(ctx, task.NewTaskInput{
			Crawler:  crawler,
			S3Client: s3Client,
			Secrets:  secrets,
		})
		if err != nil {
			var scraperErr *apperr.ScraperError
			if errors.As(err, &scraperErr) && scraperErr.ErrorCode() == apperr.TaskAlreadyExists {
				slog.Warn(scraperErr.Message)
			} else {
				failedTaskHandler(crawler.GetTarget(), err)
			}
			continue
		}

		tasks = append(tasks, task)
	}

	var wg sync.WaitGroup
	for _, t := range tasks {
		wg.Add(1)
		go func(t *task.Task) {
			defer wg.Done()
			err := t.Start()
			if err != nil {
				failedTaskHandler(t.Crawler.GetTarget(), err)
			}
		}(t)
	}
	wg.Wait()

	slog.Info(fmt.Sprintf("%d Scrape task(s) are completed.", len(tasks)))
	return nil
}

func decryptRequest(req Request, secret []byte) (crawler.Target, error) {
	url, err := decodedecrypt(req.URL, secret)
	if err != nil {
		return crawler.Target{}, err
	}

	var title string
	if req.Title != "" {
		title, err = decodedecrypt(req.Title, secret)
		if err != nil {
			return crawler.Target{}, err
		}
	}

	strEp, err := decodedecrypt(req.Episode, secret)
	if err != nil {
		return crawler.Target{}, err
	}

	ep, err := strconv.Atoi(strEp)
	if err != nil {
		return crawler.Target{}, &apperr.ScraperError{
			Code:    apperr.TypeConversionFailed,
			Message: fmt.Sprintf("Couldn't convert episode value '%v' to integer; not an integer string.", strEp),
			Err:     err.Error(),
		}
	}

	return crawler.Target{
		URL:     url,
		Title:   title,
		Episode: ep,
	}, nil
}

func decodedecrypt(b64EncStr string, secret []byte) (string, error) {
	encStr, err := base64.StdEncoding.DecodeString(b64EncStr)
	if err != nil {
		return "", &apperr.ScraperError{
			Code:    apperr.DecodeFailed,
			Message: "unable to decode Base64 string.",
			Err:     err.Error(),
		}
	}

	str, err := crypto.DecryptBlob(encStr, secret)
	if err != nil {
		return "", err
	}

	return string(str), nil
}

func validateRequest(req crawler.Target) error {
	if req.Episode < 1 {
		return &apperr.ScraperError{
			Code:    apperr.InvalidRequestBody,
			Message: fmt.Sprintf("episode must be 1 or greater, received %d.", req.Episode),
		}
	}

	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" || !strings.Contains(parsedURL.Host, ".") {
		return &apperr.ScraperError{
			Code:    apperr.InvalidRequestBody,
			Message: fmt.Sprintf("url provided is malformed or invalid: %s", req.URL),
		}
	}

	return nil
}

func failedTaskHandler(target crawler.Target, err error) {
	var scraperErr *apperr.ScraperError
	if !errors.As(err, &scraperErr) {
		slog.Warn("Error is not of type ScraperError; deregistering task with the default error.")
		scraperErr = &apperr.ScraperError{
			Code:    apperr.ScrapeTaskFailed,
			Message: "Scrape task failed.",
			Err:     err.Error(),
		}
	}

	hash := sha256.Sum256(fmt.Appendf(nil, "%s_%d", target.Title, target.Episode))
	taskID := hex.EncodeToString(hash[:])

	deregInput := deregInputTmpl
	deregInput.TaskID = taskID
	deregInput.URL = target.URL
	deregInput.Title = target.Title
	deregInput.Episode = target.Episode
	deregInput.Error = scraperErr
	e := task.DeregisterTask(deregInput)
	if e != nil {
		slog.Error("Failed to deregister task.",
			slog.String("error", e.Error()),
			slog.String("task_error", scraperErr.Error()),
		)
	}
}
