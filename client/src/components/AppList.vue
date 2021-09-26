<template>
  <q-list v-if="$store.state.user" dense class="lnbits-drawer__q-list">
    <q-item-label header>Apps</q-item-label>
    <q-item
      v-for="app in $store.state.user.apps"
      :key="app"
      clickable
      :active="$store.app?.id === app"
      tag="a"
      @click="clickApp(app)"
    >
      <q-item-section side>
        <q-avatar
          size="md"
          :color="
            $store.app?.id === app
              ? $q.dark.isActive
                ? 'primary'
                : 'primary'
              : 'grey-5'
          "
        >
          <q-icon
            :name="app.icon"
            :size="$q.dark.isActive ? '21px' : '20px'"
            :color="$q.dark.isActive ? 'blue-grey-10' : 'grey-3'"
          ></q-icon>
        </q-avatar>
      </q-item-section>
      <q-item-section>
        <q-item-label lines="1">{{
          app.split('/').slice(-1)[0].slice(0, -4)
        }}</q-item-label>
      </q-item-section>
      <q-item-section v-show="app.isActive" side>
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
        <q-item-label lines="1" class="text-caption">Add an app</q-item-label>
      </q-item-section>
    </q-item>
    <q-item v-if="showForm">
      <q-item-section>
        <q-form @submit="addApp">
          <q-input v-model="appURL" filled dense label="App URL *">
            <template #append>
              <q-btn
                round
                dense
                flat
                icon="send"
                size="sm"
                :disable="appURL === ''"
                @click="addApp"
              ></q-btn>
            </template>
          </q-input>
        </q-form>
      </q-item-section>
    </q-item>
  </q-list>
</template>

<script>
import {Buffer} from 'buffer'

import {addApp, notifyError} from '../helpers'

export default {
  data() {
    return {
      showForm: false,
      appURL: ''
    }
  },
  methods: {
    clickApp(appURL) {
      this.$router.push(
        `/wallet/${this.$store.state.wallet.id}/app/${Buffer.from(
          appURL,
          'utf-8'
        ).toString('hex')}`,
        {
          query: this.$route.query
        }
      )
    },

    async addApp() {
      try {
        await addApp(this.appURL)
        const appid = Buffer.from(this.appURL, 'utf-8').toString('hex')
        this.$store.dispatch('fetchUser')
        this.$router.push(
          `/wallet/${this.$store.state.wallet.id}/app/${appid}`,
          {
            query: this.$route.query
          }
        )
        this.$store.dispatch('fetchUser')
        this.appURL = ''
      } catch (err) {
        notifyError(err)
      }
    }
  }
}
</script>
