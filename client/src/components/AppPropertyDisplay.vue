<template>
  <q-btn
    v-if="field.type === 'url'"
    flat
    dense
    size="xs"
    icon="link"
    color="light-blue"
    :title="value"
    @click.stop="
      goToURL(
        value.startsWith('http') || value.startsWith('/')
          ? value
          : `/app/${$store.state.wallet.id}/${$store.state.app.id}/${value}`
      )
    "
  ></q-btn>
  <span v-else-if="field.type === 'msatoshi'">
    {{ formatMsatToSat(value) }} sat
  </span>
  <span v-else-if="field.type === 'currency'">
    {{ value.amount }}
    {{ value.unit }}
  </span>
  <span v-else-if="field.type === 'boolean'">
    <q-icon size="sm" :name="value ? 'check_box' : 'check_box_outline_blank'" />
  </span>
  <span v-else-if="field.type === 'ref'">
    {{
      itemsMap[field.ref] &&
      itemsMap[field.ref][value] &&
      itemsMap[field.ref][value].value[field.as]
    }}
  </span>
  <span v-else>{{ value }}</span>
</template>

<script>
import {formatMsatToSat} from '../helpers'

export default {
  props: {
    field: {
      type: Object,
      required: true
    },
    value: {
      type: Object,
      required: true
    },
    itemsMap: {
      type: Object,
      required: true
    }
  },

  methods: {
    formatMsatToSat,

    goToURL: url => {
      window.open(url)
    }
  }
}
</script>
