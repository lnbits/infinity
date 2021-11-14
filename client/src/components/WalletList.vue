<template>
  <q-list
    v-if="$store.state.user && $store.state.user.wallets.length"
    dense
    class="lnbits-drawer__q-list"
  >
    <q-item-label header>Wallets</q-item-label>
    <q-item
      v-for="wallet in $store.state.user.wallets"
      :key="wallet.id"
      :active="isActive(wallet)"
      clickable
      tag="a"
      @click="goToWallet(wallet)"
    >
      <q-item-section side>
        <q-avatar
          size="md"
          :color="
            isActive(wallet)
              ? $q.dark.isActive
                ? 'primary'
                : 'primary'
              : 'grey-5'
          "
        >
          <q-icon
            name="flash_on"
            :size="$q.dark.isActive ? '21px' : '20px'"
            :color="$q.dark.isActive ? 'blue-grey-10' : 'grey-3'"
          ></q-icon>
        </q-avatar>
      </q-item-section>
      <q-item-section>
        <q-item-label lines="1">{{ wallet.name }}</q-item-label>
        <q-item-label caption
          >{{ formatMsatToSat(wallet.balance) }} sat</q-item-label
        >
      </q-item-section>
      <q-item-section
        v-show="isActive(wallet) && $route.name === 'wallet'"
        side
      >
        <q-icon name="chevron_right" color="grey-5" size="md"></q-icon>
      </q-item-section>
    </q-item>
    <q-item clickable @click="showForm = !showForm">
      <q-item-section side>
        <q-icon
          :name="showForm ? 'remove' : 'add'"
          color="grey-5"
          size="md"
        ></q-icon>
      </q-item-section>
      <q-item-section>
        <q-item-label lines="1" class="text-caption">Add a wallet</q-item-label>
      </q-item-section>
    </q-item>
    <q-item v-if="showForm">
      <q-item-section>
        <q-form @submit="createWallet">
          <q-input v-model="newWalletName" filled dense label="Name wallet *">
            <template #append>
              <q-btn
                round
                dense
                flat
                icon="send"
                size="sm"
                type="submit"
                :disable="newWalletName === ''"
                @click="createWallet"
              ></q-btn>
            </template>
          </q-input>
        </q-form>
      </q-item-section>
    </q-item>
  </q-list>
</template>

<script>
import {createWallet} from '../api'
import {notifyError, formatMsatToSat} from '../helpers'

export default {
  data() {
    return {
      showForm: false,
      newWalletName: ''
    }
  },
  methods: {
    formatMsatToSat,

    isActive(wallet) {
      return this.$store.state.wallet?.id === wallet.id
    },

    async createWallet() {
      try {
        const {wallet} = await createWallet(this.newWalletName)
        this.newWalletName = ''
        this.$store.commit('setWallet', wallet)
        this.$router.push({
          path: `/wallet/${wallet.id}`,
          query: this.$route.query
        })
        this.$store.dispatch('fetchUser')
      } catch (err) {
        notifyError(err)
      }
    },

    goToWallet(wallet) {
      if (wallet.id === this.$route.params.id && !this.$oute.params.appid)
        return

      this.$store.commit('setWallet', wallet)
      this.$router.push({
        path: `/wallet/${wallet.id}`,
        query: this.$route.query
      })
    }
  }
}
</script>
