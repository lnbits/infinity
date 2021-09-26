<template>
  <q-page
    v-if="$store.state.app && $store.state.wallet"
    class="q-px-md q-py-lg"
    :class="{'q-px-lg': $q.screen.gt.xs}"
  >
    <CustomAppModel
      v-for="model in $store.state.app.models"
      :key="model.name"
      :model="model"
    />
  </q-page>
</template>

<script>
export default {
  name: 'App',

  data() {
    return {}
  },

  methods: {}
}
</script>
