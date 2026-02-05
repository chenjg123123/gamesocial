import { createRouter, createWebHistory } from 'vue-router'

import { getToken } from '../lib/storage'

import AdminPage from '../views/admin/AdminPage.vue'
import UserLayout from '../views/user/UserLayout.vue'
import UserHomePage from '../views/user/UserHomePage.vue'
import UserTasksPage from '../views/user/UserTasksPage.vue'
import UserShopPage from '../views/user/UserShopPage.vue'
import UserTournamentsPage from '../views/user/UserTournamentsPage.vue'
import UserMePage from '../views/user/UserMePage.vue'
import UserLoginPage from '../views/user/UserLoginPage.vue'

export const ADMIN_PATH = '/gs-admin'

export const router = createRouter({
  history: createWebHistory('/web/'), // 关键：base必须是/web/
  routes: [
    { path: '/', redirect: '/user/index' },
    { path: '/login', component: UserLoginPage },
    {
      path: '/user',
      component: UserLayout,
      children: [
        { path: 'index', component: UserHomePage },
        { path: 'tasks', component: UserTasksPage },
        { path: 'shop', component: UserShopPage },
        { path: 'tournaments', component: UserTournamentsPage },
        { path: 'me', component: UserMePage },
        { path: '', redirect: '/user/index' },
      ],
    },
    { path: ADMIN_PATH, component: AdminPage },
  ],
})

router.beforeEach(to => {
  if (to.path === '/login') return true

  const needsAuth = to.path.startsWith('/user') || to.path === ADMIN_PATH
  if (!needsAuth) return true

  const token = getToken()
  if (token) return true

  return { path: '/login', query: { redirect: to.fullPath } }
})
