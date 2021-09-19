import {exportFile, Notify} from 'quasar'

import store from './store'

const request = async (path, opts = {}) => {
  opts.headers = opts.headers || {}

  if (path.startsWith('/api/wallet')) {
    opts.headers['X-API-Key'] = store.state.wallet.invoicekey
  } else {
    const key = new URLSearchParams(location.search).get('key')
    if (key) opts.headers['X-MasterKey'] = key
  }

  const r = await fetch(path, opts)
  if (!r.ok) {
    const text = await r.text()
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
  return await r.json()
}

export const loadSettings = async () => await request('/v/settings')

export const scanLnurl = async lnurl =>
  await request(`/api/wallet/scan/${lnurl}`, {})

export const loadUser = async () => await request('/api/user')

export const createWallet = async name =>
  await request('/api/create-wallet', {
    method: 'POST',
    body: JSON.stringify({name})
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

export const decryptLnurlPayAES = (success_action, preimage) => {
  let keyb = new Uint8Array(
    preimage.match(/[\da-f]{2}/gi).map(h => parseInt(h, 16))
  )

  return crypto.subtle
    .importKey('raw', keyb, {name: 'AES-CBC', length: 256}, false, ['decrypt'])
    .then(key => {
      let ivb = Uint8Array.from(window.atob(success_action.iv), c =>
        c.charCodeAt(0)
      )
      let ciphertextb = Uint8Array.from(
        window.atob(success_action.ciphertext),
        c => c.charCodeAt(0)
      )

      return crypto.subtle.decrypt({name: 'AES-CBC', iv: ivb}, key, ciphertextb)
    })
    .then(valueb => {
      let decoder = new TextDecoder('utf-8')
      return decoder.decode(valueb)
    })
}

export const exportCSV = (columns, data) => {
  function wrapCsvValue(val, formatFn) {
    var formatted = formatFn !== void 0 ? formatFn(val) : val

    formatted =
      formatted === void 0 || formatted === null ? '' : String(formatted)

    formatted = formatted.split('"').join('""')

    return `"${formatted}"`
  }

  var content = [
    columns.map(function (col) {
      return wrapCsvValue(col.label)
    })
  ]
    .concat(
      data.map(function (row) {
        return columns
          .map(function (col) {
            return wrapCsvValue(
              typeof col.field === 'function'
                ? col.field(row)
                : row[col.field === void 0 ? col.name : col.field],
              col.format
            )
          })
          .join(',')
      })
    )
    .join('\r\n')

  var status = exportFile('table-export.csv', content)

  if (status !== true) {
    Notify.create({
      message: 'Browser denied file download...',
      color: 'negative',
      icon: null
    })
  }
}
export const notifyError = (error, title, type) => {
  const caption =
    title ||
    (error.response
      ? [error.response.status, ' ', error.response.statusText]
          .join('')
          .toUpperCase()
      : null)
  type =
    type ||
    ((error.response && error.response.status) >= 500 ? 'negative' : 'warning')

  Notify.create({
    timeout: 5000,
    type,
    message: error.message || null,
    caption,
    icon: null
  })
}

export const changeColorTheme = newValue => {
  document.body.setAttribute('data-theme', newValue)
  this.$q.localStorage.set('lnbits.theme', newValue)
}
