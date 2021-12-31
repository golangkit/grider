package grider

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// placeholder holds extracted placeholders from GridColumn.Href
type placeholder struct {
	text           string
	locateToColumn int
}

var linkPrefix string

func SetLinkPrefix(s string) {
	linkPrefix = s
}

func getPlaceholders(gc []Column) map[int][]placeholder {
	res := make(map[int][]placeholder)

	for i := range gc {
		if gc[i].Type != "link" {
			continue
		}

		ru := []rune(gc[i].Href)
		start := -1
		for j := range ru {
			if ru[j] == '{' {
				start = j
				continue
			}
			if ru[j] == '}' {
				if start == -1 {
					continue
				}
				// placeholder text holds {id}
				p := placeholder{text: string(ru[start : j+1]), locateToColumn: -1}
				for k := range gc {
					if gc[k].Name == p.text[1:len(p.text)-1] {
						p.locateToColumn = k
						break
					}
				}
				res[i] = append(res[i], p)
				start = -1
			}
		}
	}

	return res
}

func (r *Grid) Excelize(fname string) (*DownloadResponse, error) {

	f := excelize.NewFile()
	// Create a new sheet.
	sch := "Sheet1"

	pos := make(map[string]int)
	phs := getPlaceholders(r.Columns)

	k := 0
	for i := range r.Columns {
		pos[r.Columns[i].Name] = i
		if r.Columns[i].Hidden {
			continue
		}

		cell, err := excelize.CoordinatesToCellName(k+1, 2)
		if err != nil {
			return nil, errors.New("excel coordinates to cell failed (columns)")
		}
		if err := f.SetCellStr(sch, cell, r.Columns[i].Name); err != nil {
			return nil, err
		}
		k++
	}

	for row := range r.Rows {
		k := 0
		for col := range r.Rows[row] {
			if r.Columns[col].Hidden {
				continue
			}
			cell, err := excelize.CoordinatesToCellName(k+1, (row)+3)
			if err != nil {
				return nil, errors.New("excel coordinates to cell failed (rows)")
			}

			if err := f.SetCellStr(sch, cell, r.Rows[row][col]); err != nil {
				return nil, err
			}

			if r.Columns[col].Type == "link" {
				// 1. get href from columns
				// 2. find to what column it has reference ({xxx})
				// 3. get value from the row named with xxx.
				// 4. replace value
				// 5. if url does not start with http, add domain names from config_params.
				ph, ok := phs[col]
				if ok {

					href := ""
					if !strings.HasPrefix(r.Columns[col].Href, "http") {
						href = linkPrefix + r.Columns[col].Href
					} else {
						href = r.Columns[col].Href
					}
					for j := range ph {
						if ph[j].locateToColumn == -1 {
							fmt.Printf("not found %#v\n", ph[j])
							continue
						}
						href = strings.Replace(href, ph[j].text, r.Rows[row][ph[j].locateToColumn], -1)
					}

					if err := f.SetCellHyperLink(sch, cell, href, "External"); err != nil {
						return nil, err
					}
					// Set underline and font color style for the cell.
					style, err := f.NewStyle(`{"font":{"color":"#1265BE","underline":"single"}}`)
					if err == nil {
						err = f.SetCellStyle(sch, cell, cell, style)
					}
					if err != nil {
						return nil, err
					}
				}
			}
			k++
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	resp := DownloadResponse{
		FileName:    fname,
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Content:     base64.StdEncoding.EncodeToString(buf.Bytes()),
	}
	buf.Reset()
	return &resp, nil
}
