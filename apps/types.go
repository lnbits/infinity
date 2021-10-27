package apps

import (
	"fmt"
	"reflect"

	"github.com/fiatjaf/lunatico"
	models "github.com/lnbits/lnbits/models"
)

type Settings struct {
	Code     string                           `json:"code"`
	Models   []Model                          `json:"models"`
	Triggers map[string]*lunatico.LuaFunction `json:"triggers"`
	Actions  map[string]Action                `json:"actions"`
	Files    map[string]string                `json:"files"`
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
	if len(s.Models) == 0 {
		return fmt.Errorf("no models declared")
	}

	for m, model := range s.Models {
		if model.Name == "" {
			return fmt.Errorf("models[%d].name is not provided", m)
		}

		if !nameValidator.MatchString(model.Name) {
			return fmt.Errorf("models[%d].name=%s is invalid", m, model.Name)
		}

		if len(model.Fields) == 0 {
			return fmt.Errorf("model %s has no fields", model.Name)
		}

		for f, field := range model.Fields {
			if err := field.validate(s.Models); err != nil {
				return fmt.Errorf("model %s.fields[%d] validation error: %w",
					model.Name, f, err)
			}
		}
	}

	for name, def := range s.Actions {
		if !nameValidator.MatchString(name) {
			return fmt.Errorf("action name '%s' is invalid", name)
		}

		if def.Fields == nil {
			return fmt.Errorf("action %s must have a .fields array, even if empty", name)
		}

		for f, field := range def.Fields {
			if err := field.validate(nil); err != nil {
				return fmt.Errorf("action %s.fields[%d] validation error: %w",
					name, f, err)
			}
		}

		if def.Handler == nil {
			return fmt.Errorf("action %s must have a handler function", name)
		}

	}

	return nil
}

type Model struct {
	Name    string  `json:"name"`
	Display string  `json:"display,omitempty"`
	Plural  string  `json:"plural,omitempty"`
	Fields  []Field `json:"fields"`
}

func (m Model) validateItem(item models.AppDataItem) error {
	if len(m.Fields) == 0 {
		return fmt.Errorf("unknown model")
	}

	if item.Value == nil {
		return fmt.Errorf("empty item value")
	}

	missingRequired := make(map[string]bool)

	for fieldName, fieldValue := range item.Value {
		fieldExpected := false
		for _, field := range m.Fields {
			if field.Computed != nil {
				// we don't expected computed fields
				continue
			}

			if field.Required {
				if _, seenYet := missingRequired[field.Name]; !seenYet {
					missingRequired[field.Name] = true
				}
			}

			if field.Name == fieldName {
				fieldExpected = true
				if err := field.validateValue(fieldValue,
					item.WalletID, item.App); err != nil {
					return err
				}

				missingRequired[field.Name] = false

				break
			}
		}
		if fieldExpected == false {
			return fmt.Errorf("unexpected field '%s'", fieldName)
		}
	}

	for fieldName, missing := range missingRequired {
		if missing {
			return fmt.Errorf("field '%s' is required, but wasn't provided", fieldName)
		}
	}

	return nil
}

type Action struct {
	Handler *lunatico.LuaFunction `json:"handler"`
	Fields  []Field               `json:"fields"`
}

func (action Action) validateParams(params map[string]interface{}) error {
	missingRequired := make(map[string]bool)

	for fieldName, fieldValue := range params {
		fieldExpected := false

		for _, field := range action.Fields {
			if field.Computed != nil {
				// we don't expected computed fields
				continue
			}
			if field.Type == "ref" {
				// we don't expected ref fields
				continue
			}
			if field.Required {
				if _, seenYet := missingRequired[field.Name]; !seenYet {
					missingRequired[field.Name] = true
				}
			}

			if field.Name == fieldName {
				fieldExpected = true
				if err := field.validateValue(fieldValue, "", ""); err != nil {
					return err
				}

				missingRequired[field.Name] = false
				break
			}
		}
		if fieldExpected == false {
			return fmt.Errorf("unexpected field '%s'", fieldName)
		}
	}

	for fieldName, missing := range missingRequired {
		if missing {
			return fmt.Errorf("field '%s' is required, but wasn't provided", fieldName)
		}
	}

	return nil
}

