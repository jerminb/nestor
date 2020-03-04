package testserver

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	vaultAPI "github.com/hashicorp/vault/api"
	vaultHttp "github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
)

//StatusCodeFunc is a function type to define returned status code from test handler
type StatusCodeFunc func(req *http.Request) int

var (
	//DefaultHandlerContentLength is the default content length
	DefaultHandlerContentLength = 1 << 20
	//DefaultHandlerHTTPErroResponseCode is the default response code that handler returns for failed responses
	DefaultHandlerHTTPErroResponseCode = http.StatusNotFound
)

type handler struct {
	statusCodeFunc        StatusCodeFunc
	methodWhitelist       []string
	contentLength         int
	acceptRanges          bool
	ttfb                  time.Duration
	lastModified          time.Time
	rateLimiter           *time.Ticker
	currentRequestCount   int
	maxErrorCount         int
	httpErrorResponseCode int
}

//NewHandler is the constructore function for private handler class
func NewHandler(options ...HandlerOption) (http.Handler, error) {
	h := &handler{
		statusCodeFunc:        func(req *http.Request) int { return http.StatusOK },
		methodWhitelist:       []string{"GET", "HEAD"},
		contentLength:         DefaultHandlerContentLength,
		acceptRanges:          true,
		currentRequestCount:   0,
		maxErrorCount:         -1,
		httpErrorResponseCode: DefaultHandlerHTTPErroResponseCode,
	}
	for _, option := range options {
		if err := option(h); err != nil {
			return nil, err
		}
	}
	return h, nil
}

//WithTestServer is the entry point to test handler. It sets up a test handler and execute the main thread passed as function f.
func WithTestServer(t *testing.T, f func(url string), options ...HandlerOption) {
	h, err := NewHandler(options...)
	if err != nil {
		t.Fatalf("unable to create test server handler: %v", err)
		return
	}
	s := httptest.NewServer(h)
	defer func() {
		h.(*handler).close()
		s.Close()
	}()
	f(s.URL)
}

func (h *handler) close() {
	if h.rateLimiter != nil {
		h.rateLimiter.Stop()
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// delay response
	if h.ttfb > 0 {
		time.Sleep(h.ttfb)
	}
	// validate request method
	allowed := false
	for _, m := range h.methodWhitelist {
		if r.Method == m {
			allowed = true
			break
		}
	}
	if !allowed {
		httpError(w, http.StatusMethodNotAllowed)
		return
	}
	if h.maxErrorCount > -1 {
		if h.currentRequestCount < h.maxErrorCount {
			h.currentRequestCount++
			httpError(w, h.httpErrorResponseCode)
			return
		}
	}

	// set last modified timestamp
	lastMod := time.Now()
	if !h.lastModified.IsZero() {
		lastMod = h.lastModified
	}
	w.Header().Set("Last-Modified", lastMod.Format(http.TimeFormat))

	// set content-length
	offset := 0
	if h.acceptRanges {
		if reqRange := r.Header.Get("Range"); reqRange != "" {
			if _, err := fmt.Sscanf(reqRange, "bytes=%d-", &offset); err != nil {
				httpError(w, http.StatusBadRequest)
				return
			}
			if offset >= h.contentLength {
				httpError(w, http.StatusRequestedRangeNotSatisfiable)
				return
			}
		}
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", h.contentLength-offset))

	// send header and status code
	w.WriteHeader(h.statusCodeFunc(r))

	// send body
	if r.Method == "GET" {
		// use buffered io to reduce overhead on the reader
		bw := bufio.NewWriterSize(w, 4096)
		for i := offset; !isRequestClosed(r) && i < h.contentLength; i++ {
			bw.Write([]byte{byte(i)})
			if h.rateLimiter != nil {
				<-h.rateLimiter.C
			}
		}
		if !isRequestClosed(r) {
			bw.Flush()
		}
	}
}

// isRequestClosed returns true if the client request has been canceled.
func isRequestClosed(r *http.Request) bool {
	ctx := r.Context()
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func httpError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

//WithTestVaultServer sets up a test vault server with a vault token with secrets/* path access
func WithTestVaultServer(t *testing.T, f func(url string, listner net.Listener, token string)) {
	t.Helper()

	core, keys, rootToken := vault.TestCoreUnsealed(t)

	for _, key := range keys {
		if _, err := core.Unseal(key); err != nil {
			t.Fatalf("unseal err: %s", err)
		}
	}
	sealed := core.Sealed()
	if sealed {
		t.Fatal("should not be sealed")
	}

	listner, addr := vaultHttp.TestServer(t, core)

	// Create client to vault for configuration
	cfg := vaultAPI.DefaultConfig()
	cfg.Address = addr
	c, err := vaultAPI.NewClient(cfg)
	if err != nil {
		t.Fatalf("Error creating client in mock vault setup: %v\n", err)
	}
	c.SetToken(rootToken)
	// Set policy to allow use of anything /secrets/*
	rules := `path "secret/*" {
		capabilities = ["create", "read", "update", "delete", "list"]
  }`
	err = c.Sys().PutPolicy("allsecrets", rules)
	if err != nil {
		t.Fatalf("Error applying policy: %v", err)
	}

	_, err = c.Logical().Write("secret/client-uuid/sgid/sid/bps-db/password", map[string]interface{}{
		"value": "averysecretpassword",
	})
	if err != nil {
		t.Fatalf("Error setting up secret: %v", err)
	}
	// This is just ridiculous. Renewable is a *bool and there is no easy way to pass
	// a boolean pointer. https://stackoverflow.com/questions/32364027/reference-a-boolean-for-assignment-in-a-struct
	trueBool := true
	tokenCreateOpts := &vaultAPI.TokenCreateRequest{
		Policies:  []string{"allsecrets"},
		Renewable: &trueBool,
		TTL:       "2m",
	}
	customToken, err := c.Auth().Token().Create(tokenCreateOpts)
	if err != nil {
		t.Fatalf("Error creating custom token: %v", err)
	}
	f(addr, listner, customToken.Auth.ClientToken)
}
