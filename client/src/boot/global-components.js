import WalletList from '../components/WalletList'
import AppList from '../components/AppList'
import PaymentDetails from '../components/PaymentDetails'
import LnurlPaySuccessAction from '../components/LnurlPaySuccessAction'

export default ({app}) => {
  app.component('WalletList', WalletList)
  app.component('AppList', AppList)
  app.component('PaymentDetails', PaymentDetails)
  app.component('LnurlPaySuccessAction', LnurlPaySuccessAction)
}
