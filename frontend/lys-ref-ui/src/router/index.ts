import { nextTick } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import auth from '@/auth'
import ax from '@/api'

// https://router.vuejs.org/guide/advanced/meta.html
declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth: boolean
  }
}

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/Login.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    redirect: 'home',
    component: () => import('@/layouts/Main.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: 'home',
        name: 'Home',
        component: () => import('@/pages/index.vue'),
      },
      {
        path: 'core',
        children: [
          {
            path: 'arrays',
            name: 'Arrays',
            component: () => import('@/pages/CoreArrayTypes.vue'),
          },
          {
            path: 'default-values',
            name: 'Default values',
            component: () => import('@/pages/CoreDefaultValues.vue'),
          },
          {
            path: 'mandatory-values',
            name: 'Mandatory values',
            component: () => import('@/pages/CoreMandatoryValues.vue'),
          },
          {
            path: 'optional-values',
            name: 'Optional values',
            component: () => import('@/pages/CoreOptionalValues.vue'),
          },
          {
            path: 'variants',
            name: 'Variants',
            component: () => import('@/pages/CoreVariantTypes.vue'),
          },
        ]
      },
      {
        path: 'digital-marketing',
        children: [
          {
            path: 'campaigns',
            name: 'Campaigns',
            component: () => import('@/pages/DmCampaigns.vue'),
          },
          {
            path: 'campaign-charts',
            name: 'Campaign charts',
            component: () => import('@/pages/DmCampaignCharts.vue'),
          },
          {
            path: 'campaign-optimizer',
            name: 'Campaign optimizer',
            component: () => import('@/pages/DmCampaignOpt.vue'),
          },
          {
            path: 'campaign-performance',
            name: 'Campaign performance',
            component: () => import('@/pages/DmCampaignPerf.vue'),
          },
          {
            path: 'mcp-server',
            name: 'MCP Server',
            component: () => import('@/pages/DmMcpServer.vue'),
          },
          {
            path: 'verticals',
            name: 'Verticals',
            component: () => import('@/pages/DmVerticals.vue'),
          },
        ]
      },
      {
        path: 'ecb',
        children: [
          {
            path: 'currencies',
            name: 'Currencies',
            component: () => import('@/pages/EcbCurrencies.vue'),
          },
          {
            path: 'exchange-rates',
            name: 'Exchange rates',
            component: () => import('@/pages/EcbExchangeRates.vue'),
          },
          {
            path: 'xr-performance',
            name: 'XR performance',
            component: () => import('@/pages/EcbXrPerformance.vue'),
          },
        ]
      },
      {
        path: 'maxmind',
        children: [
          {
            path: 'geo-ip',
            name: 'Geo IP',
            component: () => import('@/pages/MmGeoIp.vue'),
          },
        ]
      },
      {
        path: 'monitoring',
        children: [
          {
            path: 'application',
            name: 'Application',
            component: () => import('@/pages/AppMonitor.vue'),
          },
          {
            path: 'database',
            name: 'Database',
            component: () => import('@/pages/PgMonitor.vue'),
          },
        ]
      },
      {
        path: '/process/flows',
        children: [
          {
            path: '',
            name: 'Flows',
            component: () => import('@/pages/ProcFlows.vue'),
          },
          {
            path: ':id/steps',
            props: (route: any) => {
              return {
                flow_id: parseInt(route.params.id),
              }
            },
            name: 'Steps',
            component: () => import('@/pages/ProcSteps.vue'),
          },
        ]
      },
      {
        path: '/process/runs',
        children: [
          {
            path: '',
            name: 'Runs',
            component: () => import('@/pages/ProcRuns.vue'),
          },
          {
            path: ':id/points',
            props: (route: any) => {
              return {
                run_id: parseInt(route.params.id),
              }
            },
            name: 'Points',
            component: () => import('@/pages/ProcPoints.vue'),
          },
        ]
      },
      {
        path: 'publisher',
        children: [
          {
            path: 'authors',
            name: 'Authors',
            component: () => import('@/pages/PubAuthors.vue'),
          },
          {
            path: 'books',
            name: 'Books',
            component: () => import('@/pages/PubBooks.vue'),
          },
        ]
      },
      {
        path: 'server-push',
        children: [
          {
            path: 'notifications',
            name: 'Notifications',
            component: () => import('@/pages/ServPNotifications.vue'),
          },
        ]
      },
      {
        path: 'supplier',
        children: [
          {
            path: 'internal-view',
            name: 'Internal view',
            component: () => import('@/pages/SuppInternalView.vue'),
          },
          {
            path: 'tenant-view',
            name: 'Tenant view',
            component: () => import('@/pages/SuppTenantView.vue'),
          },
        ]
      },
    ]
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach(async (to, from, next) => {
  
  // ensure auth is ready before checking if user is authenticated
  // should not be needed, since auth is bootstrapped in main.ts before router is plugged in, but just in case
  if (!auth.ready) {
    await auth.bootstrap()
  }

  // nav guard: redirect to /login if the page requires auth and the user is not authed
  if (to.matched.some(r => r.meta.requiresAuth) && !auth.user.authenticated) {
    return next({
      path: '/login',
      query: {
        to: encodeURIComponent(JSON.stringify({ name: to.name, params: to.params })),
      },
    })
  }

  next()
})

router.afterEach((to, from) => {
  nextTick(() => {
    // set the page title to be the route name followed by the appName
    document.title = to.name?.toString() + ' | LoveYourStack - Reference'
  })
})

export default router
