package apps

type Settings struct {
	Models  []Model                `json:"models"`
	On      map[string]interface{} `json:"on"`      // functions
	Actions map[string]interface{} `json:"actions"` // functions
}

func (s Settings) getModel(modelName string) Model {
	for _, m := range s.Models {
		if m.Name == modelName {
			return m
		}
	}
	return Model{}
}

type Model struct {
	Name   string      `json:"name"`
	Fields []Field     `json:"fields"`
	Filter interface{} `json:"filter"` // a function
}

type Field struct {
	Name     string      `json:"name,omitempty"`
	Display  string      `json:"display"`
	Type     string      `json:"type"`
	Required bool        `json:"required,omitempty"`
	Default  interface{} `json:"default,omitempty"`
	Ref      string      `json:"ref,omitempty"`
	Hidden   bool        `json:"hidden,omitempty"`
	Computed interface{} `json:"computed,omitempty"` // a function
}
