package apps

import (
	"fmt"

	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
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
			CodeToRun: fmt.Sprintf(
				"internal.get_trigger('%s')(internal.arg)",
				trigger,
			),
			InjectedGlobals: &map[string]interface{}{"arg": structToMap(payment)},
		})
		if err != nil {
			log.Warn().Err(err).Str("app", app).
				Str("trigger", trigger).
				Interface("payment", payment).
				Msg("failed to call trigger")
		}
	}
}
