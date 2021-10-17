package apps

import (
	"fmt"

	models "github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
	"github.com/rs/zerolog/log"
)

func TriggerEvent(trigger string, payment models.Payment) {
	// get all apps from this user
	var user models.User
	result := storage.DB.Model(&models.User{}).
		Select("apps").
		Joins("Wallet").
		Where("wallet_id = ?", payment.WalletID).
		First(&user)
	if result.Error != nil {
		log.Error().Err(result.Error).Interface("payment", payment).
			Msg("failed to load apps for payment")
		return
	}

	for _, app := range user.Apps {
		_, err := runlua(RunluaParams{
			AppID: app,
			FunctionToRun: fmt.Sprintf(
				"get_trigger('%s')(payment)",
				trigger,
			),
			InjectedGlobals: &map[string]interface{}{"payment": payment},
		})
		if err != nil {
			log.Warn().Err(err).Str("app", app).
				Str("trigger", trigger).
				Interface("payment", payment).
				Msg("failed to call trigger")
		}
	}
}
