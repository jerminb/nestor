package nestor

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jerminb/nestor/lexer"
)

const (
	defaultPollerHTTPMethod   string = "GET"
	defaultPollerHTTPResponse string = "200 OK"
)

//ExecuteAt is an enum to identify the execution step.
// At the moment there are only two possibilities; Startup time or at a scheduled frequency
type ExecuteAt int

const (
	//ExecuteAtStartUp is the startup-time enum
	ExecuteAtStartUp ExecuteAt = iota
	//ExecuteAtScheduled is the scheduled frequency enum
	ExecuteAtScheduled
)

// Executable is a generic interface for tasks to implement
type Executable interface {
	Execute(params ...interface{}) (result []reflect.Value, err error)
}

//ExecuteFromStatement executes a lexer.Statement using executable mapping which is implemented in getExecutableFromStatement
func ExecuteFromStatement(stmt lexer.Statement) (result []reflect.Value, err error) {
	exec, err := getExecutableFromStatement(stmt)
	if err != nil {
		return nil, err
	}
	params, err := getParameterForExecutable(stmt)
	if err != nil {
		return nil, err
	}
	return exec.Execute(params...)
}

func getExecutableFromStatement(stmt lexer.Statement) (Executable, error) {
	expected := []string{"PollStatement", "DownloadStatement", "SQLExecuteStatement"}
	switch v := stmt.(type) {
	case *(lexer.PollStatement):
		return NewPoller(), nil
	case *(lexer.DownloadStatement):
		return NewDownloader(), nil
	default:
		return nil, fmt.Errorf("found %v. expected %s", v, strings.Join(expected, ", "))
	}
}

func getParameterForExecutable(stmt lexer.Statement) ([]interface{}, error) {
	expected := []string{"PollStatement", "DownloadStatement", "SQLExecuteStatement"}
	switch v := stmt.(type) {
	case *(lexer.PollStatement):
		return getPollerExecutableParameters(stmt.(*(lexer.PollStatement)))
	case *(lexer.DownloadStatement):
		return getDownloaderExecutableParameters(stmt.(*(lexer.DownloadStatement)))
	default:
		return nil, fmt.Errorf("found %v. expected %s", v, strings.Join(expected, ", "))
	}
}

func getPollerExecutableParameters(pollstmt *lexer.PollStatement) ([]interface{}, error) {
	maxRetryCount, err := strconv.Atoi(pollstmt.MaxRetryCount)
	if err != nil {
		return nil, err
	}
	interval, err := time.ParseDuration(pollstmt.Interval)
	if err != nil {
		return nil, err
	}
	params := make([]interface{}, 0)
	params = append(params, pollstmt.URL)
	params = append(params, defaultPollerHTTPMethod)
	params = append(params, maxRetryCount)
	params = append(params, defaultPollerHTTPResponse)
	params = append(params, interval)
	return params, nil
}

func getDownloaderExecutableParameters(dlstmt *lexer.DownloadStatement) ([]interface{}, error) {
	params := make([]interface{}, 0)
	params = append(params, dlstmt.FilePath)
	params = append(params, dlstmt.URL)
	return params, nil
}
