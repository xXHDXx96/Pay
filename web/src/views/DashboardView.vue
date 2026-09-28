<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  ChartBarIcon,
  ShoppingCartIcon,
  CurrencyDollarIcon,
  QrCodeIcon,
  CheckCircleIcon,
  XCircleIcon,
  ArrowPathIcon,
} from '@heroicons/vue/24/outline'
import { paymentApi, Order } from '@/api/client'

const loading = ref(true)
const error = ref<string | null>(null)
const healthStatus = ref<'online' | 'offline'>('offline')

const stats = ref([
  { name: 'Total Orders', value: 0, icon: ShoppingCartIcon, color: 'bg-blue-100 text-blue-600' },
  { name: 'Paid Orders', value: 0, icon: CheckCircleIcon, color: 'bg-green-100 text-green-600' },
  { name: 'Pending Orders', value: 0, icon: ChartBarIcon, color: 'bg-yellow-100 text-yellow-600' },
  { name: 'Failed Orders', value: 0, icon: XCircleIcon, color: 'bg-red-100 text-red-600' },
  { name: 'Total Volume (CNY)', value: 0, icon: CurrencyDollarIcon, color: 'bg-purple-100 text-purple-600' },
])

const recentOrders = ref<Order[]>([])
const paymentChannels = ref([
  { name: 'Alipay', value: 'ALIPAY', color: 'bg-alipay-500' },
  { name: 'WeChat Pay', value: 'WECHAT', color: 'bg-wechat-500' },
  { name: 'UnionPay', value: 'UNIONPAY', color: 'bg-unionpay-500' },
  { name: 'Crypto', value: 'CRYPTO', color: 'bg-crypto-500' },
  { name: 'Bank Card', value: 'BANK_CARD', color: 'bg-gray-500' },
])

async function loadData() {
  loading.value = true
  error.value = null
  try {
    const response = await paymentApi.healthCheck()
    healthStatus.value = 'online'

    const ordersRes = await paymentApi.listOrders()
    const orders = ordersRes.data.orders

    stats.value[0].value = orders.length
    stats.value[1].value = orders.filter(o => o.status === 'PAID' || o.status === 'SETTLED').length
    stats.value[2].value = orders.filter(o => o.status === 'PENDING').length
    stats.value[3].value = orders.filter(o => o.status === 'FAILED' || o.status === 'CLOSED').length
    stats.value[4].value = orders
      .filter(o => o.status === 'PAID' || o.status === 'SETTLED')
      .reduce((sum, o) => sum + o.amount_cny, 0)

    recentOrders.value = orders.slice(0, 5)
  } catch (err: any) {
    error.value = err.message || 'Failed to load data'
    healthStatus.value = 'offline'
  } finally {
    loading.value = false
  }
}

function getChannelClass(channel: string): string {
  const map: Record<string, string> = {
    ALIPAY: 'channel-alipay',
    WECHAT: 'channel-wechat',
    UNIONPAY: 'channel-unionpay',
    CRYPTO: 'channel-crypto',
    BANK_CARD: 'bg-gray-100 text-gray-800',
  }
  return map[channel] || 'bg-gray-100 text-gray-800'
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => {
  loadData()
  const interval = setInterval(loadData, 30000)
  return () => clearInterval(interval)
})
</script>

<template>
  <div>
    <div v-if="loading" class="flex justify-center items-center h-64">
      <ArrowPathIcon class="w-8 h-8 animate-spin text-primary-600" />
    </div>

    <div v-else-if="error" class="bg-red-50 text-red-700 p-4 rounded-lg">
      {{ error }}
    </div>

    <div v-else class="space-y-6">
      <div class="flex items-center gap-2 mb-4">
        <span class="text-sm text-gray-500">System Status:</span>
        <span
          :class="
            healthStatus === 'online'
              ? 'text-green-600 font-medium'
              : 'text-red-600 font-medium'
          "
        >
          {{ healthStatus === 'online' ? 'Online' : 'Offline' }}
        </span>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-4">
        <div
          v-for="stat in stats"
          :key="stat.name"
          class="bg-white rounded-xl p-4 border border-gray-200"
        >
          <div class="flex items-center gap-3">
            <div :class="stat.color" class="p-2 rounded-lg">
              <component :is="stat.icon" class="w-5 h-5" />
            </div>
            <div>
              <p class="text-2xl font-bold text-gray-900">{{ stat.value }}</p>
              <p class="text-sm text-gray-600">{{ stat.name }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white rounded-xl border border-gray-200">
        <div class="p-4 border-b">
          <h3 class="text-lg font-semibold text-gray-900">Payment Channels</h3>
        </div>
        <div class="p-4">
          <div class="flex flex-wrap gap-3">
            <div
              v-for="channel in paymentChannels"
              :key="channel.value"
              class="flex items-center gap-2 px-3 py-2 rounded-lg border"
            >
              <span :class="channel.color" class="w-3 h-3 rounded-full"></span>
              <span class="text-sm font-medium">{{ channel.name }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white rounded-xl border border-gray-200">
        <div class="p-4 border-b flex justify-between items-center">
          <h3 class="text-lg font-semibold text-gray-900">Recent Orders</h3>
          <router-link
            to="/orders"
            class="text-sm text-primary-600 hover:text-primary-700 font-medium"
          >
            View all
          </router-link>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left">
            <thead>
              <tr class="border-b">
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Order No</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Subject</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Amount</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Channel</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Status</th>
                <th class="px-4 py-2 text-xs font-medium text-gray-500 uppercase">Created</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="order in recentOrders"
                :key="order.id"
                class="border-b"
              >
                <td class="px-4 py-2 text-sm text-gray-900">{{ order.out_order_no }}</td>
                <td class="px-4 py-2 text-sm text-gray-700">{{ order.subject }}</td>
                <td class="px-4 py-2 text-sm font-medium">¥{{ order.amount_cny.toFixed(2) }}</td>
                <td class="px-4 py-2">
                  <span :class="getChannelClass(order.channel)" class="status-badge">
                    {{ order.channel }}
                  </span>
                </td>
                <td class="px-4 py-2">
                  <span
                    :class="
                      order.status === 'PAID' || order.status === 'SETTLED'
                        ? 'bg-green-100 text-green-800'
                        : order.status === 'PENDING'
                        ? 'bg-yellow-100 text-yellow-800'
                        : 'bg-red-100 text-red-800'
                    "
                    class="status-badge"
                  >
                    {{ order.status }}
                  </span>
                </td>
                <td class="px-4 py-2 text-sm text-gray-500">{{ formatDate(order.created_at) }}</td>
              </tr>
              <tr v-if="recentOrders.length === 0">
                <td colspan="6" class="px-4 py-8 text-center text-gray-500">
                  No orders yet
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
