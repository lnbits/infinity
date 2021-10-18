<template>
  <q-page
    v-if="$store.state.app && $store.state.wallet"
    class="q-px-md q-py-lg"
    :class="{'q-px-lg': $q.screen.gt.xs}"
  >
    <CustomAppModel
      v-for="model in $store.state.app?.models"
      :key="model.name"
      :model="model"
      :items="($store.state.app || {}).items || []"
      :items-map="itemsMap"
    />
  </q-page>
</template>

<script>
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
    }
  }
}
</script>
