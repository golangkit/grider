package idataset

import (
	"errors"
	"fmt"
	"hash/crc32"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gopkg.in/labstack/echo.v1"
)

func (s *Service) ExecHandler(rm RestMediator) func(*echo.Context) error {
	return func(c *echo.Context) error {

		idset := s.ByID(c.Param("id"))
		if idset == nil {
			return rm.ResponseWithError(c, 404, errors.New("interactive report not found"))
		}

		flog := s.log.Enter("id", idset.ID)
		defer flog.Exit()

		display := "table"
		var (
			//qparams []interface{} // holds SQL params $1, $2...
			// pval holds parsed value from interactive_datasets.query_params
			// [position]name. Position starts from 1.
			pval map[int]string

			filter     Filter
			err        error
			isDownload bool
		)

		if k := len(idset.Params); k > 0 {
			filter.Params = make([]interface{}, k)
			pval = make(map[int]string, k)
		}

		for key, p := range c.Request().URL.Query() {
			switch key {
			case "download":
				isDownload = true
				continue
			case "Display":
				if len(p) > 0 {
					display = p[0]
				}
				continue
			case "pageNumber":
				if len(p) == 0 {
					filter.PageNumber = 0
					continue
				}
				filter.PageNumber, err = strconv.Atoi(p[0])
				if err != nil {
					return rm.ResponseWithError(c, 400, errors.New("pageNumber got invalid value"))
				}
				continue
			case "pageSize":
				if len(p) == 0 {
					continue
				}
				filter.PageSize, err = strconv.Atoi(p[0])
				if err != nil {
					return rm.ResponseWithError(c, 400, errors.New("pageSize got invalid value"))
				}
				continue
			case "sortBy":
				if len(p) == 0 {
					continue
				}
				filter.SortBy = p[0]
				continue
			default:
				break
			}

			if pval != nil {
				pos, ok := idset.Params[key]
				if ok {
					pval[pos-1] = strings.Join(p, ":")
				}
			}
		}

		for pos, v := range pval {
			filter.Params[pos] = v
		}

		sp := ""
		sep := ""
		for i := range filter.Params {
			sp += fmt.Sprintf("$%d=%v%s", i+1, filter.Params[i], sep)
			sep = ";"
		}

		filter.SetDefaults()

		e := flog.Debug().
			Int("pageNumber", filter.PageNumber).
			Int("pageSize", filter.PageSize).
			Str("sqlparams", sp).Str("display", display)
		if filter.SortBy != "" {
			e = e.Str("sortBy", filter.SortBy).Bool("desc", filter.Desc)
		}
		e.Msg("interactive execution request")

		var (
			result []byte
			hash   uint32
			total  int
		)

		if display == "table" {
			res, err := s.Grid(idset.ID, nil, hash, filter)
			if err != nil {
				return rm.ResponseWithError(c, 500, err)
			}

			if isDownload {
				grid, err := res.ToGrid()
				if err != nil {
					return rm.ResponseWithError(c, 500, err)
				}
				buf, err := grid.Excelize(fmt.Sprintf("%s-%s.xlsx", idset.ID, time.Now().Format("200601021504")))
				if err != nil {
					return rm.ResponseWithError(c, 500, err)
				}
				return c.JSON(http.StatusOK, buf)
			}

			total = len(res.Rows)
			from, to := filter.Cut(len(res.Rows))
			res.Rows = res.Rows[from:to]

			// resp := Result{
			// 	Columns:        res.Columns,
			// 	Formats:        res.Formats,
			// 	Rows:           res.Rows[from:to],
			// 	IsDownloadable: true,
			// }
			g, err := res.ToGrid()
			if err != nil {
				return rm.ResponseWithError(c, 500, err)
			}

			//fmt.Printf("$$%#v\n", g)
			result, err = g.JSON()
			if err != nil {
				return rm.ResponseWithError(c, 500, err)
			}
		}

		if display == "chart" {
			result, err = s.Chart(idset.ID, hash, filter)
			if err != nil {
				return rm.ResponseWithError(c, 500, err)
			}
		}

		hash = crc32.ChecksumIEEE(result)

		c.Response().Header().Add("X-cabinets-hash", strconv.Itoa(int(hash)))
		c.Response().Header().Add("X-Total-Count", strconv.Itoa(total))
		return c.JSONBlob(http.StatusOK, result)
	}
}

func (s *Service) TypesHandler(rm RestMediator) func(*echo.Context) error {
	return func(c *echo.Context) error {
		page := c.Query("page")
		places := []string{page}
		if page == "reports" {
			places = append(places, "dashboardReports")
		}
		return c.JSON(http.StatusOK, s.Available(places, rm.Lang(c)))
	}
}

func (s *Service) RefreshCache(rm RestMediator) func(*echo.Context) error {
	return func(c *echo.Context) error {

		if err := s.repo.RefreshCache(c.Context); err != nil {
			return rm.ResponseWithError(c, 500, err)
		}

		return c.JSON(http.StatusOK, struct {
			Result bool `json:"result"`
		}{true})
	}
}
