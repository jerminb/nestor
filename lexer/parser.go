package lexer

import (
	"fmt"
	"io"
	"strings"
)

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

// parsePollStatement parses a POLL statement.
func (p *Parser) parsePollStatement() (*PollStatement, error) {
	stmt := &PollStatement{}
	p.unscan()

	// First token should be a "POLL" keyword.
	if tok, lit := p.scanIgnoreWhitespace(); tok != POLL {
		return nil, newParseError(Tokstr(tok, lit), []string{"POLL"})
	}

	// Next we should read a URL.
	tok, lit := p.scanIgnoreWhitespace()
	if tok != IDENT {
		return nil, newParseError(Tokstr(tok, lit), []string{"URL"})
	}
	stmt.URL = lit

	// Next we should see the "EVERY" keyword.
	if tok, lit := p.scanIgnoreWhitespace(); tok != EVERY {
		return nil, newParseError(Tokstr(tok, lit), []string{"EVERY"})
	}

	// Next we should read polling interval.
	tok, lit = p.scanIgnoreWhitespace()
	if tok != IDENT {
		return nil, newParseError(Tokstr(tok, lit), []string{"PollingInterval"})
	}
	stmt.Interval = lit

	// If the next token is an After then look for InitialWaitTime.
	tok, _ = p.scanIgnoreWhitespace()
	if tok == AFTER {
		tokafter, litafter := p.scanIgnoreWhitespace()
		if tokafter != IDENT {
			return nil, newParseError(Tokstr(tokafter, litafter), []string{"InitialWaitTime"})
		}
		stmt.InitalWaitTime = litafter
	} else {
		p.unscan()
	}

	// Next we should read MaxRetryCount.
	tok, lit = p.scanIgnoreWhitespace()
	if tok != IDENT {
		return nil, newParseError(Tokstr(tok, lit), []string{"MaxRetryCount"})
	}
	stmt.MaxRetryCount = lit

	// Next we should see the "TIMES" keyword.
	if tok, lit := p.scanIgnoreWhitespace(); tok != TIMES {
		return nil, newParseError(Tokstr(tok, lit), []string{"TIMES"})
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseDownloadStatement parses a DOWNLOAD statement.
func (p *Parser) parseDownloadStatement() (*DownloadStatement, error) {
	stmt := &DownloadStatement{}
	p.unscan()

	// First token should be a "DOWNLOAD" keyword.
	if tok, lit := p.scanIgnoreWhitespace(); tok != DOWNLOAD {
		return nil, newParseError(Tokstr(tok, lit), []string{"DOWNLOAD"})
	}

	// Next we should read FROM.
	if tok, lit := p.scanIgnoreWhitespace(); tok != FROM {
		return nil, newParseError(Tokstr(tok, lit), []string{"FROM"})
	}

	// Next we should read a URL.
	tok, lit := p.scanIgnoreWhitespace()
	if tok != IDENT {
		return nil, newParseError(Tokstr(tok, lit), []string{"URL"})
	}
	stmt.URL = lit

	// Next we should read SAVE.
	if tok, lit := p.scanIgnoreWhitespace(); tok != SAVE {
		return nil, newParseError(Tokstr(tok, lit), []string{"SAVE"})
	}

	// Next we should read TO.
	if tok, lit := p.scanIgnoreWhitespace(); tok != TO {
		return nil, newParseError(Tokstr(tok, lit), []string{"To"})
	}

	// And finally, we should read a filepath.
	tok, lit = p.scanIgnoreWhitespace()
	if tok != IDENT {
		return nil, newParseError(Tokstr(tok, lit), []string{"FILEPATH"})
	}
	stmt.FilePath = lit

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseSQLExecuteStatement parses a SQLExecuteStatement statement.
func (p *Parser) parseSQLExecuteStatement() (*SQLExecuteStatement, error) {
	stmt := &SQLExecuteStatement{}
	p.unscan()

	// First token should be a "DOWNLOAD" keyword.
	if tok, lit := p.scanIgnoreWhitespace(); tok != SQLEXECUTE {
		return nil, newParseError(Tokstr(tok, lit), []string{"SQLEXECUTE"})
	}

	// Next we should read FROM.
	if tok, lit := p.scanIgnoreWhitespace(); tok != FROM {
		return nil, newParseError(Tokstr(tok, lit), []string{"FROM"})
	}

	// Next we should read a URL.
	tok, lit := p.scanIgnoreWhitespace()
	if tok != IDENT {
		return nil, newParseError(Tokstr(tok, lit), []string{"FILEPATH"})
	}
	stmt.FilePath = lit

	// Next we should read SAVE.
	if tok, lit := p.scanIgnoreWhitespace(); tok != INTO {
		return nil, newParseError(Tokstr(tok, lit), []string{"INTO"})
	}

	// And finally, we should read a filepath.
	tok, lit = p.scanIgnoreWhitespace()
	if tok != IDENT {
		return nil, newParseError(Tokstr(tok, lit), []string{"DB"})
	}
	stmt.DBConnectionString = lit

	// Return the successfully parsed statement.
	return stmt, nil
}

// ParseStatement parses an Gorsian string and returns a Statement AST object.
func (p *Parser) ParseStatement() (Statement, error) {
	// Inspect the first token.
	tok, lit := p.scanIgnoreWhitespace()
	switch tok {
	case POLL:
		return p.parsePollStatement()
	case DOWNLOAD:
		return p.parseDownloadStatement()
	case SQLEXECUTE:
		return p.parseSQLExecuteStatement()
	default:
		return nil, newParseError(Tokstr(tok, lit), []string{"POLL", "DOWNLOAD", "SQLEXECUTE"})
	}
}

// ParseQuery parses an query string and returns a Query AST object.
func (p *Parser) ParseQuery() (*Query, error) {
	var statements Statements
	semi := true
	for {
		if tok, lit := p.scanIgnoreWhitespace(); tok == EOF {
			return &Query{Statements: statements}, nil
		} else if tok == SEMICOLON {
			semi = true
		} else {
			if !semi {
				return nil, newParseError(Tokstr(tok, lit), []string{";"})
			}
			p.unscan()
			s, err := p.ParseStatement()
			if err != nil {
				return nil, err
			}
			statements = append(statements, s)
			semi = false
		}
	}
}

// ParseError represents an error that occurred during parsing.
type ParseError struct {
	Message  string
	Found    string
	Expected []string
	Pos      Pos
}

// newParseError returns a new instance of ParseError.
func newParseError(found string, expected []string) *ParseError {
	return &ParseError{Found: found, Expected: expected}
}

// Error returns the string representation of the error.
func (e *ParseError) Error() string {
	if e.Message != "" {
		//return fmt.Sprintf("%s at line %d, char %d", e.Message, e.Pos.Line+1, e.Pos.Char+1)
		return fmt.Sprintf("%s", e.Message)
	}
	//return fmt.Sprintf("found %s, expected %s at line %d, char %d", e.Found, strings.Join(e.Expected, ", "), e.Pos.Line+1, e.Pos.Char+1)
	return fmt.Sprintf("found %s, expected %s", e.Found, strings.Join(e.Expected, ", "))
}
