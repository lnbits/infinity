const routes = [
  {
    path: '/',
    component: () => import('layouts/MainLayout.vue'),
    children: [
      {
        path: '',
        component: () => import('pages/Index.vue'),
        name: 'index'
      },
      {
        path: '/wallet/:id',
        component: () => import('pages/Wallet.vue'),
        name: 'wallet'
      },
      {
        path: '/wallet/:id/app/:appid',
        component: () => import('pages/CustomApp.vue'),
        name: 'app'
      }
    ]
  },
  {
    path: '/:catchAll(.*)*',
    component: () => import('pages/Error404.vue')
  }
]

export default routes
