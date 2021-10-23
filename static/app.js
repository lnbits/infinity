const api = {}

api._es = null

api._openEventSource = function () {
  const base = location.href.split('/').slice(0, 6).join('/')
  api._es = new EventSource(base + '/sse')
}

api._listeners = new Map()

api.on = function (eventName, cb) {
  if (!api._es) {
    api._openEventSource()
  }

  const fn = ev => cb(JSON.parse(ev.data))
  api._listeners.set(cb, fn)
  api._es.addEventListener(eventName, fn)
}

api.once = function (eventName, cb) {
  if (!api._es) {
    api._openEventSource()
  }

  function once(ev) {
    cb(JSON.parse(ev.data))
    api.removeEventListener(once)
  }

  api._es.addEventListener(eventName, once)
}

api.off = function (eventName, cb) {
  const fn = api._listeners.get(cb)
  api._es.removeEventListener(eventName, fn)
}

api.action = async (actionName, params) => {
  const base = location.href.split('/').slice(0, 6).join('/')

  const r = await fetch(base + '/action/' + actionName, {
    method: 'POST',
    body: params ? JSON.stringify(params) : undefined
  })
  const text = await r.text()

  if (!r.ok) {
    let message
    let data = {}
    try {
      data = JSON.parse(text)
      message = data.message
    } catch (_) {
      message = text
    }
    const error = new Error(message)
    error.data = data
    error.response = r
    throw error
  }

  if (text && text.length) {
    return JSON.parse(text)
  } else {
    return
  }
}

if (typeof module !== 'undefined' && module.exports) module.exports = api
else if (typeof exports !== 'undefined') exports = api
else window.bitsapp = api
