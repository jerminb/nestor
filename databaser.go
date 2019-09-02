package nestor

import (
	"bufio"
	"database/sql"
	"io"
	"os"
	"reflect"
	"regexp"
)

// Execer is an interface used by Exec.
type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

//DatabaserResult is an encapsulation of possible outcomes of a
//db.Exec to allow for multiple command execution
type DatabaserResult struct {
	SQLResult sql.Result
	Error     error
}

//Databaser applies content of a file to a given database.
// To make the class as generic as possible, sql db is injected.
// The implementation of inverted logic can be found in SQLDBFactory.
type Databaser struct {
	queries map[string]string
}

//ApplyWithSection applies one or more sections of a file to a db.
//regex is to allow for annotation and selevtive application like:
// -- name: insert-profile -> Exec(insert-profile)
func (d *Databaser) ApplyWithSection(filepath string, regex string, namedQueryRegex string, db *sql.DB) ([]DatabaserResult, error) {
	err := d.LoadFromFile(filepath, namedQueryRegex)
	if err != nil {
		return nil, err
	}
	return d.Exec(db, regex)
}

// Load imports sql queries from any io.Reader.
func (d *Databaser) Load(r io.Reader, namedQueryRegex string) error {
	scanner := NewDtabaserScanner(namedQueryRegex)
	queries := scanner.Run(bufio.NewScanner(r))

	d.queries = queries

	return nil
}

// LoadFromFile imports SQL queries from the file.
func (d *Databaser) LoadFromFile(sqlFile string, namedQueryRegex string) error {
	f, err := os.Open(sqlFile)
	if err != nil {
		return err
	}
	defer f.Close()

	return d.Load(f, namedQueryRegex)
}

func (d *Databaser) lookupQuery(name string) (query []string, err error) {
	r, err := regexp.Compile(name)
	if err != nil {
		return nil, err
	}
	query = make([]string, 0, len(d.queries))
	for _, v := range d.queries {
		if r.MatchString(v) {
			query = append(query, v)
		}
	}
	return query, nil
}

// Exec is a wrapper for database/sql's Exec(), using dotsql named query.
func (d *Databaser) Exec(db Execer, name string) ([]DatabaserResult, error) {
	query, err := d.lookupQuery(name)
	if err != nil {
		return nil, err
	}
	result := make([]DatabaserResult, 0)
	for _, q := range query {
		r, e := db.Exec(q)
		result = append(result, DatabaserResult{
			SQLResult: r,
			Error:     e,
		})
	}
	return result, nil
}

//Execute executes Databaser's ApplyWithSection to implement Executable interface
func (d *Databaser) Execute(params ...interface{}) (result []reflect.Value, err error) {
	return execute(d.ApplyWithSection, params...)
}

//NewDatabaser is the constructor Databaser class with IoC
func NewDatabaser() *Databaser {
	return &Databaser{}
}
