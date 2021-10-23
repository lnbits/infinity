function API({
  masterKey,
  walletKey,
  baseURL = location.href.split('/').slice(0, 3).join('/')
}) {
  const api = {}

  api.request = async (path, opts = {}) => {
    opts.method = 'POST'
    opts.headers = opts.headers || {}

    if (path.startsWith('/api/wallet')) {
      if (!walletKey)
        throw new Error("can't make request: walletKey not provided.")

      opts.headers['X-API-Key'] = walletkey
    } else if (path.startsWith('/api/user')) {
      if (!masterKey)
        throw new Error("can't make request: user masterKey not provided.")

      if (key) opts.headers['X-MasterKey'] = masterKey
    }

    const r = await fetch(path, opts)
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

  api.loadUser = async () => await request('/api/user')

  api.createWallet = async name =>
    await request('/api/user/create-wallet', {
      method: 'POST',
      body: JSON.stringify({name})
    })

  api.addApp = async url =>
    await request('/api/user/add-app', {
      method: 'POST',
      body: JSON.stringify({url})
    })

  api.removeApp = async url =>
    await request('/api/user/remove-app', {
      method: 'POST',
      body: JSON.stringify({url})
    })

  api.loadWallet = async () => await request(`/api/wallet`)

  api.createInvoice = async params =>
    await request(`/api/wallet/create-invoice`, {
      method: 'POST',
      body: JSON.stringify(params)
    })

  api.payInvoice = async ({invoice, customAmount = 0}) =>
    await request(`/api/wallet/pay-invoice`, {
      method: 'POST',
      body: JSON.stringify({invoice, customAmount})
    })

  api.payLnurl = async params =>
    await request(`/api/wallet/pay-lnurl`, {
      method: 'POST',
      body: JSON.stringify({params})
    })

  api.authLnurl = async callback =>
    await request(`/api/wallet/lnurlauth`, {
      method: 'POST',
      body: JSON.stringify({callback})
    })

  api.renameWallet = async name =>
    await request(`/api/wallet/${name}`, {method: 'POST'})

  api.deleteWallet = async () =>
    await request(`/api/wallet/delete`, {
      method: 'POST'
    })

  api.scanLnurl = async lnurl => await request(`/api/wallet/scan/${lnurl}`, {})

  api.appInfo = async appid => {
    const appSettings = await request(`/api/wallet/app/${appid}`)
    appSettings.id = appid
    appSettings.url = atob(appid)
    return appSettings
  }

  api.listAppItems = async (appURL, model) =>
    await request(`/api/wallet/app/${btoa(appURL)}/list/${model}`)

  api.getAppItem = async (appURL, model, key) =>
    await request(`/api/wallet/app/${btoa(appURL)}/set/${model}/${key}`)

  api.setAppItem = async (appURL, model, key, value) =>
    await request(`/api/wallet/app/${btoa(appURL)}/set/${model}/${key}`, {
      method: 'POST',
      body: JSON.stringify(value)
    })

  api.addAppItem = async (appURL, model, value) =>
    await request(`/api/wallet/app/${btoa(appURL)}/add/${model}`, {
      method: 'POST',
      body: JSON.stringify(value)
    })

  api.delAppItem = async (appURL, model, key) =>
    await request(`/api/wallet/app/${btoa(appURL)}/del/${model}/${key}`)

  return api
}

if (typeof module !== 'undefined' && module.exports) module.exports = api
else if (typeof exports !== 'undefined') exports = API
else window.LNbits = API
