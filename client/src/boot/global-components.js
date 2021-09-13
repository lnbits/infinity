import VueQrcodeReader from 'vue-qrcode-reader'
import VueQrcode from 'vue-qrcode'

import APIDocs from '../components/APIDocs'
import WalletList from '../components/WalletList'
import AppList from '../components/AppList'
import PaymentDetails from '../components/PaymentDetails'
import LnurlPaySuccessAction from '../components/LnurlPaySuccessAction'

export default ({app}) => {
  app.use(VueQrcodeReader)
  app.component(VueQrcode)

  app.component('APIDocs', APIDocs)
  app.component('WalletList', WalletList)
  app.component('AppList', AppList)
  app.component('PaymentDetails', PaymentDetails)
  app.component('LnurlPaySuccessAction', LnurlPaySuccessAction)
}
