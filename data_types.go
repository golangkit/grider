package grider

import (
	"fmt"
	"strconv"
	"time"

	"github.com/axkit/date"
	"gopkg.in/guregu/null.v3"
)

type NullTime null.Time
type Time time.Time
type Int null.Int
type String null.String
type Float null.Float
type Date date.Date

type Formatter interface {
	ConvertToString(layout string) string
}

func (t Time) ConvertToString(layout string) string {
	d := time.Time(t)
	if !d.IsZero() {
		return ""
	}
	return d.Format(layout)
}

func (t NullTime) ConvertToString(layout string) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(layout)
}

func (t Int) ConvertToString(layout string) string {
	if !t.Valid {
		return ""
	}
	return strconv.FormatInt(t.Int64, 10)
}

func (t Float) ConvertToString(layout string) string {
	if !t.Valid {
		return ""
	}
	return fmt.Sprintf(layout, t.Float64)
}

func (t String) ConvertToString(layout string) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func (t Date) ConvertToString(layout string) string {

	d := date.Date(t)
	if !d.Valid() {
		return ""
	}

	if layout == "" {
		layout = "02.01.2006"
	}

	return d.Format(layout)
}
