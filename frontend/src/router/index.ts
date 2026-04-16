import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';
import { useAuthStore } from '@/store/auth';

const Login = () => import('@/views/Login.vue');
const CertificateList = () => import('@/views/CertificateList.vue');
const CertificateApply = () => import('@/views/CertificateApply.vue');
const CertificateDetail = () => import('@/views/CertificateDetail.vue');
const CertificateGenerateDns = () => import('@/views/CertificateGenerateDns.vue');
const Balance = () => import('@/views/Balance.vue');
const Recharge = () => import('@/views/Recharge.vue');

const AdminLogin = () => import('@/views/admin/AdminLogin.vue');
const AdminCertificates = () => import('@/views/admin/CertificateManage.vue');
const AdminCertCreate = () => import('@/views/admin/CertificateCreate.vue');
const AdminCertEdit = () => import('@/views/admin/CertificateEdit.vue');

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: Login
  },
  {
    path: '/',
    redirect: '/certificates'
  },
  {
    path: '/',
    component: () => import('@/layouts/UserLayout.vue'),
    children: [
      {
        path: '/certificates',
        name: 'CertificateList',
        component: CertificateList,
        meta: { requiresAuth: true }
      },
      {
        path: '/certificates/apply',
        name: 'CertificateApply',
        component: CertificateApply,
        meta: { requiresAuth: true }
      },
      {
        path: '/certificates/:id',
        name: 'CertificateDetail',
        component: CertificateDetail,
        meta: { requiresAuth: true }
      },
      {
        path: '/certificates/generate-dns',
        name: 'CertificateGenerateDns',
        component: CertificateGenerateDns,
        meta: { requiresAuth: true }
      },
      {
        path: '/balance',
        name: 'Balance',
        component: Balance,
        meta: { requiresAuth: true }
      },
      {
        path: '/balance/recharge',
        name: 'Recharge',
        component: Recharge,
        meta: { requiresAuth: true }
      }
    ]
  },
  {
    path: '/admin/login',
    name: 'AdminLogin',
    component: AdminLogin
  },
  {
    path: '/admin',
    component: () => import('@/layouts/AdminLayout.vue'),
    children: [
      {
        path: '/admin/certificates',
        name: 'AdminCertificates',
        component: AdminCertificates,
        meta: { requiresAdmin: true }
      },
      {
        path: '/admin/certificates/create',
        name: 'AdminCertCreate',
        component: AdminCertCreate,
        meta: { requiresAdmin: true }
      },
      {
        path: '/admin/certificates/:id/edit',
        name: 'AdminCertEdit',
        component: AdminCertEdit,
        meta: { requiresAdmin: true }
      }
    ]
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore();
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    next({ name: 'Login' });
    return;
  }
  if (to.meta.requiresAdmin && (!auth.isLoggedIn || auth.user?.role !== 'admin')) {
    next({ name: 'AdminLogin' });
    return;
  }
  next();
});

export default router;


