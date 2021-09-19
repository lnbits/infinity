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
      clickable
      :active="$store.state.wallet && $store.state.wallet.id === wallet.id"
      tag="a"
      :href="wallet.url"
    >
      <q-item-section side>
        <q-avatar
          size="md"
          :color="
            activeWallet && activeWallet.id === wallet.id
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
        <q-item-label caption>{{ wallet.balance }} sat</q-item-label>
      </q-item-section>
      <q-item-section
        v-show="activeWallet && activeWallet.id === wallet.id"
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
          <q-input v-model="walletName" filled dense label="Name wallet *">
            <template #append>
              <q-btn
                round
                dense
                flat
                icon="send"
                size="sm"
                :disable="walletName === ''"
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
import {createWallet, notifyApiError} from '../helpers'

export default {
  data() {
    return {
      showForm: false,
      walletName: ''
    }
  },
  methods: {
    methods: {
      async createWallet() {
        try {
          const {wallet} = await createWallet(this.walletName)
          this.$store.commit('setWallet', wallet)
          this.$router.push({
            path: `/wallet/${wallet.id}`,
            query: this.$route.query
          })
        } catch (err) {
          notifyApiError(err)
        }
      }
    }
  }
}
</script>
