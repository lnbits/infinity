package runlua

import (
	"errors"

	"github.com/aarzilli/golua/lua"
	"github.com/fiatjaf/lunatico"
)

func RunLua(actualCode, functionToRun string) (interface{}, error) {
	L := lua.NewState()
	defer L.Close()
	L.OpenLibs()

	lunatico.SetGlobals(L, map[string]interface{}{
		"code": actualCode + "\nreturn " + functionToRun + "()",
	})

	code := `
sandbox_env = {
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

_calls = 0
function count ()
  _calls = _calls + 1
  if _calls > 100 then
    error('too many operations!')
  end
end
debug.sethook(count, 'c')

ret = load(code, 'call', 't', sandbox_env)()
    `

	err := L.DoString(code)
	if err != nil {
		st := stackTraceWithCode(err.Error(), actualCode)
		err = errors.New(st)
		return nil, err
	}

	ret, _ := lunatico.GetGlobals(L, "ret")["ret"]
	return ret, nil
}
