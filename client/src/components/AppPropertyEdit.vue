<template>
  <q-input
    v-if="field.type === 'string' || field.type === 'url'"
    filled
    dense
    :type="field.type === 'url' ? 'url' : 'text'"
    :label="fieldLabel(field)"
    :model-value="value"
    @update:model-value="handleChange($event.trim())"
  />
  <q-input
    v-if="field.type === 'number'"
    filled
    dense
    type="number"
    :label="fieldLabel(field)"
    :model-value="value"
    @update:model-value="handleChange(parseFloat($event))"
  />
  <q-input
    v-if="field.type === 'msatoshi'"
    filled
    dense
    type="text"
    suffix="satoshis"
    :label="fieldLabel(field)"
    :model-value="value > 0 ? value / 1000 : ''"
    @update:model-value="handleChange((parseInt($event) || 0) * 1000)"
  />
  <q-input
    v-if="field.type === 'currency'"
    filled
    dense
    type="text"
    :label="fieldLabel(field)"
    :model-value="value.amount > 0 ? value.amount : ''"
    @update:model-value="
      handleChange({...value, amount: parseInt($event) || 0})
    "
  >
    <template #after>
      <q-select
        label="Unit"
        filled
        dense
        options-dense
        use-input
        type="text"
        :options="currencyOptions"
        style="max-width: 200px"
        :model-value="value.unit"
        @update:model-value="handleChange({...value, unit: $event})"
        @filter="currencyFilter"
      />
    </template>
  </q-input>
  <q-select
    v-if="field.type === 'select'"
    :options="field.options"
    :label="fieldLabel(field)"
    emit-value
    filled
    dense
    :model-value="value"
    @update:model-value="handleChange($event)"
  />
  <q-toggle
    v-if="field.type === 'boolean'"
    :label="fieldLabel(field)"
    :indeterminate-value="'INDETERMINATE'"
    :model-value="value"
    @update:model-value="handleChange($event)"
  />
  <q-select
    v-if="field.type === 'ref'"
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
    :model-value="value"
    @update:model-value="handleChange($event)"
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
      type: Object,
      required: true
    },
    data: {
      type: Object,
      required: true
    }
  },

  emits: ['update:data'],

  data() {
    return {
      currencyOptions: this.$store.state.settings.currencies
    }
  },

  computed: {
    value() {
      return this.data[this.field.name]
    }
  },

  methods: {
    fieldLabel,

    currencyFilter(search, update) {
      if (search === '') {
        update(() => {
          this.currencyOptions = this.$store.state.settings.currencies
        })
        return
      }
      update(() => {
        this.currencyOptions = this.$store.state.settings.currencies.filter(
          v => v.toLowerCase().indexOf(search.toLowerCase()) !== -1
        )
      })
    },

    handleChange(value) {
      this.$emit('update:data', {...this.data, [this.field.name]: value})
    }
  }
}
</script>
