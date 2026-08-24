package apperr

import (
	"fmt"
)

type ScraperError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Err     string    `json:"error,omitempty"`
}

func (e *ScraperError) ErrorCode() ErrorCode { return e.Code }

func (e *ScraperError) ErrorMessage() string { return e.Message }

func (e *ScraperError) Error() string {
	if e.Err != "" {
		return fmt.Sprintf("scraper error %s: %s (caused by: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("scraper error %s: %s", e.Code, e.Message)
}

// Allow Go's standard errors.Is() and errors.As()
// to be able to see through this struct to the original error
func (e *ScraperError) Unwrap() error {
	return e
}
