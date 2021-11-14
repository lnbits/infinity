<template>
  <q-input
    v-if="field.type === 'string' || field.type === 'url'"
    v-model.trim="value"
    filled
    dense
    :type="field.type === 'url' ? 'url' : 'text'"
    :label="fieldLabel(field)"
  />
  <q-input
    v-if="field.type === 'number'"
    v-model.number="value"
    filled
    dense
    type="number"
    :label="fieldLabel(field)"
  />
  <q-input
    v-if="field.type === 'msatoshi'"
    filled
    dense
    type="text"
    suffix="satoshis"
    :label="fieldLabel(field)"
    :model-value="value > 0 ? value / 1000 : ''"
    @update:model-value="value = (parseInt($event) || 0) * 1000"
  />
  <q-input
    v-if="field.type === 'currency'"
    filled
    dense
    type="text"
    :label="fieldLabel(field)"
    :model-value="value.amount > 0 ? value.amount : ''"
    @update:model-value="value.amount = parseInt($event) || 0"
  >
    <template #after>
      <q-select
        v-model="value.unit"
        :options="$store.state.settings.currencies"
        label="Unit"
        filled
        dense
      />
    </template>
  </q-input>
  <q-select
    v-if="field.type === 'select'"
    v-model="value"
    :options="field.options"
    :label="fieldLabel(field)"
    emit-value
    filled
    dense
  />
  <q-toggle
    v-if="field.type === 'boolean'"
    v-model="value"
    :label="fieldLabel(field)"
    :indeterminate-value="'INDETERMINATE'"
  />
  <q-select
    v-if="field.type === 'ref'"
    v-model="value"
    filled
    dense
    use-input
    emit-value
    map-options
    input-debounce="0"
    behavior="dialog"
    :options="
      items[field.ref].map(item => ({
        label: item.value[field.as],
        value: item.key
      }))
    "
    :label="fieldLabel(field)"
  />
</template>

<script>
import {fieldLabel} from '../helpers'

export default {
  props: {
    field: {
      type: Object,
      required: true
    },
    items: {
      type: Array,
      required: true
    }
  },

  data() {
    return {
      value: undefined
    }
  },

  methods: {
    fieldLabel
  }
}
</script>
