<template>
  <div>
    <p class="q-mb-sm">
      {{ successAction.message || successAction.description }}
    </p>
    <code v-if="decryptedValue" class="text-h6 q-mt-sm q-mb-none">
      {{ decryptedValue }}
    </code>
    <p v-else-if="successAction.url" class="text-h6 q-mt-sm q-mb-none">
      <a target="_blank" style="color: inherit" :href="successAction.url">{{
        successAction.url
      }}</a>
    </p>
  </div>
</template>

<script>
import {decryptLnurlPayAES} from '../helpers'

export default {
  props: {
    payment: {type: Object, required: true},
    successAction: {type: Object, required: true}
  },
  data() {
    return {
      decryptedValue: this.successAction.ciphertext
    }
  },
  mounted: function () {
    if (this.successAction.tag !== 'aes') return null

    decryptLnurlPayAES(this.successAction, this.payment.preimage).then(
      value => {
        this.decryptedValue = value
      }
    )
  }
}
</script>
