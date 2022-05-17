package apps

import (
	"fmt"

	"github.com/lnbits/infinity/models"
)

func fillComputedValues(item models.AppDataItem) error {
	settings, err := GetAppSettings(item.App, false)
	if err != nil {
		return fmt.Errorf("failed to get app settings for %s: %w", item.App, err)
	}

	model := settings.getModel(item.Model)

	for _, field := range model.Fields {
		if field.Computed != nil {
			var err error
			item.Value[field.Name], err = runlua(RunluaParams{
				AppURL: item.App,
				CodeToRun: fmt.Sprintf(
					"internal.get_model_field('%s', '%s').computed(internal.arg)",
					model.Name, field.Name,
				),
				InjectedGlobals: &map[string]interface{}{"arg": structToMap(item)},
			})
			if err != nil {
				log.Debug().Err(err).Interface("item", item).
					Str("model", model.Name).Str("field", field.Name).
					Msg("failed to run compute")
			}
		}
	}

	return nil
}
