package idataset

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"sync"
	"time"

	"bitbucket.org/elephantsoft/rdk/dbw"
	"bitbucket.org/rtg365/lms/common/language"
	"github.com/axkit/flogger"
	"github.com/rs/zerolog"
)

type Service struct {
	log  flogger.FuncLogger
	repo Repository
	//cr   croner.Servicer

	cache struct {
		mux    sync.RWMutex
		result map[string]QueryResult
	}
}

type QueryResult struct {
	CreatedAt  time.Time
	AccessedAt time.Time
	ExpiredAt  time.Time
	*Result
}

func NewService(l *zerolog.Logger, repo Repository) *Service {
	s := Service{
		log:  flogger.New(l, "service", "idataset"),
		repo: repo}

	s.cache.result = make(map[string]QueryResult)
	return &s
}

var _ Servicer = (*Service)(nil)

func (s *Service) Init(ctx context.Context) error {
	return s.repo.Init(ctx)
}

func (s *Service) Start(ctx context.Context) error {
	go s.cacheCleanRunner(ctx)
	return nil
}

func (s *Service) cacheCleanRunner(ctx context.Context) {

	flog := s.log.EnterSilent()

	t := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.cache.mux.Lock()
			for k, v := range s.cache.result {
				if time.Now().After(v.ExpiredAt) {
					delete(s.cache.result, k)
					flog.Debug().Str("key", k).Msg("cached query resultset cleaned")
				}
			}
			s.cache.mux.Unlock()
		}
	}
}

func (s *Service) Traverse(f func(*InteractiveDataset)) {
	s.repo.Traverse(f)
}

func (s *Service) ByID(id string) *InteractiveDataset {
	return s.repo.ByID(id)
}

func (s *Service) Available(place []string, lang language.Index) []InteractiveDatasetHeader {
	var res []InteractiveDatasetHeader
	s.repo.Traverse(func(idata *InteractiveDataset) {
		if idata.DeletedAt.Valid() {
			return
		}

		if len(place) > 0 {
			found := false
			for i := range place {
				if idata.PresentationPlace == place[i] {
					found = true
					break
				}
			}
			if !found {
				return
			}
		}

		r := InteractiveDatasetHeader{ID: idata.ID, Title: idata.Title.String}
		if r.Title == "" {
			r.Title = r.ID
		}

		r.Displays = []string{"table"}
		if len(idata.Chart.String) > 0 {
			r.Displays = append(r.Displays, "chart")
		}
		res = append(res, r)
	})

	return res
}

func (s *Service) Grid(id string, o []Optioner, hash uint32, filter Filter) (*Result, error) {
	idset := s.repo.ByID(id)

	if idset == nil {
		return nil, errors.New("unknown interactive dataset id")
	}

	res, err := s.queryDataset(idset, filter)
	if err != nil {
		return nil, err
	}

	if len(idset.Columns) == 0 {
		s.repo.AssignColumns(id, res.Columns)
	}

	return res, nil
}

// func (s *Service) QueryDataset(idset *InteractiveDataset, params ...interface{}) (*Result, error) {
// 	return s.queryDataset(idset, params...)
// }

func (s *Service) queryDataset(idset *InteractiveDataset, filter Filter) (*Result, error) {

	flog := s.log.Enter("id", idset.ID, "params", filter.Params)
	defer flog.Exit()

	var key string
	if idset == nil {
		return nil, errors.New("unknown interactive dataset id")
	}

	if idset.Query.Valid == false {
		return nil, errors.New("query of interactive dataset is not specified")
	}

	sqlText := idset.Query.String
	ob := ""
	if filter.SortBy != "" {
		ob := " order by " + filter.SortBy
		if filter.Desc {
			ob += " desc"
		}
	}

	sqlText = replaceSortPlaceholder(sqlText, ob)

	// if pos := strings.Index(sqlText, "/*sort"); pos >= 0 {
	// 	if sqlText[pos+6] == ':' {

	// 	}
	// 	sqlText = strings.Replace(sqlText, "/*sort*/", ob, 1)

	// }

	flog.Debug().Str("sql", dbw.CleanSpace(sqlText)).Msg("sort required")

	if idset.CacheExpirationDuration != 0 {
		key = calcHash(sqlText, filter.Params...)
		s.cache.mux.Lock()
		qr, ok := s.cache.result[key]
		if ok && time.Now().Before(qr.ExpiredAt) {
			qr.AccessedAt = time.Now()
			s.cache.result[key] = qr
			s.cache.mux.Unlock()
			flog.Debug().Str("key", key).Msg("cached resultset found")
			return qr.Result, nil
		}
		s.cache.mux.Unlock()
	}

	now := time.Now()
	res, err := s.repo.Query(sqlText, filter.Params...)
	if err != nil {
		return nil, err
	}

	if idset.CacheExpirationDuration == 0 {
		return res, nil
	}

	s.cache.mux.Lock()
	now = time.Now()
	s.cache.result[key] = QueryResult{CreatedAt: now, AccessedAt: now, Result: res,
		ExpiredAt: now.Add(time.Second * time.Duration(idset.CacheExpirationDuration)),
	}
	s.cache.mux.Unlock()
	return res, nil
}

func calcHash(qry string, params ...interface{}) string {
	str := qry + fmt.Sprintf("%v", params)
	h := sha256.Sum256([]byte(str))
	return fmt.Sprintf("%x", h)
}

func (s *Service) Chart(id string, hash uint32, filter Filter) ([]byte, error) {

	idset := s.repo.ByID(id)
	res, err := s.queryDataset(idset, filter)
	if err != nil {
		return nil, err
	}

	if idset.Chart.Valid == false {
		return nil, errors.New("chart template of interactive dataset is not specified")
	}

	t := template.Must(template.New("chart").Parse(idset.Chart.String))
	buf := bytes.NewBuffer(nil)
	if err := t.Execute(buf, res.Rows); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// replaceSortPlaceholder replaces substring /*sort*/ by "order by {ob}" or
// /*sort:id*/ by "order by {ob}" or "order by id"
func replaceSortPlaceholder(sqlText string, ob string) string {

	idx := strings.Index(sqlText, "/*sort")

	if idx < 0 {
		return sqlText
	}

	if sqlText[idx+6] == ':' {
		// there is default sort
		cbrack := strings.Index(sqlText[idx+7:], "*/")
		if cbrack < 0 {
			// empty default column
			return sqlText
		}
		defaultob := sqlText[idx+7 : idx+7+cbrack]
		if ob == "" {
			ob = defaultob
		}

		if ob == "" {
			return sqlText
		}
		return sqlText[0:idx] + "order by " + ob + sqlText[idx+7+cbrack+2:]
	}
	// no default sort
	if ob == "" {
		return sqlText
	}
	return strings.Replace(sqlText, "/*sort*/", "order by "+ob, -1)
}
