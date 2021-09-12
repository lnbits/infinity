export const loadSettings = async () =>
  await fetch('/api/settings').then(r => r.json())

export const loadUser = async () =>
  await fetch('/api/user', {
    headers: getAuthHeaders()
  }).then(r => r.json())

export const loadWallet = async walletID =>
  await fetch(`/api/wallet/${walletID}`, {
    headers: getAuthHeaders()
  }).then(r => r.json())

export const createWallet = async () =>
  await fetch('/api/create-wallet', {
    method: 'POST',
    headers: getAuthHeaders() // fine if not logged, we will just create a new user
  }).then(r => r.json())

const getAuthHeaders = () => ({
  'X-MasterKey': new URLSearchParams(location.search).get('key')
})

export const changeColorTheme = newValue => {
  document.body.setAttribute('data-theme', newValue)
  this.$q.localStorage.set('lnbits.theme', newValue)
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
