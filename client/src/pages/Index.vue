<template>
  <q-page class="q-px-md q-py-lg" :class="{'q-px-lg': $q.screen.gt.xs}">
    <div class="row q-col-gutter-md justify-center">
      <div class="col-12 col-md-7 col-lg-6 q-gutter-y-md">
        <q-card>
          <q-card-section>
            <div v-if="lnurlvoucher" class="q-gutter-md">
              <h5 class="q-my-md">
                You have received a voucher containing some satoshis!
              </h5>
              <p>
                You can click the button below to instantly create a wallet on
                this site and redeem the voucher.
              </p>
              <q-btn
                unelevated
                color="primary"
                type="a"
                :href="'/lnurlwallet?lightning=' + lnurlvoucher"
                @click="processing"
              >
                Claim your satoshis
              </q-btn>
            </div>
            <q-form v-else class="q-gutter-md" @submit="createWallet">
              <q-input
                v-model="walletName"
                filled
                dense
                :label="`Name your ${$store.state.settings.siteTitle} wallet *`"
              ></q-input>
              <q-btn
                unelevated
                color="primary"
                :disable="walletName == ''"
                type="submit"
                >Add a new wallet</q-btn
              >
            </q-form>
          </q-card-section>
        </q-card>

        <q-card v-if="storedKeys.length > 0">
          <q-card-section>
            <h5 class="q-my-md">Accounts previously used on this browser</h5>
            <div class="q-pa-md" style="max-width: 350px">
              <q-list bordered separator>
                <q-item
                  v-for="key in storedKeys"
                  :key="key"
                  v-ripple
                  clickable
                  @click="useKey(key)"
                >
                  <q-item-section>
                    <q-item-label>
                      <span v-if="walletsForKey[key]">Wallets: </span>
                      <span
                        v-for="wname in walletsForKey[key]"
                        :key="wname"
                        class="q-mr-xs text-weight-bold"
                        >"{{ wname }}"
                      </span>
                    </q-item-label>
                    <q-item-label caption>
                      Master Key: {{ key.slice(0, 7) }}...
                    </q-item-label>
                  </q-item-section>
                </q-item>
              </q-list>
            </div>
          </q-card-section>
        </q-card>

        <q-card>
          <q-card-section>
            <h3 class="q-my-none">{{ $store.state.settings.siteTitle }}</h3>
            <h5 class="q-my-md">{{ $store.state.settings.siteTagline }}</h5>
            <div v-if="$store.state.settings.siteTitle == 'LNbits'">
              <p>
                Easy to set up and lightweight, LNbits can run on any
                lightning-network funding source, currently supporting LND,
                c-lightning, OpenNode, lntxbot, LNPay and even LNbits itself!
              </p>
              <p>
                You can run LNbits for yourself, or easily offer a custodian
                solution for others.
              </p>
              <p>
                Each wallet has its own API keys and there is no limit to the
                number of wallets you can make. Being able to partition funds
                makes LNbits a useful tool for money management and as a
                development tool.
              </p>
              <p>
                Extensions add extra functionality to LNbits so you can
                experiment with a range of cutting-edge technologies on the
                lightning network. We have made developing extensions as easy as
                possible, and as a free and open-source project, we encourage
                people to develop and submit their own.
              </p>
              <div class="row q-mt-md q-gutter-sm">
                <q-btn
                  outline
                  color="grey"
                  type="a"
                  href="https://github.com/lnbits/lnbits-infinity"
                  target="_blank"
                  rel="noopener"
                  >View project in GitHub</q-btn
                >
                <q-btn
                  outline
                  color="grey"
                  type="a"
                  href="https://legend.lnbits.com/paywall/GAqKguK5S8f6w5VNjS9DfK"
                  target="_blank"
                  rel="noopener"
                  >Donate</q-btn
                >
              </div>
            </div>
            <p v-else>{{ $store.state.settings.siteDescription }}</p>
          </q-card-section>
        </q-card>
      </div>

      <div v-if="'{{SITE_TITLE}}' == 'LNbits'" class="col-12 col-md-3 col-lg-3">
        <div class="row q-col-gutter-lg justify-center">
          <div class="col-6 col-sm-4 col-md-8 q-gutter-y-sm">
            <q-btn
              flat
              color="secondary"
              label="Runs on"
              class="full-width"
            ></q-btn>
            <div class="row">
              <div class="col">
                <a href="https://github.com/ElementsProject/lightning">
                  <q-img
                    contain
                    :src="
                      $q.dark.isActive
                        ? '/static/images/cln.png'
                        : '/static/images/clnl.png'
                    "
                  ></q-img>
                </a>
              </div>
              <div class="col q-pl-md">
                <a href="https://github.com/lightningnetwork/lnd">
                  <q-img
                    contain
                    :src="
                      $q.dark.isActive
                        ? '/static/images/lnd.png'
                        : '/static/images/lnd.png'
                    "
                  ></q-img>
                </a>
              </div>
            </div>

            <div class="row">
              <div class="col">
                <a href="https://opennode.com">
                  <q-img
                    contain
                    :src="
                      $q.dark.isActive
                        ? '/static/images/opennode.png'
                        : '/static/images/opennodel.png'
                    "
                  ></q-img>
                </a>
              </div>
              <div class="col q-pl-md">
                <a href="https://lnpay.co/">
                  <q-img
                    contain
                    :src="
                      $q.dark.isActive
                        ? '/static/images/lnpay.png'
                        : '/static/images/lnpayl.png'
                    "
                  ></q-img>
                </a>
              </div>
            </div>

            <div class="row">
              <div class="col">
                <a href="https://github.com/shesek/spark-wallet">
                  <q-img
                    contain
                    :src="
                      $q.dark.isActive
                        ? '/static/images/spark.png'
                        : '/static/images/sparkl.png'
                    "
                  ></q-img>
                </a>
              </div>
              <div class="col q-pl-md">
                <a href="https://t.me/lntxbot">
                  <q-img
                    contain
                    :src="
                      $q.dark.isActive
                        ? '/static/images/lntxbot.png'
                        : '/static/images/lntxbotl.png'
                    "
                  ></q-img>
                </a>
              </div>
            </div>
            <div class="row">
              <div class="col">
                <a href="https://github.com/rootzoll/raspiblitz">
                  <q-img
                    contain
                    :src="
                      $q.dark.isActive
                        ? '/static/images/blitz.png'
                        : '/static/images/blitzl.png'
                    "
                  ></q-img>
                </a>
              </div>
              <div class="col q-pl-md">
                <a href="https://getumbrel.com/">
                  <q-img
                    contain
                    :src="
                      $q.dark.isActive
                        ? '/static/images/umbrel.png'
                        : '/static/images/umbrell.png'
                    "
                  ></q-img>
                </a>
              </div>
            </div>
            <div class="row">
              <div class="col">
                <a href="https://mynodebtc.com/">
                  <q-img
                    contain
                    :src="
                      $q.dark.isActive
                        ? '/static/images/mynode.png'
                        : '/static/images/mynodel.png'
                    "
                  ></q-img>
                </a>
              </div>
              <div class="col q-pl-md"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </q-page>
