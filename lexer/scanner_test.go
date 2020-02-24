package lexer_test

import (
	"strings"
	"testing"

	"github.com/jerminb/nestor/lexer"
)

func TestScanner(t *testing.T) {
	var tests = []struct {
		query string
		want  lexer.Token
		val   string
	}{
		{"POLL", lexer.POLL, "POLL"},
		{"EVERY", lexer.EVERY, "EVERY"},
		{"AFTER", lexer.AFTER, "AFTER"},
		{"DOWNLOAD", lexer.DOWNLOAD, "DOWNLOAD"},
		{"FROM", lexer.FROM, "FROM"},
		{"REFRESH", lexer.REFRESH, "REFRESH"},
		{"TOKEN", lexer.TOKEN, "TOKEN"},
		{"CERTIFICATE", lexer.CERTIFICATE, "CERTIFICATE"},
		{"    ", lexer.WS, "    "},
		{"\"foo\"", lexer.IDENT, "foo"},
		{"\"foo", lexer.BADSTRING, "foo"},
		{"\"12\"", lexer.IDENT, "12"},
		{"{\"foo\":\"bar\"}", lexer.IDENT, "{\"foo\":\"bar\"}"},
		{"", lexer.EOF, ""},
		{"&", lexer.AMPERSAND, "&"},
	}
	for _, c := range tests {
		scanner := lexer.NewScanner(strings.NewReader(c.query))
		tok, str := scanner.Scan()
		if tok != c.want {
			t.Errorf("expected %s . got %s", c.want.String(), tok)
		}
		if str != c.val {
			t.Errorf("expected %v . got %s", c.val, str)
		}
	}
}
