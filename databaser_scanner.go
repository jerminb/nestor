package nestor

import (
	"bufio"
	"regexp"
	"strings"
)

//DatabaserScanner is used to find named queries in a sql file based on a configurable regex
type DatabaserScanner struct {
	regex   string
	line    string
	queries map[string]string
	current string
}

type stateFn func(*DatabaserScanner) stateFn

//GetTag returns a matching tag for a given regex
func GetTag(line string, regex string) string {
	re := regexp.MustCompile(regex)
	matches := re.FindStringSubmatch(line)
	if matches == nil {
		return ""
	}
	return matches[1]
}

func initialState(s *DatabaserScanner) stateFn {
	if tag := GetTag(s.line, s.regex); len(tag) > 0 {
		s.current = tag
		return queryState
	}
	return initialState
}

func queryState(s *DatabaserScanner) stateFn {
	if tag := GetTag(s.line, s.regex); len(tag) > 0 {
		s.current = tag
	} else {
		s.appendQueryLine()
	}
	return queryState
}

func (s *DatabaserScanner) appendQueryLine() {
	current := s.queries[s.current]
	line := strings.Trim(s.line, " \t")
	reg := regexp.MustCompile(`^-* *$`)
	line = reg.ReplaceAllString(line, "${2}")
	if len(line) == 0 {
		return
	}

	if len(current) > 0 {
		current = current + "\n"
	}
	current = current + line
	s.queries[s.current] = current
}

//Run iterates through a bufio scanner and finds named queires
func (s *DatabaserScanner) Run(io *bufio.Scanner) map[string]string {
	s.queries = make(map[string]string)

	for state := initialState; io.Scan(); {
		s.line = io.Text()
		state = state(s)
	}

	return s.queries
}

//NewDtabaserScanner is constructor for DtabaserScanner class
func NewDtabaserScanner(regex string) *DatabaserScanner {
	return &DatabaserScanner{
		regex: regex,
	}
}
