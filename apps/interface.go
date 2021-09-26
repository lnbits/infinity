package apps

import (
	"encoding/json"
	"fmt"

	"github.com/lnbits/lnbits/apps/runlua"
)

func getSettings(code string) (settings Settings, err error) {
	returned, err := runlua.RunLua(code, "init")
	if err != nil {
		return settings, err
	}

	j, err := json.Marshal(returned)
	if err != nil {
		return settings, fmt.Errorf("failed to json-encode settings: %w", err)
	}

	if err := json.Unmarshal(j, &settings); err != nil {
		return settings, fmt.Errorf("failed to decode '%s' into settings: %w",
			string(j), err)
	}

	return settings, nil
}
