<template>
  <q-page
    v-if="$store.state.wallet"
    class="q-px-md q-py-lg"
    :class="{'q-px-lg': $q.screen.gt.xs}"
  >
    <div class="row q-col-gutter-md">
      <div class="col-12 col-md-7 q-gutter-y-md">
        <q-card>
          <q-card-section>
            <h3 class="q-my-none">
              <strong>{{ $store.state.wallet.balance }}</strong> sat
            </h3>
          </q-card-section>
          <div class="row q-pb-md q-px-md q-col-gutter-md">
            <div class="col">
              <q-btn
                unelevated
                color="primary"
                class="full-width"
                @click="showParseDialog"
                >Paste Request</q-btn
              >
            </div>
            <div class="col">
              <q-btn
                unelevated
                color="primary"
                class="full-width"
                @click="showReceiveDialog"
                >Create Invoice</q-btn
              >
            </div>
            <div class="col">
              <q-btn
                unelevated
                color="secondary"
                icon="photo_camera"
                @click="showCamera"
                >scan
                <q-tooltip>Use camera to scan an invoice/QR</q-tooltip>
              </q-btn>
            </div>
          </div>
        </q-card>

        <q-card>
          <q-card-section>
            <div class="row items-center no-wrap q-mb-sm">
              <div class="col">
                <h5 class="text-subtitle1 q-my-none">Transactions</h5>
              </div>
              <div class="col-auto">
                <q-btn flat color="grey" @click="exportCSV"
                  >Export to CSV</q-btn
                >
                <q-btn
                  dense
                  flat
                  round
                  icon="show_chart"
                  color="grey"
                  @click="showChart"
                >
                  <q-tooltip>Show chart</q-tooltip>
                </q-btn>
              </div>
            </div>
            <q-input
              v-if="payments.length > 10"
              v-model="paymentsTable.filter"
              filled
              dense
              clearable
              debounce="300"
              placeholder="Search by tag, description, amount"
              class="q-mb-md"
            >
            </q-input>
            <q-table
              v-model:pagination="paymentsTable.pagination"
              dense
              flat
              :rows="$store.state.wallet.payments"
              :row-key="paymentTableRowKey"
              :columns="paymentsTable.columns"
              no-data-label="No transactions made yet"
              :filter="paymentsTable.filter"
              :filter-method="paymentsTable.filterMethod"
            >
              <template #header="props">
                <q-tr :props="props">
                  <q-th auto-width></q-th>
                  <q-th
                    v-for="col in props.cols"
                    :key="col.name"
                    :props="props"
                    >{{ col.label }}</q-th
                  >
                </q-tr>
              </template>
              <template #body="props">
                <q-tr :props="props">
                  <q-td auto-width class="text-center">
                    <q-icon
                      v-if="props.row.isPaid"
                      size="14px"
                      :name="props.row.isOut ? 'call_made' : 'call_received'"
                      :color="props.row.isOut ? 'pink' : 'green'"
                      @click="props.expand = !props.expand"
                    ></q-icon>
                    <q-icon
                      v-else
                      name="settings_ethernet"
                      color="grey"
                      @click="props.expand = !props.expand"
                    >
                      <q-tooltip>Pending</q-tooltip>
                    </q-icon>
                  </q-td>
                  <q-td
                    key="description"
                    :props="props"
                    style="white-space: normal; word-break: break-all"
                  >
                    {{ props.row.description }}
                  </q-td>
                  <q-td key="date" auto-width :props="props">
                    <q-tooltip>{{
                      formatDate(props.row.date, true)
                    }}</q-tooltip>
                    {{ formatDate(props.row.date) }}
                  </q-td>
                  <q-td key="amount" auto-width :props="props">
                    {{ formatMsatToSat(props.row.amount) }}
                  </q-td>
                  <q-td key="fee" auto-width :props="props">
                    {{ props.row.fee.toString() }}
                  </q-td>
                </q-tr>

                <q-dialog v-model="props.expand" :props="props">
                  <q-card class="q-pa-lg q-pt-xl lnbits__dialog-card">
                    <div class="text-center q-mb-lg">
                      <div v-if="props.row.isIn && props.row.pending">
                        <q-icon name="settings_ethernet" color="grey"></q-icon>
                        Invoice waiting to be paid
                        <PaymentDetails :payment="props.row" />
                        <div
                          v-if="props.row.bolt11"
                          class="text-center q-mb-lg"
                        >
                          <a :href="'lightning:' + props.row.bolt11">
                            <q-responsive :ratio="1" class="q-mx-xl">
                              <QRCode
                                :size="500"
                                :value="props.row.bolt11"
                                :options="{width: 340}"
                                class="rounded-borders"
                              ></QRCode>
                            </q-responsive>
                          </a>
                        </div>
                        <div class="row q-mt-lg">
                          <q-btn
                            outline
                            color="grey"
                            @click="copyText(props.row.bolt11)"
                            >Copy invoice</q-btn
                          >
                          <q-btn
                            v-close-popup
                            flat
                            color="grey"
                            class="q-ml-auto"
                            >Close</q-btn
                          >
                        </div>
                      </div>
                      <div v-else-if="props.row.isPaid && props.row.isIn">
                        <q-icon
                          size="18px"
                          :name="'call_received'"
                          :color="'green'"
                        ></q-icon>
                        Payment Received
                        <PaymentDetails :payment="props.row" />
                      </div>
                      <div v-else-if="props.row.isPaid && props.row.isOut">
                        <q-icon
                          size="18px"
                          :name="'call_made'"
                          :color="'pink'"
                        ></q-icon>
                        Payment Sent
                        <PaymentDetails :payment="props.row" />
                      </div>
                      <div v-else-if="props.row.isOut && props.row.pending">
                        <q-icon name="settings_ethernet" color="grey"></q-icon>
                        Outgoing payment pending
                        <PaymentDetails :payment="props.row" />
                      </div>
                    </div>
                  </q-card>
                </q-dialog>
              </template>
            </q-table>
          </q-card-section>
        </q-card>
      </div>
      <div class="col-12 col-md-5 q-gutter-y-md">
        <q-card>
          <q-card-section>
            <h6 class="text-subtitle1 q-mt-none q-mb-sm">
              {{ $store.state.settings.siteTitle }} wallet:
              <strong
                ><em>{{ $store.state.wallet.name }}</em></strong
              >
            </h6>
          </q-card-section>
          <q-card-section class="q-pa-none">
            <q-separator></q-separator>

            <q-list>
              <q-separator>
                <q-expansion-item
                  group="extras"
                  icon="swap_vertical_circle"
                  label="API info"
                  :content-inset-level="0.5"
                >
                  <q-card-section>
                    <strong>Wallet ID: </strong
                    ><em>{{ $store.state.wallet.id }}</em
                    ><br />
                    <strong>Admin key: </strong
                    ><em>{{ $store.state.wallet.adminkey }}</em
                    ><br />
                    <strong>Invoice/read key: </strong
                    ><em>{{ $store.state.wallet.invoicekey }}</em>
                  </q-card-section>
                </q-expansion-item>
                <APIDocs />
              </q-separator>

              <q-expansion-item
                group="extras"
                icon="crop_free"
                label="Drain Funds"
              >
                <q-card>
                  <q-card-section class="text-center">
                    <p>
                      This is an LNURL-withdraw QR code for slurping everything
                      from this wallet. Do not share with anyone.
                    </p>
                    <a href="lightning:{{wallet.lnurlwithdraw_full}}">
                      <QRCode
                        :size="500"
                        value="{{wallet.lnurlwithdraw_full}}"
                        :options="{width: 240}"
                      ></QRCode>
                    </a>
                    <p>
                      It is compatible with <code>balanceCheck</code> and
                      <code>balanceNotify</code> so your wallet may keep pulling
                      the funds continuously from here after the first withdraw.
                    </p>
                  </q-card-section>
                </q-card>
              </q-expansion-item>
              <q-separator></q-separator>

              <q-expansion-item
                group="extras"
                icon="settings_cell"
                label="Export to Phone with QR Code"
              >
                <q-card>
                  <q-card-section class="text-center">
                    <p>
                      This QR code contains your wallet URL with full access.
                      You can scan it from your phone to open your wallet from
                      there.
                    </p>
                    <QRCode
                      :size="500"
                      :value="
                        '{{request.url_root}}' +
                        'wallet?usr={{user.id}}&wal={{wallet.id}}'
                      "
                      :options="{width: 240}"
                    ></QRCode>
                  </q-card-section>
                </q-card>
              </q-expansion-item>
              <q-separator></q-separator>
              <q-expansion-item
                group="extras"
                icon="edit"
                label="Rename wallet"
              >
                <q-card>
                  <q-card-section>
                    <div class="" style="max-width: 320px">
                      <q-input
                        v-model.trim="newName"
                        filled
                        label="Label"
                        dense="dense"
                        @update:model-value="e => console.log(e)"
                      />
                    </div>
                    <q-btn
                      :disable="!newName.length"
                      unelevated
                      class="q-mt-sm"
                      color="primary"
                      @click="updateWalletName()"
                      >Update name</q-btn
                    >
                  </q-card-section>
                </q-card>
              </q-expansion-item>
              <q-separator></q-separator>
              <q-expansion-item
                group="extras"
                icon="remove_circle"
                label="Delete wallet"
              >
                <q-card>
                  <q-card-section>
                    <p>
                      This whole wallet will be deleted, the funds will be
                      <strong>UNRECOVERABLE</strong>.
                    </p>
                    <q-btn unelevated color="red-10" @click="deleteWallet"
                      >Delete wallet</q-btn
                    >
                  </q-card-section>
                </q-card>
              </q-expansion-item>
            </q-list>
          </q-card-section>
        </q-card>
      </div>
    </div>

    <q-dialog v-model="receive.show" @hide="closeReceiveDialog">
      <q-card
        v-if="!receive.paymentReq"
        class="q-pa-lg q-pt-xl lnbits__dialog-card"
      >
        <q-form class="q-gutter-md" @submit="createInvoice">
          <p v-if="receive.lnurl" class="text-h6 text-center q-my-none">
            <b>{{ receive.lnurl.domain }}</b> is requesting an invoice:
          </p>

          <q-select
            v-model="receive.unit"
            filled
            dense
            type="text"
            label="Unit"
            :options="['sat'].concat($store.state.settings.currencies)"
          ></q-select>
          <q-input
            v-model.number="receive.data.amount"
            filled
            dense
            type="number"
            :label="`Amount (${receive.unit}) *`"
            :step="receive.unit != 'sat' ? '0.001' : '1'"
            :min="receive.minMax[0]"
            :max="receive.minMax[1]"
            :readonly="receive.lnurl && receive.lnurl.fixed"
          ></q-input>
          <q-input
            v-model.trim="receive.data.description"
            filled
            dense
            label="Description *"
            placeholder="LNbits invoice"
          ></q-input>
          <div v-if="receive.status == 'pending'" class="row q-mt-lg">
            <q-btn
              unelevated
              color="primary"
              :disable="
                receive.data.description == null ||
                receive.data.amount == null ||
                receive.data.amount <= 0
              "
              type="submit"
            >
              <span v-if="receive.lnurl">
                Withdraw from {{ receive.lnurl.domain }}
              </span>
              <span v-else>Create invoice</span>
            </q-btn>
            <q-btn v-close-popup flat color="grey" class="q-ml-auto"
              >Cancel</q-btn
            >
          </div>
          <q-spinner
            v-if="receive.status == 'loading'"
            color="primary"
            size="2.55em"
          ></q-spinner>
        </q-form>
      </q-card>
      <q-card v-else class="q-pa-lg q-pt-xl lnbits__dialog-card">
        <div class="text-center q-mb-lg">
          <a :href="'lightning:' + receive.paymentReq">
            <q-responsive :ratio="1" class="q-mx-xl">
              <QRCode
                :size="500"
                :value="receive.paymentReq"
                :options="{width: 340}"
                class="rounded-borders"
              ></QRCode>
            </q-responsive>
          </a>
        </div>
        <div class="row q-mt-lg">
          <q-btn outline color="grey" @click="copyText(receive.paymentReq)"
            >Copy invoice</q-btn
          >
          <q-btn v-close-popup flat color="grey" class="q-ml-auto">Close</q-btn>
        </div>
      </q-card>
    </q-dialog>

    <q-dialog v-model="parse.show" @hide="closeParseDialog">
      <q-card class="q-pa-lg q-pt-xl lnbits__dialog-card">
        <div v-if="parse.invoice">
          <h6 class="q-my-none">{{ parse.invoice.sat }} sat</h6>
          <q-separator class="q-my-sm"></q-separator>
          <p class="text-wrap">
            <strong>Description:</strong> {{ parse.invoice.description }}<br />
            <strong>Expire date:</strong> {{ parse.invoice.expireDate }}<br />
            <strong>Hash:</strong> {{ parse.invoice.hash }}
          </p>
          <div v-if="canPay" class="row q-mt-lg">
            <q-btn unelevated color="primary" @click="payInvoice">Pay</q-btn>
            <q-btn v-close-popup flat color="grey" class="q-ml-auto"
              >Cancel</q-btn
            >
          </div>
          <div v-else class="row q-mt-lg">
            <q-btn unelevated disabled color="yellow" text-color="black"
              >Not enough funds!</q-btn
            >
            <q-btn v-close-popup flat color="grey" class="q-ml-auto"
              >Cancel</q-btn
            >
          </div>
        </div>
        <div v-else-if="parse.lnurlauth">
          <q-form class="q-gutter-md" @submit="authLnurl">
            <p class="q-my-none text-h6">
              Authenticate with <b>{{ parse.lnurlauth.domain }}</b
              >?
            </p>
            <q-separator class="q-my-sm"></q-separator>
            <p>
              For every website and for every LNbits wallet, a new keypair will
              be deterministically generated so your identity can't be tied to
              your LNbits wallet or linked across websites. No other data will
              be shared with {{ parse.lnurlauth.domain }}.
            </p>
            <p>
              Your public key for <b>{{ parse.lnurlauth.domain }}</b> is:
            </p>
            <p class="q-mx-xl">
              <code class="text-wrap"> {{ parse.lnurlauth.pubkey }} </code>
            </p>
            <div class="row q-mt-lg">
              <q-btn unelevated color="primary" type="submit">Login</q-btn>
              <q-btn v-close-popup flat color="grey" class="q-ml-auto"
                >Cancel</q-btn
              >
            </div>
          </q-form>
        </div>
        <div v-else-if="parse.lnurlpay">
          <q-form class="q-gutter-md" @submit="payLnurl">
            <p v-if="parse.lnurlpay.fixed" class="q-my-none text-h6">
              <b>{{ parse.lnurlpay.domain }}</b> is requesting
              {{ (parse.lnurlpay.maxSendable / 1000).toFixed(0) }} sat
              <span v-if="parse.lnurlpay.commentAllowed > 0">
                <br />
                and a {{ parse.lnurlpay.commentAllowed }}-char comment
              </span>
            </p>
            <p v-else class="q-my-none text-h6 text-center">
              <b>{{ parse.lnurlpay.targetUser || parse.lnurlpay.domain }}</b> is
              requesting <br />
              between
              <b>{{ parse.lnurlpay.minSendable.toFixed(0) }}</b> and
              <b>{{ parse.lnurlpay.maxSendable.toFixed(0) }}</b> sat
              <span v-if="parse.lnurlpay.commentAllowed > 0">
                <br />
                and a {{ parse.lnurlpay.commentAllowed }}-char comment
              </span>
            </p>
            <q-separator class="q-my-sm"></q-separator>
            <div class="row">
              <p class="col text-justify text-italic">
                {{ parse.lnurlpay.description }}
              </p>
              <p v-if="parse.lnurlpay.image" class="col-4 q-pl-md">
                <q-img :src="parse.lnurlpay.image" />
              </p>
            </div>
            <div class="row">
              <div class="col">
                <q-input
                  v-model.number="parse.data.amount"
                  filled
                  dense
                  type="number"
                  label="Amount (sat) *"
                  :min="parse.lnurlpay.minSendable / 1000"
                  :max="parse.lnurlpay.maxSendable / 1000"
                  :readonly="parse.lnurlpay.fixed"
                ></q-input>
              </div>
              <div
                v-if="parse.lnurlpay.commentAllowed > 0"
                class="col-8 q-pl-md"
              >
                <q-input
                  v-model="parse.data.comment"
                  filled
                  dense
                  :type="
                    parse.lnurlpay.commentAllowed > 64 ? 'textarea' : 'text'
                  "
                  label="Comment (optional)"
                  :maxlength="parse.lnurlpay.commentAllowed"
                ></q-input>
              </div>
            </div>
            <div class="row q-mt-lg">
              <q-btn unelevated color="primary" type="submit"
                >Send satoshis</q-btn
              >
              <q-btn v-close-popup flat color="grey" class="q-ml-auto"
                >Cancel</q-btn
              >
            </div>
          </q-form>
        </div>
        <div v-else>
          <q-form
            v-if="!parse.camera.show"
            class="q-gutter-md"
            @submit="decodeRequest"
          >
            <q-input
              v-model.trim="parse.data.request"
              filled
              dense
              type="textarea"
              label="Paste an invoice, payment request or lnurl code *"
            >
            </q-input>
            <div class="row q-mt-lg">
              <q-btn
                unelevated
                color="primary"
                :disable="parse.data.request == ''"
                type="submit"
                >Read</q-btn
              >
              <q-btn v-close-popup flat color="grey" class="q-ml-auto"
                >Cancel</q-btn
              >
            </div>
          </q-form>
          <div v-else>
            <q-responsive :ratio="1">
              <qrcode-stream
                class="rounded-borders"
                @decode="decodeQR"
              ></qrcode-stream>
            </q-responsive>
            <div class="row q-mt-lg">
              <q-btn flat color="grey" class="q-ml-auto" @click="closeCamera">
                Cancel
              </q-btn>
            </div>
          </div>
        </div>
      </q-card>
    </q-dialog>

    <q-dialog v-model="parse.camera.show">
      <q-card class="q-pa-lg q-pt-xl">
        <div class="text-center q-mb-lg">
          <qrcode-stream
            class="rounded-borders"
            @decode="decodeQR"
          ></qrcode-stream>
        </div>
        <div class="row q-mt-lg">
          <q-btn flat color="grey" class="q-ml-auto" @click="closeCamera"
            >Cancel</q-btn
          >
        </div>
      </q-card>
    </q-dialog>

    <q-dialog v-model="paymentsChart.show">
      <q-card class="q-pa-sm" style="width: 800px; max-width: unset">
        <q-card-section>
          <canvas ref="canvas" width="600" height="400"></canvas>
        </q-card-section>
      </q-card>
    </q-dialog>

    <div ref="disclaimer"></div>
    <q-dialog v-model="disclaimerDialog.show">
      <q-card class="q-pa-lg">
        <h6 class="q-my-md text-deep-purple">Warning</h6>
        <p>
          Login functionality to be released in v0.2, for now,
          <strong
            >make sure you bookmark this page for future access to your
            wallet</strong
          >!
        </p>
        <p>
          This service is in BETA, and we hold no responsibility for people
          losing access to funds. To encourage you to run your own LNbits
          installation, any balance on {{ disclaimerDialog.location.host }} will
          incur a charge of <strong>{{ service_fee }}% service fee</strong> per
          week.
        </p>
        <div class="row q-mt-lg">
          <q-btn
            outline
            color="grey"
            @click="copyText(disclaimerDialog.location.href)"
            >Copy wallet URL</q-btn
          >
          <q-btn v-close-popup flat color="grey" class="q-ml-auto"
            >I understand</q-btn
          >
        </div>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script>
