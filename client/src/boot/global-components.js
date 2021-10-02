import VueQrcodeReader from 'vue-qrcode-reader'
import VueQrcode from 'vue-qrcode'

import APIDocs from '../components/APIDocs'
import AppList from '../components/AppList'
import WalletList from '../components/WalletList'
import CustomAppModel from '../components/CustomAppModel'
import PaymentDetails from '../components/PaymentDetails'
import LnurlPaySuccessAction from '../components/LnurlPaySuccessAction'

export default ({app}) => {
  app.use(VueQrcodeReader)
  app.component(VueQrcode)

  app.component('APIDocs', APIDocs)
  app.component('AppList', AppList)
  app.component('WalletList', WalletList)
  app.component('CustomAppModel', CustomAppModel)
  app.component('PaymentDetails', PaymentDetails)
  app.component('LnurlPaySuccessAction', LnurlPaySuccessAction)
}
