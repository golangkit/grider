package idataset

import (
	"context"
	"fmt"
	"os"
	"testing"

	"bitbucket.org/elephantsoft/rdk/dbw"
	"github.com/rs/zerolog"
)

var (
	zl  zerolog.Logger
	db  *dbw.DB
	ctx = context.Background()
)

func TestMain(m *testing.M) {
	constr, ok := os.LookupEnv("LMS_DB_CONNECTION")
	if !ok {
		fmt.Println("set env var LMS_DB_CONNECTION")
		fmt.Println("as \"user=x password=x host=127.0.0.1 port=5342 dbname=x sslmode='disable' search_path='x' bytea_output='hex'\"")
		os.Exit(2)
	}

	var err error
	db, err = dbw.Open("postgres", constr, &zl)

	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	zl = zerolog.New(os.Stdout)

	os.Exit(m.Run())
}

func TestInteractiveDataset_GetRefBook(t *testing.T) {

	srv := NewService(&zl, NewRepository(db))

	if err := srv.Init(ctx); err != nil {
		t.Error(err)
	}
	srv.Start(ctx)
}

func TestConvertStringSliceToJSON(t *testing.T) {

	r := Result{
		Columns: []string{"id", "read_at", "is_mandatory"},
		Formats: []string{`{"align": "left", "hidden": true}`, "{}", "{}"},
		Rows: [][]string{
			{"1", "01.01.2020", "true"},
			{"2", "02.01.2020", "false"},
		}}

	buf := r.JSON()
	println(string(buf))

}
func TestExtractColumnName(t *testing.T) {

	var tcase = []struct {
		in  string
		exp string
		ok  bool
	}{{"/api/{id}", "id", true},
		{"/api/{hello}/summary", "hello", true}}

	t.Log("started")
	for i := range tcase {
		res, ok := extractColumnName(tcase[i].in)
		if tcase[i].ok != ok || tcase[i].exp != res {
			t.Errorf("tcase %d failed. got %s, expected %s", i, res, tcase[i].exp)
		}
	}
}

func TestReplaceSortPlaceholder(t *testing.T) {
	var tcase = []struct {
		sqltext string
		ob      string
		exp     string
	}{
		{"select * from mountings /*sort*/", "", "select * from mountings /*sort*/"},
		{"select * from mountings /*sort*/", "id", "select * from mountings order by id"},
		{"select * from mountings /*sort:*/", "", "select * from mountings /*sort:*/"},
		{"select * from mountings /*sort:*/", "id", "select * from mountings order by id"},
		{"select * from mountings /*sort:id*/", "", "select * from mountings order by id"},
		{"select * from mountings /*sort:id*/", "name", "select * from mountings order by name"},
		{"select * from mountings /* sort:id*/", "name", "select * from mountings /* sort:id*/"},
	}

	t.Log("started")
	for i := range tcase {
		res := replaceSortPlaceholder(tcase[i].sqltext, tcase[i].ob)
		if tcase[i].exp != res {
			t.Errorf("tcase %d failed. got %s, expected %s", i, res, tcase[i].exp)
		}
	}
}
