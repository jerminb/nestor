package nestor_test

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jerminb/nestor"
)

func TestApply(t *testing.T) {
	db, mock, err := sqlmock.New()
	dber := nestor.NewDatabaser()

	mock.ExpectExec("INSERT INTO eda_sp_permission;\nINSERT INTO eda_sp_permission").WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := dber.ApplyWithSection("assets/databaser_test_insert.sql", "eda_sp_permission", "^\\s*--\\s*Data for Name:\\s*(\\S+);", db)
	if err != nil {
		t.Fatalf("expected nil. got %v", err)
	}
	if result == nil {
		t.Fatalf("expected [1,1]. got nil")
	}
	id, _ := result[0].SQLResult.LastInsertId()
	if id != 1 {
		t.Fatalf("expected 1. got %d", id)
	}
}

func TestApplyMultiple(t *testing.T) {
	db, mock, err := sqlmock.New()
	dber := nestor.NewDatabaser()

	mock.ExpectExec("INSERT INTO eda_sp_permission;").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO eda_sp_perm;").WillReturnResult(sqlmock.NewResult(2, 2))

	result, err := dber.ApplyWithSection("assets/databaser_test_regex.sql", "(.*)", "^\\s*--\\s*Data for Name:\\s*(\\S+);", db)
	if err != nil {
		t.Fatalf("expected nil. got %v", err)
	}
	if result == nil {
		t.Fatalf("expected [1,1]. got nil")
	}
	id, _ := result[0].SQLResult.LastInsertId()
	if id != 1 {
		t.Fatalf("expected 1. got %d", id)
	}
	id, _ = result[1].SQLResult.LastInsertId()
	if id != 2 {
		t.Fatalf("expected 1. got %d", id)
	}
}

func TestExecute_Databaser(t *testing.T) {
	db, mock, err := sqlmock.New()
	dber := nestor.NewDatabaser()

	mock.ExpectExec("INSERT INTO eda_sp_permission;").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO eda_sp_perm;").WillReturnResult(sqlmock.NewResult(2, 2))

	result, err := dber.Execute("assets/databaser_test_regex.sql", "(.*)", "^\\s*--\\s*Data for Name:\\s*(\\S+);", db)
	if err != nil {
		t.Fatalf("expected nil. got %v", err)
	}
	if result == nil {
		t.Fatalf("expected [1,1]. got nil")
	}
	if len(result) != 2 {
		t.Fatalf("expected two result. got %d", len(result))
	}
}
