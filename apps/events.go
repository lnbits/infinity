package apps

import (
	"fmt"

	"github.com/lnbits/lnbits/apps/runlua"
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
		code, err := getAppCode(app)
		if err != nil {
			log.Warn().Err(err).Str("app", app).
				Str("trigger", trigger).
				Interface("payment", payment).
				Msg("couldn't get app code to trigger event")
			return
		}

		_, err = runlua.RunLua(runlua.Params{
			AppID:   app,
			AppCode: code,
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
