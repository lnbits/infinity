import {LocalStorage, Dark} from 'quasar'
import {createStore} from 'vuex'

import {
  changeColorTheme,
  loadSettings,
  loadWallet,
  loadUser,
  appInfo
} from './helpers'

export default createStore({
  state() {
    return {
      settings: {},
      user: null,
      wallet: null,
      app: null,
      hasListeners: {} // { [walletID]: true }
    }
  },
  mutations: {
    setSettings(state, settings) {
      state.settings = settings
    },
    setUser(state, user) {
      state.user = user
    },
    setWallet(state, wallet) {
      state.wallet = wallet
    },
    setApp(state, app) {
      state.app = app
    },
    ackListeners(state, walletID) {
      state.hasListeners[walletID] = true
    }
  },
  actions: {
    async init({dispatch, commit}) {
      const settings = await loadSettings()

      // set dark mode
      Dark.set(LocalStorage.getItem('lnbits.darkMode'))

      // failsafe if admin changes themes halfway
      if (
        LocalStorage.getItem('lnbits.theme') &&
        !settings.allowedThemes.includes(LocalStorage.getItem('lnbits.theme'))
      ) {
        console.log('allowedThemes changed by admin', settings.allowedThemes)
        changeColorTheme(settings.allowedThemes[0])
      }

      // set theme
      if (LocalStorage.getItem('lnbits.theme')) {
        document.body.setAttribute(
          'data-theme',
          LocalStorage.getItem('lnbits.theme')
        )
      }

      // commit settings
      commit('setSettings', settings)
    },
    async fetchUser({state, dispatch, commit}) {
      if (!new URLSearchParams(location.search).get('key')) return

      const user = await loadUser()
      commit('setUser', user)

      if (!state.wallet) {
        commit('setWallet', user.wallets[0])
        dispatch('fetchWallet', user.wallets[0].id)
      }
    },
    async fetchWallet({commit, dispatch}, walletID) {
      const wallet = await loadWallet(walletID)
      commit('setWallet', wallet)
      dispatch('listenForPayments')
    },
    async listenForPayments({commit, state}) {
      if (state.wallet.id in state.hasListeners) return

      // prevent listening for events of this same wallet twice
      commit('ackListeners', state.wallet.id)

      // listen for payments sent and received, and failures
      const es = new EventSource(
        `/api/wallet/sse?api-key=${state.wallet.adminkey}`
      )

      es.addEventListener('payment-sent', ev => {
        window.events.emit('payment-sent', JSON.parse(ev.data))
      })
      es.addEventListener('payment-failed', ev => {
        window.events.emit('payment-failed', JSON.parse(ev.data))
      })
      es.addEventListener('payment-received', ev => {
        window.events.emit('payment-received', JSON.parse(ev.data))
      })
    },
    async fetchApp({state, commit}, appID) {
      const app = await appInfo(appID)
      commit('setApp', app)
    }
  }
})
