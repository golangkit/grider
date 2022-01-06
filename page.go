package grider

// Page:Header
// 		[*Tab1:Header] [Tab2:Header]
// 			Tab1:Widget		Tab1:Widget
// 				Line1			LineA
// 				Line2			LineB
//

// Page describes a single object page.
type Page struct {
	ID     int     `json:"id,omitempty"`
	Header *Header `json:"header,omitempty"`

	Widgets []Widgeter `json:"widgets,omitempty"`

	// Tabs содержит описание подчиненных объектов.
	Tabs []Tab `json:"tabs,omitempty"`

	Action ActionSet `json:"action,omitempty"`

	// PageActions holds actions available in the drop down list on the page level.
	PageActions []ActionCode `json:"pageActions,omitempty"`

	// Footer описывает содержимое нижней части окна.
	//Footer *Footer `json:"footer,omitempty"`
}

// Tab описывает содержимое одного связанного объекта.
type Tab struct {
	Header     *Header      `json:"header,omitempty"`
	TabActions []ActionCode `json:"tabActions,omitempty"`

	Widgets []Widgeter `json:"widgets,omitempty"`

	IsActive       bool `json:"isActive,omitempty"`
	IsInitRequired bool `json:"isInitRequired,omitempty"`
	IsDisabled     bool `json:"isDisabled,omitempty"`
}

type WidgetType int

const (
	AttrValueType WidgetType = 1
	MediaType     WidgetType = 2
	GridType      WidgetType = 3
	MapType       WidgetType = 4
	ChartType     WidgetType = 5
	CustomType    WidgetType = 6
	LazyType      WidgetType = 7
	ContentType   WidgetType = 8
)

func (wt WidgetType) String() string {
	switch wt {
	case AttrValueType:
		return "attrval"
	case MediaType:
		return "media"
	case MapType:
		return "map"
	case ChartType:
		return "chart"
	case GridType:
		return "grid"
	case CustomType:
		return "custom"
	case LazyType:
		return "lazy"
	case ContentType:
		return "content"
	}
	return ""
}

func (wt WidgetType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + wt.String() + `"`), nil
}

type ContentBodyType int

const (
	Text     ContentBodyType = 1
	Html     ContentBodyType = 2
	Markdown ContentBodyType = 3
)

func (wt ContentBodyType) String() string {
	switch wt {
	case Text:
		return "text"
	case Html:
		return "html"
	case Markdown:
		return "markdown"
	}
	return ""
}

func (wt ContentBodyType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + wt.String() + `"`), nil
}

type Widgeter interface {
	WidgetType() WidgetType
}

type Widget struct {
	ID     int        `json:"id,omitempty"`
	Type   WidgetType `json:"type"`
	Header *Header    `json:"header,omitempty"`
	// Row     int          `json:"row"`
	// Col     int          `json:"col"`
	Width   int          `json:"width"`
	Actions []ActionCode `json:"widgetActions,omitempty"`

	// Action must be only filled if widget generated independent as
	// result of request from lazy widget.
	Action ActionSet `json:"action,omitempty"`

	// Object is an object in JSON. The object can be consumed
	// by customized UI logic. As instance: take data to init modal window.
	Object interface{} `json:"object,omitempty"`
}

type AttrValueWidget struct {
	*Widget
	Lines []Line `json:"lines,omitempty"`
}

func (AttrValueWidget) WidgetType() WidgetType {
	return AttrValueType
}

type ContentWidget struct {
	*Widget
	Type ContentBodyType `json:"type"`
	Body string          `json:"body"`
}

func (ContentWidget) WidgetType() WidgetType {
	return ContentType
}

type LazyWidget struct {
	*Widget
	URL string `json:"url"`
}

func (LazyWidget) WidgetType() WidgetType {
	return LazyType
}

type MediaWidget struct {
	*Widget
	Media []Media `json:"media,omitempty"`
}

func (MediaWidget) WidgetType() WidgetType {
	return MediaType
}

type GridWidget struct {
	*Widget
	Grid *Grid `json:"grid,omitempty"`
}

func (GridWidget) WidgetType() WidgetType {
	return GridType
}

// Header describes header of the Page or Tab.
type Header struct {
	ID int `json:"id,omitempty"`

	// Icon параметры иконки отображемой в заголовке.
	LeftIcons []Icon `json:"leftIcons,omitempty"`

	// Title может содержать код из ресурсов %слово%
	Title string `json:"title,omitempty"`

	// SubTitle может содержать код из ресурсов %слово%
	SubTitle string `json:"subTitle,omitempty"`

	RightIcons []Icon `json:"rightIcons,omitempty"`

	// URL для перехода на страницу объекта, если значение заполнено.
	URL string `json:"url,omitempty"`

	// BgColor содержит цвета фона заголовка в формате HTML: red, #fff или #fefefe.
	BgColor string `json:"bgColor,omitempty"`
}

// Icon describes fa-icon properties.
type Icon struct {
	// Name is name of font-awesome icon.
	Name string `json:"name"`

	// Color in HTML format.
	Color string `json:"color,omitempty"`
}

