package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Peerapon966/blackbox/scraper/internal/apperr"
)

type FetchInput struct {
	Ctx     context.Context
	URL     string
	Headers map[string]string
}

func Fetch(params FetchInput) ([]byte, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var req *http.Request
	var err error
	if params.Ctx == nil {
		req, err = http.NewRequest("GET", params.URL, nil)
	} else {
		req, err = http.NewRequestWithContext(params.Ctx, "GET", params.URL, nil)
	}
	if err != nil {
		return nil, &apperr.ScraperError{
			Code:    apperr.CreateRequestFailed,
			Message: fmt.Sprintf("Error creating request for '%s'.", params.URL),
			Err:     err.Error(),
		}
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36")
	for k, v := range params.Headers {
		req.Header.Set(k, v)
	}

	resp, err := RetryWithResponse(func() (*http.Response, error) {
		return client.Do(req)
	}, RetryConfig{
		Attempts: 3,
		Delay:    5 * time.Second,
	})
	if err != nil {
		if os.IsTimeout(err) {
			return nil, &apperr.ScraperError{
				Code:    apperr.TargetTimeout,
				Message: fmt.Sprintf("Error requesting website: '%s' The request timed out.", params.URL),
				Err:     err.Error(),
			}
		} else {
			return nil, &apperr.ScraperError{
				Code:    apperr.TargetRequestFailed,
				Message: fmt.Sprintf("Error requesting website: '%s' Network or protocol error occurred.", params.URL),
				Err:     err.Error(),
			}
		}
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &apperr.ScraperError{
			Code:    apperr.ReadResponseFailed,
			Message: fmt.Sprintf("Error reading response from '%s'.", params.URL),
			Err:     err.Error(),
		}
	}
	if strings.Contains(string(body), "you have been blocked") {
		return nil, &apperr.ScraperError{
			Code:    apperr.TargetBlockedRequest,
			Message: fmt.Sprintf("Error requesting website: '%s' Request blocked by security policy.", params.URL),
		}
	}

	return body, nil
}
