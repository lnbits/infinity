package apps

type Settings struct {
	Models []Model `json:"models"`
}

type Model struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Default  string `json:"default"`
	Ref      string `json:"ref"`
}
