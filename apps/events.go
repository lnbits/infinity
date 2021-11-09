package apps

import (
	"fmt"

	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
)

func TriggerGenericEvent(trigger string, data interface{}) {
	// get all apps from from all users
	var users []models.User
	result := storage.DB.Model(&models.User{}).Select("id", "apps").Find(&users)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("failed to load apps for all users")
		return
	}

	for _, user := range users {
		// get all wallets from all users
		var wallets []string
		storage.DB.Model(&models.Wallet{}).Select("id").Where("user_id", user.ID).
			Find(&wallets)

		for _, walletID := range wallets {
			for _, app := range user.Apps {
				_, err := runlua(RunluaParams{
					AppURL:   app,
					WalletID: walletID,
					CodeToRun: fmt.Sprintf(
						"internal.get_trigger('%s')(internal.arg)",
						trigger,
					),
					InjectedGlobals: &map[string]interface{}{"arg": structToMap(data)},
				})
				if err != nil {
					log.Warn().Err(err).Str("app", app).
						Str("trigger", trigger).
						Interface("data", data).
						Msg("failed to call trigger")
				}
			}
		}
	}
}

func TriggerPaymentEvent(trigger string, payment models.Payment) {
	// get all apps from this user
	var user models.User
	result := storage.DB.Model(&models.User{}).
		Select("apps").
		Joins("INNER JOIN wallets ON users.id = wallets.user_id").
		Where("wallets.id = ?", payment.WalletID).
		First(&user)
	if result.Error != nil {
		log.Error().Err(result.Error).Interface("payment", payment).
			Msg("failed to load apps for payment")
		return
	}

	for _, app := range user.Apps {
		_, err := runlua(RunluaParams{
			AppURL:   app,
			WalletID: payment.WalletID,
			CodeToRun: fmt.Sprintf(
				"internal.get_trigger('%s')(internal.enhance_payment(internal.arg))",
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
