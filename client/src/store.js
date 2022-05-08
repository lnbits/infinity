import {LocalStorage, Dark} from 'quasar'
import {createStore} from 'vuex'

import {notifyError} from './helpers'
import {loadSettings, loadWallet, loadUser, appInfo, listAppItems} from './api'

export default createStore({
  state() {
    return {
      settings: {},
      user: null,
      wallet: null,
      app: null,
      hasListeners: {}, // { [walletID]: true },
      debugMessages: []
    }
  },
  getters: {
    getDebugMessages (state) {
      return state.debugMessages
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
    setAppItems(state, {modelName, items}) {
      if (!state.app) return
      state.app.items[modelName] = items
    },
    ackListeners(state, walletID) {
      state.hasListeners[walletID] = true
    },
    addDebugMessage(state, message) {
      const date = new Date(message.time)
      state.debugMessages.push({text: message.text, time: date.toLocaleString()})
    }
  },
  actions: {
    async init({dispatch, commit}) {
      const settings = await loadSettings()

      // set dark mode
      Dark.set(LocalStorage.getItem('lnbits.darkMode'))

      // set theme
      if (LocalStorage.getItem('lnbits.theme')) {
        document.body.setAttribute(
          'data-theme',
          LocalStorage.getItem('lnbits.theme')
        )
      }

      // commit settings
      commit('setSettings', settings)

      // set title
      document.querySelector('title').innerHTML = settings.siteTitle
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
        app.items = Object.fromEntries(
          await Promise.all(
            app.models.map(async model => [
              model.name,
              await listAppItems(app.url, model.name)
            ])
          )
        )
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
        `/api/wallet/app/sse?api-key=${state.wallet.adminkey}`
      )
      apps.addEventListener('print', ev => {
        const item = JSON.parse(ev.data)

        if (
          !state.app ||
          item.wallet_id !== state.wallet.id ||
          item.app !== state.app.url
        ) {
          return
        }
        commit('addDebugMessage', {text: item.values, time: item.time})
      })

      apps.addEventListener('item', ev => {
        const item = JSON.parse(ev.data)

        if (
          !state.app ||
          item.walletID !== state.wallet.id ||
          item.app !== state.app.url ||
          !(item.model in state.app.items)
        ) {
          return
        }

        const items = state.app.items[item.model]
        const index = items.findIndex(({key}) => item.key === key)
        if (!item.value && index !== -1) {
          // deleted
          commit('setAppItems', {
            modelName: item.model,
            items: [...items.slice(0, index), ...items.slice(index + 1)]
          })
        } else if (index !== -1) {
          // updated
          commit('setAppItems', {
            modelName: item.model,
            items: [...items.slice(0, index), item, ...items.slice(index + 1)]
          })
        } else {
          // added
          commit('setAppItems', {
            modelName: item.model,
            items: [...items, item]
          })
        }
      })
    }
  }
})
