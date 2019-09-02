package lexer_test

import (
	"strings"
	"testing"

	"github.com/jerminb/nestor/lexer"
)

func TestParsePositive(t *testing.T) {
	var tests = []struct {
		query string
		stmt  lexer.Statement
	}{
		{
			"poll \"http://foo.bar\" every \"2s\" after \"10m\" \"10\" times",
			&lexer.PollStatement{
				Interval:       "2s",
				URL:            "http://foo.bar",
				InitalWaitTime: "10m",
				MaxRetryCount:  "10",
			},
		},
		{
			"poll \"http://foo.bar\" every \"2s\" \"10\" times",
			&lexer.PollStatement{
				Interval:      "2s",
				URL:           "http://foo.bar",
				MaxRetryCount: "10",
			},
		},
		{
			"download from \"http://foo.bar\" save to \"/path/to/file\"",
			&lexer.DownloadStatement{
				URL:      "http://foo.bar",
				FilePath: "/path/to/file",
			},
		},
		{
			"sqlexecute from \"/path/to/file\" into \"jdbc://foo.bar?ssl=true\"",
			&lexer.SQLExecuteStatement{
				FilePath:           "/path/to/file",
				DBConnectionString: "jdbc://foo.bar?ssl=true",
			},
		},
	}
	for _, c := range tests {
		parser := lexer.NewParser(strings.NewReader(c.query))
		stmt, err := parser.ParseStatement()
		if err != nil {
			t.Fatalf("expected nil . got %v", err)
		}
		if stmt.String() != c.stmt.String() {
			t.Fatalf("expected %s . got %s", stmt.String(), c.stmt.String())
		}
	}
}

func TestParseNegative(t *testing.T) {
	var tests = []struct {
		query string
		err   string
	}{
		{"poll every \"2 seconds\" after \"10 minutes\" \"10\" times", "found every, expected URL"},
		{"poll \"URL\" every \"2 seconds\" after \"10 minutes\" ", "found EOF, expected MaxRetryCount"},
	}
	for _, c := range tests {
		parser := lexer.NewParser(strings.NewReader(c.query))
		_, err := parser.ParseStatement()
		if err == nil {
			t.Fatalf("expected error .got nil")
		}
		if c.err != err.Error() {
			t.Fatalf("expected %s . got %s", c.err, err.Error())
		}
	}
}

func TestParseQuery(t *testing.T) {
	s := `download from "http://foo.bar" save to "/path/to/file"; download from "http://bar.foo" save to "/path/to/another/file"`
	q, err := lexer.NewParser(strings.NewReader(s)).ParseQuery()
	if err != nil {
		t.Fatalf("expected nil . got %v", err)
	} else if len(q.Statements) != 2 {
		t.Fatalf("expected 2 statements. got %d", len(q.Statements))
	}
}

func TestTrailingSemicolon(t *testing.T) {
	s := `download from "http://foo.bar" save to "/path/to/file";`
	q, err := lexer.NewParser(strings.NewReader(s)).ParseQuery()
	if err != nil {
		t.Fatalf("expected nil . got %v", err)
	} else if len(q.Statements) != 1 {
		t.Fatalf("expected 1 statement. got %d", len(q.Statements))
	}
}

// Ensure the parser can parse an empty query.
func TestParserEmpty(t *testing.T) {
	q, err := lexer.NewParser(strings.NewReader(``)).ParseQuery()
	if err != nil {
		t.Fatalf("expected nil . got %v", err)
	} else if len(q.Statements) != 0 {
		t.Fatalf("expected 0 statements. got %d", len(q.Statements))
	}
}
