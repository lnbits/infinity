<template>
  <q-page class="q-px-md q-py-lg" :class="{'q-px-lg': $q.screen.gt.xs}">
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
                <!--<q-btn v-if="pendingPaymentsExist" dense flat round icon="update" color="grey" @click="checkPendingPayments">
                <q-tooltip>Check pending</q-tooltip>
              </q-btn>-->
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
              placeholder="Search by tag, memo, amount"
              class="q-mb-md"
            >
            </q-input>
            <q-table
              v-model:pagination="paymentsTable.pagination"
              dense
              flat
              :data="$store.state.wallet.payments"
              :row-key="paymentTableRowKey"
              :columns="paymentsTable.columns"
              no-data-label="No transactions made yet"
              :filter="paymentsTable.filter"
              :filter-method="paymentsTable.filterMethod"
            >
              {% raw %}
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
                    key="memo"
                    :props="props"
                    style="white-space: normal; word-break: break-all"
                  >
                    <q-badge
                      v-if="props.row.tag"
                      color="yellow"
                      text-color="black"
                    >
                      <a
                        class="inherit"
                        :href="
                          [
                            '/',
                            props.row.tag,
                            '/?usr=',
                            $store.state.user.id
                          ].join('')
                        "
                      >
                        #{{ props.row.tag }}
                      </a>
                    </q-badge>
                    {{ props.row.memo }}
                  </q-td>
                  <q-td key="date" auto-width :props="props">
                    <q-tooltip>{{ props.row.date }}</q-tooltip>
                    {{ props.row.dateFrom }}
                  </q-td>
                  <q-td key="sat" auto-width :props="props">
                    {{ props.row.sat }}
                  </q-td>
                  <q-td key="fee" auto-width :props="props">
                    {{ props.row.fee }}
                  </q-td>
                </q-tr>

                <q-dialog v-model="props.expand" :props="props">
                  <q-card class="q-pa-lg q-pt-xl lnbits__dialog-card">
                    <div class="text-center q-mb-lg">
                      <div v-if="props.row.isIn && props.row.pending">
                        <q-icon name="settings_ethernet" color="grey"></q-icon>
                        Invoice waiting to be paid
                        <lnbits-payment-details
                          :payment="props.row"
                        ></lnbits-payment-details>
                        <div
                          v-if="props.row.bolt11"
                          class="text-center q-mb-lg"
                        >
                          <a :href="'lightning:' + props.row.bolt11">
                            <q-responsive :ratio="1" class="q-mx-xl">
                              <qrcode
                                :value="props.row.bolt11"
                                :options="{width: 340}"
                                class="rounded-borders"
                              ></qrcode>
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
                        <lnbits-payment-details
                          :payment="props.row"
                        ></lnbits-payment-details>
                      </div>
                      <div v-else-if="props.row.isPaid && props.row.isOut">
                        <q-icon
                          size="18px"
                          :name="'call_made'"
                          :color="'pink'"
                        ></q-icon>
                        Payment Sent
                        <lnbits-payment-details
                          :payment="props.row"
                        ></lnbits-payment-details>
                      </div>
                      <div v-else-if="props.row.isOut && props.row.pending">
                        <q-icon name="settings_ethernet" color="grey"></q-icon>
                        Outgoing payment pending
                        <lnbits-payment-details
                          :payment="props.row"
                        ></lnbits-payment-details>
                      </div>
                    </div>
                  </q-card>
                </q-dialog>
              </template>
              {% endraw %}
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
                ><em>{{ wallet.name }}</em></strong
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
                      <qrcode
                        value="{{wallet.lnurlwithdraw_full}}"
                        :options="{width: 240}"
                      ></qrcode>
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
                    <qrcode
                      :value="
                        '{{request.url_root}}' +
                        'wallet?usr={{user.id}}&wal={{wallet.id}}'
                      "
                      :options="{width: 240}"
                    ></qrcode>
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
                    <q-btn
                      unelevated
                      color="red-10"
                      @click="deleteWallet('{{ wallet.id }}', '{{ user.id }}')"
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
      {% raw %}
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
            :options="$store.settings.currencies"
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
            v-model.trim="receive.data.memo"
            filled
            dense
            label="Memo *"
            placeholder="LNbits invoice"
          ></q-input>
          <div v-if="receive.status == 'pending'" class="row q-mt-lg">
            <q-btn
              unelevated
              color="primary"
              :disable="
                receive.data.memo == null ||
                receive.data.amount == null ||
                receive.data.amount <= 0
              "
              type="submit"
            >
              <span v-if="receive.lnurl">
                Withdraw from {{ receive.lnurl.domain }}
              </span>
              <span v-else> Create invoice </span>
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
              <qrcode
                :value="receive.paymentReq"
                :options="{width: 340}"
                class="rounded-borders"
              ></qrcode>
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
      {% endraw %}
    </q-dialog>

    <q-dialog v-model="parse.show" @hide="closeParseDialog">
      <q-card class="q-pa-lg q-pt-xl lnbits__dialog-card">
        <div v-if="parse.invoice">
          {% raw %}
          <h6 class="q-my-none">{{ parse.invoice.sat }} sat</h6>
          <q-separator class="q-my-sm"></q-separator>
          <p class="text-wrap">
            <strong>Description:</strong> {{ parse.invoice.description }}<br />
            <strong>Expire date:</strong> {{ parse.invoice.expireDate }}<br />
            <strong>Hash:</strong> {{ parse.invoice.hash }}
          </p>
          {% endraw %}
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
          {% raw %}
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
          {% endraw %}
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

    {% if service_fee > 0 %}
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
          installation, any balance on {% raw %}{{
            disclaimerDialog.location.host
          }}{% endraw %} will incur a charge of
          <strong>{{ service_fee }}% service fee</strong> per week.
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
import groupBy from 'lodash.groupby'
import Chart from 'chart.js'
import bolt11 from 'bolt11'

