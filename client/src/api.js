import store from './store'
import {appURLToId} from './helpers'

const request = async (path, opts = {}, key = null) => {
  opts.headers = opts.headers || {}

  if (path.startsWith('/api/wallet')) {
    opts.headers['X-API-Key'] = key || store.state.wallet.adminkey
  } else if (path.startsWith('/api/user')) {
    opts.headers['X-MasterKey'] =
      key || new URLSearchParams(location.search).get('key')
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
    try {
      return JSON.parse(text)
    } catch (err) {
      return text
    }
  } else {
    return
  }
}

export const loadSettings = async () => await request('/v/settings')

export const loadUser = async key => await request('/api/user', {}, key)

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
  await request(`/api/wallet/rename/${name}`, {method: 'POST'})

export const deleteWallet = async () =>
  await request(`/api/wallet/delete`, {
    method: 'POST'
  })

export const scanLnurl = async lnurl =>
  await request(`/api/wallet/lnurlscan/${lnurl}`, {})

export const appInfo = async appid => {
  const appSettings = await request(`/api/wallet/app/${appid}`)
  appSettings.id = appid
  appSettings.url = atob(appid)
  return appSettings
}

export const appRefresh = async appid =>
  await request(`/api/wallet/app/${appid}/refresh`)

export const appClearData = async appid =>
  await request(`/api/wallet/app/${appid}/clear-data`)

export const listAppItems = async (appURL, model) =>
  await request(`/api/wallet/app/${appURLToId(appURL)}/list/${model}`)

export const getAppItem = async (appURL, model, key) =>
  await request(`/api/wallet/app/${appURLToId(appURL)}/set/${model}/${key}`)

export const setAppItem = async (appURL, model, key, value) =>
  await request(`/api/wallet/app/${appURLToId(appURL)}/set/${model}/${key}`, {
    method: 'POST',
    body: JSON.stringify(value)
  })

export const addAppItem = async (appURL, model, value) =>
  await request(`/api/wallet/app/${appURLToId(appURL)}/add/${model}`, {
    method: 'POST',
    body: JSON.stringify(value)
  })

export const delAppItem = async (appURL, model, key) =>
  await request(`/api/wallet/app/${appURLToId(appURL)}/del/${model}/${key}`)

export const callAppAction = async (wallet, appid, action, params) =>
  await request(`/ext/${wallet}/${appid}/action/${action}`, {
    method: 'POST',
    body: JSON.stringify(params)
  })
