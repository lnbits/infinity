package apps

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aarzilli/golua/lua"
	"github.com/fiatjaf/lunatico"
	"github.com/lnbits/lnbits/services"
	"github.com/lnbits/lnbits/utils"
)

type RunluaParams struct {
	AppID            string
	WalletID         string
	CodeToRun        string
	InjectedGlobals  *map[string]interface{}
	ExtractedGlobals interface{}
}

func runlua(params RunluaParams) (interface{}, error) {
	L := lua.NewState()
	defer L.Close()
	L.OpenLibs()

	appCode, err := getAppCode(params.AppID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app code: %w", err)
	}

	code := appCode + "\n" + CUSTOM_ENV_DEF + "\n"
	if params.CodeToRun != "" {
		code += "return " + params.CodeToRun
	} else {
		code += "return {models=models, triggers=triggers, actions=actions}"
	}

	globalsToInject := map[string]interface{}{
		"app_id": params.AppID,
		"code":   code,

		"debug_print": luaPrint,

		"random_hex":              utils.RandomHex,
		"perform_key_auth_flow":   utils.PerformKeyAuthFlow,
		"get_msats_per_fiat_unit": utils.GetMsatsPerFiatUnit,
	}

	if params.InjectedGlobals != nil {
		for k, v := range *params.InjectedGlobals {
			globalsToInject[k] = v
		}
	}

	// this means this code can create invoices, pay invoices and do crud operations
	// on entries for this wallet/app db
	if params.WalletID != "" {
		walletDependentGlobals := map[string]interface{}{
			"wallet_id": params.WalletID,

			"auth_key":       services.AuthKey,
			"pay_invoice":    services.PayInvoice,
			"create_invoice": services.CreateInvoice,

			"db_get":    DBGet,
			"db_set":    DBSet,
			"db_add":    DBAdd,
			"db_update": DBUpdate,
			"db_delete": DBDelete,
		}

		for k, v := range walletDependentGlobals {
			globalsToInject[k] = v
		}
	}

	// inject globals into lua
	lunatico.SetGlobals(L, globalsToInject)

	// expose them to sandbox
	sandboxGlobalsInjector := "injected_globals = {\n"
	for k, _ := range globalsToInject {
		sandboxGlobalsInjector += "  " + k + " = " + k + ",\n"
	}
	sandboxGlobalsInjector += "}\n"

	err = L.DoString(sandboxGlobalsInjector + `
local sandbox = require('apps.sandbox')
ret = sandbox.run(code, { quota = 300, env = injected_globals })
    `)
	if err != nil {
		if luaError, ok := err.(*lua.LuaError); ok {
			err = errors.New(stacktrace(luaError))
		}

		return nil, fmt.Errorf("lua error: %w", err)
	}

	ret := lunatico.GetGlobals(L, "ret")["ret"]

	if params.CodeToRun != "" {
		return ret, nil
	} else if params.ExtractedGlobals != nil {
		j, err := json.Marshal(ret)
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

const CUSTOM_ENV_DEF = `
wallet = {
  auth_key = function (domain) return auth_key(wallet_id, domain) end,
  pay_invoice = function (params)
    params.tag = app_id
    return pay_invoice(wallet_id, params)
  end,
  create_invoice = function (params)
    params.tag = app_id
    return create_invoice(wallet_id, params)
  end,
}

utils = {
  random_hex = random_hex,
  perform_key_auth_flow = perform_key_auth_flow,
  get_msats_per_fiat_unit = get_msats_per_fiat_unit,
}

db = setmetatable({}, {
  __index = function (_, model_name)
    return {
      get = function (key)
        return db_get(wallet_id, app_id, model_name, key)
      end,
      set = function (key, value)
        return db_set(wallet_id, app_id, key, model_name, value)
      end,
      add = function (value)
        return db_add(wallet_id, app_id, model_name, value)
      end,
      update = function (key, updates)
        return db_update(wallet_id, app_id, model_name, key, updates)
      end,
      delete = function (key)
        return db_delete(wallet_id, app_id, model_name, key)
      end,
    }
  end,
})

print = function (...)
  debug_print(wallet_id, ...)
end

internal = {
  get_model = function (model_name)
    for _, model in ipairs(models) do
      if model_name == model.name then
        return model
      end
    end
  end,
  get_model_field = function (model_name, field_name)
    for _, model in ipairs(models) do
      if model_name == model.name then
        for _, field in ipairs(model.fields) do
          if field_name == field.name then
            return field
          end
        end
      end
    end
  end,
  get_trigger = function (trigger_name)
    if trigger == nil or trigger[trigger_name] ~= 'function' then
      return function () end
    end
    return trigger[trigger_name]
  end,
  arg = arg and setmetatable(arg, {
    __index = function (base, attr)
      local value = base.value

      if base.key and value then -- it's an item
        if attr ~= 'value' and attr ~= 'key' and attr ~= 'wallet' and attr ~= 'model' then
          return value[attr]
        end
      end

      return base[attr]
    end,
  }),
}
`

func luaPrint(walletID string, args ...interface{}) {
	prints := []interface{}{"lua print: "}
	for _, arg := range args {
		if j, err := json.Marshal(arg); err != nil {
			prints = append(prints, arg)
		} else {
			prints = append(prints, string(j))
		}
	}
	fmt.Println(prints...)

	if walletID != "" {
		SendPrintSSE(walletID, prints)
	}
}
