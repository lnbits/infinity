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
          <div
            v-for="field in selectedAction.fields"
            :key="field.name"
            class="row"
          >
            <q-input
              v-if="field.type === 'string' || field.type === 'url'"
              v-model.trim="params[field.name]"
              filled
              dense
              :type="field.type === 'url' ? 'url' : 'text'"
              :label="fieldLabel(field)"
            />
            <q-input
              v-if="field.type === 'number'"
              v-model.number="params[field.name]"
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
                params[field.name] > 0 ? params[field.name] / 1000 : ''
              "
              @update:model-value="
                params[field.name] = (parseInt($event) || 0) * 1000
              "
            />
            <q-toggle
              v-if="field.type === 'boolean'"
              v-model="params[field.name]"
              :label="fieldLabel(field)"
              :indeterminate-value="'INDETERMINATE'"
            />
          </div>
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
import {fieldLabel, notifyAppError} from '../helpers'
import {callAppAction} from '../api'

export default {
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
      this.params = {}
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
        notifyAppError(err)
      }
    }
  }
}
</script>