import {
  notifyApiError,
  createInvoice,
  deleteWallet,
  renameWallet,
  payInvoice,
  exportCSV,
  authLnurl,
  scanLnurl,
  payLnurl
} from '../helpers'

function generateChart(canvas, payments) {
  var txs = []
  var n = 0
  var data = {
    labels: [],
    income: [],
    outcome: [],
    cumulative: []
  }

  payments
    .filter(p => !p.pending)
    .sort((a, b) => a.time - b.time)
    .forEach(tx => {
      txs.push({
        hour: this.$q.utils.date.formatDate(tx.date, 'YYYY-MM-DDTHH:00'),
        sat: tx.sat
      })
    })

  groupBy(txs, 'hour').forEach((value, day) => {
    var income = value.reduce(
      (memo, tx) => (tx.sat >= 0 ? memo + tx.sat : memo),
      0
    )
    var outcome = value.reduce(
      (memo, tx) => (tx.sat < 0 ? memo + Math.abs(tx.sat) : memo),
      0
    )
    n = n + income - outcome
    data.labels.push(day)
    data.income.push(income)
    data.outcome.push(outcome)
    data.cumulative.push(n)
  })

  new Chart(canvas.getContext('2d'), {
    type: 'bar',
    data: {
      labels: data.labels,
      datasets: [
        {
          data: data.cumulative,
          type: 'line',
          label: 'balance',
          backgroundColor: '#673ab7', // deep-purple
          borderColor: '#673ab7',
          borderWidth: 4,
          pointRadius: 3,
          fill: false
        },
        {
          data: data.income,
          type: 'bar',
          label: 'in',
          barPercentage: 0.75,
          backgroundColor: window.Color('rgb(76,175,80)').alpha(0.5).rgbString() // green
        },
        {
          data: data.outcome,
          type: 'bar',
          label: 'out',
          barPercentage: 0.75,
          backgroundColor: window.Color('rgb(233,30,99)').alpha(0.5).rgbString() // pink
        }
      ]
    },
    options: {
      title: {
        text: 'Chart.js Combo Time Scale'
      },
      tooltips: {
        mode: 'index',
        intersect: false
      },
      scales: {
        xAxes: [
          {
            type: 'time',
            display: true,
            offset: true,
            time: {
              minUnit: 'hour',
              stepSize: 3
            }
          }
        ]
      },
      // performance tweaks
      animation: {
        duration: 0
      },
      elements: {
        line: {
          tension: 0
        }
      }
    }
  })
}

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
          amount: null,
          memo: ''
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
        paymentChecker: null,
        camera: {
          show: false,
          camera: 'auto'
        }
      },
      payments: [],
      paymentsTable: {
        columns: [
          {
            name: 'memo',
            align: 'left',
            label: 'Memo',
            field: 'memo'
          },
          {
            name: 'date',
            align: 'left',
            label: 'Date',
            field: 'date',
            sortable: true
          },
          {
            name: 'sat',
            align: 'right',
            label: 'Amount (sat)',
            field: 'sat',
            sortable: true
          },
          {
            name: 'fee',
            align: 'right',
            label: 'Fee (msat)',
            field: 'fee'
          }
        ],
        pagination: {
          rowsPerPage: 10
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
    },
    pendingPaymentsExist() {
      return this.payments.findIndex(payment => payment.pending) !== -1
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
    this.$events.on('invoice-paid', payment => {
      // TODO
      // if (this.receive.paymentHash === paymentHash) {
      //   this.receive.show = false
      //   this.receive.paymentHash = null
      //   clearInterval(this.receive.paymentChecker)
      // }
    })
    this.$events.on('payment-complete', payment => {
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
    this.$events.on('payment-failed', payment => {
      // TODO
    })
  },

  methods: {
    paymentTableRowKey(row) {
      return row.payment_hash + row.amount
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
      this.receive.data.memo = null
      this.receive.unit = 'sat'
      this.receive.paymentChecker = null
      this.receive.minMax = [0, 2100000000000000]
      this.receive.lnurl = null
    },
    showParseDialog() {
      this.parse.show = true
      this.parse.invoice = null
      this.parse.lnurlpay = null
      this.parse.lnurlauth = null
      this.parse.data.request = ''
      this.parse.data.comment = ''
      this.parse.data.paymentChecker = null
      this.parse.camera.show = false
    },
    closeReceiveDialog() {
      setTimeout(() => {
        clearInterval(this.receive.paymentChecker)
      }, 10000)
    },
    closeParseDialog() {
      setTimeout(() => {
        clearInterval(this.parse.paymentChecker)
      }, 10000)
    },
    async createInvoice() {
      this.receive.status = 'loading'

      try {
        const response = await createInvoice(this.$store.state.wallet.id, {
          msatoshi: this.receive.data.amount * 1000,
          memo: this.receive.data.memo,
          unit: this.receive.unit,
          lnurlCallback: this.receive.lnurl && this.receive.lnurl.callback
        })

        this.receive.status = 'success'
        this.receive.paymentReq = response.data.payment_request
        this.receive.paymentHash = response.data.payment_hash

        if (response.data.lnurl_response !== null) {
          if (response.data.lnurl_response === false) {
            response.data.lnurl_response = `Unable to connect`
          }

          if (typeof response.data.lnurl_response === 'string') {
            // failure
            this.$q.notify({
              timeout: 5000,
              type: 'warning',
              message: `${this.receive.lnurl.domain} lnurl-withdraw call failed.`,
              caption: response.data.lnurl_response
            })
            return
          } else if (response.data.lnurl_response === true) {
            // success
            this.$q.notify({
              timeout: 5000,
              message: `Invoice sent to ${this.receive.lnurl.domain}!`,
              spinner: true
            })
          }
        }
      } catch (err) {
        notifyApiError(err)
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
        this.parse.data.request.toLowerCase().startsWith('lnurl1') ||
        this.parse.data.request.match(/[\w.+-~_]+@[\w.+-~_]/)
      ) {
        try {
          const response = await scanLnurl(
            this.$store.state.wallet.id,
            this.parse.data.request
          )

          let data = response.data

          if (data.status === 'ERROR') {
            this.$q.notify({
              timeout: 5000,
              type: 'warning',
              message: `${data.domain} lnurl call failed.`,
              caption: data.reason
            })
            return
          }

          if (data.kind === 'pay') {
            this.parse.lnurlpay = Object.freeze(data)
            this.parse.data.amount = data.minSendable / 1000
          } else if (data.kind === 'auth') {
            this.parse.lnurlauth = Object.freeze(data)
          } else if (data.kind === 'withdraw') {
            this.parse.show = false
            this.receive.show = true
            this.receive.status = 'pending'
            this.receive.paymentReq = null
            this.receive.paymentHash = null
            this.receive.data.amount = data.maxWithdrawable / 1000
            this.receive.data.memo = data.defaultDescription
            this.receive.minMax = [
              data.minWithdrawable / 1000,
              data.maxWithdrawable / 1000
            ]
            this.receive.lnurl = {
              domain: data.domain,
              callback: data.callback,
              fixed: data.fixed
            }
          }
        } catch (err) {
          notifyApiError(err)
        }
        return
      }

      try {
        const invoice = bolt11.decode(this.parse.data.request)

        let cleanInvoice = {
          msat: invoice.human_readable_part.amount,
          sat: invoice.human_readable_part.amount / 1000
        }

        invoice.data.tags.forEach(tag => {
          if (tag && 'description' in tag) {
            if (tag.description === 'payment_hash') {
              cleanInvoice.hash = tag.value
            } else if (tag.description === 'description') {
              cleanInvoice.description = tag.value
            } else if (tag.description === 'expiry') {
              var expireDate = new Date(
                (invoice.data.time_stamp + tag.value) * 1000
              )
              cleanInvoice.expireDate = this.$q.utils.date.formatDate(
                expireDate,
                'YYYY-MM-DDTHH:mm:ss.SSSZ'
              )
              cleanInvoice.expired = false // TODO
            }
          }
        })

        this.parse.invoice = cleanInvoice
      } catch (error) {
        this.$q.notify({
          timeout: 3000,
          type: 'warning',
          message: error + '.',
          caption: '400 BAD REQUEST'
        })
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
        await payInvoice(this.$store.state.wallet.id, this.parse.data.request)
      } catch (err) {
        dismissPaymentMsg()
        notifyApiError(err)
      }
    },
    async payLnurl() {
      let dismissPaymentMsg = this.$q.notify({
        timeout: 0,
        message: 'Processing payment...'
      })

      try {
        await payLnurl(this.$store.state.wallet.id, {
          callback: this.parse.lnurlpay.callback,
          descriptionHash: this.parse.lnurlpay.description_hash,
          msatoshi: this.parse.data.amount * 1000,
          description: this.parse.lnurlpay.description.slice(0, 120),
          comment: this.parse.data.comment
        })

        this.parse.show = false
      } catch (err) {
        dismissPaymentMsg()
        notifyApiError(err)
      }
    },
    async authLnurl() {
      let dismissAuthMsg = this.$q.notify({
        timeout: 10,
        message: 'Performing authentication...'
      })

      try {
        await authLnurl(
          this.$store.state.wallet.id,
          this.parse.lnurlauth.callback
        )
        dismissAuthMsg()
        this.$q.notify({
          message: `Authentication successful.`,
          type: 'positive',
          timeout: 3500
        })
        this.parse.show = false
      } catch (err) {
        dismissAuthMsg()
        if (err.response.data.reason) {
          this.$q.notify({
            message: `Authentication failed. ${this.parse.lnurlauth.domain} says:`,
            caption: err.response.data.reason,
            type: 'warning',
            timeout: 5000
          })
        } else {
          notifyApiError(err)
        }
      }
    },
    async updateWalletName() {
      let newName = this.newName
      if (!newName || !newName.length) return

      try {
        await renameWallet(this.$store.state.wallet.id, newName)

        this.newName = ''
        this.$q.notify({
          message: `Wallet named updated.`,
          type: 'positive',
          timeout: 3500
        })

        this.$store.dispatch('fetchWallet', this.$store.state.wallets.id)
      } catch (err) {
        this.newName = ''
        notifyApiError(err)
      }
    },
    deleteWallet(walletId, user) {
      this.$q.plugins.Dialog.create({
        message: 'Are you sure you want to delete this wallet?',
        ok: {
          flat: true,
          color: 'orange'
        },
        cancel: {
          flat: true,
          color: 'grey'
        }
      }).onOk(async () => {
        await deleteWallet(walletId)
        location.href = `/?${location.search}`
      })
    },
    exportCSV() {
      exportCSV(this.paymentsTable.columns, this.payments)
    }
  }
}
</script>
