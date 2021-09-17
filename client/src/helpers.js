import {exportFile, Notify} from 'quasar'

const request = async (url, opts = {}) => {
  opts.headers = opts.headers || {}
  const key = new URLSearchParams(location.search).get('key')
  if (key) opts.headers['X-MasterKey'] = key
  const r = await fetch(url, opts)
  if (!r.ok) throw new Error(await r.text())
  return await r.json()
}

export const loadSettings = async () => await request('/v/settings')

export const loadUser = async () => await request('/api/user')

export const createWallet = async name =>
  await request('/api/create-wallet', {
    method: 'POST',
    body: JSON.stringify({name})
  })

export const loadWallet = async walletID =>
  await request(`/api/wallet/${walletID}`)

export const createInvoice = async (walletID, params) =>
  await request(`/api/wallet/${walletID}/create-invoice`, {
    method: 'POST',
    body: JSON.stringify(params)
  })

export const payInvoice = async (walletID, invoice) =>
  await request(`/api/wallet/${walletID}/pay-invoice`, {
    method: 'POST',
    body: JSON.stringify({invoice})
  })

export const payLnurl = async (walletID, params) =>
  await request(`/api/wallet/${walletID}/pay-lnurl`, {
    method: 'POST',
    body: JSON.stringify({params})
  })

export const scanLnurl = async (walletID, request) =>
  await request(`/api/wallets/${walletID}/lnurlscan/${request}`, {})

export const authLnurl = async (walletID, callback) =>
  await request(`/api/wallets/${walletID}/lnurlauth`, {
    method: 'POST',
    body: JSON.stringify({callback})
  })

export const renameWallet = async (walletID, name) =>
  await request(`/api/wallets/${walletID}/rename`, {
    method: 'POST',
    body: JSON.stringify({name})
  })

export const deleteWallet = async walletID =>
  await request(`/api/wallets/${walletID}/delete`, {
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
export const notifyApiError = error => {
  var types = {
    400: 'warning',
    401: 'warning',
    500: 'negative'
  }
  Notify.create({
    timeout: 5000,
    type: types[error.response.status] || 'warning',
    message: error.response.data.message || null,
    caption:
      [error.response.status, ' ', error.response.statusText]
        .join('')
        .toUpperCase() || null,
    icon: null
  })
}

export const changeColorTheme = newValue => {
  document.body.setAttribute('data-theme', newValue)
  this.$q.localStorage.set('lnbits.theme', newValue)
}
