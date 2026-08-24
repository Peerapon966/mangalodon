package utils

import (
	"errors"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/Peerapon966/blackbox/scraper/internal/apperr"
)

type RetryConfig struct {
	Attempts    int
	Delay       time.Duration
	FatalErrors []apperr.ErrorCode
}

// For function with response
func RetryWithResponse[T any](fn func() (T, error), config RetryConfig) (T, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	var zero T

	if config.Attempts <= 0 {
		logger.Warn("Invalid max retry attempts. Using default value.",
			slog.Int("provided_value", config.Attempts),
			slog.Int("fallback_value", 5),
		)
		config.Attempts = 5
	}
	if config.Delay <= 0 {
		logger.Warn("Invalid delay value. Using default value.",
			slog.Any("provided_value", config.Delay),
			slog.Duration("fallback_value", 500*time.Millisecond),
		)
		config.Delay = 500 * time.Millisecond
	}

	var result T
	var err error
	for i := 0; i < config.Attempts; i++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}
		var scraperErr *apperr.ScraperError
		if errors.As(err, &scraperErr) && slices.Contains(config.FatalErrors, scraperErr.ErrorCode()) {
			return zero, err
		}

		time.Sleep(config.Delay)
	}

	return zero, err
}

// For function with no response
func RetryWithNoResponse(fn func() error, config RetryConfig) error {
	// Reuses the generic Retry function by passing a dummy bool
	_, err := RetryWithResponse(func() (bool, error) {
		return false, fn()
	}, config)
	return err
}
