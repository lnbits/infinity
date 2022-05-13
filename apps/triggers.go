package apps

import (
	"fmt"

	"github.com/lnbits/infinity/models"
	"github.com/lnbits/infinity/storage"
)

type AppWallet struct {
	WalletID string
	URL      string
}

func TriggerPaymentEvent(trigger string, payment models.Payment) {
	// get all apps/wallets from from the user that owns this payment's wallet
	var appWalletCombinations []AppWallet
	result := storage.DB.Raw(`
      SELECT wallets.id AS wallet_id, url
      FROM user_apps
      LEFT OUTER JOIN users ON user_apps.user_id = users.id
      LEFT OUTER JOIN wallets ON wallets.user_id = users.id
      WHERE users.id = (SELECT user_id FROM wallets WHERE id = ?)
    `, payment.WalletID).Scan(&appWalletCombinations)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("failed to load apps for payment user")
		return
	}

	for _, appWallet := range appWalletCombinations {
		if settings := getCachedAppSettings(appWallet.URL); settings != nil {
			if _, exists := settings.Triggers[trigger]; !exists {
				continue
			}
		}

		_, err := runlua(RunluaParams{
			AppURL:   appWallet.URL,
			WalletID: appWallet.WalletID,
			CodeToRun: fmt.Sprintf(
				"internal.get_trigger('%s')(internal.arg)",
				trigger,
			),
			InjectedGlobals: &map[string]interface{}{"arg": structToMap(payment)},
		})
		if err != nil {
			log.Warn().Err(err).
				Str("wallet", appWallet.WalletID).
				Str("app", appWallet.URL).
				Str("trigger", trigger).
				Interface("payment", payment).
				Msg("failed to call trigger")
		}
	}
}

func TriggerGlobalEvent(trigger string, data interface{}) {
	// get all apps/wallets from from all users
	var appWalletCombinations []AppWallet
	result := storage.DB.Raw(`
      SELECT wallets.id AS wallet_id, url
      FROM user_apps
      LEFT OUTER JOIN users ON user_apps.user_id = users.id
      LEFT OUTER JOIN wallets ON wallets.user_id = users.id
    `).Scan(&appWalletCombinations)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("failed to load apps for all users")
		return
	}

	for _, appWallet := range appWalletCombinations {
		TriggerEventOnSpecificAppWallet(appWallet, trigger, data)
	}
}

func TriggerEventOnSpecificAppWallet(
	appWallet AppWallet,
	trigger string,
	data interface{},
) {
	if settings := getCachedAppSettings(appWallet.URL); settings != nil {
		if _, exists := settings.Triggers[trigger]; !exists {
			return
		}
	}

	_, err := runlua(RunluaParams{
		AppURL:   appWallet.URL,
		WalletID: appWallet.WalletID,
		CodeToRun: fmt.Sprintf(
			"internal.get_trigger('%s')(internal.arg)",
			trigger,
		),
		InjectedGlobals: &map[string]interface{}{"arg": structToMap(data)},
	})
	if err != nil {
		log.Warn().Err(err).
			Str("wallet", appWallet.WalletID).
			Str("app", appWallet.URL).
			Str("trigger", trigger).
			Interface("data", data).
			Msg("failed to call trigger")
	}
}
