package grider

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FieldTagLabel holds struct field tag key.
var FieldTagLabel = "grid"

// Column describes grid column's properties.
type Column struct {
	Name       string `json:"name"`
	Hidden     bool   `json:"hidden,omitempty"`     // default false
	Sortable   bool   `json:"sortable,omitempty"`   // default false
	Filterable bool   `json:"filterable,omitempty"` // default false
	Title      string `json:"title,omitempty"`      // default ""
	Perm       string `json:"perm,omitempty"`       // default not permission
	Type       string `json:"type,omitempty"`       // default "" (regular text)
	Href       string `json:"href,omitempty"`       // default "" (no link)
	Align      string `json:"align,omitempty"`      // default "" ("left") cell align
	Caption    string `json:"caption,omitempty"`    // default ""
	Method     string `json:"method,omitempty"`     // ?
	Icons      string `json:"icons,omitempty"`      // comma separated fa-* icon names
	IconsAlign string `json:"ialign,omitempty"`     // default "" ("left") "right" - after text
	Target     string `json:"target,omitempty"`     // default "" browser window target for opening link
}

// Grid describes data and metadata for presenting grid.
type Grid struct {
	Columns        []Column       `json:"columns"`
	Rows           [][]string     `json:"rows"`
	RowObjects     []interface{}  `json:"rowObjects,omitempty"`
	RowIDs         []int          `json:"rowIds,omitempty"`
	RowActions     [][]ActionCode `json:"rowActions,omitempty"`
	GridActions    []ActionCode   `json:"gridActions,omitempty"`
	Action         ActionSet      `json:"action,omitempty"`
	IsDownloadable bool           `json:"isDownloadable"`
	IsFilterable   bool           `json:"isFilterable"`
	NoPagination   bool           `json:"noPagination,omitempty"`
	PaginationType PaginationType `json:"paginationType"`
	option         Option
}

type DownloadResponse struct {
	FileName    string
	ContentType string
	Content     string // base64
}

type Option struct {
	titlePrefix    string
	isDownloadable bool
	multiLang      bool
}

func WitTitlePrefix(prefix string) func(*Option) {
	return func(s *Option) {
		s.titlePrefix = prefix
	}
}

func WithDownloadOption(b bool) func(*Option) {
	return func(s *Option) {
		s.isDownloadable = b
	}
}

func WithI18n() func(*Option) {
	return func(s *Option) {
		s.multiLang = true
	}
}

func New(opts ...func(*Option)) *Grid {
	g := Grid{}
	for _, f := range opts {
		f(&g.option)
	}
	return &g
}

// DeleteColumns deletes columns with exact names in cols.
func (g *Grid) DeleteColumns(col []string) {

	// finds columns positions.
	idx := make(map[int]struct{}, len(col))
	for i := range col {
		for ci := range g.Columns {
			if g.Columns[ci].Name == col[i] {
				idx[ci] = struct{}{}
			}
		}
	}

	// delete elements from Columns and Rows.
	k := 0
	for i := range g.Columns {
		if _, ok := idx[i]; ok {
			continue
		}
		g.Columns[k] = g.Columns[i]

		for r := range g.Rows {
			g.Rows[r][k] = g.Rows[r][i]
		}
		k++
	}
	g.Columns = g.Columns[:k]
	for r := range g.Rows {
		g.Rows[r] = g.Rows[r][:k]
	}

	return
}

func (g *Grid) JSON() ([]byte, error) {
	return json.Marshal(g)
}

func (g *Grid) ReplaceCellWithFullLinks() error {

	pos := make(map[string]int)
	phs := getPlaceholders(g.Columns)

	for i := range g.Columns {
		pos[g.Columns[i].Name] = i
	}

	for row := range g.Rows {
		for col := range g.Rows[row] {
			if g.Columns[col].Hidden {
				continue
			}

			if g.Columns[col].Type != "link" {
				continue
			}
			// 1. get href from columns
			// 2. find to what column it has reference ({xxx})
			// 3. get value from the row named with xxx.
			// 4. replace value
			// 5. if url does not start with http, add domain names from config_params.
			ph, ok := phs[col]
			if ok {
				href := ""
				if !strings.HasPrefix(g.Columns[col].Href, "http") {
					href = "<a href=\"" + linkPrefix + g.Columns[col].Href + "\">" + g.Rows[row][col] + "</a>"
				} else {
					href = "<a href=\"" + g.Columns[col].Href + "\">" + g.Rows[row][col] + "</a>"
				}
				for j := range ph {
					if ph[j].locateToColumn == -1 {
						return fmt.Errorf("invalid placeholder %s", ph[j].text)
					}

					// replaces placeholder {col} with text
					href = strings.Replace(href, ph[j].text, g.Rows[row][ph[j].locateToColumn], -1)
					g.Rows[row][col] = href
				}
			}
		}
	}
	return nil
}

func (g *Grid) AssignActionSet(as ActionSet) error {
	g.Action = NewActionSet()
	g.Action.Add(g.GridActions)
	for i := range g.RowActions {
		g.Action.Add(g.RowActions[i])
	}

	return g.Action.AssignActionValues(as)
}

type PaginationType int

const (
	PaginationServer  PaginationType = 0
	PaginationClient  PaginationType = 1
	PaginationWithout PaginationType = 2
)

func (pt PaginationType) String() string {
	switch pt {
	case 0:
		return "server"
	case 1:
		return "client"
	case 2:
		return "without"
	}
	return ""
}

func (pt PaginationType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + pt.String() + `"`), nil
}
