package grider

import (
	"reflect"
	"regexp"
	"strings"
)

// ApplySliceOfStruct converts slice of any struct to Grid,
// slice of column and rows.
func (g *Grid) ApplySliceOfStruct(src interface{}) *Grid {

	s := reflect.ValueOf(src)
	t := s.Type()

	if t.Kind() != reflect.Slice {
		panic("Convert's parameter src expected to be a slice")
	}

	if s.Len() == 0 {
		// if src empty we have to create empty slice element.
		// and generate values for Columns attribute.
		g.Columns = extractMeta(g.titlePrefix, "", reflect.Zero(t.Elem()))
		return g
	}

	//	fmt.Printf("s.Len()=%d\n", s.Len())
	for i := 0; i < s.Len(); i++ {
		row := s.Index(i)
		if i == 0 {
			g.Columns = extractMeta(g.titlePrefix, "", row)
		}
		g.Rows = append(g.Rows, convertStructValues(row))
		//	fmt.Printf("dst=%v\n", res.Rows)

		ofunc := row.Addr().MethodByName("Object")
		//fmt.Printf("ofunc=%#v\n", ofunc)
		if ofunc.IsValid() && !ofunc.IsZero() {
			obj := ofunc.Call([]reflect.Value{})
			g.Objects = append(g.Objects, obj[0].Interface())
		}

		// aif := row.Addr().MethodByName("Icon")
		// //fmt.Printf("ofunc=%#v\n", ofunc)
		// if aif.IsValid() && !aif.IsZero() {
		// 	icon := aif.Call([]reflect.Value{})
		// 	res.Icons = append(res.Icons, icon[0].Interface())
		// }
	}

	return g
}

