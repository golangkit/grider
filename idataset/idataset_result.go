package idataset

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"

	"bitbucket.org/rtg365/lms/backend/service/ui/grid"
	"gopkg.in/guregu/null.v3"
)

// в нижнем регистре!
type Result struct {
	Columns        []string
	Formats        []string
	Rows           [][]string
	OrderedBy      null.Int
	IsDownloadable bool
}

type ColumnType string

const (
	ColumnTypeText ColumnType = "text"
	ColumnTypeLink ColumnType = "link"
)

type ColumnAlign string

const (
	ColumnAlignLeft  ColumnAlign = "left"
	ColumnAlignRight ColumnAlign = "right"
)

type ColumnFormat struct {
	Title  string      `json:"title"`
	Hidden bool        `json:"hidden"`
	Type   ColumnType  `json:"type"`
	Align  ColumnAlign `json:"align"`
	Href   string      `json:"href"`
	Perm   string      `json:"perm"`
	Target string      `json:"target"`
}

type ParsedResult struct {
	Columns []string
	Formats []ColumnFormat
	Rows    [][]string
}

func (r *Result) Parse() (*ParsedResult, error) {

	var res ParsedResult

	res.Columns = r.Columns
	res.Rows = append(res.Rows, r.Rows...)
	res.Formats = make([]ColumnFormat, len(r.Formats))

	for i := range r.Formats {
		if err := json.Unmarshal([]byte(r.Formats[i]), &res.Formats[i]); err != nil {
			return nil, err
		}
	}
	for i := range res.Formats {
		if res.Formats[i].Type == "" {
			res.Formats[i].Type = ColumnTypeText
		}

		if res.Formats[i].Align == "" {
			res.Formats[i].Align = ColumnAlignLeft
		}

		if res.Formats[i].Type == ColumnTypeLink {
			cn, ok := extractColumnName(res.Formats[i].Href)
			if !ok {
				continue
			}

			found := false
			ci := -1
			for ci = range res.Columns {
				if res.Columns[ci] == cn {
					found = true
					break
				}
			}
			if !found {
				continue
			}

			for ri := range res.Rows {
				if res.Rows[ri][ci] != "" {
					res.Rows[ri][i] = fmt.Sprintf(`<a href="%s">%s</a>`, strings.Replace(res.Formats[i].Href, "{"+cn+"}", res.Rows[ri][ci], 1), res.Rows[ri][i])
				}
			}
		}
	}

	return &res, nil
}

func (r *Result) ToGrid() (*grid.Grid, error) {

	res := grid.Grid{
		Rows:           r.Rows,
		Columns:        make([]grid.GridColumn, len(r.Formats)),
		IsDownloadable: r.IsDownloadable,
	}

	for i := range r.Formats {
		if err := json.Unmarshal([]byte(r.Formats[i]), &res.Columns[i]); err != nil {
			return nil, err
		}
		res.Columns[i].Name = r.Columns[i]

	}
	return &res, nil
}

func extractColumnName(s string) (string, bool) {

	from, to := -1, -1
	for i, c := range []rune(s) {
		if c == '{' && from == -1 {
			from = i
		}
		if from != -1 && c == '}' {
			to = i
			break
		}
	}

	if from < 0 || to < 0 {
		return "", false
	}

	return s[from+1 : to], true

}

func (r *Result) JSON() []byte {

	//fmt.Printf("JSON()= %#v\n", r)

	buf := bytes.NewBuffer(nil)
	buf.WriteString(`{"columns":[`)
	sep := ""
	for i := range r.Columns {
		buf.WriteString(sep)
		buf.WriteString(`{"name":"`)
		buf.WriteString(r.Columns[i]) // name
		buf.WriteString(`"`)

		if len(r.Formats[i]) > 2 { // format
			buf.WriteString(",")
			buf.WriteString(r.Formats[i][1 : len(r.Formats[i])-1])
		}
		buf.WriteString(`}`)
		sep = ","
	}

	buf.WriteString(`],"rows":[`)
	sepr := ""
	for i := range r.Rows {

		buf.WriteString(sepr)
		buf.WriteString(`[`)
		sepc := ""
		for j := range r.Rows[i] {
			buf.WriteString(sepc)
			buf.WriteString(`"`)
			buf.WriteString(r.Rows[i][j])
			buf.WriteString(`"`)
			sepc = ","
		}
		buf.WriteString(`]`)
		sepr = ","
	}
	buf.WriteString(`], "isDownloadable": ` + map[bool]string{false: "false", true: "true"}[r.IsDownloadable] + "}")

	return buf.Bytes()
}

// GenHTML renders HTML using template tmplBody and writes it into dest.
// ifields holds independend fields to be used in the template.
func (r *Result) GenHTML(tmplBody string, ifield map[string]interface{}, dest *bytes.Buffer) error {

	g, err := r.ToGrid()
	if err != nil {
		return err
	}

	g.ReplaceCellWithFullLinks()

	d := struct {
		*grid.Grid
		GeneratedAt string
		Attr        map[string]interface{}
	}{g, time.Now().Format("02.01.2006 15:04 MST"), ifield}

	tmpl, err := template.New("test").Funcs(template.FuncMap{
		"htmlSafe": func(html string) template.HTML {
			return template.HTML(html)
		}}).Parse(tmplBody)

	if err != nil {
		return err
	}

	return tmpl.Execute(dest, &d)
}
