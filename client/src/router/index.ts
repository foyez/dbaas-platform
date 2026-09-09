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

import DocsLayout from '@/components/docs/DocsLayout.vue'
import DocsOverviewPage from '@/pages/docs/DocsOverviewPage.vue'
import DocsGettingStartedPage from '@/pages/docs/DocsGettingStartedPage.vue'
import DocsArchitecturePage from '@/pages/docs/DocsArchitecturePage.vue'
import DocsApiPage from '@/pages/docs/DocsApiPage.vue'
import DocsSdkPage from '@/pages/docs/DocsSdkPage.vue'

const router = createRouter({
  history: createWebHistory(),

  routes: [
    // public - marketing homepage
    { path: '/', name: 'landing', component: LandingPage },

    // public - documentation
    {
      path: '/docs',
      component: DocsLayout,
      children: [
        { path: '', name: 'docs-overview', component: DocsOverviewPage },
        { path: 'getting-started', name: 'docs-getting-started', component: DocsGettingStartedPage },
        { path: 'architecture', name: 'docs-architecture', component: DocsArchitecturePage },
        { path: 'api', name: 'docs-api', component: DocsApiPage },
        { path: 'sdk', name: 'docs-sdk', component: DocsSdkPage },
      ],
    },

    // protected - the actual app
    {
      path: '/dashboard',
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

    { path: '/auth/callback', name: 'callback', component: CallbackPage },
  ],
})

router.beforeEach(async (to) => {
  if (!to.matched.some((record) => record.meta.requiresAuth)) {
    return true
  }

  const user = await userManager.getUser()

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