package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDuration parses a duration string (e.g., "30m", "5m", "10s", "1m30s")
// Supports both standard Go duration format and legacy integer+unit format
func ParseDuration(duration string) (time.Duration, error) {
	if d, err := time.ParseDuration(duration); err == nil {
		return d, nil
	}

	// Fallback: handle only pure integer + unit (legacy)
	switch {
	case strings.HasSuffix(duration, "s"):
		val, err := strconv.Atoi(strings.TrimSuffix(duration, "s"))
		if err != nil {
			return 0, err
		}
		return time.Duration(val) * time.Second, nil
	case strings.HasSuffix(duration, "m"):
		val, err := strconv.Atoi(strings.TrimSuffix(duration, "m"))
		if err != nil {
			return 0, err
		}
		return time.Duration(val) * time.Minute, nil
	case strings.HasSuffix(duration, "h"):
		val, err := strconv.Atoi(strings.TrimSuffix(duration, "h"))
		if err != nil {
			return 0, err
		}
		return time.Duration(val) * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid duration: %s", duration)
	}
}
