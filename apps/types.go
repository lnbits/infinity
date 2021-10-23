package apps

import (
	"fmt"
	"reflect"

	models "github.com/lnbits/lnbits/models"
)

type Settings struct {
	Code     string                 `json:"code"`
	Models   []Model                `json:"models"`
	Triggers map[string]interface{} `json:"triggers"` // functions
	Actions  map[string]interface{} `json:"actions"`  // functions
	Files    map[string]string      `json:"files"`
}

func (s Settings) getModel(modelName string) Model {
	for _, m := range s.Models {
		if m.Name == modelName {
			return m
		}
	}
	return Model{}
}

func (s Settings) validate() error {
	validTypes := []string{
		"string",
		"number",
		"boolean",
		"ref",
		"msatoshi",
		"url",
	}

	if len(s.Models) == 0 {
		return fmt.Errorf("no models declared")
	}

	for m, model := range s.Models {
		if model.Name == "" {
			return fmt.Errorf("models[%d].name is not provided", m)
		}

		if len(model.Fields) == 0 {
			return fmt.Errorf("model %s has no fields", model.Name)
		}

		for f, field := range model.Fields {
			if field.Name == "" {
				return fmt.Errorf("models[%d].fields[%d].name is not provided", m, f)
			}

			typeIsValid := false
			for _, validType := range validTypes {
				if validType == field.Type {
					typeIsValid = true
					break
				}
			}
			if typeIsValid == false {
				return fmt.Errorf("%s.%s's type cannot be '%s', must be one of %v",
					model.Name, field.Name, field.Type, validTypes)
			}

			if field.Type == "ref" {
				if field.Ref == "" {
					return fmt.Errorf("%s.%s has type='ref', but ref is not provided", model.Name, field.Name)
				}

				if field.As == "" {
					return fmt.Errorf("%s.%s's as not provided, must be the name of a property in the referred model", model.Name, field.Name)
				}

				// check if referred model exists
				refExists := false
				for _, refModel := range s.Models {
					if field.Ref == refModel.Name {
						// check if the "as" property refers to a field
						// on the referred model that does exist
						asFieldExistsAsRefModelField := false
						for _, refModelField := range refModel.Fields {
							if refModelField.Name == field.As {
								asFieldExistsAsRefModelField = true
								break
							}
						}
						if asFieldExistsAsRefModelField == false {
							return fmt.Errorf("%s.%s's field as='%s', but model '%s' doesn't have a field '%s'", model.Name, field.Name, field.As, refModel.Name, field.As)
						}

						refExists = true
						break
					}
				}
				if refExists == false {
					return fmt.Errorf("%s.%s's ref '%s' doesn't exist as a model",
						model.Name, field.Name, field.Ref)
				}
			}
		}
	}

	return nil
}

type Model struct {
	Name    string      `json:"name"`
	Display string      `json:"display,omitempty"`
	Plural  string      `json:"plural,omitempty"`
	Fields  []Field     `json:"fields"`
	Filter  interface{} `json:"filter,omitempty"` // in lua this is a function, here we just check for its presence
}

func (m Model) validateItem(item models.AppDataItem) error {
	if len(m.Fields) == 0 {
		return fmt.Errorf("unknown model")
	}

	if item.Value == nil {
		return fmt.Errorf("empty item value")
	}

	for fieldName, fieldValue := range item.Value {
		fieldExpected := false
		for _, field := range m.Fields {
			if field.Computed != nil {
				// we don't expected computed fields
				continue
			}

			if field.Name == fieldName {
				fieldExpected = true

				fieldValueType := reflect.TypeOf(fieldValue)
				if fieldValueType == nil {
					return fmt.Errorf("%s=%v has unexpected type %v",
						field.Name, fieldValue, fieldValueType)
				}

				switch field.Type {
				case "string", "url":
					if fieldValueType.Name() != "string" {
						return fmt.Errorf("%s=%v is not a string", field.Name, fieldValue)
					}
				case "number":
					if fieldValueType.Name() != "float64" {
						return fmt.Errorf("%s=%v is not a number", field.Name, fieldValue)
					}
				case "msatoshi":
					if fieldValueType.Name() != "float64" {
						return fmt.Errorf("%s=%v is not a number", field.Name, fieldValue)
					}
					msat := int64(fieldValue.(float64))
					if float64(msat) != fieldValue.(float64) {
						return fmt.Errorf(
							"%s=%v is not an integer, msatoshi must be integer",
							field.Name, fieldValue,
						)
					}
					if msat > 100000000000 {
						return fmt.Errorf("%s=%v is way too many satoshis",
							field.Name, fieldValue)
					}
				case "boolean":
					if fieldValueType.Name() != "bool" {
						return fmt.Errorf("%s=%v is not a boolean", field.Name, fieldValue)
					}
				case "ref":
					if fieldValueType.Name() != "string" {
						return fmt.Errorf("%s=%v is not a ref string",
							field.Name, fieldValue)
					}
					ref, err := DBGet(
						item.WalletID, item.App, field.Ref, fieldValue.(string))
					if err != nil || ref == nil {
						return fmt.Errorf("%s=%v is not a valid ref",
							field.Name, fieldValue)
					}
				}

				break
			}
		}
		if fieldExpected == false {
			return fmt.Errorf("unexpected field '%s'", fieldName)
		}
	}

	return nil
}

type Field struct {
	Name     string      `json:"name"`
	Display  string      `json:"display,omitempty"`
	Type     string      `json:"type"`
	Required bool        `json:"required,omitempty"`
	Default  interface{} `json:"default,omitempty"`
	Ref      string      `json:"ref,omitempty"`
	As       string      `json:"as,omitempty"`
	Computed interface{} `json:"computed,omitempty"` // lua function, like above
}
