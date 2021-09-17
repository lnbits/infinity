import store from '../store'

const routes = [
  {
    path: '/',
    component: () => import('layouts/MainLayout.vue'),
    beforeEnter: async (to, from, next) => {
      await store.dispatch('fetchUser')
      next()
    },
    children: [
      {path: '', component: () => import('pages/Index.vue')},
      {
        path: '/wallet/:id',
        component: () => import('pages/Wallet.vue'),
        beforeEnter: async (to, from, next) => {
          await store.dispatch('fetchWallet', to.params.id)
          next()
        }
      }
    ]
  },
  {
    path: '/:catchAll(.*)*',
    component: () => import('pages/Error404.vue')
  }
]

export default routes
