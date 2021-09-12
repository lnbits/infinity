import {Quasar} from 'quasar'
import {createStore} from 'vuex'

import {
  changeColorTheme,
  loadSettings,
  loadWallet,
  loadUser,
  createWallet
} from './helpers'

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
      dispatch('getUser')
      dispatch('getWallet')

      const settings = await loadSettings()

      // set dark mode
      Quasar.dark.set(Quasar.localStorage.getItem('lnbits.darkMode'))

      // failsafe if admin changes themes halfway
      if (
        Quasar.localStorage.getItem('lnbits.theme') &&
        !settings.allowedThemes.includes(
          Quasar.localStorage.getItem('lnbits.theme')
        )
      ) {
        console.log('allowedThemes changed by admin', settings.allowedThemes)
        changeColorTheme(settings.allowedThemes[0])
      }

      // set theme
      if (Quasar.localStorage.getItem('lnbits.theme')) {
        document.body.setAttribute(
          'data-theme',
          Quasar.localStorage.getItem('lnbits.theme')
        )
      }

      // commit settings
      commit('setSettings', settings)
    },
    async getUser({commit}) {
      const user = await loadUser()
      commit('setUser', user)
      const wallet = await loadWallet(user.wallets[0])
      commit('setWallet', wallet)
    },
    async createWallet({state, commit}, {name}) {
      const {userMasterKey, wallet} = await createWallet({name})
      if (state.user) {
        commit('setWallet', wallet)
      } else {
        location.href = `/?key=${userMasterKey}`
      }
    }
  }
})
