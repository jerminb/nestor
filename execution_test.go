package nestor_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jerminb/nestor"
	"github.com/jerminb/nestor/lexer"
	"github.com/jerminb/nestor/testserver"
)

func TestExecutionPoller(t *testing.T) {
	maxErrorCount := 3
	testserver.WithTestServer(t, func(url string) {
		s := fmt.Sprintf(`poll "%s" every "1s" after "10m" "%d" times`, url, maxErrorCount)
		stmt, err := lexer.NewParser(strings.NewReader(s)).ParseStatement()
		if err != nil {
			t.Errorf("expected nil . got %v", err)
		}
		res, err := nestor.ExecuteFromStatement(stmt)
		if err != nil {
			t.Fatalf("expected no error. got %v", err)
		}
		if res == nil {
			t.Fatalf("expected result. got nil")
		}
	},
		testserver.MaxErrorCount(maxErrorCount-1))
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
	testserver.WithTestServer(t, func(url string) {
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
