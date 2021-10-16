package runlua

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aarzilli/golua/lua"
	"github.com/fiatjaf/lunatico"
)

type Params struct {
	AppID            string
	AppCode          string
	WalletID         string
	FunctionToRun    string
	InjectedGlobals  *map[string]interface{}
	ExtractedGlobals interface{}
}

func RunLua(params Params) (interface{}, error) {
	L := lua.NewState()
	defer L.Close()
	L.OpenLibs()

	var code string
	var globalsToGet []string
	if params.FunctionToRun != "" {
		code = params.AppCode + "\nreturn " + params.FunctionToRun + "()"
		globalsToGet = []string{"ret"}
	} else {
		code = params.AppCode
		globalsToGet = []string{
			"models",
			"actions",
			"on",
		}
	}

	lunatico.SetGlobals(L, map[string]interface{}{
		"app_id": params.AppID,
		"code":   code,
	})
	lunatico.SetGlobals(L, exposedUtils)

	// this means this code can create invoices, pay invoices and do crud operations
	// on entries for this wallet/app db
	if params.WalletID != "" {
		lunatico.SetGlobals(L, map[string]interface{}{
			"wallet_id": params.WalletID,
		})
		lunatico.SetGlobals(L, exposedWalletMethods)
		lunatico.SetGlobals(L, exposedDatabase)
	}

	if params.InjectedGlobals != nil {
		lunatico.SetGlobals(L, *params.InjectedGlobals)
	}

	err := L.DoString(SANDBOX_ENV + params.AppCode + `
_calls = 0
function count ()
  _calls = _calls + 1
  if _calls > 100 then
    error('too many operations!')
  end
end
debug.sethook(count, 'c')

ret = load(code, 'call', 't', sandbox_env)()
    `)
	if err != nil {
		st := stackTraceWithCode(err.Error(), params.AppCode)
		err = errors.New(st)
		return nil, err
	}

	globals := lunatico.GetGlobals(L, globalsToGet...)

	if params.FunctionToRun != "" {
		return globals["ret"], nil
	} else if params.ExtractedGlobals != nil {
		j, err := json.Marshal(globals)
		if err != nil {
			return nil, fmt.Errorf("failed to json-encode globals: %w", err)
		}

		if err := json.Unmarshal(j, params.ExtractedGlobals); err != nil {
			return nil, fmt.Errorf("failed to decode '%s' into interface: %w",
				string(j), err)
		}

		return nil, nil
	}

	return nil, nil
}
