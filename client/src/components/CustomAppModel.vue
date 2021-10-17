<template>
  <div class="row q-col-gutter-md q-mb-md">
    <div class="col-12 col-md-7 q-gutter-y-md">
      <q-card>
        <q-card-section>
          <q-btn unelevated color="primary" @click="openCreateDialog"
            >New {{ model.display || model.name }}</q-btn
          >
        </q-card-section>
        <q-card-section>
          <div class="row items-center no-wrap q-mb-md">
            <div class="col">
              <h5 class="text-subtitle1 q-my-none">
                {{
                  model.displayPlural || `${model.display}s` || `${model.name}s`
                }}
              </h5>
            </div>
          </div>

          <q-table
            v-model:pagination="table.pagination"
            dense
            flat
            binary-state-sort
            column-sort-order="da"
            :rows="items"
            row-key="key"
          >
            <template #header="props">
              <q-tr :props="props">
                <q-th
                  v-for="field in model.fields.filter(f => !f.hidden)"
                  :key="field.name"
                  auto-width
                  >{{ field.display || field.name }}</q-th
                >
                <q-th auto-width></q-th>
                <q-th auto-width></q-th>
              </q-tr>
            </template>
            <template #body="props">
              <q-tr :props="props">
                <q-td
                  v-for="field in model.fields"
                  :key="field.name"
                  auto-width
                >
                  <a
                    v-if="field.type === 'url'"
                    target="_blank"
                    :href="props.row.value[field.name]"
                  >
                    <q-icon name="link" />
                  </a>
                  <span v-else-if="props.type === 'ref'">
                    {{ json(refItemsMap[props.row.value[field.name]]) }}
                  </span>
                  <span v-else>{{ props.row.value[field.name] }}</span>
                </q-td>
                <q-td auto-width>
                  <q-btn
                    flat
                    dense
                    size="xs"
                    icon="edit"
                    color="light-blue"
                    @click="openUpdateDialog(props.row.key)"
                  ></q-btn>
                  <q-btn
                    flat
                    dense
                    size="xs"
                    icon="cancel"
                    color="pink"
                    @click="deleteItem(props.row.key)"
                  ></q-btn>
                </q-td>
              </q-tr>
            </template>
          </q-table>
        </q-card-section>
      </q-card>
    </div>
  </div>

  <q-dialog v-model="dialog.show" @hide="closeFormDialog">
    <q-card class="q-pa-lg q-pt-xl lnbits__dialog-card">
      <q-form class="q-gutter-md" @submit="saveItem">
        <template
          v-for="field in model.fields.filter(f => !f.hidden && !f.computed)"
          :key="field.name"
        >
          <q-input
            v-if="field.type === 'string' || field.type === 'url'"
            v-model.trim="dialog.item[field.name]"
            filled
            dense
            :type="field.type === 'url' ? 'url' : 'text'"
            :label="fieldLabel(field)"
          />
          <q-input
            v-if="field.type === 'number'"
            v-model.number="dialog.item[field.name]"
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
            :model-value="
              dialog.item[field.name] > 0 ? dialog.item[field.name] / 1000 : ''
            "
            @update:model-value="
              dialog.item[field.name] = (parseInt($event) || 0) * 1000
            "
          />
          <q-toggle
            v-if="field.type === 'boolean'"
            v-model="dialog.item[field.name]"
            :label="fieldLabel(field)"
          />
          <q-select
            v-if="field.type === 'ref'"
            v-model="dialog.item[field.name]"
            filled
            use-input
            input-debounce="0"
            behavior="dialog"
            :options="refItemsFiltered[field.ref]"
            :label="fieldLabel(field)"
            @filter="refOptionsFilter(field.ref)"
          />
        </template>
        <div class="row q-mt-lg">
          <q-btn v-if="dialog.item.key" unelevated color="primary" type="submit"
            >Update {{ model.name }}</q-btn
          >
          <q-btn
            v-else
            unelevated
            color="primary"
            :disabled="isFormSubmitDisabled"
            type="submit"
            >Create {{ model.name }}</q-btn
          >
          <q-btn v-close-popup flat color="grey" class="q-ml-auto"
            >Cancel</q-btn
          >
        </div>
      </q-form>
    </q-card>
  </q-dialog>
</template>

<script>
import {listAppItems, setAppItem, addAppItem, delAppItem} from '../api'
import {notifyError} from '../helpers'

export default {
  props: {
    model: {
      type: Object,
      required: true
    }
  },

  data() {
    return {
      items: [],
      table: {
        pagination: {
          rowsPerPage: 15,
          sortBy: 'created_at'
        }
      },
      dialog: {
        show: false,
        item: null
      },
      refItems: {},
      refItemsFiltered: {}
    }
  },

  computed: {
    refItemsMap() {
      const map = {}

      Object.values(this.refItems).forEach(items => {
        items.forEach(item => {
          map[item.key] = item
        })
      })

      return map
    },

    isFormSubmitDisabled() {
      return (
        this.dialog.show &&
        this.model.fields
          .filter(field => field.required)
          .filter(
            field =>
              this.dialog.item[field.name] === undefined ||
              this.dialog.item[field.name] === ''
          ).length > 0
      )
    }
  },

  mounted() {
    this.loadItems()
  },

  methods: {
    json: JSON.stringify,

    fieldLabel(field) {
      return (field.display || field.name) + (field.required ? ' *' : '')
    },

    async fetchRefItems(modelName) {
      if (!this.refItems[modelName]) {
        this.refItems[modelName] = await listAppItems(
          this.$store.state.app.id,
          modelName
        )
      }
    },

    refOptionsFilter(modelName) {
      return async (val, update, abort) => {
        update(() => {
          this.refItemsFiltered[modelName] = this.refItems[modelName].filter(
            v => v.toLowerCase().indexOf(val.trim().toLowerCase()) !== -1
          )
        })
      }
    },

    async loadItems() {
      try {
        this.items = await listAppItems(
          this.$store.state.app.id,
          this.model.name
        )
      } catch (err) {
        notifyError(err)
        return
      }

      this.model.fields.forEach(field => {
        if (field.type === 'ref') this.fetchRefItems(field.ref)
      })
    },

    openCreateDialog() {
      this.dialog.item = Object.fromEntries(
        this.model.fields.map(field => [field.name, field.default])
      )
      this.dialog.show = true
    },

    openUpdateDialog(key) {
      const item = this.items.find(item => item.key === key)
      this.dialog.item = {...item}
      this.dialog.show = true
    },

    closeFormDialog() {
      this.dialog.show = false
    },

    async saveItem() {
      try {
        if (this.dialog.item.id) {
          await setAppItem(
            this.$store.state.app.id,
            this.model.name,
            this.dialog.item.id,
            this.dialog.item
          )
        } else {
          await addAppItem(
            this.$store.state.app.id,
            this.model.name,
            this.dialog.item
          )
        }

        this.$q.notify({
          message: `${this.model.display || this.model.name} saved.`,
          type: 'positive',
          timeout: 3500
        })
      } catch (err) {
        notifyError(err)
      }
    },

    deleteItem(key) {
      this.$.plugins.Dialog.create({
        message: 'Are you sure you want to delete this item?',
        ok: {
          flat: true,
          color: 'orange'
        },
        cancel: {
          flat: true,
          color: 'grey'
        }
      }).onOk(async () => {
        try {
          await delAppItem(this.$store.state.app.id, this.model.name, key)
          this.$q.notify({
            message: `${this.model.display || this.model.name} deleted.`,
            type: 'info',
            timeout: 2500
          })
        } catch (err) {
          notifyError(err)
        }
      })
    }
  }
}
</script>
