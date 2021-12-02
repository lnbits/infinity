import {LocalStorage, exportFile, Notify, copyToClipboard} from 'quasar'
import MarkdownIt from 'markdown-it'

export const paramDefaults = fields => {
  return Object.fromEntries(
    fields
      .filter(field => !field.computed)
      .map(field => [
        field.name,
        field.default ||
          (
            {
              currency: () => ({amount: 0, unit: 'sat'}),
              boolean: () => false
            }[field.type] || (() => {})
          )()
      ])
  )
}

export const formatMsatToSat = msat => {
  const sat = msat / 1000
  const satInt = parseInt(sat)
  if (sat - satInt > 0) return sat.toFixed(3)
  else return satInt.toString()
}

export const formatDate = (timestamp, full) => {
  if (full)
    return new Date(timestamp * 1000)
      .toISOString()
      .split('.')[0]
      .replace('T', ' ')

  const now = Date.now() / 1000
  const delta = now - timestamp
  if (delta < 60 * 100) return parseInt((now - timestamp) / 60) + ' minutes ago'
  if (delta < 48 * 60 * 60)
    return parseInt((now - timestamp) / (60 * 60)) + ' hours ago'

  return new Date(timestamp * 1000).toISOString().split('T')[0]
}

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
    timeout: ((caption + (error.message || '')).length / 12) * 1000,
    type,
    message: error.message || null,
    caption,
    actions: [{icon: 'close'}]
  })
}

export const changeColorTheme = newValue => {
  document.body.setAttribute('data-theme', newValue)
  LocalStorage.set('lnbits.theme', newValue)
}

export const copyText = (text, message, position) => {
  copyToClipboard(text).then(function () {
    Notify.create({
      message: message || 'Copied to clipboard!',
      position: position || 'bottom'
    })
  })
}

export const appDisplayName = url => {
  if (url.startsWith('http') && url.endsWith('.lua'))
    return url.split('/').slice(-1)[0].slice(0, -4)
  return url
}

export const appURLToId = url =>
  btoa(url).replace(/=/g, '').replace(/\+/g, '-').replace(/\//g, '_')

export const fieldLabel = field =>
  (field.display || field.name) + (field.required ? ' *' : '')

export const md = MarkdownIt({
  linkify: true
})

md.linkify
  .set({
    fuzzyEmail: false,
    fuzzyLink: true
  })
  .tlds('onion', true)
