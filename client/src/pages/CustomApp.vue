<template>
  <q-page
    v-if="$store.state.app && $store.state.wallet"
    class="q-px-md q-py-lg"
    :class="{'q-px-lg': $q.screen.gt.xs}"
  >
    <div class="row q-col-gutter-md">
      <div class="col-12 col-md-7 q-gutter-y-md">
        <CustomAppModel
          v-for="model in $store.state.app?.models"
          :key="model.name"
          :model="model"
          :items="($store.state.app || {}).items || []"
          :items-map="itemsMap"
        />

        <CustomAppActions
          v-if="$store.state.app?.actions"
          :actions="$store.state.app?.actions"
        />
      </div>

      <div class="col-12 col-md-5 q-gutter-y-md">
        <q-list>
          <q-expansion-item
            group="extras"
            icon="swap_vertical_circle"
            label="App Info"
            default-opened
            :content-inset-level="0.5"
          >
            <q-card-section>
              <q-item>
                <q-item-section>
                  <q-item-label overline>Title</q-item-label>
                  <q-item-label>
                    {{ $store.state.app.title }}
                  </q-item-label>
                </q-item-section>
              </q-item>

              <q-item v-if="$store.state.app.description">
                <q-item-section>
                  <q-item-label overline>Description</q-item-label>
                  <q-item-label
                    :style="{
                      paddingLeft: '10px',
                      fontWeight: 'lighter',
                      borderLeft: '6px solid'
                    }"
                    class="markdown"
                    v-html="markdownDescription"
                  ></q-item-label>
                </q-item-section>
              </q-item>

              <q-item>
                <q-item-section>
                  <q-item-label overline>ID</q-item-label>
                  <q-item-label>
                    {{ $store.state.app.id }}
                  </q-item-label>
                </q-item-section>
              </q-item>

              <q-item v-ripple clickable>
                <q-menu anchor="top right" self="top left">
                  <q-list style="min-width: 100px">
                    <q-item
                      v-close-popup
                      clickable
                      @click="goToURL($store.state.app.url)"
                    >
                      <q-item-section>Open App URL</q-item-section>
                    </q-item>
                  </q-list>
                  <q-list v-if="$store.state.app.files">
                    <q-item-label header>Static Resources</q-item-label>
                    <q-item
                      v-for="(file, match) in $store.state.app.files"
                      :key="match"
                      v-close-popup
                      clickable
                      @click="
                        goToURL(
                          file.startsWith('http') || file.startsWith('/')
                            ? file
                            : `/app/${$store.state.wallet.id}/${$store.state.app.id}/${file}`
                        )
                      "
                    >
                      <q-item-section>{{ file }}</q-item-section>
                      <q-item-section side caption>{{ match }}</q-item-section>
                    </q-item>
                  </q-list>
                </q-menu>

                <q-item-section>
                  <q-item-label overline>URL</q-item-label>
                  <q-item-label>
                    {{ $store.state.app.url }}
                  </q-item-label>
                </q-item-section>
              </q-item>
            </q-card-section>
          </q-expansion-item>

          <q-separator />

          <q-expansion-item
            group="extras"
            icon="code"
            label="App Code"
            :content-inset-level="0.5"
          >
            <q-card-section>
              <pre><code>{{ $store.state.app.code }}</code></pre>
            </q-card-section>
          </q-expansion-item>
        </q-list>
      </div>
    </div>
  </q-page>
</template>

<script>
import {md} from '../helpers'

export default {
  name: 'App',

  data() {
    return {}
  },

  computed: {
    itemsMap() {
      if (!this.$store.state.app) return {}

      const map = {}

      Object.entries(this.$store.state.app.items).forEach(
        ([modelName, items]) => {
          map[modelName] = {}

          items.forEach(item => {
            map[modelName][item.key] = item
          })
        }
      )

      return map
    },

    markdownDescription() {
      if (!this.$store.state.app?.description) return ''

      return md
        .render(this.$store.state.app.description)
        .replace('$appBase', match =>
          location.pathname.replace('/app/', '/').replace('/wallet/', '/app/')
        )
        .replace('<a href="', '<a target="_blank" href="')
    }
  },

  methods: {
    goToURL: url => {
      window.open(url)
    }
  }
}
</script>