// Media описывает фото или видео для отображения в окне.
type Media struct {
	// ThumbnailURL содержит URL который необходимо использовать для отображения
	// миниатюр.
	ThumbnailURL string `json:"thumbnailUrl"`

	// URL содержит адрес полной фотографии/видео.
	URL string `json:"url"`

	// IsVideo = true, если это видео. В миниатюре будет фото.
	IsVideo bool `json:"isVideo,omitempty"`
}

// Line описывает одну информационную строчку
type Line struct {
	ID    int      `json:"id,omitempty"`
	Icon  *Icon    `json:"icon,omitempty"`
	Label string   `json:"label,omitempty"`
	Value string   `json:"value,omitempty"`
	Type  LineType `json:"type,omitempty"`
	// RefBook заполняется только если Type = "refbook"
	RefBook    *RefBookType    `json:"refBook,omitempty"`
	Suggestion *SuggestionType `json:"suggestion,omitempty"`

	// URL заполняется при Type = href или exthref
	URL     string       `json:"url,omitempty"`
	Actions []ActionCode `json:"actions,omitempty"`
}

// LineType описывает поддерживаемые типы линий.
type LineType string

const (
	// LineTypeDefault обычная строка. Показывать как текст
	LineTypeDefault LineType = ""

	// LineTypeHref показывать Value как ссылку
	LineTypeHref LineType = "href"

	// LineTypeExtHref показывать Value как ссылку на внешний ресурс.
	LineTypeExtHref LineType = "exthref"

	// LineTypeRefbook отображать Value как ссылку активирующую режим редактирования
	// параметры отображения комбобокса в структуре RefBookType.
	LineTypeRefbook LineType = "refbook"

	// LineTypeSuggestion отображать Value как ссылку активирующую режим редактирования
	// параметры отображения комбобокса в структуре RefBookType.
	LineTypeSuggestion LineType = "suggestion"
)

// RefBookType описывает параметры строчки которая является изменяемым элементом справочника.
type RefBookType struct {
	// Name содержит название справочника из /dictionary
	Name string `json:"name"`
	// SelectedID текущий ID элемента справочника.
	SelectedID int `json:"selectedId"`

	// SubmitURL адрес куда надо направить POST запрос содержащий новый выбранный id из справочника Name.
	// Структура запроса {"id" : 2}
	SubmitURL string `json:"sumbitUrl"`
}

type SuggestionType struct {
	// Name содержит название справочника из /dictionary
	Name string `json:"name"`
	// SelectedID текущий ID элемента справочника.
	SelectedID *int64 `json:"selectedId,omitempty"`

	UID string `json:"uid,omitempty"`

	// SubmitURL адрес куда надо направить POST запрос содержащий новый выбранный id из справочника Name.
	// Структура запроса {"id" : 2}
	SubmitURL string `json:"submitUrl"`
}

// Footer описывает содержимое нижней части окна.
// type Footer struct {
// 	Media []Media `json:"media,omitempty"`
// }
func (p *Page) AssignActionSet(supported ActionSet) error {
	p.Action = NewActionSet()
	p.Action.Add(p.PageActions)
	for i := range p.Tabs {
		p.Action.Add(p.Tabs[i].TabActions)
		p.assignActionCode(p.Tabs[i].Widgets)
	}

	p.assignActionCode(p.Widgets)

	return p.Action.AssignActionValues(supported)
}

func (p *Page) assignActionCode(ws []Widgeter) {
	for i := range ws {
		switch ws[i].WidgetType() {
		case AttrValueType:
			w := ws[i].(AttrValueWidget)
			p.Action.Add(w.Actions)
			for j := range w.Lines {
				p.Action.Add(w.Lines[j].Actions)
			}
		case MediaType:
			w := ws[i].(MediaWidget)
			p.Action.Add(w.Actions)
		case MapType:
			break
		case ChartType:
			break
		case ContentType:
			w := ws[i].(ContentWidget)
			p.Action.Add(w.Actions)
			break
		case CustomType:
			break
		case GridType:
			g := ws[i].(GridWidget)
			p.Action.Add(g.Grid.GridActions)
			for j := range g.Grid.RowActions {
				p.Action.Add(g.Grid.RowActions[j])
			}
		}
	}
}

func AssignActionSet(lw Widgeter, as ActionSet) error {
	var err error
	switch lw.WidgetType() {
	case AttrValueType:
		w := lw.(AttrValueWidget)
		w.Action = NewActionSet()
		w.Action.Add(w.Actions)
		for j := range w.Lines {
			w.Action.Add(w.Lines[j].Actions)
		}
		err = w.Action.AssignActionValues(as)
	case MediaType:
		w := lw.(MediaWidget)
		w.Action = NewActionSet()
		w.Action.Add(w.Actions)
		err = w.Action.AssignActionValues(as)
	case MapType:
		break
	case ChartType:
		break
	case CustomType:
		break
	case ContentType:
		w := lw.(ContentWidget)
		w.Action = NewActionSet()
		w.Action.Add(w.Actions)
		err = w.Action.AssignActionValues(as)
		break
	case GridType:
		w := lw.(GridWidget)
		w.Action = NewActionSet()
		w.Action.Add(w.Actions)
		w.Action.Add(w.Grid.GridActions)
		for i := range w.Grid.RowActions {
			w.Action.Add(w.Grid.RowActions[i])
		}
		err = w.Action.AssignActionValues(as)
	}
	return err
}
