const routes = [
  {
    path: '/',
    component: () => import('layouts/MainLayout.vue'),
    children: [
      {path: '', component: () => import('pages/Index.vue')},
      {
        path: '/wallet/:id',
        component: () => import('pages/Wallet.vue')
      },
      {
        path: '/wallet/:id/app/:appid',
        component: () => import('pages/CustomApp.vue')
      }
    ]
  },
  {
    path: '/:catchAll(.*)*',
    component: () => import('pages/Error404.vue')
  }
]

export default routes
