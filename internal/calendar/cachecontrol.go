package calendar

import (
	"strconv"
	"strings"
	"time"
)

func parseCacheDuration(cacheControl string) time.Duration {
	defaultAge := 2 * time.Hour

	for part := range strings.SplitSeq(cacheControl, ",") {
		part = strings.ToLower(strings.TrimSpace(part))

		if val, ok := strings.CutPrefix(part, maxAge); ok {
			seconds, err := strconv.Atoi(val)
			if err == nil {
				return time.Duration(seconds) * time.Second
			}
		}
	}

	return defaultAge
}
