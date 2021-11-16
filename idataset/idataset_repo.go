package idataset

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"sync"

	"bitbucket.org/elephantsoft/rdk/dbw"
	"bitbucket.org/elephantsoft/rdk/dbw/table"
	"github.com/rs/zerolog"
)

type repo struct {
	db  *dbw.DB
	log zerolog.Logger

	idsets struct {
		tbl  *table.Table
		mux  sync.RWMutex
		list []InteractiveDataset
		idx  map[string]int
	}
}

// NewRepository конструктор основного репозитория.
func NewRepository(db *dbw.DB) Repository {
	r := &repo{
		db:  db,
		log: db.Logger().With().Str("repo", "idatasets").Logger(),
	}

	r.idsets.tbl = table.NewSimple(db, "interactive_datasets", &InteractiveDataset{})

	return r
}

func (r *repo) Init(ctx context.Context) error {
	return r.cache(ctx)
}

func (r *repo) cacheInteractiveDatasets(ctx context.Context) error {
	var row InteractiveDataset
	var list []InteractiveDataset
	if err := r.idsets.tbl.DoSelectCache(func() error {
		row.Params = nil
		if len(row.QueryParams) > 0 {
			if err := json.Unmarshal(row.QueryParams, &row.Params); err != nil {
				return err
			}
		}
		list = append(list, row)
		row.QueryParams = nil
		return nil
	}, &row); err != nil {
		return err
	}

	sort.Slice(list, func(i, j int) bool {
		cmp := strings.Compare(list[i].PresentationPlace, list[j].PresentationPlace)
		if cmp == 0 {
			return list[i].PresentationOrder < list[j].PresentationOrder
		}
		return cmp < 0
	})

	idx := make(map[string]int, len(list))
	for i := range list {
		idx[list[i].ID] = i
	}

	r.idsets.mux.Lock()
	r.idsets.list = append(r.idsets.list[:0], list...)
	r.idsets.idx = idx
	r.idsets.mux.Unlock()

	return nil
}

func (r *repo) RefreshCache(ctx context.Context) error {
	return r.cacheInteractiveDatasets(ctx)
}

func (r *repo) cache(ctx context.Context) error {
	if err := r.cacheInteractiveDatasets(ctx); err != nil {
		return err
	}
	return nil
}

func (r *repo) Traverse(f func(*InteractiveDataset)) {
	r.idsets.mux.RLock()
	for i := range r.idsets.list {
		f(&r.idsets.list[i])
	}
	r.idsets.mux.RUnlock()
}

func (r *repo) ByID(id string) *InteractiveDataset {
	r.idsets.mux.RLock()
	idx, ok := r.idsets.idx[id]
	if !ok {
		r.idsets.mux.RUnlock()
		return nil
	}
	res := r.idsets.list[idx]
	res.Columns = make([]string, len(res.Columns))
	copy(res.Columns, r.idsets.list[idx].Columns)
	r.idsets.mux.RUnlock()
	return &res
}

func (r *repo) Query(qry string, params ...interface{}) (*Result, error) {

	var res Result

	rows, err := r.db.SQLDB().Query(qry, params...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	res.Columns, err = rows.Columns()
	if err != nil {
		return nil, err
	}

	row := make([]string, len(res.Columns))
	ptr := make([]interface{}, len(res.Columns))
	for j := range row {
		ptr[j] = &row[j]
	}

	first := true
	for rows.Next() {
		if err := rows.Scan(ptr...); err != nil {
			return nil, err
		}

		if first {
			first = false
			res.Formats = make([]string, len(row))
			copy(res.Formats, row)
			continue
		}
		nrow := make([]string, len(row))
		copy(nrow, row)
		res.Rows = append(res.Rows, nrow)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *repo) AssignColumns(id string, cols []string) {
	r.idsets.mux.Lock()
	idx, ok := r.idsets.idx[id]
	if ok {
		r.idsets.list[idx].Columns = cols
	}
	r.idsets.mux.Unlock()
}
