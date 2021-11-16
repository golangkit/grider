package idataset

import (
	"context"
	"strconv"

	"bitbucket.org/elephantsoft/rdk"
	"bitbucket.org/rtg365/lms/backend/basefilter"
	"bitbucket.org/rtg365/lms/common/language"
	"gopkg.in/guregu/null.v3"
	"gopkg.in/labstack/echo.v1"
)

type Servicer interface {
	Init(ctx context.Context) error
	Start(ctx context.Context) error

	// Grid returns query rows
	Grid(id string, o []Optioner, hash uint32, filter Filter) (*Result, error)

	Chart(id string, hash uint32, filter Filter) ([]byte, error)
	// Available returns list of interactive dataset codes available in place.
	Available(place []string, lang language.Index) []InteractiveDatasetHeader

	//	Excelize(fname string, r *Result) (*DownloadResponse, error)
	//	QueryDataset(idset *InteractiveDataset, params ...interface{}) (*Result, error)
	ByID(string) *InteractiveDataset

	ExecHandler(RestMediator) func(*echo.Context) error
	TypesHandler(RestMediator) func(*echo.Context) error
	RefreshCache(RestMediator) func(*echo.Context) error
}

// Repository публичный, для возможности замокать снаружи.
type Repository interface {
	Init(ctx context.Context) error
	ByID(string) *InteractiveDataset
	Traverse(f func(*InteractiveDataset))
	Query(qry string, params ...interface{}) (*Result, error)
	AssignColumns(id string, cols []string)
	RefreshCache(context.Context) error
}

type RestMediator interface {
	Lang(*echo.Context) language.Index
	ResponseWithError(c *echo.Context, code int, err error) error
}

// InteractiveDataset
type InteractiveDataset struct {
	ID                      string      `json:"id"`
	Query                   null.String `json:"query"`
	Chart                   null.String `json:"chart"`
	PresentationOrder       int         `json:"presentation_order"`
	PresentationPlace       string      `json:"-"`
	Title                   null.String `json:"title"`
	QueryParams             []byte      `json:"-"`
	HtmlTemplate            null.String `json:"-"`
	CacheExpirationDuration int         // duration in seconds. zero - no caching
	rdk.SystemColumns

	// calculated field
	IsOK    bool     `json:"-" dbw:"-"`
	Columns []string `json:"-" dbw:"-"`

	// Params holds unmarshalled QueryParams.
	Params map[string]int `dbw:"-"`
}

type InteractiveDatasetHeader struct {
	ID                string `json:"id"`
	Title             string `json:"title"`
	presentationOrder int
	Displays          []string `json:"displays"`
}

type OptionKind int

const (
	Limit    OptionKind = 1
	Offset   OptionKind = 2
	OrderBy  OptionKind = 3
	OrderDir OptionKind = 4
)

type Optioner interface {
	Part() string
	Kind() OptionKind
}

type LimitOption int
type OffsetOption int
type OrderByOption string
type OrderDirOption string

func (o LimitOption) Part() string {
	return " LIMIT " + strconv.Itoa(int(o))
}
func (o LimitOption) Kind() OptionKind {
	return Limit
}

func (o OffsetOption) Part() string {
	return " OFFSET " + strconv.Itoa(int(o))
}

func (o OffsetOption) Kind() OptionKind {
	return Offset
}

func (o OrderByOption) Part() string {
	return " ORDER BY " + string(o)
}

func (o OrderByOption) Kind() OptionKind {
	return OrderBy
}

func (o OrderDirOption) Part() string {
	return string(o)
}

func (o OrderDirOption) Kind() OptionKind {
	return OrderDir
}

type Filter struct {
	basefilter.BaseFilter
	Params []interface{}
}

type DownloadResponse struct {
	FileName    string
	ContentType string
	Content     string // base64
}

type HTMLResponse struct {
}
