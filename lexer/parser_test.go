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
				Interval:        "2s",
				URL:             "http://foo.bar",
				InitialWaitTime: "10m",
				MaxRetryCount:   "10",
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
			"poll \"http://foo.bar\" every \"2s\" after \"10m\" \"10\" times &",
			&lexer.PollStatement{
				lexer.BaseStatement{
					IsBackground: true,
				},
				"2s",
				"http://foo.bar",
				"10m",
				"10",
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
			"download from \"http://foo.bar\" save to \"/path/to/file\" &",
			&lexer.DownloadStatement{
				lexer.BaseStatement{
					IsBackground: true,
				},
				"http://foo.bar",
				"/path/to/file",
			},
		},
		{
			"sqlexecute from \"/path/to/file\" into \"jdbc://foo.bar?ssl=true\"",
			&lexer.SQLExecuteStatement{
				FilePath:           "/path/to/file",
				DBConnectionString: "jdbc://foo.bar?ssl=true",
			},
		},
		{
			"refresh token from \"/path/to/file\" every \"24h\"",
			&lexer.RefreshStatement{
				Artifact: "token",
				Path:     "/path/to/file",
				Interval: "24h",
			},
		},
		{
			"refresh token from \"/path/to/file\" every \"24h\" &",
			&lexer.RefreshStatement{
				lexer.BaseStatement{
					IsBackground: true,
				},
				"token",
				"/path/to/file",
				"24h",
			},
		},
		{
			"(sqlexecute from \"/path/to/file\" into \"jdbc://foo.bar?ssl=true\"; sqlexecute from \"/path/to/file\" into \"jdbc://foo.bar?ssl=true\")",
			&lexer.Query{
				Statements: []lexer.Statement{
					&lexer.SQLExecuteStatement{
						FilePath:           "/path/to/file",
						DBConnectionString: "jdbc://foo.bar?ssl=true",
					},
					&lexer.SQLExecuteStatement{
						FilePath:           "/path/to/file",
						DBConnectionString: "jdbc://foo.bar?ssl=true",
					},
				},
			},
		},
		{
			"(sqlexecute from \"/path/to/file\" into \"jdbc://foo.bar?ssl=true\"; sqlexecute from \"/path/to/file\" into \"jdbc://foo.bar?ssl=true\") &",
			&lexer.Query{
				lexer.BaseStatement{
					IsBackground: true,
				},
				[]lexer.Statement{
					&lexer.SQLExecuteStatement{
						FilePath:           "/path/to/file",
						DBConnectionString: "jdbc://foo.bar?ssl=true",
					},
					&lexer.SQLExecuteStatement{
						FilePath:           "/path/to/file",
						DBConnectionString: "jdbc://foo.bar?ssl=true",
					},
				},
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
		{"download from save ", "found save, expected URL"},
		{"download \"URL\" save ", "found URL, expected FROM"},
		{"(sqlexecute from \"/path/to/file\" into \"jdbc://foo.bar?ssl=true\" sqlexecute from \"/path/to/file\" into \"jdbc://foo.bar?ssl=true\")", "found sqlexecute, expected ;"},
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

func TestNestedParseQuery(t *testing.T) {
	s := `download from "http://foo.bar" save to "/path/to/file"; (download from "http://foo.bar" save to "/path/to/file"; download from "http://bar.foo" save to "/path/to/another/file") &`
	q, err := lexer.NewParser(strings.NewReader(s)).ParseQuery()
	if err != nil {
		t.Fatalf("expected nil . got %v", err)
	} else if len(q.Statements) != 2 {
		t.Fatalf("expected 2 statements. got %d", len(q.Statements))
	}
	stmts, err := ((lexer.Statements)(q.Statements)).Get(1)
	if err != nil {
		t.Fatalf("expected nil . got %v", err)
	}
	qstmts, ok := stmts.(*lexer.Query)
	if !ok {
		t.Fatalf("expected QUERY as second parameter . got %v", stmts)
	}
	if len(qstmts.Statements) != 2 {
		t.Fatalf("expected 2 statements. got %d", len(qstmts.Statements))
	}
	if !qstmts.IsBackground {
		t.Fatalf("expected background flag set . got false")
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
