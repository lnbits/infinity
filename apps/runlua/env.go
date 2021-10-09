package runlua

import (
	"github.com/lnbits/lnbits/services"
	"github.com/lnbits/lnbits/utils"
)

var exposedWalletMethods = map[string]interface{}{
	"auth_key":       services.AuthKey,
	"pay_invoice":    services.PayInvoice,
	"create_invoice": services.CreateInvoice,
}

var exposedUtils = map[string]interface{}{
	"random_hex":              utils.RandomHex,
	"perform_key_auth_flow":   utils.PerformKeyAuthFlow,
	"get_msats_per_fiat_unit": utils.GetMsatsPerFiatUnit,
}

const SANDBOX_ENV = `
sandbox_env = {
  wallet = {
    auth_key = function (domain) return auth_key(wallet_id, domain) end,
    pay_invoice = function (params) return pay_invoice(wallet_id, params) end,
    create_invoice = function (params) return create_invoice(wallet_id, params) end,
  },
  utils = {
    random_hex = random_hex,
    perform_key_auth_flow = perform_key_auth_flow,
    get_msats_per_fiat_unit = get_msats_per_fiat_unit,
  },
  internal = {
    get_model = function (model_name) {
      for _, model in ipairs(model) do
        if model_name == model.name then
          return model
        end
      end
    },
    get_model_field = function (model_name, field_name) {
      for _, model in ipairs(model) do
        if model_name == model.name then
          for _, field in ipairs(model) do
            if field_name == field.name then
              return field
            end
          end
        end
      end
    },
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
