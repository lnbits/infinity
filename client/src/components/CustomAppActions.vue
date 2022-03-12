<template>
  <q-card v-if="actionOptions.length">
    <q-card-section>
      <div class="text-h6">Public Actions</div>
    </q-card-section>

    <q-card-actions class="q-mx-xs q-pb-md">
      <q-btn-toggle
        :model-value="selectedAction?.name"
        :options="actionOptions"
        flat
        clearable
        @update:model-value="selectAction"
        @clear="cancelAction"
      />
    </q-card-actions>

    <q-card-section v-if="selectedAction" class="q-px-xl">
      <div class="row items-center">
        <div class="col-2 text-bold">API Call:</div>
        <div class="col">
          <code style="word-break: break-all">
            curl '{{ appPublic }}/action/{{ selectedAction.name }}' -s
            <span v-if="hasParams"
              >-H 'Content-Type: application/json' -s -d '{{
                paramsPrettyJSON
              }}'</span
            >
          </code>
        </div>
      </div>
    </q-card-section>

    <q-card-section v-if="selectedAction">
      <q-form class="q-gutter-md" @submit="callAction">
        <div class="row wrap q-gutter-md">
          <template
            v-for="field in selectedAction.fields"
            :key="field.name"
            class="row"
          >
            <AppPropertyEdit
              v-model:data="params"
              :field="field"
              :items="items"
            />
          </template>
        </div>
        <div class="row wrap q-gutter-md q-pb-md">
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
      type: Object,
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
    },

    appPublic() {
      return (
        location.protocol +
        '//' +
        location.host +
        location.pathname.replace('/app/', '/').replace('/wallet/', '/ext/')
      )
    },

    hasParams() {
      return this.paramsPrettyJSON !== '{}'
    },

    paramsPrettyJSON() {
      return JSON.stringify(this.params, null, 2)
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

        var text
        try {
          text = JSON.stringify(resp, null, 2)
        } catch (err) {
          text = resp
        }

        this.$q.notify({
          position: 'center',
          timeout: 0,
          closeBtn: true,
          html: true,
          message: `<pre style="white-space: pre-wrap; word-wrap: break-word; word-break: break-all;"><code>${text
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')}</code></pre>`,
          caption: `Response from <code>${this.selectedAction.name}</code> call.`
        })
      } catch (err) {
        notifyError(err)
      }
    }
  }
}
</script>
