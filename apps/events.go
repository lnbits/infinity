package apps

import (
	"fmt"

	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
)

type AppWalletCombination struct {
	WalletID string
	URL      string
}

func TriggerGenericEvent(trigger string, data interface{}) {
	// get all apps/wallets from from all users
	var appWalletCombinations []AppWalletCombination
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
}

func TriggerPaymentEvent(trigger string, payment models.Payment) {
	// get all apps/wallets from from the user that owns this payment's wallet
	var appWalletCombinations []AppWalletCombination
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