import bolt11 from 'light-bolt11-decoder'
import {date} from 'quasar'

import {generateChart} from '../chart'
import {
  notifyError,
  exportCSV,
  formatDate,
  formatMsatToSat,
  copyText
} from '../helpers'
import {
  createInvoice,
  deleteWallet,
  renameWallet,
  payInvoice,
  authLnurl,
  scanLnurl,
  payLnurl
} from '../api'

export default {
  name: 'Wallet',

  data() {
    return {
      receive: {
        show: false,
        status: 'pending',
        paymentReq: null,
        paymentHash: null,
        minMax: [0, 2100000000000000],
        lnurl: null,
        unit: 'sat',
        data: {
          unit: 'sat',
          amount: null,
          description: ''
        }
      },
      parse: {
        show: false,
        invoice: null,
        lnurlpay: null,
        lnurlauth: null,
        data: {
          request: '',
          amount: 0,
          comment: ''
        },
        camera: {
          show: false,
          camera: 'auto'
        }
      },
      payments: [],
      paymentsTable: {
        columns: [
          {
            name: 'description',
            align: 'left',
            label: 'Description'
          },
          {
            name: 'date',
            align: 'left',
            label: 'Date',
            sortable: true
          },
          {
            name: 'amount',
            align: 'right',
            label: 'Amount (sat)',
            sortable: true
          },
          {
            name: 'fee',
            align: 'right',
            label: 'Fee (msat)'
          }
        ],
        pagination: {
          rowsPerPage: 15
        },
        filter: null,
        filterMethod: (rows, terms) => {
          const queries = terms.toLowerCase().split(' ')
          return rows.filter(
            row => queries.filter(q => JSON.stringify(row).indexOf(q)) > 0
          )
        }
      },
      paymentsChart: {
        show: false
      },
      disclaimerDialog: {
        show: false,
        location: window.location
      },
      balance: 0,
      newName: ''
    }
  },

  computed: {
    canPay() {
      if (!this.parse.invoice) return false
      return this.parse.invoice.sat <= this.balance
    }
  },

  mounted() {
    // show disclaimer
    if (
      this.$refs.disclaimer &&
      !this.$q.localStorage.getItem('lnbits.disclaimerShown')
    ) {
      this.disclaimerDialog.show = true
      this.$q.localStorage.set('lnbits.disclaimerShown', true)
    }

    // listen to events
    window.events.on('invoice-paid', payment => {
      // TODO
      // if (this.receive.paymentHash === paymentHash) {
      //   this.receive.show = false
      //   this.receive.paymentHash = null
      // }
    })
    window.events.on('payment-complete', payment => {
      // TODO
      // this.parse.show = false
      // dismissPaymentMsg()
      // TODO
      // show lnurlpay success action
      // if (response.data.success_action) {
      //   switch (response.data.success_action.tag) {
      //     case 'url':
      //       this.$q.notify({
      //         message: `<a target="_blank" style="color: inherit" href="${response.data.success_action.url}">${response.data.success_action.url}</a>`,
      //         caption: response.data.success_action.description,
      //         html: true,
      //         type: 'positive',
      //         timeout: 0,
      //         closeBtn: true
      //       })
      //       break
      //     case 'message':
      //       this.$q.notify({
      //         message: response.data.success_action.message,
      //         type: 'positive',
      //         timeout: 0,
      //         closeBtn: true
      //       })
      //       break
      //     case 'aes':
      //       LNbits.api
      //         .getPayment(this.g.wallet, response.data.payment_hash)
      //         .then(({data: payment}) =>
      //           decryptLnurlPayAES(
      //             response.data.success_action,
      //             payment.preimage
      //           )
      //         )
      //         .then(value => {
      //           this.$q.notify({
      //             message: value,
      //             caption: response.data.success_action.description,
      //             html: true,
      //             type: 'positive',
      //             timeout: 0,
      //             closeBtn: true
      //           })
      //         })
      //       break
      //   }
      // }
    })
    window.events.on('payment-failed', payment => {
      // TODO
    })
  },

  methods: {
    log: console.log,
    formatMsatToSat,
    formatDate,
    paymentTableRowKey(row) {
      return row.hash + row.amount
    },
    closeCamera() {
      this.parse.camera.show = false
    },
    showCamera() {
      this.parse.camera.show = true
    },
    showChart() {
      this.paymentsChart.show = true
      this.$nextTick(() => {
        generateChart(this.$refs.canvas, this.payments)
      })
    },
    showReceiveDialog() {
      this.receive.show = true
      this.receive.status = 'pending'
      this.receive.paymentReq = null
      this.receive.paymentHash = null
      this.receive.data.amount = null
      this.receive.data.description = null
      this.receive.unit = 'sat'
      this.receive.minMax = [0, 2100000000000000]
      this.receive.lnurl = null
    },
    closeReceiveDialog() {
      this.receive.show = false
    },
    showParseDialog() {
      this.parse.show = true
      this.parse.invoice = null
      this.parse.lnurlpay = null
      this.parse.lnurlauth = null
      this.parse.data.request = ''
      this.parse.data.comment = ''
      this.parse.camera.show = false
    },
    closeParseDialog() {
      this.parse.show = false
    },
    async createInvoice() {
      this.receive.status = 'loading'

      try {
        const response = await createInvoice({
          amount: this.receive.data.amount,
          description: this.receive.data.description,
          unit: this.receive.unit,
          lnurlCallback: this.receive.lnurl && this.receive.lnurl.callback
        })

        this.receive.status = 'success'
        this.receive.paymentReq = response.bolt11
        this.receive.paymentHash = response.hash

        // if (response.lnurl_response !== null) {
        //   if (response.lnurl_response === false) {
        //     response.lnurl_response = `Unable to connect`
        //   }

        //   if (typeof response.data.lnurl_response === 'string') {
        //     // failure
        //     this.$q.notify({
        //       timeout: 5000,
        //       type: 'warning',
        //       message: `${this.receive.lnurl.domain} lnurl-withdraw call failed.`,
        //       caption: response.data.lnurl_response
        //     })
        //     return
        //   } else if (response.data.lnurl_response === true) {
        //     // success
        //     this.$q.notify({
        //       timeout: 5000,
        //       message: `Invoice sent to ${this.receive.lnurl.domain}!`,
        //       spinner: true
        //     })
        //   }
        // }
      } catch (err) {
        notifyError(err)
        this.receive.status = 'pending'
      }
    },
    decodeQR(res) {
      this.parse.data.request = res
      this.decodeRequest()
      this.parse.camera.show = false
    },
    async decodeRequest() {
      this.parse.show = true

      if (this.parse.data.request.startsWith('lightning:')) {
        this.parse.data.request = this.parse.data.request.slice(10)
      } else if (this.parse.data.request.startsWith('lnurl:')) {
        this.parse.data.request = this.parse.data.request.slice(6)
      } else if (this.parse.data.request.indexOf('lightning=lnurl1') !== -1) {
        this.parse.data.request = this.parse.data.request
          .split('lightning=')[1]
          .split('&')[0]
      }

      if (
        this.parse.data.request.match(/[\w.+-~_]+@[\w.+-~_]/) ||
        this.parse.data.request.toLowerCase().startsWith('lnurl1') ||
        this.parse.data.request.toLowerCase().startsWith('lnurlp') ||
        this.parse.data.request.toLowerCase().startsWith('lnurlw') ||
        this.parse.data.request.toLowerCase().startsWith('keyauth:') ||
        this.parse.data.request.toLowerCase().startsWith('lnurlc') ||
        this.parse.data.request.toLowerCase().startsWith('https://')
      ) {
        try {
          const response = await scanLnurl(this.parse.data.request)

          if (response.kind === 'pay') {
            this.parse.lnurlpay = Object.freeze(response)
            this.parse.response.amount = response.minSendable / 1000
          } else if (response.kind === 'auth') {
            this.parse.lnurlauth = Object.freeze(response)
          } else if (response.kind === 'withdraw') {
            this.parse.show = false
            this.receive.show = true
            this.receive.status = 'pending'
            this.receive.paymentReq = null
            this.receive.paymentHash = null
            this.receive.response.amount = response.maxWithdrawable / 1000
            this.receive.response.description = response.defaultDescription
            this.receive.minMax = [
              response.minWithdrawable / 1000,
              response.maxWithdrawable / 1000
            ]
            this.receive.lnurl = {
              domain: response.domain,
              callback: response.callback,
              fixed: response.fixed
            }
          }
        } catch (err) {
          notifyError(err, `${err?.data?.domain || 'lnurl'} call failed`)
        }
        return
      }

      try {
        const invoice = bolt11.decode(this.parse.data.request)
        let cleanInvoice = {
          msat: invoice.millisatoshis,
          sat: invoice.millisatoshis / 1000,
          hash: invoice.payment_hash,
          description: invoice.description,
          expireDate: date.formatDate(
            invoice.expiry,
            'YYYY-MM-DDTHH:mm:ss.SSSZ'
          )
        }

        this.parse.invoice = cleanInvoice
      } catch (error) {
        notifyError(error, 'Failed to parse invoice')
        this.parse.show = false
        return
      }
    },
    async payInvoice() {
      let dismissPaymentMsg = this.$q.notify({
        timeout: 0,
        message: 'Processing payment...'
      })

      try {
        await payInvoice({invoice: this.parse.data.request})
      } catch (err) {
        dismissPaymentMsg()
        notifyError(err)
      }
    },
    async payLnurl() {
      let dismissPaymentMsg = this.$q.notify({
        timeout: 0,
        message: 'Processing payment...'
      })

      try {
        await payLnurl({
          callback: this.parse.lnurlpay.callback,
          descriptionHash: this.parse.lnurlpay.description_hash,
          msatoshi: this.parse.data.amount * 1000,
          description: this.parse.lnurlpay.description.slice(0, 120),
          comment: this.parse.data.comment
        })

        this.parse.show = false
      } catch (err) {
        dismissPaymentMsg()
        notifyError(err)
      }
    },
    async authLnurl() {
      let dismissAuthMsg = this.$q.notify({
        timeout: 10,
        message: 'Performing authentication...'
      })

      try {
        await authLnurl(this.parse.lnurlauth.callback)
        dismissAuthMsg()
        this.$q.notify({
          message: `Authentication successful.`,
          type: 'positive',
          timeout: 3500
        })
        this.parse.show = false
      } catch (err) {
        dismissAuthMsg()
        notifyError(err)
      }
    },
    async updateWalletName() {
      let newName = this.newName
      if (!newName || !newName.length) return

      try {
        await renameWallet(newName)

        this.newName = ''
        this.$q.notify({
          message: `Wallet named updated.`,
          type: 'positive',
          timeout: 3500
        })

        this.$store.dispatch('fetchWallet', this.$store.state.wallets.id)
      } catch (err) {
        this.newName = ''
        notifyError(err)
      }
    },
    deleteWallet() {
      this.$q.plugins.Dialog.create({
        message: `Are you sure you want to delete the wallet '${this.$store.state.wallet.name}'?`,
        ok: {
          flat: true,
          color: 'orange'
        },
        cancel: {
          flat: true,
          color: 'grey'
        }
      }).onOk(async () => {
        await deleteWallet()
        location.href = `/?${location.search}`
      })
    },
    exportCSV() {
      exportCSV(this.paymentsTable.columns, this.payments)
    },
    copyText
  }
}
</script>
