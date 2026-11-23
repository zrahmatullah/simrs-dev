import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import MainLayout from '@/layouts/MainLayout.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/LoginView.vue'),
    meta: { public: true },
  },
  {
    path: '/display-queue',
    name: 'DisplayQueue',
    component: () => import('@/views/public/DisplayQueueView.vue'),
    meta: { public: true },
  },
  {
    path: '/verify-doc',
    name: 'VerifyDoc',
    component: () => import('@/views/public/VerifyDocView.vue'),
    meta: { public: true },
  },
  {
    path: '/',
    component: MainLayout,
    redirect: () => {
      const auth = useAuthStore()
      if (auth.isAuthenticated) {
        return '/dashboard'
      }
      return '/login'
    },
    children: [
      // 0. Welcome Page Dashboard (For all 9 roles)
      {
        path: 'dashboard',
        name: 'RoleDashboard',
        component: () => import('@/views/dashboard/RoleDashboardView.vue'),
        meta: { roles: ['SUPERADMIN', 'REGISTRATION', 'CALL_CENTER', 'NURSE', 'DOCTOR', 'PHARMACY', 'SUPPORT', 'LOGISTICS', 'CASHIER'] },
      },

      // Reports (Financial, Pharmacy, Registrations, Demographics)
      {
        path: 'reports',
        name: 'Reports',
        component: () => import('@/views/reports/ReportsView.vue'),
        meta: { roles: ['SUPERADMIN', 'REGISTRATION', 'CALL_CENTER', 'NURSE', 'DOCTOR', 'PHARMACY', 'SUPPORT', 'LOGISTICS', 'CASHIER'] },
      },
      {
        path: 'reports/financial',
        name: 'FinancialReport',
        component: () => import('@/views/reports/ReportsView.vue'),
        meta: { roles: ['SUPERADMIN', 'CASHIER'] },
      },
      {
        path: 'reports/pharmacy-stock',
        name: 'PharmacyStockReport',
        component: () => import('@/views/reports/ReportsView.vue'),
        meta: { roles: ['SUPERADMIN', 'PHARMACY', 'LOGISTICS'] },
      },
      {
        path: 'reports/patient-registrations',
        name: 'PatientRegistrationsReport',
        component: () => import('@/views/reports/ReportsView.vue'),
        meta: { roles: ['SUPERADMIN', 'REGISTRATION', 'CALL_CENTER', 'NURSE', 'DOCTOR'] },
      },
      {
        path: 'reports/demographics',
        name: 'DemographicsReport',
        component: () => import('@/views/reports/ReportsView.vue'),
        meta: { roles: ['SUPERADMIN', 'REGISTRATION', 'DOCTOR', 'NURSE'] },
      },

      // 1. Superadmin
      {
        path: 'admin',
        name: 'SuperadminDashboard',
        component: () => import('@/views/admin/SuperadminView.vue'),
        meta: { roles: ['SUPERADMIN'] },
      },
      {
        path: 'admin/users',
        name: 'UserManagement',
        component: () => import('@/views/admin/UserManagementView.vue'),
        meta: { roles: ['SUPERADMIN'] },
      },
      {
        path: 'admin/master',
        name: 'MasterData',
        component: () => import('@/views/admin/MasterDataView.vue'),
        meta: { roles: ['SUPERADMIN'] },
      },
      {
        path: 'admin/audit',
        name: 'AuditTrail',
        component: () => import('@/views/admin/AuditTrailView.vue'),
        meta: { roles: ['SUPERADMIN'] },
      },

      // 2. Registration
      {
        path: 'registration',
        name: 'Registration',
        component: () => import('@/views/registration/RegistrationView.vue'),
        meta: { roles: ['SUPERADMIN', 'REGISTRATION'] },
      },
      {
        path: 'registration/patients',
        name: 'PatientsList',
        component: () => import('@/views/registration/PatientsListView.vue'),
        meta: { roles: ['SUPERADMIN', 'REGISTRATION', 'CALL_CENTER', 'NURSE', 'DOCTOR'] },
      },

      // 3. Call Center
      {
        path: 'call-center',
        name: 'CallCenter',
        component: () => import('@/views/callcenter/CallCenterView.vue'),
        meta: { roles: ['SUPERADMIN', 'CALL_CENTER'] },
      },

      // 4. Nurse
      {
        path: 'nurse',
        name: 'Nurse',
        component: () => import('@/views/nurse/NurseView.vue'),
        meta: { roles: ['SUPERADMIN', 'NURSE'] },
      },

      // 5. Doctor
      {
        path: 'doctor',
        name: 'Doctor',
        component: () => import('@/views/doctor/DoctorView.vue'),
        meta: { roles: ['SUPERADMIN', 'DOCTOR'] },
      },

      // 6. Pharmacy
      {
        path: 'pharmacy',
        name: 'Pharmacy',
        component: () => import('@/views/pharmacy/PharmacyView.vue'),
        meta: { roles: ['SUPERADMIN', 'PHARMACY'] },
      },

      // 7. Support (Penunjang Lab/Rad)
      {
        path: 'support',
        name: 'Support',
        component: () => import('@/views/support/SupportView.vue'),
        meta: { roles: ['SUPERADMIN', 'SUPPORT'] },
      },

      // 8. Logistics
      {
        path: 'logistics',
        name: 'Logistics',
        component: () => import('@/views/logistics/LogisticsView.vue'),
        meta: { roles: ['SUPERADMIN', 'LOGISTICS'] },
      },

      // 9. Cashier
      {
        path: 'cashier',
        name: 'Cashier',
        component: () => import('@/views/cashier/CashierView.vue'),
        meta: { roles: ['SUPERADMIN', 'CASHIER'] },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Frontend Router Guard (Section 2 RBAC Matrix)
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  if (to.meta.public) {
    if (authStore.isAuthenticated && to.path === '/login') {
      next('/')
      return
    }
    next()
    return
  }

  if (!authStore.isAuthenticated) {
    next('/login')
    return
  }

  // Check role authorization
  const allowedRoles = to.meta.roles as string[] | undefined
  if (allowedRoles && !authStore.hasRole(allowedRoles)) {
    // Redirect to permitted default route
    next('/')
    return
  }

  next()
})
