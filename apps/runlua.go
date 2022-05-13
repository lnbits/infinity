package apps

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strconv"
	"time"

	"github.com/aarzilli/golua/lua"
	"github.com/fiatjaf/go-lnurl"
	"github.com/fiatjaf/lunatico"
	"github.com/lnbits/infinity/services"
	"github.com/lnbits/infinity/utils"
	"github.com/lnbits/infinity/utils/nostr_utils"
)

//go:embed sandbox.lua
var sandboxCode string

type RunluaParams struct {
	Code             string
	AppURL           string
	WalletID         string
	CodeToRun        string
	InjectedGlobals  *map[string]interface{}
	ExtractedGlobals interface{}
}

func runlua(params RunluaParams) (interface{}, error) {
	L := lua.NewState()
	defer L.Close()
	L.OpenLibs()

	if params.Code == "" {
		appCode, err := getAppCode(params.AppURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get app code: %w", err)
		}
		params.Code = appCode
	}

	code := CUSTOM_ENV_DEF + "\n" + params.Code + "\n"
	if params.CodeToRun != "" {
		code += "return " + params.CodeToRun
	} else {
		code += `
return {
  title = title,
  description = description,
  models = models,
  triggers = triggers,
  actions = actions,
  files = files
}`
	}

	globalsToInject := map[string]interface{}{
		"app_id":         params.AppURL,
		"app_encoded_id": appURLToID(params.AppURL),
		"code":           code,

		"debug_print": luaPrint,

		"sha256":                  utils.Sha256String,
		"random_hex":              utils.RandomHex,
		"aes_encrypt":             utils.AESEncrypt,
		"aes_decrypt":             utils.AESDecrypt,
		"perform_key_auth_flow":   utils.PerformKeyAuthFlow,
		"currencies":              utils.CURRENCIES,
		"get_msats_per_fiat_unit": utils.GetMsatsPerFiatUnit,
		"parse_date":              utils.DateStringToTimestamp,
		"http_get":                utils.HTTPGet,
		"http_put":                utils.HTTPPut,
		"http_post":               utils.HTTPPost,
		"http_patch":              utils.HTTPPatch,
		"http_delete":             utils.HTTPDelete,
		"http_request":            utils.HTTP,
		"qs_parse":                utils.ParseQueryString,
		"qs_encode":               utils.EncodeQueryString,
		"json_parse":              utils.JSONParse,
		"json_encode":             utils.JSONEncode,
		"snigirev_encrypt":        utils.SnigirevEncrypt,
		"snigirev_decrypt":        utils.SnigirevDecrypt,
		"lnurl_bech32_encode":     lnurl.LNURLEncode,
		"lnurl_bech32_decode":     lnurl.LNURLDecode,
		"lnurl_successaction_aes": utils.AESSuccessAction,
		"feed_parse":              utils.ParseFeed,
		"html_escape":             html.EscapeString,
		"html_unescape":           html.UnescapeString,
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

			"emit_public_event": emitPublicEvent,

			"auth_key":             services.AuthKey,
			"pay_invoice":          services.PayInvoiceFromApp,
			"create_invoice":       services.CreateInvoiceFromApp,
			"internal_transfer":    services.Transfer,
			"get_wallet_payment":   services.GetWalletPayment,
			"load_wallet_balance":  services.LoadWalletBalance,
			"load_wallet_payments": services.LoadWalletPayments,
			"nostr_publish":        nostr_utils.Publish,

			"db_get":    DBGet,
			"db_set":    DBSet,
			"db_add":    DBAdd,
			"db_list":   DBList,
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
	for k := range globalsToInject {
		sandboxGlobalsInjector += "  " + k + " = " + k + ",\n"
	}
	sandboxGlobalsInjector += "}\n"

	err := L.DoString(sandboxGlobalsInjector + `
local sandbox = (function () ` + sandboxCode + `end)()
ret = sandbox.run(code, { quota = ` + strconv.Itoa(LuaQuota) + `, env = injected_globals })
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
		j, err := utils.JSONMarshal(ret)
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
  id = wallet_id,
  balance = function () return load_wallet_balance(wallet_id) end,
  payments = function () return load_wallet_payments(wallet_id) end,
  get_payment = function (checking_id_or_hash)
    return get_wallet_payment(wallet_id, checking_id_or_hash)
  end,
  auth_key = function (domain) return auth_key(wallet_id, domain) end,
  pay_invoice = function (params)
    params.tag = app_id
    return pay_invoice(wallet_id, params)
  end,
  create_invoice = function (params)
    params.tag = app_id
    return create_invoice(wallet_id, params)
  end,
  transfer = function (to_wallet, msatoshi, description)
    params.tag = app_id
    return internal_transfer(wallet_id, to_wallet, msatoshi, description)
  end,
}

app = {
  id = app_id,
  encoded_id = app_encoded_id,
  emit_event = function (name, data)
    emit_public_event(wallet_id, app_id, name, data)
  end,
}

http = {
  request = http_request,
  get = http_get,
  put = http_put,
  post = http_post,
  patch = http_patch,
  delete = http_delete,
}

qs = {
  parse = qs_parse,
  encode = qs_encode,
}

json = {
  parse = json_parse,
  encode = json_encode,
}

lnurl = {
  bech32_encode = lnurl_bech32_encode,
  bech32_decode = lnurl_bech32_decode,
  successaction_aes = lnurl_successaction_aes,
}

utils = {
  qs = qs,
  json = json,
  http = http,
  sha256 = sha256,
  feed_parse = feed_parse,
  currencies = currencies,
  parse_date = parse_date,
  random_hex = random_hex,
  aes_encrypt = aes_encrypt,
  aes_decrypt = aes_decrypt,
  html_escape = html_escape,
  html_unescape = html_unescape,
  snigirev_encrypt = snigirev_encrypt,
  snigirev_decrypt = snigirev_decrypt,
  perform_key_auth_flow = perform_key_auth_flow,
  get_msats_per_fiat_unit = get_msats_per_fiat_unit,
}

db = setmetatable({}, {
  __index = function (_, model_name)
    return {
      get = function (key)
        if internal.get_model(model_name).single then
          key = 'single'
        end

        return db_get(wallet_id, app_id, model_name, key)
      end,
      set = function (key, value)
        if internal.get_model(model_name).single then
          key = 'single'
        end

        return db_set(wallet_id, app_id, model_name, key, value)
      end,
      add = function (value)
        if internal.get_model(model_name).single then
          error("can't .add() because " .. model_name .. " is 'single'")
        end

        return db_add(wallet_id, app_id, model_name, value)
      end,
      list = function (args)
        if internal.get_model(model_name).single then
          error("can't .list() because " .. model_name .. " is 'single'")
        end

        args = args or {}
        return db_list(wallet_id, app_id, model_name, args.startkey or "", args.endkey or "")
      end,
      update = function (key, updates)
        if internal.get_model(model_name).single then
          key = 'single'
        end

        return db_update(wallet_id, app_id, model_name, key, updates)
      end,
      delete = function (key)
        if internal.get_model(model_name).single then
          key = 'single'
        end

        return db_delete(wallet_id, app_id, model_name, key)
      end,
    }
  end,
})

print = function (...)
  debug_print(wallet_id, app_id, ...)
end

emptyarray = function ()
  return { __emptyarray = 1 }
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
    if triggers == nil or type(triggers[trigger_name]) ~= 'function' then
      return function () end
    end
    return triggers[trigger_name]
  end,
  arg = arg
}
`

type LuaPrintDTO struct {
	WalletID string        `json:"wallet_id"`
	App      string        `json:"app"`
	Time     time.Time     `json:"time"`
	Values   []interface{} `json:"values"`
}

func luaPrint(walletID string, app string, args ...interface{}) {
	prints := LuaPrintDTO{
		WalletID: walletID,
		App:      app,
		Time:     time.Now(),
		Values:   args,
	}

	toSSE, err := json.Marshal(prints)
	if err != nil {
		log.Error().Err(err).Msg("lua print unmarshaling error")
		return
	}

	log.Debug().RawJSON("lua_print", toSSE).Send()

	if walletID != "" {
		SendPrintSSE(walletID, toSSE)
	}
}
