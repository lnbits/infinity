<template>
  <q-list v-if="$store.state.user" dense class="lnbits-drawer__q-list">
    <q-item-label header>Apps</q-item-label>
    <q-item
      v-for="app in $store.state.user.apps"
      :key="app"
      :active="isActive(app)"
      clickable
      tag="a"
      @click="clickApp(app)"
    >
      <q-item-section side>
        <q-avatar
          size="md"
          :color="
            isActive(app)
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
        <q-item-label lines="1">
          {{ appDisplayName(app) }}
          <q-menu context-menu>
            <q-list dense style="min-width: 100px">
              <q-item v-close-popup clickable @click="refresh(app)">
                <q-item-section>Refresh Cache</q-item-section>
              </q-item>
              <q-item v-close-popup clickable @click="remove(app)">
                <q-item-section>Remove</q-item-section>
              </q-item>
            </q-list>
          </q-menu>
        </q-item-label>
      </q-item-section>
      <q-item-section v-show="isActive(app)" side>
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
        <q-form @submit="add">
          <q-input v-model="appURL" filled dense label="App URL *">
            <template #append>
              <q-btn
                round
                dense
                flat
                icon="send"
                size="sm"
                :disable="appURL === ''"
                @click="add"
              ></q-btn>
            </template>
          </q-input>
        </q-form>
      </q-item-section>
    </q-item>
  </q-list>
</template>

<script>
import {addApp, removeApp, appRefresh} from '../api'
import {appDisplayName, notifyError} from '../helpers'

export default {
  data() {
    return {
      showForm: false,
      appURL: ''
    }
  },

  methods: {
    appDisplayName,

    isActive(app) {
      return this.$store.state.app?.url === app && this.$route.name === 'app'
    },

    clickApp(appURL) {
      this.$router.push({
        path: `/wallet/${this.$store.state.wallet.id}/app/${btoa(appURL)}`,
        query: this.$route.query
      })
    },

    async add() {
      try {
        await addApp(this.appURL)
        this.$store.dispatch('fetchUser')
        this.$router.push({
          path: `/wallet/${this.$store.state.wallet.id}/app/${btoa(
            this.appURL
          )}`,
          query: this.$route.query
        })
        this.appURL = ''
      } catch (err) {
        notifyError(err)
      }
    },

    async remove(appURL) {
      try {
        await removeApp(appURL)
        this.$store.dispatch('fetchUser')
        this.$router.push({
          path: `/wallet/${this.$store.state.wallet.id}`,
          query: this.$route.query
        })
      } catch (err) {
        notifyError(err)
      }
    },

    async refresh(appURL) {
      try {
        await appRefresh(btoa(appURL))
        if (
          this.$route.pathname.indexOf('/app/') !== -1 &&
          this.$store.state.app?.url === appURL
        ) {
          this.$store.dispatch('fetchApp', this.$store.state.app.id)
        }
      } catch (err) {
        notifyError(err)
      }
    }
  }
}
</script>
