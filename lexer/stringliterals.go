package lexer

// Token represents a lexical token.
type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	WS

	//Operators
	// ADD and the following are Gorsian Operators
	ADD      // +
	SUB      // -
	MUL      // *
	DIV      // /
	AND      // AND
	OR       // OR
	EQ       // =
	NEQ      // !=
	EQREGEX  // =~
	NEQREGEX // !~
	LT       // <
	LTE      // <=
	GT       // >
	GTE      // >=

	// Literals
	IDENT     // timestamps, intervals, urls, filepaths
	STRING    // "abc"
	BADSTRING // "abc
	BADESCAPE // \q

	// Misc characters
	ASTERISK  // *
	COMMA     // ,
	SEMICOLON // ;

	// Keywords
	POLL
	EVERY
	AFTER
	DOWNLOAD
	FROM
	SAVE
	TO
	TIMES
	SQLEXECUTE
	INTO
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	WS:      "WS",

	ADD:      "+",
	SUB:      "-",
	MUL:      "*",
	DIV:      "/",
	AND:      "AND",
	OR:       "OR",
	EQ:       "=",
	NEQ:      "!=",
	EQREGEX:  "=~",
	NEQREGEX: "!~",
	LT:       "<",
	LTE:      "<=",
	GT:       ">",
	GTE:      ">=",

	ASTERISK:  "*",
	COMMA:     ",",
	SEMICOLON: ";",

	IDENT:      "IDENT",
	POLL:       "POLL",
	EVERY:      "EVERY",
	AFTER:      "AFTER",
	DOWNLOAD:   "DOWNLOAD",
	FROM:       "FROM",
	SAVE:       "SAVE",
	TO:         "TO",
	TIMES:      "TIMES",
	SQLEXECUTE: "SQLEXECUTE",
	INTO:       "INTO",
}

// String returns the string representation of the token.
func (tok Token) String() string {
	if tok >= 0 && tok < Token(len(tokens)) {
		return tokens[tok]
	}
	return ""
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '.'
}
func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}