type Field struct {
	Name     string                `json:"name"`
	Display  string                `json:"display,omitempty"`
	Type     string                `json:"type"`
	Required bool                  `json:"required,omitempty"`
	Default  interface{}           `json:"default,omitempty"`
	Ref      string                `json:"ref,omitempty"`
	As       string                `json:"as,omitempty"`
	Computed *lunatico.LuaFunction `json:"computed,omitempty"`
}

var validTypes = []string{
	"string",
	"number",
	"boolean",
	"ref",
	"msatoshi",
	"url",
}

func (field Field) validate(models []Model) error {
	if field.Name == "" {
		return fmt.Errorf("name is not provided")
	}

	if !nameValidator.MatchString(field.Name) {
		return fmt.Errorf("name '%s' is invalid", field.Name)
	}

	typeIsValid := false
	for _, validType := range validTypes {
		if validType == field.Type {
			typeIsValid = true
			break
		}
	}
	if typeIsValid == false {
		return fmt.Errorf("%s's type cannot be '%s', must be one of %v",
			field.Name, field.Type, validTypes)
	}

	if field.Type == "ref" {
		if models == nil {
			return fmt.Errorf("%s has type='ref', but we don't accept ref types in this context", field.Name)
		}

		if field.Ref == "" {
			return fmt.Errorf("%s has type='ref', but ref is not provided", field.Name)
		}

		if field.As == "" {
			return fmt.Errorf("%s's as not provided, must be the name of a property in the referred model", field.Name)
		}

		// check if referred model exists
		refExists := false
		for _, refModel := range models {
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
					return fmt.Errorf("%s's field as='%s', but model '%s' doesn't have a field '%s'", field.Name, field.As, refModel.Name, field.As)
				}

				refExists = true
				break
			}
		}
		if refExists == false {
			return fmt.Errorf("%s's ref '%s' doesn't exist as a model",
				field.Name, field.Ref)
		}
	}

	return nil
}

func (field Field) validateValue(value interface{}, walletID, app string) error {
	valueType := reflect.TypeOf(value)
	if valueType == nil {
		return fmt.Errorf("%s=%v has unexpected type %v",
			field.Name, value, valueType)
	}

	switch field.Type {
	case "string", "url":
		if valueType.Name() != "string" {
			return fmt.Errorf("%s=%v is not a string", field.Name, value)
		}
	case "number":
		if valueType.Name() != "float64" {
			return fmt.Errorf("%s=%v is not a number", field.Name, value)
		}
	case "msatoshi":
		if valueType.Name() != "float64" {
			return fmt.Errorf("%s=%v is not a number", field.Name, value)
		}
		msat := int64(value.(float64))
		if float64(msat) != value.(float64) {
			return fmt.Errorf(
				"%s=%v is not an integer, msatoshi must be integer",
				field.Name, value,
			)
		}
		if msat > 100000000000 {
			return fmt.Errorf("%s=%v is way too many satoshis",
				field.Name, value)
		}
	case "boolean":
		if valueType.Name() != "bool" {
			return fmt.Errorf("%s=%v is not a boolean", field.Name, value)
		}
	case "ref":
		if valueType.Name() != "string" {
			return fmt.Errorf("%s=%v is not a ref string",
				field.Name, value)
		}
		if walletID == "" {
			return fmt.Errorf("%s=%v is is a ref but we don't accept refs here",
				field.Name, value)
		} else {
			ref, err := DBGet(
				walletID, app, field.Ref, value.(string))
			if err != nil || ref == nil {
				return fmt.Errorf("%s=%v is not a valid ref",
					field.Name, value)
			}
		}
	default:
		return fmt.Errorf("unknown type %s", field.Type)
	}

	return nil
}
