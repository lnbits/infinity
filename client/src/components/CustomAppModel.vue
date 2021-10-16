<template>
  <div class="row q-col-gutter-md q-mb-md">
    <div class="col-12 col-md-7 q-gutter-y-md">
      <q-card>
        <q-card-section>
          <q-btn unelevated color="primary" @click="openCreateDialog"
            >New {{ model.name }}</q-btn
          >
        </q-card-section>
        <q-card-section>
          <div class="row items-center no-wrap q-mb-md">
            <div class="col">
              <h5 class="text-subtitle1 q-my-none">
                {{ model.namePlural || `${model.name}s` }}
              </h5>
            </div>
          </div>

          <q-table
            v-model:pagination="table.pagination"
            dense
            flat
            :data="items"
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
                    :href="props.row[field.name]"
                  >
                    <q-icon name="http" />
                  </a>
                  <span v-else-if="props.type === 'ref'">
                    {JSON.stringify(refItemsMap[props.row[field.name]])}
                  </span>
                  <span v-else>{props.row[field.name]}</span>
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

  <q-dialog v-model="formDialog.show" @hide="closeFormDialog">
    <q-card class="q-pa-lg q-pt-xl lnbits__dialog-card">
      <q-form class="q-gutter-md" @submit="setItem">
        <template
          v-for="field in model.fields.filter(f => !f.hidden && !f.computed)"
          :key="field.name"
        >
          <q-input
            v-if="field.type === 'string' || field.type === 'url'"
            v-model.trim="formDialog.item[field.name]"
            filled
            dense
            type="text"
            :label="field.name + (field.required ? ' *' : '')"
          />
          <q-input
            v-if="field.type === 'number'"
            v-model.number="formDialog.item[field.name]"
            filled
            dense
            type="number"
            :label="field.name + (field.required ? ' *' : '')"
          />
          <q-toggle
            v-if="field.type === 'boolean'"
            v-model="formDialog.item[field.name]"
            :label="field.name + (field.required ? ' *' : '')"
          />
          <q-select
            v-if="field.type === 'ref'"
            v-model="formDialog.item[field.name]"
            filled
            use-input
            input-debounce="0"
            behavior="dialog"
            :options="refItemsFiltered[field.ref]"
            :label="field.name + (field.required ? ' *' : '')"
            @filter="refOptionsFilter(field.ref)"
          />
        </template>
        <div class="row q-mt-lg">
          <q-btn
            v-if="formDialog.item.key"
            unelevated
            color="primary"
            type="submit"
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
import {listAppItems, setAppItem, delAppItem} from '../api'
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
          rowsPerPage: 15
        }
      },
      formDialog: {
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
        this.formDialog.show &&
        this.model.fields
          .filter(field => field.required)
          .filter(
            field =>
              this.formDialog.item[field.name] === undefined ||
              this.formDialog.item[field.name] === ''
          ).length > 0
      )
    }
  },

  methods: {
    async fetchRefItems(modelName) {
      if (!this.refItems[modelName]) {
        this.refItems[modelName] = await listAppItems(
          this.$store.wallet.id,
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
        this.items = await listAppItems(this.$store.wallet.id, this.model.name)
      } catch (err) {
        notifyError(err)
        return
      }

      this.model.fields.forEach(field => {
        if (field.type === 'ref') this.fetchRefItems(field.ref)
      })
    },

    openCreateDialog() {
      this.formDialog.item = Object.fromEntries(
        this.model.fields.map(field => [field.name, field.default])
      )
      this.formDialog.show = true
    },

    openUpdateDialog(key) {
      const item = this.items.find(item => item.key === key)
      this.formDialog.item = {...item}
      this.formDialog.show = true
    },

    closeFormDialog() {
      this.formDialog.show = false
    },

    async setItem() {
      try {
        await setAppItem(
          this.$store.wallet.id,
          this.formDialog.item.id,
          this.formDialog.item
        )
        this.$q.notify({
          message: `${this.model.name} saved.`,
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
          await delAppItem(this.$store.wallet.id, key)
          this.$q.notify({
            message: `${this.model.name} deleted.`,
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
