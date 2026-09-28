import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import DashboardView from '@/views/DashboardView.vue'
import OrdersView from '@/views/OrdersView.vue'
import OrderDetailView from '@/views/OrderDetailView.vue'
import QRCodeView from '@/views/QRCodeView.vue'
import AggregationView from '@/views/AggregationView.vue'
import ExchangeView from '@/views/ExchangeView.vue'
import WebhookView from '@/views/WebhookView.vue'
import SettlementView from '@/views/SettlementView.vue'
import ReconciliationView from '@/views/ReconciliationView.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Dashboard',
    component: DashboardView,
    meta: { title: 'Dashboard', icon: 'chart-bar' },
  },
  {
    path: '/orders',
    name: 'Orders',
    component: OrdersView,
    meta: { title: 'Orders', icon: 'shopping-cart' },
  },
  {
    path: '/orders/:id',
    name: 'OrderDetail',
    component: OrderDetailView,
    meta: { title: 'Order Detail', icon: 'shopping-cart' },
    props: true,
  },
  {
    path: '/qrcode',
    name: 'QRCode',
    component: QRCodeView,
    meta: { title: 'QR Code Payment', icon: 'qrcode' },
  },
  {
    path: '/aggregation',
    name: 'Aggregation',
    component: AggregationView,
    meta: { title: 'Aggregation Payment', icon: 'collection' },
  },
  {
    path: '/exchange',
    name: 'Exchange',
    component: ExchangeView,
    meta: { title: 'Crypto Exchange', icon: 'currency-dollar' },
  },
  {
    path: '/webhook',
    name: 'Webhook',
    component: WebhookView,
    meta: { title: 'Webhook Testing', icon: 'webhook' },
  },
  {
    path: '/settlement',
    name: 'Settlement',
    component: SettlementView,
    meta: { title: 'Settlement', icon: 'bank' },
  },
  {
    path: '/reconciliation',
    name: 'Reconciliation',
    component: ReconciliationView,
    meta: { title: 'Reconciliation', icon: 'clipboard-list' },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})

router.beforeEach((to, _from, next) => {
  const title = to.meta.title ? `${to.meta.title} - Alipay Payment Gateway` : 'Alipay Payment Gateway'
  document.title = title
  next()
})

export default router
