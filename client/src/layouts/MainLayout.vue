<template>
  <q-layout v-cloak id="vue" view="hHh lpR lfr">
    <q-header bordered class="bg-marginal-bg">
      <q-toolbar>
        <q-btn
          v-if="$store.state.user"
          dense
          flat
          round
          icon="menu"
          @click="visibleDrawer = !visibleDrawer"
        ></q-btn>
        <q-toolbar-title>
          <q-btn flat no-caps dense size="lg" type="a" href="/">
            <span v-if="$store.state.settings.siteTitle">
              {{ $store.state.settings.siteTitle }}
            </span>
            <span v-else> <strong>LN</strong>bits </span>
          </q-btn>
        </q-toolbar-title>
        <q-badge color="yellow" text-color="black" class="q-mr-md">
          <span
            ><span v-show="$q.screen.gt.sm"
              >USE WITH CAUTION - {{ $store.state.settings.siteTitle }} wallet
              is still in </span
            >BETA</span
          >
        </q-badge>
        <q-btn-dropdown
          v-if="
            $store.state.settings.allowedThemes &&
            $store.state.settings.allowedThemes.length > 1
          "
          dense
          flat
          round
          size="sm"
          icon="dashboard_customize"
          class="q-pl-md"
        >
          <div class="row no-wrap q-pa-md">
            <q-btn
              v-if="$store.state.settings.allowedThemes.includes('classic')"
              dense
              flat
              icon="format_color_fill"
              color="deep-purple"
              size="md"
              @click="changeColor('classic')"
              ><q-tooltip>classic</q-tooltip> </q-btn
            ><q-btn
              v-if="$store.state.settings.allowedThemes.includes('mint')"
              dense
              flat
              icon="format_color_fill"
              color="green"
              size="md"
              @click="changeColor('mint')"
              ><q-tooltip>mint</q-tooltip> </q-btn
            ><q-btn
              v-if="$store.state.settings.allowedThemes.includes('autumn')"
              dense
              flat
              icon="format_color_fill"
              color="brown"
              size="md"
              @click="changeColor('autumn')"
              ><q-tooltip>autumn</q-tooltip>
            </q-btn>
            <q-btn
              v-if="$store.state.settings.allowedThemes.includes('monochrome')"
              dense
              flat
              icon="format_color_fill"
              color="grey"
              size="md"
              @click="changeColor('monochrome')"
              ><q-tooltip>monochrome</q-tooltip>
            </q-btn>
            <q-btn
              v-if="$store.state.settings.allowedThemes.includes('salvador')"
              dense
              flat
              icon="format_color_fill"
              color="blue-10"
              size="md"
              @click="changeColor('salvador')"
              ><q-tooltip>elSalvador</q-tooltip>
            </q-btn>
            <q-btn
              v-if="$store.state.settings.allowedThemes.includes('flamingo')"
              dense
              flat
              icon="format_color_fill"
              color="pink-3"
              size="md"
              @click="changeColor('flamingo')"
              ><q-tooltip>flamingo</q-tooltip>
            </q-btn>
          </div>
        </q-btn-dropdown>

        <q-btn
          dense
          flat
          round
          size="sm"
          :icon="$q.dark.isActive ? 'brightness_3' : 'wb_sunny'"
          @click="toggleDarkMode"
        >
          <q-tooltip>Toggle Dark Mode</q-tooltip>
        </q-btn>
      </q-toolbar>
    </q-header>

    <q-drawer
      v-if="$store.state.user"
      v-model="visibleDrawer"
      side="left"
      :width="$q.screen.lt.md ? 260 : 230"
      show-if-above
      :elevated="$q.screen.lt.md"
    >
      <WalletList />
      <AppList class="q-pb-xl" />
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>

    <q-footer
      class="bg-transparent q-px-lg q-py-md"
      :class="{'text-dark': !$q.dark.isActive}"
    >
      <q-toolbar>
        <q-toolbar-title class="text-caption">
          {{ $store.state.settings.siteTitle }},
          {{ $store.state.settings.siteTagline }}
          <br />
          <small>Commit version: {{ $store.state.settings.siteVersion }}</small>
        </q-toolbar-title>
        <q-space></q-space>
        <q-btn
          flat
          dense
          :color="$q.dark.isActive ? 'white' : 'primary'"
          icon="code"
          type="a"
          href="https://github.com/lnbits/lnbits"
          target="_blank"
          rel="noopener"
        >
          <q-tooltip>View project in GitHub</q-tooltip>
        </q-btn>
      </q-toolbar>
    </q-footer>
  </q-layout>
</template>

<script>
import {changeColorTheme} from '../helpers'

export default {
  name: 'MainLayout',

  data() {
    return {
      visibleDrawer: false
    }
  },

  beforeCreate() {
    this.$store.dispatch('init')
    this.$store.dispatch('fetchUser')
  },

  methods: {
    changeColor(newValue) {
      changeColorTheme(newValue)
    },

    toggleDarkMode() {
      this.$q.dark.toggle()
      this.$q.localStorage.set('lnbits.darkMode', this.$q.dark.isActive)
    }
  }
}
</script>
