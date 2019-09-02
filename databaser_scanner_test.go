package nestor_test

import (
	"bufio"
	"strings"
	"testing"

	"github.com/jerminb/nestor"
)

func TestGetTag(t *testing.T) {
	var tests = []struct {
		regex string
		line  string
		want  string
	}{
		{"^\\s*--\\s*name:\\s*(\\S+)", "SELECT 1+1", ""},
		{"^\\s*--\\s*name:\\s*(\\S+)", "-- Some Comment", ""},
		{"^\\s*--\\s*name:\\s*(\\S+)", "-- name:  ", ""},
		{"^\\s*--\\s*name:\\s*(\\S+)", "-- name: find-users-by-name", "find-users-by-name"},
		{"^\\s*--\\s*name:\\s*(\\S+)", "  --  name:  save-user ", "save-user"},
		{"^\\s*--\\s*Data for Name:\\s*(\\S+);", "-- Data for Name: eda_sp_permission;", "eda_sp_permission"},
		{"^\\s*--\\s*Data for Name:\\s*(\\S+);", "-- Data for Name: eda_sp_permission; Type: TABLE DATA; Schema: eda_tenant1; Owner: -", "eda_sp_permission"},
		{"\\--(.*)", "-- Data for Name: eda_sp_permission; Type: TABLE DATA; Schema: eda_tenant1; Owner: -", " Data for Name: eda_sp_permission; Type: TABLE DATA; Schema: eda_tenant1; Owner: -"},
	}

	for _, c := range tests {
		got := nestor.GetTag(c.line, c.regex)
		if got != c.want {
			t.Errorf("isTag('%s') == %s, expect %v", c.line, got, c.want)
		}
	}
}

func TestScannerRun(t *testing.T) {
	sqlFile := `
	--
	-- Data for Name: eda_sp_permission; Type: TABLE DATA; Schema: eda_tenant1; Owner: -
	--

	INSERT INTO eda_tenant1.eda_sp_permission VALUES ('4da8563e-767b-42f8-93fb-a923121761eb', 0, 'PERM_DATAVIEWER');
	INSERT INTO eda_tenant1.eda_sp_permission VALUES ('dfd7b02f-6713-4fdf-a9f5-7ef7d98654aa', 0, 'PERM_LETTER_TPL_LST');
	INSERT INTO eda_tenant1.eda_sp_permission VALUES ('42dd4f1e-1ea5-4beb-836c-0dd9260db23f', 0, 'PERM_ORPHAN_BRANCH_ACCESS');

	--
	-- Data for Name: eda_sp_perm; Type: TABLE DATA; Schema: eda_tenant1; Owner: -
	--

	INSERT INTO eda_tenant1.eda_sp_perm VALUES ('db63e1dc-456e-4909-811f-747254df5aed', '4da8563e-767b-42f8-93fb-a923121761eb');
	INSERT INTO eda_tenant1.eda_sp_perm VALUES ('db63e1dc-456e-4909-811f-747254df5aed', '42dd4f1e-1ea5-4beb-836c-0dd9260db23f');
	`

	scanner := nestor.NewDtabaserScanner("^\\s*--\\s*Data for Name:\\s*(\\S+);")
	queries := scanner.Run(bufio.NewScanner(strings.NewReader(sqlFile)))

	numberOfQueries := len(queries)
	expectedQueries := 2
	if numberOfQueries != expectedQueries {
		t.Errorf("Scanner/Run() has %d queries instead of %d",
			numberOfQueries, expectedQueries)
	}
}
