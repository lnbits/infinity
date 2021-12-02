package apps

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fiatjaf/lunatico"
	models "github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/utils"
)

type Settings struct {
	URL         string                           `json:"url"`
	Title       string                           `json:"title"`
	Description string                           `json:"description,omitempty"`
	Code        string                           `json:"code"`
	Models      []Model                          `json:"models"`
	Triggers    map[string]*lunatico.LuaFunction `json:"triggers"`
	Actions     map[string]Action                `json:"actions"`
	Files       map[string]string                `json:"files"`
}

func (s *Settings) normalize() {
	if s.Title == "" {
		spl := strings.Split(s.URL, "/")
		s.Title = spl[len(spl)-1]
	}

	for m, _ := range s.Models {
		for _, filter := range s.Models[m].DefaultFiltersLua {
			if len(filter) == 3 {
				s.Models[m].DefaultFiltersJS = append(s.Models[m].DefaultFiltersJS,
					Filter{
						filter[0].(string),
						filter[1].(string),
						filter[2],
					},
				)
			}
		}
		s.Models[m].DefaultFiltersLua = nil

		if s.Models[m].DefaultSortLua != "" {
			spl := strings.Split(strings.TrimSpace(
				strings.ToLower(s.Models[m].DefaultSortLua),
			), " ")
			s.Models[m].DefaultSortJS.SortBy = spl[0]
			s.Models[m].DefaultSortJS.Descending = len(spl) == 2 && spl[1] == "desc"
			s.Models[m].DefaultSortLua = ""
		}
	}
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

		for f, filter := range s.Models[m].DefaultFiltersLua {
			if len(filter) != 3 {
				return fmt.Errorf(
					"model %s.default_filters[%d] doesn't have 3 elements", model.Name, f)
			}

			if _, ok := filter[0].(string); !ok {
				return fmt.Errorf(
					"model %s.default_filters[%d]'s first field is not a string",
					model.Name, f)
			}

			if _, ok := filter[1].(string); !ok {
				return fmt.Errorf(
					"model %s.default_filters[%d]'s second field is not a string",
					model.Name, f)
			}
		}
	}

	for name, def := range s.Actions {
		if !nameValidator.MatchString(name) {
			return fmt.Errorf("action name '%s' is invalid", name)
		}

		if def.Fields == nil {
			def.Fields = []Field{}
		}

		for f, field := range def.Fields {
			if err := field.validate(s.Models); err != nil {
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
	Single  bool    `json:"single,omitempty"`

	DefaultSortLua string `json:"default_sort,omitempty"`
	DefaultSortJS  struct {
		SortBy     string `json:"sortBy,omitempty"`
		Descending bool   `json:"descending,omitempty"`
	} `json:"defaultSort,omitempty"`

	DefaultFiltersLua [][]interface{} `json:"default_filters,omitempty"`
	DefaultFiltersJS  []Filter        `json:"defaultFilters,omitempty"`
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

func (action Action) validateParams(
	params map[string]interface{},
	querystringFields map[string]struct{},
	walletID string,
	app string,
) error {
	missingRequired := make(map[string]bool)

	for fieldName, fieldValue := range params {
		fieldExpected := false

		for _, field := range action.Fields {
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
				if err := field.validateValue(fieldValue, walletID, app); err != nil {
					return err
				}

				missingRequired[field.Name] = false
				break
			}
		}

		// we disallow unexpected body params, but querystrings can contain extra stuff
		if _, isQSField := querystringFields[fieldName]; !fieldExpected && !isQSField {
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
	Options  []interface{}         `json:"options,omitempty"`
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
	"currency",
	"select",
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

	switch field.Type {
	case "select":
		if len(field.Options) == 0 {
			return fmt.Errorf("%s has type='select', but no options were defined",
				field.Name)
		}
	case "ref":
		if models == nil {
			return fmt.Errorf("%s has type='ref', but we don't accept ref types in this context", field.Name)
		}

		if field.Ref == "" {
			return fmt.Errorf("%s has type='ref', but ref is not provided", field.Name)
		}

		if field.As == "" {
			return fmt.Errorf("%s has type='ref', needs a field as='' with the value equal to the name of a property in the '%s' model", field.Name, field.Ref)
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
					return fmt.Errorf("%s has as='%s', but model '%s' doesn't have a field '%s'", field.Name, field.As, refModel.Name, field.As)
				}

				refExists = true
				break
			}
		}
		if refExists == false {
			return fmt.Errorf("%s has ref='%s', but that doesn't exist as a model",
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
		amount := int64(value.(float64))
		if float64(amount) != value.(float64) {
			return fmt.Errorf(
				"%s=%v is not an integer, %d must be an integer",
				field.Type, field.Name, value,
			)
		}
		if amount > 100000000000 {
			return fmt.Errorf("%s=%v is way too many satoshis",
				field.Name, value)
		}
	case "currency":
		v, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%s=%v is not a map", field.Name, value)
		}
		if len(v) != 2 {
			return fmt.Errorf("%s=%v has more than 2 fields, should have {amount, unit}",
				field.Name, value)
		}
		if amt, ok := v["amount"]; !ok {
			return fmt.Errorf("%s=%v is missing the 'amount' field", field.Name, value)
		} else {
			if _, ok := amt.(float64); !ok {
				return fmt.Errorf("%s=%v amount is not a number", field.Name, value)
			}
		}

		if un, ok := v["unit"]; !ok {
			return fmt.Errorf("%s=%v is missing the 'unit' field", field.Name, value)
		} else {
			if unit, ok := un.(string); !ok {
				return fmt.Errorf("%s=%v unit is not a string", field.Name, value)
			} else if unit != "sat" {
				bad := true
				for _, curr := range utils.CURRENCIES {
					if strings.ToUpper(unit) == curr {
						bad = false
						break
					}
				}
				if bad {
					return fmt.Errorf("%s=%v unit='%v', invalid currency",
						field.Name, value, unit)
				}
			}
		}
	case "select":
		// anything goes
	case "boolean":
		if valueType.Name() != "bool" {
			return fmt.Errorf("%s=%v is not a boolean", field.Name, value)
		}
	case "ref":
		if valueType.Name() != "string" {
			return fmt.Errorf("%s=%v is not a ref string",
				field.Name, value)
		}
		if walletID == "" || app == "" {
			return fmt.Errorf("%s=%v is a ref but we don't accept refs here",
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

type Filter struct {
	Field string      `json:"field"`
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
}
