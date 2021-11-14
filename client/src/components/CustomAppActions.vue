<template>
  <q-card>
    <q-card-section>
      <div class="text-h6">Public Actions</div>
    </q-card-section>

    <q-card-actions>
      <q-btn-toggle
        :model-value="selectedAction?.name"
        :options="actionOptions"
        flat
        clearable
        @update:model-value="selectAction"
        @clear="cancelAction"
      />
    </q-card-actions>

    <q-card-section v-if="selectedAction">
      <q-form class="q-gutter-md" @submit="callAction">
        <div class="row wrap q-gutter-md">
          <template
            v-for="field in selectedAction.fields"
            :key="field.name"
            class="row"
          >
            <AppPropertyEdit
              v-model:value="params[field.name]"
              :field="field"
              :items="items"
            />
          </template>
        </div>
        <div class="row wrap q-gutter-md">
          <div class="col">
            <q-btn unelevated color="primary" type="submit">Call</q-btn>
          </div>
          <div class="col">
            <q-btn flat color="grey" class="q-ml-auto" @click="cancelAction">
              Cancel
            </q-btn>
          </div>
        </div>
      </q-form>
    </q-card-section>
  </q-card>
</template>

<script>
import {callAppAction} from '../api'
import {paramDefaults, fieldLabel, notifyError} from '../helpers'

export default {
  props: {
    items: {
      type: Array,
      required: true
    },
    actions: {
      type: Object,
      required: true
    }
  },

  data() {
    return {
      selectedAction: null,
      params: {}
    }
  },

  computed: {
    actionOptions() {
      return Object.entries(this.$store.state.app.actions).map(
        ([name, action]) => ({
          label: name,
          value: name
        })
      )
    }
  },

  methods: {
    fieldLabel,

    selectAction(actionName) {
      this.selectedAction = this.$store.state.app.actions[actionName]
      this.selectedAction.name = actionName
      this.params = paramDefaults(this.selectedAction.fields)
    },

    cancelAction() {
      this.selectedAction = null
      this.params = {}
    },

    async callAction(ev) {
      ev.preventDefault()

      try {
        const resp = await callAppAction(
          this.$store.state.wallet.id,
          this.$store.state.app.id,
          this.selectedAction.name,
          this.params
        )

        const jsonResp = JSON.stringify(resp, null, 2)

        this.$q.notify({
          position: 'center',
          timeout: 0,
          closeBtn: true,
          html: true,
          message: `<pre style="white-space: pre-wrap; word-wrap: break-word; word-break: break-all;"><code>${jsonResp}</code></pre>`,
          caption: `Response from <code>${this.selectedAction.name}</code> call.`
        })
      } catch (err) {
        notifyError(err)
      }
    }
  }
}
</script>
