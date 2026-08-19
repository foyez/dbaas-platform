import { createRouter, createWebHistory } from 'vue-router'

import CreateInstancePage from '@/pages/CreateInstancePage.vue'
import InstanceDetailPage from '@/pages/InstanceDetailPage.vue'
import InstancesPage from '@/pages/InstancesPage.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import DashboardPage from '@/pages/DashboardPage.vue'
import LandingPage from '@/pages/LandingPage.vue'
import CallbackPage from '@/pages/auth/CallbackPage.vue'
import { userManager } from '@/auth/oidc'
import AuditLogsPage from '@/pages/AuditLogsPage.vue'
import { hasRole } from '@/auth/roles'

const router = createRouter({
  history: createWebHistory(),

  routes: [
    // protected
    {
      path: '/',
      component: AppLayout,
      meta: { requiresAuth: true },

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

        {
          path: 'audit-logs',
          name: 'audit-logs',
          component: AuditLogsPage,
          meta: { requiresAuth: true, requiresAdmin: true }, // guard not yet enforced - see note above
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
  console.log(user?.profile)

  if (!user || user.expired) {
    await userManager.signinRedirect({ state: { returnUrl: to.fullPath } })
    return false
  }

  if (to.meta.requiresAdmin && !hasRole(user, 'dbaas.admin')) {
    return { name: 'dashboard' } // or a dedicated 403 page, if you have one
  }

  return true
})

export default router
