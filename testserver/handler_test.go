package testserver_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/jerminb/nestor/testserver"
)

func TestHandlerDefaults(t *testing.T) {
	testserver.WithTestServer(t, func(url string) {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			t.Fatalf("expected nil. got %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("expected nil. got %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected %d. got %d", http.StatusOK, resp.StatusCode)
		}
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("expected nil. got %v", err)
		}
	})
}

func TestHandlerMethodWhitelist(t *testing.T) {
	tests := []struct {
		Whitelist        []string
		Method           string
		ExpectStatusCode int
	}{
		{[]string{"GET", "HEAD"}, "GET", http.StatusOK},
		{[]string{"GET", "HEAD"}, "HEAD", http.StatusOK},
		{[]string{"GET"}, "HEAD", http.StatusMethodNotAllowed},
		{[]string{"HEAD"}, "GET", http.StatusMethodNotAllowed},
	}

	for _, test := range tests {
		testserver.WithTestServer(t, func(url string) {
			req, err := http.NewRequest(test.Method, url, nil)
			if err != nil {
				t.Fatalf("expected nil. got %v", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("expected nil. got %v", err)
			}
			if resp.StatusCode != test.ExpectStatusCode {
				t.Fatalf("expected %d. got %d", test.ExpectStatusCode, resp.StatusCode)
			}
			if err := resp.Body.Close(); err != nil {
				t.Fatalf("expected nil. got %v", err)
			}
		}, testserver.MethodWhitelist(test.Whitelist...))
	}
}

func TestHandlerContentLength(t *testing.T) {
	tests := []struct {
		Method          string
		ContentLength   int
		ExpectHeaderLen int64
		ExpectBodyLen   int
	}{
		{"GET", 321, 321, 321},
		{"HEAD", 321, 321, 0},
		{"GET", 0, 0, 0},
		{"HEAD", 0, 0, 0},
	}

	for _, test := range tests {
		testserver.WithTestServer(t, func(url string) {
			req, err := http.NewRequest(test.Method, url, nil)
			if err != nil {
				t.Fatalf("expected nil. got %v", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("expected nil. got %v", err)
			}

			expected := fmt.Sprintf("%d", test.ExpectHeaderLen)
			actual := resp.Header.Get("Content-Length")
			if expected != actual {
				t.Errorf("expected %s. got %s", expected, actual)
			}
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("expected nil. got %v", err)
			}
			if len(b) != test.ExpectBodyLen {
				t.Errorf(
					"expected body length: %v, got: %v, in: %v",
					test.ExpectBodyLen,
					len(b),
					test,
				)
			}
		},
			testserver.ContentLength(test.ContentLength),
		)
	}
}
