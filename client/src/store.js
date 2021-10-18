import {LocalStorage, Dark} from 'quasar'
import {createStore} from 'vuex'

import {changeColorTheme, notifyError} from './helpers'
import {loadSettings, loadWallet, loadUser, appInfo} from './api'

export default createStore({
  state() {
    return {
      settings: {},
      user: null,
      wallet: null,
      app: null,
      hasListeners: {} // { [walletID]: true },
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
      if (!wallet) return
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

      let user
      try {
        user = await loadUser()
      } catch (err) {
        notifyError(err)
      }
      commit('setUser', user)

      if (!state.wallet) {
        commit('setWallet', user.wallets[0])
        dispatch('fetchWallet', user.wallets[0].id)
      }
    },
    async fetchWallet({state, commit, dispatch}, walletID) {
      try {
        const wallet = await loadWallet(walletID)
        commit('setWallet', wallet)
        dispatch('listen')
      } catch (err) {
        notifyError(err)
      }
    },
    async fetchApp({state, commit}, appID) {
      try {
        const app = await appInfo(appID)
        commit('setApp', app)
      } catch (err) {
        notifyError(err)
      }
    },
    async listen({commit, state, dispatch}) {
      if (state.wallet.id in state.hasListeners) return

      // prevent listening for events of this same wallet twice
      commit('ackListeners', state.wallet.id)

      // listen for payments sent and received, and failures
      const payments = new EventSource(
        `/api/wallet/sse?api-key=${state.wallet.adminkey}`
      )

      payments.addEventListener('payment-sent', ev => {
        const payment = JSON.parse(ev.data)
        window.events.emit('payment-sent', payment)
        dispatch('fetchWallet', payment.walletID)
        dispatch('fetchUser')
      })
      payments.addEventListener('payment-failed', ev => {
        const payment = JSON.parse(ev.data)
        window.events.emit('payment-failed', payment)
        dispatch('fetchWallet', payment.walletID)
      })
      payments.addEventListener('payment-received', ev => {
        const payment = JSON.parse(ev.data)
        window.events.emit('payment-received', payment)
        dispatch('fetchWallet', payment.walletID)
        dispatch('fetchUser')
      })

      // listen for app db changes (all apps for this wallet)
      const apps = new EventSource(
        `/api/wallet/apps/sse?api-key=${state.wallet.adminkey}`
      )
      apps.addEventListener('item', ev => {
        const item = JSON.parse(ev.data)
        window.events.emit('item', item)
      })
    }
  }
})
