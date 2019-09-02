package lexer

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
)

var errBadString = errors.New("bad string")
var errBadEscape = errors.New("bad escape")

const eof = rune(0)

//Scanner represents a lexical scanner
type Scanner struct {
	r *bufio.Reader
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() { _ = s.r.UnreadRune() }

// Scan returns the next token and literal value.
func (s *Scanner) Scan() (tok Token, lit string) {
	// Read the next rune.
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter then consume as an ident or reserved word.
	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) {
		s.unread()
		return s.scanIdent()
	}

	// Otherwise read the individual character.
	switch ch {
	case eof:
		return EOF, ""
	case '"':
		s.unread()
		return s.scanIdent()
	case '{':
		s.unread()
		return s.scanIdent()
	case '#':
		return ASTERISK, string(ch)
	case ',':
		return COMMA, string(ch)
	case ';':
		return SEMICOLON, string(ch)
	case '+':
		return ADD, ""
	case '-':
		return SUB, ""
	case '*':
		return MUL, ""
	case '/':
		return DIV, ""
	case '=':
		if ch1 := s.read(); ch1 == '~' {
			return EQREGEX, ""
		}
		s.unread()
		return EQ, ""
	case '!':
		if ch1 := s.read(); ch1 == '=' {
			return NEQ, ""
		} else if ch1 == '~' {
			return NEQREGEX, ""
		}
		s.unread()
	case '>':
		if ch1 := s.read(); ch1 == '=' {
			return GTE, ""
		}
		s.unread()
		return GT, ""
	case '<':
		if ch1 := s.read(); ch1 == '=' {
			return LTE, ""
		} else if ch1 == '>' {
			return NEQ, ""
		}
		s.unread()
		return LT, ""
	}

	return ILLEGAL, string(ch)
}

// scanString reads a quoted string from a rune reader.
func scanString(r *bufio.Reader) (string, error) {
	var buf bytes.Buffer
	ending, _, err := r.ReadRune()
	if err != nil {
		return "", errBadString
	}
	//this is for json objects. Look for curly brackets
	if ending == '{' {
		buf.WriteRune('{')
		ending = '}'
	}
	for {
		ch0, _, err := r.ReadRune()
		if ch0 == ending {
			if ending == '}' {
				buf.WriteRune('}')
			}
			return buf.String(), nil
		} else if err != nil || ch0 == '\n' {
			return buf.String(), errBadString
		} else if ch0 == '\\' {
			// If the next character is an escape then write the escaped char.
			// If it's not a valid escape then return an error.
			ch1, _, _ := r.ReadRune()
			if ch1 == 'n' {
				_, _ = buf.WriteRune('\n')
			} else if ch1 == '\\' {
				_, _ = buf.WriteRune('\\')
			} else if ch1 == '"' {
				_, _ = buf.WriteRune('"')
			} else if ch1 == '\'' {
				_, _ = buf.WriteRune('\'')
			} else {
				return string(ch0) + string(ch1), errBadEscape
			}
		} else {
			_, _ = buf.WriteRune(ch0)
		}
	}
}

// scanString consumes a contiguous string of non-quote characters.
// Quote characters can be consumed if they're first escaped with a backslash.
func (s *Scanner) scanString() (tok Token, lit string) {
	var err error
	lit, err = scanString(s.r)
	if err == errBadString {
		return BADSTRING, lit
	} else if err == errBadEscape {
		return BADESCAPE, lit
	}
	return STRING, lit
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return WS, buf.String()
}

// scanIdent consumes the current rune and all contiguous ident runes.
func (s *Scanner) scanIdent() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	//buf.WriteRune(s.read())

	// Read every ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if ch == '"' {
			s.unread()
			tok0, lit0 := s.scanString()
			if tok0 == BADSTRING || tok0 == BADESCAPE {
				return tok0, lit0
			}
			buf.WriteString(lit0)
			break
		} else if ch == '{' {
			s.unread()
			tok0, lit0 := s.scanString()
			if tok0 == BADSTRING || tok0 == BADESCAPE {
				return tok0, lit0
			}
			buf.WriteString(lit0)
			break
		} else if !isLetter(ch) && !isDigit(ch) && ch != '_' {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// If the string matches a keyword then return that keyword.
	switch strings.ToUpper(buf.String()) {
	case "POLL":
		return POLL, buf.String()
	case "EVERY":
		return EVERY, buf.String()
	case "AFTER":
		return AFTER, buf.String()
	case "DOWNLOAD":
		return DOWNLOAD, buf.String()
	case "FROM":
		return FROM, buf.String()
	case "SAVE":
		return SAVE, buf.String()
	case "TO":
		return TO, buf.String()
	case "TIMES":
		return TIMES, buf.String()
	case "SQLEXECUTE":
		return SQLEXECUTE, buf.String()
	case "INTO":
		return INTO, buf.String()
	}

	// Otherwise return as a regular identifier.
	return IDENT, buf.String()
}
