package grid

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/axkit/flogger"
	"github.com/rs/zerolog"
)

// FieldTagLabel holds struct field tag key.
var FieldTagLabel = "grid"

// GridColumn describes grid column's properties.
type GridColumn struct {
	Name       string `json:"name"`
	Hidden     bool   `json:"hidden,omitempty"`
	Sortable   bool   `json:"sortable,omitempty"`
	Filterable bool   `json:"filterable,omitempty"`
	Title      string `json:"title,omitempty"`
	Perm       string `json:"perm,omitempty"`
	Type       string `json:"type,omitempty"`
	Href       string `json:"href,omitempty"`
	Align      string `json:"align,omitempty"`
	Caption    string `json:"caption,omitempty"`
	Method     string `json:"method,omitempty"`
	Icons      string `json:"icons,omitempty"`
	IconsAlign string `json:"ialign,omitempty"`
	Target     string `json:"target,omitempty"` // target for opening link
}

type Grid struct {
	Columns        []GridColumn  `json:"columns"`
	Rows           [][]string    `json:"rows"`
	Objects        []interface{} `json:"objects,omitempty"`
	IsDownloadable bool          `json:"isDownloadable"`
	titlePrefix    string
	log            flogger.FuncLogger
}

type DownloadResponse struct {
	FileName    string
	ContentType string
	Content     string // base64
}

func New(titlePrefix string) *Grid {
	return &Grid{titlePrefix: titlePrefix}
}

func (g *Grid) Logger(l *zerolog.Logger) *Grid {
	g.log = flogger.New(l, "grid", g.titlePrefix)
	return g
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
