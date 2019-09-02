package nestor_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cavaliercoder/grab/grabtest"
	"github.com/jerminb/nestor"
	"github.com/jerminb/nestor/lexer"
)

func TestExecutionPoller(t *testing.T) {
	counter := 0
	maxErrorCount := 3
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := map[string]interface{}{
			"bar": "bar",
			"foo": "foo",
		}

		if counter < (maxErrorCount - 1) {
			counter++
			w.WriteHeader(http.StatusNotFound)
			return
		}

		b, err := json.Marshal(d)
		if err != nil {
			t.Error(err)
		}
		io.WriteString(w, string(b))
		w.WriteHeader(http.StatusOK)
	})
	backend := httptest.NewServer(http.HandlerFunc(handlerFunc))
	s := fmt.Sprintf(`poll "%s" every "2s" after "2s" "%d" times`, backend.URL, maxErrorCount)
	stmt, err := lexer.NewParser(strings.NewReader(s)).ParseStatement()
	if err != nil {
		t.Errorf("expected nil . got %v", err)
	}
	res, err := nestor.ExecuteFromStatement(stmt)
	if err != nil {
		t.Errorf("expected nil . got %v", err)
	}
	if res == nil {
		t.Fatalf("expected result. got nil")
	}
}

func TestExecutionPollerNegative(t *testing.T) {
	s := `poll "http://foo.bar" every "2" after "10m" "10" times`
	stmt, err := lexer.NewParser(strings.NewReader(s)).ParseStatement()
	if err != nil {
		t.Errorf("expected nil . got %v", err)
	}
	_, err = nestor.ExecuteFromStatement(stmt)
	if err == nil {
		t.Errorf("expected error . got nil")
	}
	s = `poll "http://foo.bar" every "2s" after "10m" "foo" times`
	stmt, err = lexer.NewParser(strings.NewReader(s)).ParseStatement()
	if err != nil {
		t.Errorf("expected nil . got %v", err)
	}
	_, err = nestor.ExecuteFromStatement(stmt)
	if err == nil {
		t.Errorf("expected error . got nil")
	}
}

func TestExecutionDownloader(t *testing.T) {
	now := time.Now()
	nanos := now.UnixNano()
	filename := fmt.Sprintf("/tmp/nestor_tests/%d", nanos)
	defer os.Remove(filename)
	grabtest.WithTestServer(t, func(url string) {
		s := fmt.Sprintf(`download from "%s" save to "%s"`, url, filename)
		stmt, err := lexer.NewParser(strings.NewReader(s)).ParseStatement()
		if err != nil {
			t.Fatalf("expected nil . got %v", err)
		}
		res, err := nestor.ExecuteFromStatement(stmt)
		if err != nil {
			t.Fatalf("expected no error. got %v", err)
		}
		if res == nil {
			t.Fatalf("expected result. got nil")
		}
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Fatalf("expected file in %s. got nil", filename)
		}
	})
}
