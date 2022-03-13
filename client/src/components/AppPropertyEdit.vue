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
  <q-input
    v-if="field.type === 'datetime'"
    readonly
    filled
    dense
    :model-value="formatDate(value, true)"
    :label="fieldLabel(field)"
  >
    <template #append>
      <q-icon name="event" class="cursor-pointer">
        <q-popup-proxy cover transition-show="scale" transition-hide="scale">
          <q-date
            mask="datetimeMask"
            :model-value="formatTimestampMask(value)"
            @update:model-value="handleChangeDateTime"
          >
            <div class="row items-center justify-end">
              <q-btn v-close-popup label="Done" color="primary" flat />
            </div>
          </q-date>
        </q-popup-proxy>
      </q-icon>

      <q-icon name="access_time" class="cursor-pointer">
        <q-popup-proxy cover transition-show="scale" transition-hide="scale">
          <q-time
            mask="datetimeMask"
            format24h
            :model-value="formatTimestampMask(value)"
            @update:model-value="handleChangeDateTime"
          >
            <div class="row items-center justify-end">
              <q-btn v-close-popup label="Done" color="primary" flat />
            </div>
          </q-time>
        </q-popup-proxy>
      </q-icon>
    </template>
  </q-input>
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
import {date} from 'quasar'

import {formatDate, fieldLabel} from '../helpers'

const datetimeMask = 'ddd DD MMM YYYY HH:mm:ss'

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
      datetimeMask,
      currencyOptions: this.$store.state.settings.currencies
    }
  },

  computed: {
    value() {
      return this.data[this.field.name]
    }
  },

  methods: {
    formatDate,
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
      if (value)
        this.$emit('update:data', {...this.data, [this.field.name]: value})
    },

    handleChangeDateTime(...args) {
      var current = new Date(this.value * 1000)
      let change = args[args.length - 1]
      delete change.timezoneOffset
      delete change.dateHash
      delete change.timeHash
      delete change.changed
      let updated = date.adjustDate(current, change)
      this.handleChange(Math.round(updated.getTime() / 1000))
    },

    formatTimestampMask(ts) {
      return date.formatDate(new Date(ts * 1000), datetimeMask)
    }
  }
}
</script>
