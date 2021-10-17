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
	FunctionToRun    string
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
	var code string
	var globalsToGet []string
	if params.FunctionToRun != "" {
		code = appCode + "\nreturn " + params.FunctionToRun + "()"
		globalsToGet = []string{"ret"}
	} else {
		code = appCode
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

	if params.InjectedGlobals != nil {
		lunatico.SetGlobals(L, *params.InjectedGlobals)
	}
	lunatico.SetGlobals(L, map[string]interface{}{
		"random_hex":              utils.RandomHex,
		"perform_key_auth_flow":   utils.PerformKeyAuthFlow,
		"get_msats_per_fiat_unit": utils.GetMsatsPerFiatUnit,
	})

	// this means this code can create invoices, pay invoices and do crud operations
	// on entries for this wallet/app db
	if params.WalletID != "" {
		lunatico.SetGlobals(L, map[string]interface{}{
			"wallet_id": params.WalletID,
		})
		lunatico.SetGlobals(L, map[string]interface{}{
			"auth_key":       services.AuthKey,
			"pay_invoice":    services.PayInvoice,
			"create_invoice": services.CreateInvoice,
		})
		lunatico.SetGlobals(L, map[string]interface{}{
			"db_get":    DBGet,
			"db_set":    DBSet,
			"db_add":    DBAdd,
			"db_update": DBUpdate,
			"db_delete": DBDelete,
		})
	}

	err = L.DoString(SANDBOX_ENV + appCode + `
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
		st := stackTraceWithCode(err.Error(), appCode)
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

const SANDBOX_ENV = `
sandbox_env = {
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
  },
  utils = {
    random_hex = random_hex,
    perform_key_auth_flow = perform_key_auth_flow,
    get_msats_per_fiat_unit = get_msats_per_fiat_unit,
  },
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
  }),
  internal = {
    get_model = function (model_name)
      for _, model in ipairs(model) do
        if model_name == model.name then
          return model
        end
      end
    end,
    get_model_field = function (model_name, field_name)
      for _, model in ipairs(model) do
        if model_name == model.name then
          for _, field in ipairs(model) do
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
  },
  ipairs = ipairs,
  next = next,
  pairs = pairs,
  error = error,
  tonumber = tonumber,
  tostring = tostring,
  type = type,
  unpack = unpack,
  utf8 = utf8,
  string = { byte = string.byte, char = string.char, find = string.find,
      format = string.format, gmatch = string.gmatch, gsub = string.gsub,
      len = string.len, lower = string.lower, match = string.match,
      rep = string.rep, reverse = string.reverse, sub = string.sub,
      upper = string.upper },
  table = { insert = table.insert, maxn = table.maxn, remove = table.remove,
      sort = table.sort, pack = table.pack },
  math = { abs = math.abs, acos = math.acos, asin = math.asin,
      atan = math.atan, atan2 = math.atan2, ceil = math.ceil, cos = math.cos,
      cosh = math.cosh, deg = math.deg, exp = math.exp, floor = math.floor,
      fmod = math.fmod, frexp = math.frexp, huge = math.huge,
      ldexp = math.ldexp, log = math.log, log10 = math.log10, max = math.max,
      min = math.min, modf = math.modf, pi = math.pi, pow = math.pow,
      rad = math.rad, random = math.random, randomseed = math.randomseed,
      sin = math.sin, sinh = math.sinh, sqrt = math.sqrt, tan = math.tan, tanh = math.tanh  },
  os = { clock = os.clock, difftime = os.difftime, time = os.time, date = os.date },
}
`
