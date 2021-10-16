package apps

type Settings struct {
	Models   []Model                `json:"models"`
	Triggers map[string]interface{} `json:"triggers"` // functions
	Actions  map[string]interface{} `json:"actions"`  // functions
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
	Filter interface{} `json:"filter"` // in lua this is a function, just check for presence
}

type Field struct {
	Name     string      `json:"name,omitempty" valid:"required"`
	Display  string      `json:"display"`
	Type     string      `json:"type" valid:"in(string|number|boolean|url)"`
	Required bool        `json:"required,omitempty"`
	Default  interface{} `json:"default,omitempty"`
	Ref      string      `json:"ref,omitempty"`
	Hidden   bool        `json:"hidden,omitempty"`
	Computed interface{} `json:"computed,omitempty"` // lua function, like above
}
