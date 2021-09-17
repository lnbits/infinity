import {LocalStorage, Dark} from 'quasar'
import {createStore} from 'vuex'

import {changeColorTheme, loadSettings, loadWallet, loadUser} from './helpers'

export default createStore({
  state() {
    return {
      settings: {},
      user: null,
      wallet: null
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

      // listeners
      dispatch('listenForPayments')
    },
    async fetchUser({dispatch, commit}) {
      if (!new URLSearchParams(location.search).get('key')) return

      const user = await loadUser()
      commit('setUser', user)
    },
    async fetchWallet({commit}, walletID) {
      const wallet = await loadWallet(walletID)
      commit('setWallet', wallet)
    },
    async listenForPayments({dispatch}) {
      // TODO: listen for payments sent and received, and failures
      // call callbacks
      // refresh wallet
    }
  }
})
