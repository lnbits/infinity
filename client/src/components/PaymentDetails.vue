<template>
  <div class="q-py-md" style="text-align: left">
    <div class="row justify-center q-mb-md">
      <q-badge v-if="hasTag" color="yellow" text-color="black">
        #{{ payment.tag }}
      </q-badge>
    </div>
    <div class="row">
      <div class="col-3"><b>Date</b>:</div>
      <div class="col-9">{{ payment.date }} ({{ payment.dateFrom }})</div>
    </div>
    <div class="row">
      <div class="col-3"><b>Description</b>:</div>
      <div class="col-9">{{ payment.memo }}</div>
    </div>
    <div class="row">
      <div class="col-3"><b>Amount</b>:</div>
      <div class="col-9">{{ (payment.amount / 1000).toFixed(3) }} sat</div>
    </div>
    <div class="row">
      <div class="col-3"><b>Fee</b>:</div>
      <div class="col-9">{{ (payment.fee / 1000).toFixed(3) }} sat</div>
    </div>
    <div class="row">
      <div class="col-3"><b>Payment hash</b>:</div>
      <div class="col-9 text-wrap mono">{{ payment.payment_hash }}</div>
    </div>
    <div v-if="payment.webhook" class="row">
      <div class="col-3"><b>Webhook</b>:</div>
      <div class="col-9 text-wrap mono">
        {{ payment.webhook }}
        <q-badge :color="webhookStatusColor" text-color="white">
          {{ webhookStatusText }}
        </q-badge>
      </div>
    </div>
    <div v-if="hasPreimage" class="row">
      <div class="col-3"><b>Payment proof</b>:</div>
      <div class="col-9 text-wrap mono">{{ payment.preimage }}</div>
    </div>
    <div v-for="entry in extras" :key="entry.key" class="row">
      <div class="col-3">
        <q-badge v-if="hasTag" color="secondary" text-color="white">
          extra
        </q-badge>
        <b>{{ entry.key }}</b
        >:
      </div>
      <div class="col-9 text-wrap mono">{{ entry.value }}</div>
    </div>
    <div v-if="hasSuccessAction" class="row">
      <div class="col-3"><b>Success action</b>:</div>
      <div class="col-9">
        <LnurlPaySuccessAction
          :payment="payment"
          :success-action="payment.extra.success_action"
        ></LnurlPaySuccessAction>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    payment: {
      type: Object,
      required: true
    }
  },
  computed: {
    hasPreimage() {
      return (
        this.payment.preimage &&
        this.payment.preimage !==
          '0000000000000000000000000000000000000000000000000000000000000000'
      )
    },
    hasSuccessAction() {
      return (
        this.hasPreimage &&
        this.payment.extra &&
        this.payment.extra.success_action
      )
    },
    webhookStatusColor() {
      return this.payment.webhook_status >= 300 ||
        this.payment.webhook_status < 0
        ? 'red-10'
        : !this.payment.webhook_status
        ? 'cyan-7'
        : 'green-10'
    },
    webhookStatusText() {
      return this.payment.webhook_status
        ? this.payment.webhook_status
        : 'not sent yet'
    },
    hasTag() {
      return this.payment.extra && !!this.payment.extra.tag
    },
    extras() {
      if (!this.payment.extra) return []
      return Object.keys(this.payment.extra)
        .filter(key => key !== 'tag' && key !== 'success_action')
        .map(key => ({key, value: this.payment.extras[key]}))
    }
  }
}
</script>
