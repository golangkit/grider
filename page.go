package grider

// Page:Header
// 		[*Tab1:Header] [Tab2:Header]
// 			Tab1:Widget		Tab1:Widget
// 				Line1			LineA
// 				Line2			LineB
//

// Page describes a single object page.
type Page struct {
	Header *Header `json:"header,omitempty"`

	Widgets []Widgeter `json:"widgets,omitempty"`

	// Tabs содержит описание подчиненных объектов.
	Tabs []Tab `json:"tabs,omitempty"`

	Action ActionSet `json:"action,omitempty"`

	// Footer описывает содержимое нижней части окна.
	//Footer *Footer `json:"footer,omitempty"`
}

// Tab описывает содержимое одного связанного объекта.
type Tab struct {
	Header *Header `json:"header,omitempty"`

	Widgets []Widgeter `json:"widgets,omitempty"`

	IsActive       bool `json:"isActive,omitempty"`
	IsInitRequired bool `json:"isInitRequired,omitempty"`
	IsDisabled     bool `json:"isDisabled,omitempty"`
}

type WidgetType int

const (
	AttrValueType WidgetType = 1
	MediaType     WidgetType = 2
	MapType       WidgetType = 3
	ChartType     WidgetType = 4
)

type Widgeter interface {
	WidgetType() WidgetType
}

type WidgetPos struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type AttrValueWidget struct {
	WidgetPos
	Lines []Line `json:"lines,omitempty"`
}

func (AttrValueWidget) WidgetType() WidgetType {
	return AttrValueType
}

type MediaWidget struct {
	WidgetPos
	Media []Media `json:"media,omitempty"`
}

func (MediaWidget) WidgetType() WidgetType {
	return MediaType
}

// Header describes header of the Page or Tab.
type Header struct {
	ID int `json:"id"`

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
	BgColor string `bgColor,omitempty`
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
	Icon  *Icon    `json:"icon,omitempty"`
	Label string   `json:"label,omitempty"`
	Value string   `json:"value,omitempty"`
	Type  LineType `json:"type,omitempty"`
	// RefBook заполняется только если Type = "refbook"
	RefBook    *RefBookType    `json:"refBook,omitempty"`
	Suggestion *SuggestionType `json:"suggestion,omitempty"`

	// URL заполняется при Type = href или exthref
	URL     string   `json:"url,omitempty"`
	Actions []Action `json:"actions,omitempty"`
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
