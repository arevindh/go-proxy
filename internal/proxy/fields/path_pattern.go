package fields

import (
	"regexp"

	E "github.com/yusing/go-proxy/internal/error"
)

type PathPattern string
type PathPatterns = []PathPattern

func NewPathPattern(s string) (PathPattern, E.NestedError) {
	if len(s) == 0 {
		return "", E.Invalid("path", "must not be empty")
	}
	if !pathPattern.MatchString(s) {
		return "", E.Invalid("path pattern", s)
	}
	return PathPattern(s), nil
}

func ValidatePathPatterns(s []string) (PathPatterns, E.NestedError) {
	if len(s) == 0 {
		return []PathPattern{"/"}, nil
	}
	pp := make(PathPatterns, len(s))
	for i, v := range s {
		if pattern, err := NewPathPattern(v); err.HasError() {
			return nil, err
		} else {
			pp[i] = pattern
		}
	}
	return pp, nil
}

var pathPattern = regexp.MustCompile(`^(/[-\w./]*({\$\})?|((GET|POST|DELETE|PUT|HEAD|OPTION) /[-\w./]*({\$\})?))$`)
