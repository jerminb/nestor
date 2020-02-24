package lexer

import (
	"bytes"
	"fmt"
	"strings"
)

// Statement represents a single command in InfluxQL.
type Statement interface {
	Node
	// stmt is unexported to ensure implementations of Statement
	// can only originate in this package.
	stmt()
}

//BaseStatement is a base Statement implementation as a container
// for common properties
type BaseStatement struct {
	IsBackground bool
}

func (b *BaseStatement) String() string {
	if b.IsBackground {
		return "&"
	}
	return ""
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

//Get returns a specifc index of Statements array if available
func (a Statements) Get(index int) (Statement, error) {
	if index < len(a) {
		return a[index], nil
	}
	return nil, fmt.Errorf("index out of rangs. statements length is %d, index is %d", len(a), index)
}

// Query represents a collection of ordered statements.
type Query struct {
	BaseStatement
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

func (*BaseStatement) node()       {}
func (*PollStatement) node()       {}
func (*DownloadStatement) node()   {}
func (*SQLExecuteStatement) node() {}
func (*RefreshStatement) node()    {}

func (*Query) stmt() {}

func (*BaseStatement) stmt()       {}
func (*PollStatement) stmt()       {}
func (*DownloadStatement) stmt()   {}
func (*SQLExecuteStatement) stmt() {}
func (*RefreshStatement) stmt()    {}

// PollStatement represents a command for polling an endpoint.
type PollStatement struct {
	BaseStatement
	// Polling internval
	Interval        string
	URL             string
	InitialWaitTime string
	MaxRetryCount   string
}

// String returns a string representation of the poll statement.
func (p *PollStatement) String() string {
	var buf bytes.Buffer
	_, _ = buf.WriteString("POLL ")
	_, _ = buf.WriteString(p.URL)
	_, _ = buf.WriteString(" EVERY ")
	_, _ = buf.WriteString(p.Interval)

	if p.InitialWaitTime != "" {
		_, _ = buf.WriteString(" AFTER ")
		_, _ = buf.WriteString(p.InitialWaitTime)
	}

	_, _ = buf.WriteString(" ")
	_, _ = buf.WriteString(p.MaxRetryCount)
	_, _ = buf.WriteString(" TIMES")

	if p.IsBackground {
		_, _ = buf.WriteString(" &")
	}

	return buf.String()
}

// DownloadStatement represents a command to download files and store them in a path.
type DownloadStatement struct {
	BaseStatement
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

	if d.IsBackground {
		_, _ = buf.WriteString(" &")
	}
	return buf.String()
}

// SQLExecuteStatement represents a command execute a sql file against a db.
type SQLExecuteStatement struct {
	BaseStatement
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
	if s.IsBackground {
		_, _ = buf.WriteString(" &")
	}
	return buf.String()
}

//RefreshStatement represents a refresh command to renew tokens and cerificates
//Artifact specifies whether it is token or certificate that is being refreshed
type RefreshStatement struct {
	BaseStatement
	Artifact string
	Path     string
	// Refresh internval
	Interval string
}

// String returns a string representation of the refresh statement.
func (r *RefreshStatement) String() string {
	var buf bytes.Buffer
	_, _ = buf.WriteString("REFRESH ")
	_, _ = buf.WriteString(r.Artifact)
	_, _ = buf.WriteString(" FROM ")
	_, _ = buf.WriteString(r.Path)
	_, _ = buf.WriteString(" EVERY ")
	_, _ = buf.WriteString(r.Interval)
	if r.IsBackground {
		_, _ = buf.WriteString(" &")
	}
	return buf.String()
}
