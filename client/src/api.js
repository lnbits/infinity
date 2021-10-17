import store from './store'

const request = async (path, opts = {}) => {
  opts.headers = opts.headers || {}

  if (path.startsWith('/api/wallet')) {
    opts.headers['X-API-Key'] = store.state.wallet.adminkey
  } else if (path.startsWith('/api/user')) {
    const key = new URLSearchParams(location.search).get('key')
    if (key) opts.headers['X-MasterKey'] = key
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

export const loadSettings = async () => await request('/v/settings')

export const loadUser = async () => await request('/api/user')

export const createWallet = async name =>
  await request('/api/user/create-wallet', {
    method: 'POST',
    body: JSON.stringify({name})
  })

export const addApp = async url =>
  await request('/api/user/add-app', {
    method: 'POST',
    body: JSON.stringify({url})
  })

export const removeApp = async url =>
  await request('/api/user/remove-app', {
    method: 'POST',
    body: JSON.stringify({url})
  })

export const loadWallet = async () => await request(`/api/wallet`)

export const createInvoice = async params =>
  await request(`/api/wallet/create-invoice`, {
    method: 'POST',
    body: JSON.stringify(params)
  })

export const payInvoice = async ({invoice, customAmount = 0}) =>
  await request(`/api/wallet/pay-invoice`, {
    method: 'POST',
    body: JSON.stringify({invoice, customAmount})
  })

export const payLnurl = async params =>
  await request(`/api/wallet/pay-lnurl`, {
    method: 'POST',
    body: JSON.stringify({params})
  })

export const authLnurl = async callback =>
  await request(`/api/wallet/lnurlauth`, {
    method: 'POST',
    body: JSON.stringify({callback})
  })

export const renameWallet = async name =>
  await request(`/api/wallet/${name}`, {method: 'POST'})

export const deleteWallet = async () =>
  await request(`/api/wallet/delete`, {
    method: 'POST'
  })

export const scanLnurl = async lnurl =>
  await request(`/api/wallet/scan/${lnurl}`, {})

export const appInfo = async appid => {
  const appSettings = await request(`/api/wallet/app/${appid}`)
  appSettings.id = appid
  return appSettings
}

export const listAppItems = async (id, model) =>
  await request(`/api/wallet/app/${id}/list/${model}`)

export const getAppItem = async (id, model, key) =>
  await request(`/api/wallet/app/${id}/set/${model}/${key}`)

export const setAppItem = async (id, model, key, value) =>
  await request(`/api/wallet/app/${id}/set/${model}/${key}`, {
    method: 'POST',
    body: JSON.stringify(value)
  })

export const addAppItem = async (id, model, value) =>
  await request(`/api/wallet/app/${id}/add/${model}`, {
    method: 'POST',
    body: JSON.stringify(value)
  })

export const delAppItem = async (id, model, key) =>
  await request(`/api/wallet/app/${id}/del/${model}/${key}`)
