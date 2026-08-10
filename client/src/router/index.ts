import { createRouter, createWebHistory } from 'vue-router'

import CreateInstancePage from '@/pages/CreateInstancePage.vue'
import Home from '@/pages/Home.vue'
import InstanceDetailPage from '@/pages/InstanceDetailPage.vue'
import InstancesPage from '@/pages/InstancesPage.vue'

const router = createRouter({
  history: createWebHistory(),

  routes: [
    {
      path: '/instances',
      component: InstancesPage,
    },
    {
      path: '/instances/new',
      component: CreateInstancePage,
    },
    {
      path: '/instances/:id',
      name: 'instance-detail',
      component: InstanceDetailPage,
    },
    {
      path: '/',
      name: 'home',
      component: Home,
    },
  ],
})

export default router
