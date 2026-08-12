import { createRouter, createWebHistory } from 'vue-router'

import CreateInstancePage from '@/pages/CreateInstancePage.vue'
import InstanceDetailPage from '@/pages/InstanceDetailPage.vue'
import InstancesPage from '@/pages/InstancesPage.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import DashboardPage from '@/pages/DashboardPage.vue'
import LandingPage from '@/pages/LandingPage.vue'
import CallbackPage from '@/pages/auth/CallbackPage.vue'
import { userManager } from '@/auth/oidc'

const router = createRouter({
  history: createWebHistory(),

  routes: [
    // protected
    {
      path: '/',
      component: AppLayout,
      meta: {requiresAuth: true},

      children: [
        {
          path: '',
          name: 'dashboard',
          component: DashboardPage,
        },

        {
          path: 'instances',
          name: 'instances',
          component: InstancesPage,
        },

        {
          path: 'instances/new',
          name: 'instance-create',
          component: CreateInstancePage,
        },

        {
          path: 'instances/:id',
          name: 'instance-detail',
          component: InstanceDetailPage,
        },
      ],
    },

    // public routes - outside the protected layout
    { path: '/landing', name: 'landing', component: LandingPage },
    { path: '/auth/callback', name: 'callback', component: CallbackPage },
  ],
})

router.beforeEach(async (to) => {
  if (!to.matched.some((record) => record.meta.requiresAuth)) {
    return true
  }

  const user = await userManager.getUser()

  if (user && !user.expired) {
    return true
  }

  await userManager.signinRedirect({
    state: {
      returnUrl: to.fullPath,
    },
  })

  return false
})

export default router