func convertStructValues(s reflect.Value) []string {
	//println("excludeTag", excludeTag)
	//s := reflect.ValueOf(model).Elem()
	t := s.Type()
	if t.Kind() != reflect.Struct {
		panic("convertStructValues's parameter src expected to be a struct")
	}

	var res []string

	for i := 0; i < s.NumField(); i++ {
		sf := s.Field(i)
		tf := t.Field(i)

		// ignore private fields.

		if tf.Name[0] >= 'a' && tf.Name[0] <= 'z' {
			//if sf.CanAddr() == false {
			continue
		}
		tag := tf.Tag.Get(FieldTagLabel)
		//println("fieldName=", tf.Name, "tag", tag)
		if tag == "-" {
			continue
		}

		if tf.Type.Name() == "" || tf.Anonymous {
			if tf.Type.Kind() != reflect.Ptr {
				res = append(res, convertStructValues(sf)...)
			} else {
				if sf.IsNil() {
					sf = reflect.New(tf.Type.Elem())
				}
				res = append(res, convertStructValues(sf)...)
			}
			continue
		}

		//if strings.HasPrefix(tf.Type.String(), "int") || strings.HasPrefix(tf.Type.String(), "uint") {

		//}
		if tf.Type.Kind() == reflect.Ptr && sf.IsNil() {
			res = append(res, "")
			continue
		}

		//res = append(res, fmt.Sprintf("no json %v", sf.Interface()))
		res = append(res, formatAttribute(sf, extractTagAttr(tag, "fmt")))
	}
	return res
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// ToSnakeCase converts string like RobertEgorov to robert_egorov.
func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

/*
func convertTagToJSON(prefix string, tag string) string {
	if tag == "" {
		return "{}"
	}

	res := "{"
	tags := strings.Split(tag, ",")
	sep := ""
	for i := range tags {
		lay := "%s\"%s\":\"%s\""

		k := strings.Split(tags[i], "=")
		if len(k) < 2 {
			panic("wrong tag" + tag)
		}
		k[0] = strings.TrimSpace(k[0])
		k[1] = strings.TrimSpace(k[1])

		switch k[0] {
		case "fmt":
			continue
		case "href":
			k[1] = addPrefixToColumn(k[1], prefix)
		case "hidden", "sortable":
			lay = "%s\"%s\":%s"
		}
		res += fmt.Sprintf(lay, sep, k[0], k[1])
		sep = ","
	}
	res += "}"

	//println("tag=", res)

	return res
}
*/
func extractTagAttr(s, substr string) string {

	k := len(substr)
	if k == 0 {
		return ""
	}
	substr += "="
	k++

	pos := strings.Index(s, substr)
	if pos < 0 {
		return ""
	}

	qpos := strings.Index(s[pos+k:], ",")
	if qpos < 0 {
		if len(s[pos+k:]) > 1 {
			return s[pos+k : len(s)-1]
		}
		panic("struct field tag has attr 'fmt' without value")
	}

	return s[pos+k : pos+k+qpos]
}

func addPrefixToColumn(s, substr string) string {

	pos := strings.Index(s, "{")
	if pos > 0 {
		return s[0:pos+1] + substr + s[pos+1:]
	}
	return s
}

func extractMeta(titlePrefix string, parentAttribute string, s reflect.Value) []GridColumn {

	var res []GridColumn

	//s := reflect.ValueOf(model).Elem()
	t := s.Type()

	if t.Kind() != reflect.Struct {
		panic("extractMeta's parameter expected to be a struct")
	}

	for i := 0; i < t.NumField(); i++ {

		sf := s.Field(i)
		tf := t.Field(i)

		// ignore private fields.
		if tf.Name[0] >= 'a' && tf.Name[0] <= 'z' {
			//if sf.CanAddr() == false {
			continue
		}
		tag := tf.Tag.Get(FieldTagLabel)
		if tag == "-" {
			continue
		}
		//	println(tf.Name, tf.Tag, tf.Anonymous, tf.Type.Name(), "; tag=", tag)

		//snakeName := ToSnakeCase(tf.Name)
		snakeName := tf.Name

		if tf.Type.Name() == "" || tf.Anonymous {
			//fmt.Println("struct with no type")
			var gc []GridColumn
			if tf.Type.Kind() != reflect.Ptr {
				gc = extractMeta(titlePrefix, snakeName, sf)
			} else {
				if sf.IsNil() {
					mock := reflect.New(tf.Type.Elem())
					gc = extractMeta(titlePrefix, snakeName, mock)
				} else {
					gc = extractMeta(titlePrefix, snakeName, sf)
				}
			}
			//fmt.Printf("anonym: %v\n", h)
			if len(gc) > 0 {
				res = append(res, gc...)
				continue
			}
		} else {
			res = append(res, convertTagToGridColumn(titlePrefix, parentAttribute, snakeName, tag))
			continue
		}

	}
	return res
}
func joinAttributeNames(parentAttribute, attribute string) string {
	if parentAttribute == "" {
		return attribute
	}
	return parentAttribute + attribute
}

func convertTagToGridColumn(titlePrefix string, parentAttribute, attribute string, tag string) GridColumn {

	var res = GridColumn{Name: joinAttributeNames(parentAttribute, attribute)}
	res.Title = "%" + titlePrefix + res.Name + "%"

	if tag == "" {
		return res
	}

	tags := strings.Split(tag, ",")
	for i := range tags {
		k := strings.Split(tags[i], "=")
		if len(k) < 2 {
			panic("wrong tag" + tag)
		}
		k[0] = strings.TrimSpace(k[0])
		k[1] = strings.TrimSpace(k[1])

		switch k[0] {
		case "type":
			res.Type = k[1]
		case "align":
			res.Align = k[1]
		case "href":
			res.Href = k[1]
		case "hidden":
			res.Hidden = (k[1] == "true")
		case "sortable":
			res.Sortable = (k[1] == "true")
		case "filterable":
			res.Filterable = (k[1] == "true")
		case "perm":
			res.Perm = k[1]
		case "caption":
			res.Caption = k[1]
		case "method":
			res.Method = k[1]
		case "icons":
			res.Icons = k[1]
		case "ialign":
			res.IconsAlign = k[1]
		case "target":
			res.Target = k[1]
		}
	}

	//	fmt.Printf("res=%#v\n", res)

	return res
}
