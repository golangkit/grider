package grid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"bitbucket.org/elephantsoft/rdk/type/date"
	"gopkg.in/guregu/null.v3"
)

var ffunc map[string]func(interface{}) string

var formats = map[string]string{
	"datehms": "02.01.2006\u200715:04:05",
	"datehm":  "02.01.2006\u200715:04",
	"date":    "02.01.2006",
}

func formatAttribute(src reflect.Value, layout string) string {

	format := layout

	if len(layout) > 0 && layout[0] != '%' {
		format = formats[layout]
	}

	t := src.Type()
	v := src.Interface()

	var res string
	tn := t.String()
	//fmt.Printf("formating=val %#v, layout=%s\n", layout)
	switch tn {
	case "time.Time":
		if len(format) == 0 {
			format = formats["datehm"]
		}
		res = v.(time.Time).Format(format)
	case "date.Date":
		res = (v.(*date.Date)).String()
	case "null.Time":
		t := v.(null.Time)
		if !t.Valid {
			return "-"
		}
		if len(format) == 0 {
			format = formats["datehm"]
		}
		res = t.Time.Format(format)
	default:
		if !src.CanInterface() {
			//res = fmt.Sprintf("%v", v)
			//fmt.Println("cannot interface", res)
			break
		}
		if st, ok := src.Interface().(json.Marshaler); ok {
			buf, err := st.MarshalJSON()
			if err != nil {
				panic(err)
			}
			if buf[0] == byte('"') {
				buf = buf[1 : len(buf)-1]
			}
			if bytes.Compare(buf, []byte(`null`)) == 0 {
				buf = buf[0:0]
			}
			//println("json:", string(buf))
			res = string(buf)
		} else {
			res = fmt.Sprintf("%v", v)
			//fmt.Println("cannot MarshalJSON", res)
		}
	}
	return res
}