</template>

<script>
import {LocalStorage} from 'quasar'
import {notifyError} from '../helpers'
import {createWallet, loadUser} from '../api'

export default {
  name: 'PageIndex',
  data() {
    return {
      lnurlvoucher: new URLSearchParams(location.search).get('lightning'),
      walletName: '',
      storedKeys: LocalStorage.getItem('lnbits.storedkeys') || [],
      walletsForKey: {}
    }
  },
  created() {
    this.storedKeys.forEach(async key => {
      const userData = await loadUser(key)
      this.walletsForKey[key] = userData.wallets.map(w => w.name)
    })
  },
  methods: {
    async createWallet() {
      try {
        const {userMasterKey, wallet} = await createWallet(this.walletName)
        const query = {...this.$route.query}
        if (userMasterKey) {
          query.key = userMasterKey
          this.storedKeys.unshift(userMasterKey)
          LocalStorage.set('lnbits.storedkeys', this.storedKeys)
        }
        this.$store.commit('setWallet', wallet)
        await this.$router.push({path: `/wallet/${wallet.id}`, query})
      } catch (err) {
        notifyError(err)
      }

      await this.$store.dispatch('fetchUser')
    },
    async useKey(key) {
      const query = {...this.$route.query, key}
      await this.$router.replace({
        path: this.$route.path,
        query
      })
      await this.$store.dispatch('fetchUser')
      this.$router.push({
        path: `/wallet/${this.$store.state.wallet.id}`,
        query
      })
    },
    processing() {
      this.$q.notify({
        timeout: 0,
        message: 'Processing...',
        icon: null
      })
    }
  }
}
</script>
