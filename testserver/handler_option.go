package testserver

import (
	"errors"
	"net/http"
	"time"
)

//HandlerOption is the default type for manipulating handler function values without having to have access to test handler
type HandlerOption func(*handler) error

//StatusCodeStatic returns a static http status code
func StatusCodeStatic(code int) HandlerOption {
	return func(h *handler) error {
		return StatusCode(func(req *http.Request) int {
			return code
		})(h)
	}
}

//StatusCode dynamically returns a status code function
func StatusCode(f StatusCodeFunc) HandlerOption {
	return func(h *handler) error {
		if f == nil {
			return errors.New("status code function cannot be nil")
		}
		h.statusCodeFunc = f
		return nil
	}
}

//MethodWhitelist to whitelist HTTP methods
func MethodWhitelist(methods ...string) HandlerOption {
	return func(h *handler) error {
		h.methodWhitelist = methods
		return nil
	}
}

//ContentLength sets content length that will be returned by handler
func ContentLength(n int) HandlerOption {
	return func(h *handler) error {
		if n < 0 {
			return errors.New("content length must be zero or greater")
		}
		h.contentLength = n
		return nil
	}
}

//TimeToFirstByte sets handler's ttfb
func TimeToFirstByte(d time.Duration) HandlerOption {
	return func(h *handler) error {
		if d < 1 {
			return errors.New("time to first byte must be greater than zero")
		}
		h.ttfb = d
		return nil
	}
}

//RateLimiter is to throttle connectivity
func RateLimiter(bps int) HandlerOption {
	return func(h *handler) error {
		if bps < 1 {
			return errors.New("bytes per second must be greater than zero")
		}
		h.rateLimiter = time.NewTicker(time.Second / time.Duration(bps))
		return nil
	}
}

//MaxErrorCount set the maximum number of error returned before a successful response
func MaxErrorCount(maxErrorCount int) HandlerOption {
	return func(h *handler) error {
		if maxErrorCount < 0 {
			return errors.New("max error count must not be negative")
		}
		h.maxErrorCount = maxErrorCount
		return nil
	}
}
