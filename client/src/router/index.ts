import { createRouter, createWebHistory } from 'vue-router'

import CreateInstancePage from '@/pages/CreateInstancePage.vue'
import InstanceDetailPage from '@/pages/InstanceDetailPage.vue'
import InstancesPage from '@/pages/InstancesPage.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import DashboardPage from '@/pages/DashboardPage.vue'

const router = createRouter({
  history: createWebHistory(),

  routes: [
    {
      path: '/',
      component: AppLayout,

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
  ],
})

export default router
