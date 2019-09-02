package lexer

import (
	"bytes"
	"strings"
)

// Statement represents a single command in InfluxQL.
type Statement interface {
	Node
	// stmt is unexported to ensure implementations of Statement
	// can only originate in this package.
	stmt()
}

// Statements represents a list of statements.
type Statements []Statement

// String returns a string representation of the statements.
func (a Statements) String() string {
	var str []string
	for _, stmt := range a {
		str = append(str, stmt.String())
	}
	return strings.Join(str, ";\n")
}

// Query represents a collection of ordered statements.
type Query struct {
	Statements Statements
}

// String returns a string representation of the query.
func (q *Query) String() string { return q.Statements.String() }

// Node represents a node in the InfluxDB abstract syntax tree.
type Node interface {
	// node is unexported to ensure implementations of Node
	// can only originate in this package.
	node()
	String() string
}

func (*Query) node()     {}
func (Statements) node() {}

func (*PollStatement) node()       {}
func (*DownloadStatement) node()   {}
func (*SQLExecuteStatement) node() {}

func (*PollStatement) stmt()       {}
func (*DownloadStatement) stmt()   {}
func (*SQLExecuteStatement) stmt() {}

// PollStatement represents a command for polling an endpoint.
type PollStatement struct {
	// Polling internval
	Interval       string
	URL            string
	InitalWaitTime string
	MaxRetryCount  string
}

// String returns a string representation of the poll statement.
func (p *PollStatement) String() string {
	var buf bytes.Buffer
	_, _ = buf.WriteString("POLL ")
	_, _ = buf.WriteString(p.URL)
	_, _ = buf.WriteString(" EVERY ")
	_, _ = buf.WriteString(p.Interval)

	if p.InitalWaitTime != "" {
		_, _ = buf.WriteString(" AFTER ")
		_, _ = buf.WriteString(p.InitalWaitTime)
	}

	_, _ = buf.WriteString(" ")
	_, _ = buf.WriteString(p.MaxRetryCount)
	_, _ = buf.WriteString(" TIMES")
	return buf.String()
}

// DownloadStatement represents a command to download files and store them in a path.
type DownloadStatement struct {
	URL      string
	FilePath string
}

// String returns a string representation of the download statement.
func (d *DownloadStatement) String() string {
	var buf bytes.Buffer
	_, _ = buf.WriteString("DOWNLOAD FROM ")
	_, _ = buf.WriteString(d.URL)
	_, _ = buf.WriteString(" SAVE TO ")
	_, _ = buf.WriteString(d.FilePath)
	return buf.String()
}

// SQLExecuteStatement represents a command execute a sql file against a db.
type SQLExecuteStatement struct {
	DBConnectionString string
	FilePath           string
}

// String returns a string representation of the download statement.
func (s *SQLExecuteStatement) String() string {
	var buf bytes.Buffer
	_, _ = buf.WriteString("SQLEXECUTE FROM ")
	_, _ = buf.WriteString(s.FilePath)
	_, _ = buf.WriteString(" INTO ")
	_, _ = buf.WriteString(s.DBConnectionString)
	return buf.String()
}
