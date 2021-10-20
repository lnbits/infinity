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
                {{ model.plural || `${model.display}s` || `${model.name}s` }}
              </h5>
            </div>
          </div>

          <q-table
            v-model:pagination="table.pagination"
            dense
            flat
            binary-state-sort
            column-sort-order="da"
            :rows="items[model.name]"
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
                  class="text-center"
                >
                  <q-btn
                    v-if="field.type === 'url'"
                    flat
                    dense
                    size="xs"
                    icon="link"
                    color="light-blue"
                    :title="props.row.value[field.name]"
                    @click.stop="
                      goToURL(
                        props.row.value[field.name].startsWith('http') ||
                          props.row.value[field.name].startsWith('/')
                          ? props.row.value[field.name]
                          : `/app/${$store.state.wallet.id}/${
                              $store.state.app.id
                            }/${props.row.value[field.name]}`
                      )
                    "
                  ></q-btn>
                  <span v-else-if="field.type === 'msatoshi'">
                    {{ formatMsatToSat(props.row.value[field.name]) }} sat
                  </span>
                  <span v-else-if="field.type === 'ref'">
                    {{
                      itemsMap[field.ref] &&
                      itemsMap[field.ref][props.row.value[field.name]] &&
                      itemsMap[field.ref][props.row.value[field.name]].value[
                        field.as
                      ]
                    }}
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
            v-model.trim="dialog.item.value[field.name]"
            filled
            dense
            :type="field.type === 'url' ? 'url' : 'text'"
            :label="fieldLabel(field)"
          />
          <q-input
            v-if="field.type === 'number'"
            v-model.number="dialog.item.value[field.name]"
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
              dialog.item.value[field.name] > 0
                ? dialog.item.value[field.name] / 1000
                : ''
            "
            @update:model-value="
              dialog.item.value[field.name] = (parseInt($event) || 0) * 1000
            "
          />
          <q-toggle
            v-if="field.type === 'boolean'"
            v-model="dialog.item.value[field.name]"
            :label="fieldLabel(field)"
          />

          <q-select
            v-if="field.type === 'ref'"
            v-model="dialog.item.value[field.name]"
            filled
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
import {setAppItem, addAppItem, delAppItem} from '../api'
import {formatMsatToSat, notifyError} from '../helpers'

export default {
  props: {
    model: {
      type: Object,
      required: true
    },
    items: {
      type: Object,
      required: true
    },
    itemsMap: {
      type: Object,
      required: true
    }
  },

  data() {
    return {
      table: {
        pagination: {
          rowsPerPage: 15,
          sortBy: 'created_at'
        }
      },
      dialog: {
        show: false,
        item: null
      }
    }
  },

  computed: {
    isFormSubmitDisabled() {
      return (
        this.dialog.show &&
        this.model.fields
          .filter(field => field.required)
          .filter(
            field =>
              this.dialog.item.value[field.name] === undefined ||
              this.dialog.item.value[field.name] === ''
          ).length > 0
      )
    }
  },

  methods: {
    json: v => JSON.stringify(v, null, 2),

    formatMsatToSat,

    goToURL: url => {
      window.open(url)
    },

    fieldLabel(field) {
      return (field.display || field.name) + (field.required ? ' *' : '')
    },

    openCreateDialog() {
      this.dialog.item = {
        wallet: this.$store.state.wallet.id,
        model: this.model.name,
        value: Object.fromEntries(
          this.model.fields
            .filter(field => !field.computed)
            .map(field => [field.name, field.default])
        )
      }
      this.dialog.show = true
    },

    openUpdateDialog(key) {
      const item = this.items[this.model.name].find(item => item.key === key)
      this.dialog.item = {...item, value: {...item.value}}
      this.model.fields
        .filter(field => field.computed)
        .forEach(f => {
          delete this.dialog.item.value[f.name]
        })
      this.dialog.show = true
    },

    closeFormDialog() {
      this.dialog.show = false
    },

    async saveItem() {
      try {
        if (this.dialog.item.key) {
          await setAppItem(
            this.$store.state.app.url,
            this.model.name,
            this.dialog.item.key,
            this.dialog.item.value
          )
        } else {
          await addAppItem(
            this.$store.state.app.url,
            this.model.name,
            this.dialog.item.value
          )
        }

        this.$q.notify({
          message: `${this.model.display || this.model.name} saved.`,
          type: 'positive',
          timeout: 3500
        })

        this.closeFormDialog()
      } catch (err) {
        notifyError(err)
      }
    },

    deleteItem(key) {
      this.$q
        .dialog({
          message: 'Are you sure you want to delete this item?',
          ok: {
            flat: true,
            color: 'orange'
          },
          cancel: {
            flat: true,
            color: 'grey'
          }
        })
        .onOk(async () => {
          try {
            await delAppItem(this.$store.state.app.url, this.model.name, key)
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
