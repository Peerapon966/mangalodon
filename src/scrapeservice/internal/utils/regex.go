package utils

import (
	"fmt"
	"regexp"

	"github.com/Peerapon966/blackbox/scraper/internal/apperr"
)

func CompileRegex(regStr string) (re *regexp.Regexp, err error) {
	defer func() {
		if r := recover(); r != nil {
			var panicErr error
			if e, ok := r.(error); ok {
				panicErr = e
			} else {
				panicErr = fmt.Errorf("%v", r)
			}

			err = &apperr.ScraperError{
				Code:    apperr.InvalidRegexPattern,
				Message: "Couldn't compile regex; regex pattern is invalid.",
				Err:     panicErr.Error(),
			}
		}
	}()

	re = regexp.MustCompile(regStr)
	return re, nil
}
